package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	"github.com/xiaolfeng/Lumina/internal/entity"
	qaQueue "github.com/xiaolfeng/Lumina/internal/qa"
)

// AnswerFormatContext 封装格式化回答所需的 question 级上下文
type AnswerFormatContext struct {
	Description       string                 // 问题级描述（question.description）
	Config            map[string]interface{} // 问题配置（plan sections / diff before-after 等）
	OptionSupplements map[string]string      // 选项级 supplement 映射（key=optionID, value=markdown 内容，HTML 已过滤）
}

// parseQuestionMetadata 从 QaQuestion 实体提取题型、选项列表、配置和描述
func parseQuestionMetadata(q *entity.QaQuestion) (string, []map[string]interface{}, map[string]interface{}, string) {
	if q == nil {
		return "", nil, nil, ""
	}
	questionType := q.Type
	var options []map[string]interface{}
	if q.Options != nil {
		_ = json.Unmarshal(q.Options, &options)
	}
	var config map[string]interface{}
	if q.Config != nil {
		_ = json.Unmarshal(q.Config, &config)
	}
	return questionType, options, config, q.Description
}

// formatAnswerString 将多条回答格式化为人类可读字符串（P-18 优化）
//
// 优化要点：skip/supplement 标记由 WebSocket handler 直接推入字符串（"[SKIPPED]" /
// "[NEED_SUPPLEMENT] ..."），正常回答推入 map。通过类型判断直接区分两种路径，
// 消除原来对每个回答先走一次 default 格式化再查 DB 再格式化的性能浪费。
// 同时为 select/multi-select 题型注入选项级 supplement 映射。
func (l *QaLogic) formatAnswerString(ctx context.Context, answers []qaQueue.Answer) string {
	var sb strings.Builder
	for i, a := range answers {
		if i > 0 {
			sb.WriteString("\n")
		}

		var marker, content string

		// skip/supplement 标记是字符串类型（WebSocket handler 直接推入字符串）
		if dataStr, ok := a.Data.(string); ok {
			switch {
			case dataStr == "[SKIPPED]":
				marker = "[SKIPPED]"
				content = "用户跳过了此问题"
			case strings.HasPrefix(dataStr, "[NEED_SUPPLEMENT]"):
				marker = "[NEED_SUPPLEMENT]"
				content = strings.TrimPrefix(dataStr, "[NEED_SUPPLEMENT] ")
			default:
				marker = "[ANSWER]"
				content = dataStr
			}
		} else {
			// 正常回答（map 类型）— 查询 DB 获取题型元数据后格式化
			marker = "[ANSWER]"
			parsedQID, err := xSnowflake.ParseSnowflakeID(a.QuestionID)
			if err == nil {
				question, xErr := l.repo.question.GetByID(ctx, parsedQID)
				if xErr == nil {
					questionType, questionOptions, config, description := parseQuestionMetadata(question)
					fmtCtx := AnswerFormatContext{
						Description:       description,
						Config:            config,
						OptionSupplements: l.buildOptionSupplementMap(ctx, question.SessionID.String(), a.QuestionID),
					}
					content = formatAnswerData(a.QuestionID, a.Data, questionType, questionOptions, fmtCtx)
					// image/file 类型：追加 OTP 下载令牌（P-11）
					content = l.enhanceMediaAnswerWithOTP(ctx, questionType, question.SessionID.String(), content, a.Data)
				} else {
					content = formatAnswerData(a.QuestionID, a.Data, "", nil, AnswerFormatContext{})
				}
			} else {
				content = formatAnswerData(a.QuestionID, a.Data, "", nil, AnswerFormatContext{})
			}
		}

		sb.WriteString(fmt.Sprintf("--- question:%s ---\n%s %s\n", a.QuestionID, marker, content))
	}
	return sb.String()
}

// buildOptionSupplementMap 查询问题关联的选项级 supplement，构建 map[optionID]content
// 仅包含 content_type=markdown 的 supplement（HTML 格式不返回给 AI）
func (l *QaLogic) buildOptionSupplementMap(ctx context.Context, sessionID string, questionID string) map[string]string {
	parsedSID, err := xSnowflake.ParseSnowflakeID(sessionID)
	if err != nil {
		return nil
	}
	supplements, xErr := l.repo.supplement.GetBySessionID(ctx, parsedSID)
	if xErr != nil {
		return nil
	}
	result := make(map[string]string)
	for _, s := range supplements {
		if s.TargetType != "option" || s.ContentType != "markdown" {
			continue
		}
		result[s.TargetID.String()] = s.Content
	}
	return result
}

// enhanceMediaAnswerWithOTP 为 image/file 类型的格式化回答追加 OTP 下载令牌链接
//
// 后处理注入策略：formatFileLikeAnswer 是纯函数，无法访问 Redis，
// 后处理注入策略：formatFileLikeAnswer 是纯函数，无法访问缓存层，
// 故在 QaLogic 方法层（通过注入的 downloadToken 服务）对格式化结果做后处理，为每个文件生成一次性令牌。
//
// 输出标记（P-11 核心交付物）：
//   - [DOWNLOAD_URL]  每行一个 OTP 下载链接（<domain>/api/v1/qa/download/<token>）
//   - [IMPORTANT]     一次性令牌提示
//   - [TIP]           curl/wget 下载指引
//   - [GIT_TIP]       .lumina/cache 忽略建议
//
// 参数:
//   - ctx: 上下文（用于 Redis 读写和 DB 查询）
//   - questionType: 题型（仅 "image"/"file" 触发增强，其他类型直接返回原字符串）
//   - sessionID: 会话 ID（当前未参与令牌生成，预留扩展）
//   - formatted: formatAnswerData 已格式化的回答字符串
//   - answerData: 原始回答数据（用于提取 images/files 列表）
//
// 返回增强后的字符串；若无需增强（非 media 类型 / 无文件 / 令牌生成失败）则原样返回。
func (l *QaLogic) enhanceMediaAnswerWithOTP(ctx context.Context, questionType, sessionID, formatted string, answerData any) string {
	// 仅 image/file 类型需要追加 OTP 令牌
	if questionType != "image" && questionType != "file" {
		return formatted
	}

	m, ok := answerData.(map[string]interface{})
	if !ok {
		return formatted
	}

	fieldKey := "images"
	if questionType == "file" {
		fieldKey = "files"
	}
	items, ok := m[fieldKey].([]interface{})
	if !ok || len(items) == 0 {
		return formatted
	}

	// 读取运行时域名配置（Info 表 key=runtime.domain）
	domain := "http://localhost:8080"
	if domainVal, xErr := l.repo.info.GetByKey(ctx, "runtime.domain"); xErr == nil && domainVal != "" {
		domain = strings.TrimRight(domainVal, "/")
	}

	// 为每个文件生成 OTP 令牌（使用注入的 downloadToken 服务，避免 logic 直连 Redis）
	var tokenUrls []string
	for _, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		filePath, _ := itemMap["filePath"].(string)
		if filePath == "" {
			continue
		}
		filename, _ := itemMap["filename"].(string)
		if filename == "" {
			filename, _ = itemMap["name"].(string)
		}
		mimeType, _ := itemMap["mimeType"].(string)

		token, xErr := l.repo.downloadToken.GenerateToken(ctx, filePath, filename, mimeType)
		if xErr != nil {
			l.log.Warn(ctx, fmt.Sprintf("enhanceMediaAnswerWithOTP - 生成下载令牌失败 [file=%s]: %v", filePath, xErr))
			continue
		}
		tokenUrls = append(tokenUrls, fmt.Sprintf("    - %s/api/v1/qa/download/%s", domain, token))
	}

	if len(tokenUrls) == 0 {
		return formatted
	}

	var sb strings.Builder
	sb.WriteString(formatted)
	sb.WriteString("\n[DOWNLOAD_URL]")
	for _, url := range tokenUrls {
		sb.WriteString("\n")
		sb.WriteString(url)
	}
	sb.WriteString("\n[IMPORTANT] 下载链接为一次性令牌，使用后立即失效。若下载失败需重新下载，请重新调用 qa_reget_answer 获取新的下载链接。")
	sb.WriteString("\n[TIP] 下载并保存文件时，最终路径 = DOWNLOAD_PATH + FILE_NAME：")
	sb.WriteString("\n  - Mac/Linux: curl -o \"<DOWNLOAD_PATH><FILE_NAME>\" <url>")
	sb.WriteString("\n  - Windows:   Invoke-WebRequest -Uri <url> -OutFile \"<DOWNLOAD_PATH><FILE_NAME>\"")
	sb.WriteString("\n  AI 引用该文件时，使用 <DOWNLOAD_PATH><FILE_NAME> 作为完整路径。")
	sb.WriteString("\n[GIT_TIP] 若存在 git 等版本管理项目，需要把 .lumina/cache/* 加入忽略上传（如 .gitignore 使用分类模式：\n    ### Lumina ###\n    .lumina/cache/）")

	return sb.String()
}

// formatAnswerData 格式化单条回答数据
func formatAnswerData(questionID string, data any, questionType string, options []map[string]interface{}, fmtCtx AnswerFormatContext) string {
	if data == nil {
		return ""
	}

	m, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Sprintf("%v", data)
	}

	switch questionType {
	case "select":
		return formatSelectAnswer(m, options, fmtCtx)
	case "multi-select":
		return formatMultiSelectAnswer(m, options, fmtCtx)
	case "text":
		return formatTextAnswer(m)
	case "boolean":
		return formatBooleanAnswer(m)
	case "code":
		return formatCodeAnswer(m)
	case "image":
		return formatFileLikeAnswer(m, "images")
	case "file":
		return formatFileLikeAnswer(m, "files")
	case "plan":
		return formatPlanAnswer(m, fmtCtx.Config)
	case "diff":
		return formatDiffAnswer(m, fmtCtx.Config)
	case "review":
		return formatReviewAnswer(m)
	case "options":
		return formatOptionsAnswer(m, options, fmtCtx)
	case "slider":
		return formatSliderAnswer(m)
	case "rank":
		return formatRankAnswer(m, options)
	case "rate":
		return formatRateAnswer(m, options)
	default:
		return formatGenericAnswer(m)
	}
}

// formatSelectAnswer 格式化单选回答
//
// 输出格式（P-05）：
//   - 用户选择：<label>
//   - [DESCRIPTION] <question.description>          （问题级，可选）
//   - [OPTION_DESCRIPTION] <option.description>      （选项级，可选）
//   - [SUPPLEMENT] <agent option supplement | other> （选项级补充或自定义输入，可选）
//
// 每级信息为空时整行省略。
func formatSelectAnswer(m map[string]interface{}, options []map[string]interface{}, fmtCtx AnswerFormatContext) string {
	selected, exists := m["selected"]
	if !exists {
		return fmt.Sprintf("%v", m)
	}

	selectedID := getOptionID(selected)
	label, optionDesc := resolveOptionLabel(selected, options)
	if label == "" {
		label = selectedID
	}

	var sb strings.Builder
	sb.WriteString("用户选择：" + label)

	if strings.TrimSpace(fmtCtx.Description) != "" {
		sb.WriteString("\n[DESCRIPTION] ")
		sb.WriteString(fmtCtx.Description)
	}
	if optionDesc != "" {
		sb.WriteString("\n[OPTION_DESCRIPTION] ")
		sb.WriteString(optionDesc)
	}
	if supp, ok := fmtCtx.OptionSupplements[selectedID]; ok && strings.TrimSpace(supp) != "" {
		sb.WriteString("\n[SUPPLEMENT] ")
		sb.WriteString(supp)
	}
	if other, ok := m["other"].(string); ok && strings.TrimSpace(other) != "" {
		sb.WriteString("\n[SUPPLEMENT] ")
		sb.WriteString(other)
	}

	return sb.String()
}

// formatMultiSelectAnswer 格式化多选回答
//
// 输出格式（P-08）：
//   - 用户选择 N 项
//   - [DESCRIPTION] <question.description>                  （问题级，可选）
//   - ---
//   - [OPTION] <label>                                      （逐项）
//   - [OPTION_DESCRIPTION] <option.description>             （选项级，可选）
//   - [SUPPLEMENT] <agent option supplement>                （选项级补充，可选）
//
// 每级信息为空时整行省略；选项之间以 `---` 分隔。
func formatMultiSelectAnswer(m map[string]interface{}, options []map[string]interface{}, fmtCtx AnswerFormatContext) string {
	selected, exists := m["selected"]
	if !exists {
		return fmt.Sprintf("%v", m)
	}

	var selectedList []interface{}
	switch v := selected.(type) {
	case []interface{}:
		selectedList = v
	default:
		selectedList = []interface{}{selected}
	}

	otherCount := 0
	if others, ok := m["other"].([]interface{}); ok {
		otherCount = len(others)
	}
	totalCount := len(selectedList) + otherCount

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("用户选择 %d 项", totalCount))

	if strings.TrimSpace(fmtCtx.Description) != "" {
		sb.WriteString("\n[DESCRIPTION] ")
		sb.WriteString(fmtCtx.Description)
	}

	for _, item := range selectedList {
		selectedID := getOptionID(item)
		label, optionDesc := resolveOptionLabel(item, options)
		if label == "" {
			label = selectedID
		}

		sb.WriteString("\n---")
		sb.WriteString("\n[OPTION] ")
		sb.WriteString(label)

		if optionDesc != "" {
			sb.WriteString("\n[OPTION_DESCRIPTION] ")
			sb.WriteString(optionDesc)
		}
		if supp, ok := fmtCtx.OptionSupplements[selectedID]; ok && strings.TrimSpace(supp) != "" {
			sb.WriteString("\n[SUPPLEMENT] ")
			sb.WriteString(supp)
		}
	}

	if others, ok := m["other"].([]interface{}); ok && len(others) > 0 {
		for _, other := range others {
			s, ok := other.(string)
			if !ok || strings.TrimSpace(s) == "" {
				continue
			}
			sb.WriteString("\n---")
			sb.WriteString("\n[OPTION] __other__")
			sb.WriteString("\n[SUPPLEMENT] ")
			sb.WriteString(s)
		}
	}

	return sb.String()
}

// resolveOptionLabel 从选项列表中解析选项的标签和描述
func resolveOptionLabel(selected any, options []map[string]interface{}) (string, string) {
	// 直接是 map（含 id + label）
	if selMap, ok := selected.(map[string]interface{}); ok {
		label, _ := selMap["label"].(string)
		desc, _ := selMap["description"].(string)
		return label, desc
	}

	// 纯 ID 字符串 → 从 options 反查
	selStr, ok := selected.(string)
	if !ok || len(options) == 0 {
		return fmt.Sprintf("%v", selected), ""
	}

	for _, opt := range options {
		if optID, _ := opt["id"].(string); optID == selStr {
			label, _ := opt["label"].(string)
			desc, _ := opt["description"].(string)
			return label, desc
		}
	}
	return selStr, ""
}

// resolveLabelByID 从选项列表反查 label，找不到则返回 ID 本身（降级）
//
// 用于 rank/rate 等仅持有 ID 列表的场景，避免直接向 Agent 暴露雪花 ID。
func resolveLabelByID(id string, options []map[string]interface{}) string {
	for _, opt := range options {
		if optID, _ := opt["id"].(string); optID == id {
			if label, _ := opt["label"].(string); label != "" {
				return label
			}
		}
	}
	return id
}

// getOptionID 从 selected 值中提取 optionID
//
// 兼容两种前端提交格式：
//   - string（纯 ID，主流格式）
//   - map[string]interface{}（含 id 字段，历史兼容）
func getOptionID(selected any) string {
	if idStr, ok := selected.(string); ok {
		return idStr
	}
	if selMap, ok := selected.(map[string]interface{}); ok {
		if id, ok := selMap["id"].(string); ok {
			return id
		}
	}
	return fmt.Sprintf("%v", selected)
}

// formatTextAnswer 格式化文本回答
func formatTextAnswer(m map[string]interface{}) string {
	if text, ok := m["text"].(string); ok {
		return text
	}
	return fmt.Sprintf("%v", m)
}

// formatBooleanAnswer 格式化布尔回答
func formatBooleanAnswer(m map[string]interface{}) string {
	if choice, ok := m["choice"].(string); ok {
		if choice == "yes" {
			return "是"
		}
		return "否"
	}
	if choice, ok := m["choice"].(bool); ok {
		if choice {
			return "是"
		}
		return "否"
	}
	return fmt.Sprintf("%v", m)
}

// formatCodeAnswer 格式化代码回答（P-10，从 formatMediaAnswer 拆分）
//
// 输出 code + 可选的 [LANGUAGE] 标记，便于 Agent 识别代码语言。
func formatCodeAnswer(m map[string]interface{}) string {
	code, ok := m["code"].(string)
	if !ok {
		return fmt.Sprintf("%v", m)
	}
	var sb strings.Builder
	sb.WriteString(code)
	if lang, ok := m["language"].(string); ok && lang != "" {
		sb.WriteString("\n[LANGUAGE] ")
		sb.WriteString(lang)
	}
	return sb.String()
}

// formatFileLikeAnswer 格式化图片/文件类回答（P-11，从 formatMediaAnswer 拆分）
//
// 注意：OTP 令牌生成（[DOWNLOAD_URL]/[TIP]/[IMPORTANT]）在 enhanceMediaAnswerWithOTP 中后处理注入，
// 本函数仅输出文件名和内部存储路径信息。兼容 filename 和 name 两种字段。
func formatFileLikeAnswer(m map[string]interface{}, sliceKey string) string {
	items, ok := m[sliceKey].([]interface{})
	if !ok || len(items) == 0 {
		return fmt.Sprintf("%v", m)
	}

	var sb strings.Builder
	sb.WriteString("用户已上传内容")
	for _, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := itemMap["filename"].(string)
		if name == "" {
			name, _ = itemMap["name"].(string)
		}
		filePath, _ := itemMap["filePath"].(string)
		sb.WriteString(fmt.Sprintf("\n---\n[FILE_NAME] %s", name))
		if filePath != "" {
			// 输出目录级路径，去掉末尾 UUID 文件名，便于 Agent 拼接 FILE_NAME
			dir := filepath.Dir(filePath)
			if !strings.HasSuffix(dir, "/") {
				dir += "/"
			}
			sb.WriteString(fmt.Sprintf("\n[DOWNLOAD_PATH] %s", dir))
		}
	}
	return sb.String()
}

// formatSliderAnswer 格式化滑块回答
func formatSliderAnswer(m map[string]interface{}) string {
	if value, ok := m["value"]; ok {
		return fmt.Sprintf("%v", value)
	}
	return fmt.Sprintf("%v", m)
}

// formatRankAnswer 格式化排名回答
//
// 兼容前端 "ranking"（主）和 "ranked"（向后兼容）两种字段名；
// 纯 ID 列表会通过 options 反查 label，避免向 Agent 暴露雪花 ID。
// 输出形如 `1. A → 2. B → 3. C`。
func formatRankAnswer(m map[string]interface{}, options []map[string]interface{}) string {
	var ranked []interface{}
	if v, ok := m["ranking"].([]interface{}); ok {
		ranked = v
	} else if v, ok := m["ranked"].([]interface{}); ok {
		ranked = v
	}
	if len(ranked) == 0 {
		return fmt.Sprintf("%v", m)
	}

	items := make([]string, 0, len(ranked))
	for i, item := range ranked {
		var label string
		if idStr, ok := item.(string); ok {
			label = resolveLabelByID(idStr, options)
		} else if rMap, ok := item.(map[string]interface{}); ok {
			if l, ok := rMap["label"].(string); ok && l != "" {
				label = l
			} else {
				label = fmt.Sprintf("%v", item)
			}
		} else {
			label = fmt.Sprintf("%v", item)
		}
		items = append(items, fmt.Sprintf("%d. %s", i+1, label))
	}
	return strings.Join(items, " → ")
}

// formatRateAnswer 格式化评分回答
//
// 按 options 顺序输出评分（保持稳定输出顺序，避免 map 随机迭代）；
// 空 ratings 或无 options 匹配时降级为原始 map 输出。
// 无评分数据时返回友好提示而非原始 JSON。
func formatRateAnswer(m map[string]interface{}, options []map[string]interface{}) string {
	ratings, ok := m["ratings"].(map[string]interface{})
	if !ok || len(ratings) == 0 {
		return "用户未提供评分"
	}

	// 按选项顺序输出（稳定排序），避免 map 随机迭代
	parts := make([]string, 0, len(ratings))
	for _, opt := range options {
		optID, _ := opt["id"].(string)
		if optID == "" {
			continue
		}
		if score, exists := ratings[optID]; exists {
			label, _ := opt["label"].(string)
			if label == "" {
				label = optID
			}
			parts = append(parts, fmt.Sprintf("%s: %v", label, score))
		}
	}

	// 降级：没有 options 或无匹配项时直接输出原始 map
	if len(parts) == 0 {
		for k, v := range ratings {
			parts = append(parts, fmt.Sprintf("%s: %v", k, v))
		}
	}
	return strings.Join(parts, ", ")
}

// formatGenericAnswer 未知题型的通用格式化
func formatGenericAnswer(m map[string]interface{}) string {
	if selected, ok := m["selected"]; ok {
		return fmt.Sprintf("%v", selected)
	}
	if text, ok := m["text"].(string); ok {
		return text
	}
	return fmt.Sprintf("%v", m)
}

// formatPlanAnswer 格式化 Plan 计划题回答（P-07 重写）
//
// 回答结构：{ decision: "approve"|"reject"|"revise", annotations?: [...], feedback?: "..." }
// - approve → 输出 [PLAN_DETAIL]，让 Agent 知道用户批准了什么
// - reject  → 仅提示拒绝
// - revise  → 输出 [REVISIONS] 逐章节修订意见
// config 来源：question.config（含 sections）
func formatPlanAnswer(m map[string]interface{}, config map[string]interface{}) string {
	decision, _ := m["decision"].(string)
	var sb strings.Builder

	switch decision {
	case "approve":
		sb.WriteString("用户已批准该计划")
		if sections, ok := config["sections"].([]interface{}); ok && len(sections) > 0 {
			sb.WriteString("\n[PLAN_DETAIL]")
			for i, sec := range sections {
				if secMap, ok := sec.(map[string]interface{}); ok {
					title, _ := secMap["title"].(string)
					content, _ := secMap["content"].(string)
					sb.WriteString(fmt.Sprintf("\n%d. %s\n   %s", i+1, title, content))
				}
			}
		}
	case "reject":
		sb.WriteString("用户已拒绝该计划")
	case "revise":
		sb.WriteString("用户要求修改该计划")
		if annotations, ok := m["annotations"].([]interface{}); ok && len(annotations) > 0 {
			sb.WriteString("\n[REVISIONS]")
			for i, ann := range annotations {
				if annMap, ok := ann.(map[string]interface{}); ok {
					sectionID, _ := annMap["sectionId"].(string)
					content, _ := annMap["content"].(string)
					if content != "" {
						sb.WriteString(fmt.Sprintf("\n%d. [%s] %s", i+1, sectionID, content))
					}
				}
			}
		}
	default:
		sb.WriteString(fmt.Sprintf("decision=%s", decision))
	}

	if feedback, ok := m["feedback"].(string); ok && strings.TrimSpace(feedback) != "" {
		sb.WriteString(fmt.Sprintf("\n[FEEDBACK] %s", feedback))
	}

	return sb.String()
}

// formatOptionsAnswer 格式化通用 options 题型回答（P-03 新增）
//
// options 题型与 select 行为相似，但无 [DESCRIPTION] 问题级描述输出，
// 主要输出选中选项的 label + 可选的选项级 description + 用户反馈。
func formatOptionsAnswer(m map[string]interface{}, options []map[string]interface{}, fmtCtx AnswerFormatContext) string {
	selected, exists := m["selected"]
	if !exists {
		return fmt.Sprintf("%v", m)
	}

	selectedID := getOptionID(selected)
	label, desc := resolveOptionLabel(selected, options)
	if label == "" {
		label = selectedID
	}

	var sb strings.Builder
	sb.WriteString(label)
	if desc != "" {
		sb.WriteString("\n[DESCRIPTION] ")
		sb.WriteString(desc)
	}
	// 选项级 supplement（Agent 为该选项推送的 markdown 内容）
	if supp, ok := fmtCtx.OptionSupplements[selectedID]; ok && strings.TrimSpace(supp) != "" {
		sb.WriteString("\n[SUPPLEMENT] ")
		sb.WriteString(supp)
	}
	// 用户的选择理由（feedback）
	if feedback, ok := m["feedback"].(string); ok && strings.TrimSpace(feedback) != "" {
		sb.WriteString("\n[SUPPLEMENT] ")
		sb.WriteString(feedback)
	}
	return sb.String()
}

// formatDiffAnswer 格式化 Diff 决策题回答（P-09，从 formatDecisionAnswer 拆分）
//
// 回答结构：{ decision: "approve"|"reject"|"edit", edited?: "...", feedback?: "..." }
// - approve → 输出 config.after 作为 [FINAL] 最终代码，让 Agent 知道实际写入内容
// - reject  → 仅提示拒绝
// - edit    → 输出用户编辑后的 m.edited 作为 [FINAL]
func formatDiffAnswer(m map[string]interface{}, config map[string]interface{}) string {
	decision, _ := m["decision"].(string)
	var sb strings.Builder

	switch decision {
	case "approve":
		sb.WriteString("用户已批准该修改")
		if after, ok := config["after"].(string); ok && after != "" {
			sb.WriteString("\n[FINAL]\n")
			sb.WriteString(after)
		}
	case "reject":
		sb.WriteString("用户已拒绝该修改")
	case "edit":
		sb.WriteString("用户修改后提交")
		if edited, ok := m["edited"].(string); ok && edited != "" {
			sb.WriteString("\n[FINAL]\n")
			sb.WriteString(edited)
		}
	default:
		sb.WriteString(fmt.Sprintf("decision=%s", decision))
	}

	if feedback, ok := m["feedback"].(string); ok && strings.TrimSpace(feedback) != "" {
		sb.WriteString(fmt.Sprintf("\n[FEEDBACK] %s", feedback))
	}

	return sb.String()
}

// formatReviewAnswer 格式化 Review 决策题回答（P-12，从 formatDecisionAnswer 拆分）
//
// 回答结构：{ decision: "approve"|"revise", feedback?: "..." }
// 去除 [已批准]/[已拒绝] 前缀（外层 [ANSWER] 已有标记，避免语义重复）。
func formatReviewAnswer(m map[string]interface{}) string {
	decision, _ := m["decision"].(string)
	var sb strings.Builder

	switch decision {
	case "approve":
		sb.WriteString("用户批准了该修改")
	case "revise":
		sb.WriteString("用户要求修改")
	default:
		sb.WriteString(fmt.Sprintf("decision=%s", decision))
	}

	if feedback, ok := m["feedback"].(string); ok && strings.TrimSpace(feedback) != "" {
		sb.WriteString(fmt.Sprintf("\n[FEEDBACK] %s", feedback))
	}

	return sb.String()
}
