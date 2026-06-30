package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"gorm.io/datatypes"
)

// ─── MCP Tool Business Methods ──────────────────────────────────────────

// CreateSession 创建QA会话（MCP工具）
func (l *QaLogic) CreateSession(ctx context.Context, title, agent, sessionType, projectID string) (string, string, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("CreateSession - 创建QA会话 [%s]", title))

	// 解析项目ID
	var prjID xSnowflake.SnowflakeID
	if projectID != "" {
		parsedPrjID, parseErr := xSnowflake.ParseSnowflakeID(projectID)
		if parseErr != nil {
			return "", "", xError.NewError(ctx, xError.BusinessError, "无效的项目ID格式", false, nil)
		}
		prjID = parsedPrjID
	}

	// 生成雪花ID
	id := xSnowflake.GenerateID(bConst.GeneQaSession)

	// 生成 Hash（SHA256 前 16 位 hex，碰撞检测重试）
	hash := generateSessionHash(id)
	for {
		// 碰撞检测
		if _, xErr := l.repo.session.GetByHash(ctx, hash); xErr != nil {
			// 未找到说明无碰撞，跳出
			break
		}
		// 碰撞：重新生成 ID 和 Hash
		id = xSnowflake.GenerateID(bConst.GeneQaSession)
		hash = generateSessionHash(id)
	}

	// 处理会话类型和过期时间
	sessionType = strings.ToLower(strings.TrimSpace(sessionType))
	if sessionType == "" || sessionType == "temporary" {
		sessionType = "temporary"
		expiresAt := time.Now().Add(48 * time.Hour)
		entity := &entity.QaSession{
			BaseEntity: xModels.BaseEntity{ID: id},
			Title:      title,
			Agent:      agent,
			Owner:      "system",
			Type:       sessionType,
			Status:     "active",
			ProjectID:  prjID,
			Hash:       hash,
			ExpiresAt:  &expiresAt,
		}
		if xErr := l.repo.session.Create(ctx, entity); xErr != nil {
			return "", "", xErr
		}
	} else {
		// permanent 类型无过期时间
		entity := &entity.QaSession{
			BaseEntity: xModels.BaseEntity{ID: id},
			Title:      title,
			Agent:      agent,
			Owner:      "system",
			Type:       sessionType,
			Status:     "active",
			ProjectID:  prjID,
			Hash:       hash,
			ExpiresAt:  nil,
		}
		if xErr := l.repo.session.Create(ctx, entity); xErr != nil {
			return "", "", xErr
		}
	}

	// 读取运行时域名配置，生成浏览器链接
	domain, xErr := l.repo.info.GetByKey(ctx, "runtime.domain")
	if xErr != nil || domain == "" {
		domain = "http://localhost:3000"
	}
	link := fmt.Sprintf("%s/interact?session=%s", strings.TrimRight(domain, "/"), hash)

	return id.String(), link, nil
}

// PushQuestion 推送问题到会话（MCP工具）
func (l *QaLogic) PushQuestion(ctx context.Context, sessionID, qType, title, description string, options, config, batch any, groupLabel string, supplement bool) (string, map[string]string, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("PushQuestion - 推送问题 [session=%s, title=%s]", sessionID, title))

	// 解析并验证会话
	parsedSID, err := xSnowflake.ParseSnowflakeID(sessionID)
	if err != nil {
		return "", nil, xError.NewError(ctx, xError.BusinessError, "无效的会话ID", false, nil)
	}

	// 验证会话存在且为活跃状态
	session, xErr := l.repo.session.GetByID(ctx, parsedSID)
	if xErr != nil {
		return "", nil, xErr
	}
	if session.Status != "active" {
		return "", nil, xError.NewError(ctx, xError.BusinessError, "会话不是活跃状态，无法推送问题", false, nil)
	}

	// ── 题型参数校验（P-01, P-15）──
	requiresOptions := map[string]bool{
		"select": true, "multi-select": true, "rank": true,
		"rate": true, "options": true,
	}
	if requiresOptions[qType] {
		if options == nil {
			return "", nil, xError.NewError(ctx, xError.BusinessError,
				xError.ErrMessage(fmt.Sprintf("题型 %s 必须提供 options 参数", qType)), false, nil)
		}
		var optList []map[string]interface{}
		switch v := options.(type) {
		case string:
			_ = json.Unmarshal([]byte(v), &optList)
		default:
			optBytes, _ := json.Marshal(options)
			_ = json.Unmarshal(optBytes, &optList)
		}
		if len(optList) == 0 {
			return "", nil, xError.NewError(ctx, xError.BusinessError,
				xError.ErrMessage(fmt.Sprintf("题型 %s 的 options 参数不能为空", qType)), false, nil)
		}
		// options 题型额外校验 pros/cons（P-01）
		if qType == "options" {
			for _, opt := range optList {
				pros, _ := opt["pros"].([]interface{})
				cons, _ := opt["cons"].([]interface{})
				if len(pros) == 0 || len(cons) == 0 {
					return "", nil, xError.NewError(ctx, xError.BusinessError,
						"options 题型的每个选项必须包含 pros 和 cons（用于差异化对比）", false, nil)
				}
			}
		}
	}

	// 生成问题ID
	qID := xSnowflake.GenerateID(bConst.GeneQaQuestion)

	// 构建问题实体
	questionEntity := &entity.QaQuestion{
		BaseEntity:  xModels.BaseEntity{ID: qID},
		SessionID:   parsedSID,
		Type:        qType,
		Title:       title,
		Description: description,
		Options:     toJSON(options),
		Config:      toJSON(config),
		Batch:       toJSON(batch),
		GroupLabel:  groupLabel,
		Supplement:  supplement,
		Status:      "pending",
	}

	// 为每个 option 生成雪花 ID 并构建映射
	optionIDMap := make(map[string]string)
	if options != nil {
		var rawOpts string
		switch v := options.(type) {
		case string:
			rawOpts = v
		default:
			if optBytes, jsonErr := json.Marshal(options); jsonErr == nil {
				rawOpts = string(optBytes)
			}
		}
		if rawOpts != "" {
			var optList []map[string]interface{}
			if json.Unmarshal([]byte(rawOpts), &optList) == nil {
				for i, opt := range optList {
					optID := xSnowflake.GenerateID(bConst.GeneQaQuestion).String()
					opt["id"] = optID
					optList[i] = opt
					if label, ok := opt["label"].(string); ok {
						optionIDMap[label] = optID
					}
				}
				if updated, jsonErr := json.Marshal(optList); jsonErr == nil {
					questionEntity.Options = datatypes.JSON(updated)
				}
			}
		}
	}

	// 持久化
	if xErr := l.repo.question.Create(ctx, questionEntity); xErr != nil {
		return "", nil, xErr
	}

	// 通知 WebSocket 层广播新问题到在线设备
	if OnQuestionPushed != nil {
		OnQuestionPushed(sessionID, questionEntity)
	}

	return qID.String(), optionIDMap, nil
}

// PushSupplement 推送补充内容（MCP工具）
func (l *QaLogic) PushSupplement(ctx context.Context, sessionID, targetType string, targetID xSnowflake.SnowflakeID, contentType, content string) *xError.Error {
	l.log.Info(ctx, fmt.Sprintf("PushSupplement - 推送补充内容 [session=%s, target=%s/%d]", sessionID, targetType, targetID))

	// 解析会话ID
	parsedSID, err := xSnowflake.ParseSnowflakeID(sessionID)
	if err != nil {
		return xError.NewError(ctx, xError.BusinessError, "无效的会话ID", false, nil)
	}

	// 生成补充ID
	sID := xSnowflake.GenerateID(bConst.GeneQaSupplement)

	supplementEntity := &entity.QaSupplement{
		BaseEntity:  xModels.BaseEntity{ID: sID},
		SessionID:   parsedSID,
		TargetType:  targetType,
		TargetID:    targetID,
		ContentType: contentType,
		Content:     content,
	}

	// 创建或覆写
	result, xErr := l.repo.supplement.CreateOrUpdate(ctx, supplementEntity)
	if xErr != nil {
		return xErr
	}

	// 通知 WebSocket 层广播补充内容到在线设备
	if OnSupplementPushed != nil {
		OnSupplementPushed(sessionID, result)
	}

	return nil
}

// GetAnswer 阻塞获取回答（MCP工具）
func (l *QaLogic) GetAnswer(ctx context.Context, sessionID string, timeout time.Duration) (string, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("GetAnswer - 阻塞获取回答 [session=%s, timeout=%s]", sessionID, timeout))

	// 从队列消费
	answers, err := l.queue.Consume(ctx, sessionID, timeout)
	if err != nil {
		return "", xError.NewError(ctx, xError.BusinessError, "获取回答被中断", false, err)
	}

	// 无回答
	if len(answers) == 0 {
		return "", nil
	}

	// 成功消费回答，重置该会话的重试计数器
	if resetErr := l.repo.retryCache.Reset(ctx, sessionID); resetErr != nil {
		l.log.Warn(ctx, fmt.Sprintf("GetAnswer - 重置重试计数器失败（忽略）: %s", resetErr.Error()))
	}

	// 格式化回答字符串（需要查询问题元数据）
	return l.formatAnswerString(ctx, answers), nil
}

// IncrementGetAnswerRetry 增加指定会话的 qa_get_answer 重试计数器
// 返回当前计数（递增后的值）。首次调用时自动设置过期时间为 Session TTL。
func (l *QaLogic) IncrementGetAnswerRetry(ctx context.Context, sessionID string, ttlSeconds int) (int, *xError.Error) {
	ttl := time.Duration(ttlSeconds) * time.Second
	count, err := l.repo.retryCache.Increment(ctx, sessionID, ttl)
	if err != nil {
		return 0, xError.NewError(ctx, xError.UnknownError, "增加重试计数器失败", false, err)
	}
	return int(count), nil
}

// RegetAnswer 重新获取已回答的问题（MCP工具）
func (l *QaLogic) RegetAnswer(ctx context.Context, sessionID, questionID string, base64Flag bool) (string, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("RegetAnswer - 重新获取回答 [session=%s, question=%s]", sessionID, questionID))

	// 解析问题ID
	parsedQID, err := xSnowflake.ParseSnowflakeID(questionID)
	if err != nil {
		return "", xError.NewError(ctx, xError.BusinessError, "无效的问题ID", false, nil)
	}

	// 查询问题
	question, xErr := l.repo.question.GetByID(ctx, parsedQID)
	if xErr != nil {
		return "", xErr
	}

	// 检查是否已回答
	if question.Status != "answered" {
		return "该问题用户暂未回答，请使用 qa_get_answer 来获取用户最近的一次回答。", nil
	}

	// 如果需要 base64 且有媒体数据
	if base64Flag && question.Media != nil {
		mediaBytes, jsonErr := json.Marshal(question.Media)
		if jsonErr == nil && len(mediaBytes) > 0 {
			return fmt.Sprintf("--- question:%s ---\n[MEDIA] %s", questionID, string(mediaBytes)), nil
		}
	}

	// 构造单条回答格式
	var answerData interface{}
	if question.Answer != nil {
		_ = json.Unmarshal(question.Answer, &answerData)
	}
	questionType, questionOptions, config, description := parseQuestionMetadata(question)
	fmtCtx := AnswerFormatContext{
		Description:       description,
		Config:            config,
		OptionSupplements: l.buildOptionSupplementMap(ctx, sessionID, questionID),
	}
	summary := formatAnswerData(questionID, answerData, questionType, questionOptions, fmtCtx)
	// image/file 类型：追加 OTP 下载令牌（P-11）
	summary = l.enhanceMediaAnswerWithOTP(ctx, questionType, sessionID, summary, answerData)
	return fmt.Sprintf("--- question:%s ---\n[ANSWER] %s", questionID, summary), nil
}

// RegetAnswers 批量重新获取回答（MCP工具）
//
// 遍历所有 questionIDs，对每个执行：
// - 已回答 → "[ANSWERED] {question_id}: {答案摘要}"
// - 未回答 → "[PENDING] {question_id}: 该问题用户暂未回答，请使用 qa_get_answer 来获取用户最近的一次回答。"
// - 无效ID → "[ERROR] {question_id}: 无效的问题ID格式"
// - 不存在 → "[ERROR] {question_id}: 问题不存在"
func (l *QaLogic) RegetAnswers(ctx context.Context, sessionID string, questionIDs []string) (string, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("RegetAnswers - 批量重新获取回答 [session=%s, count=%d]", sessionID, len(questionIDs)))

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("批量查询结果（%d 个问题）:\n\n", len(questionIDs)))

	for _, qid := range questionIDs {
		parsedQID, err := xSnowflake.ParseSnowflakeID(qid)
		if err != nil {
			sb.WriteString(fmt.Sprintf("--- question:%s ---\n[ERROR] 无效的问题ID格式\n", qid))
			continue
		}

		question, xErr := l.repo.question.GetByID(ctx, parsedQID)
		if xErr != nil {
			sb.WriteString(fmt.Sprintf("--- question:%s ---\n[ERROR] 问题不存在\n", qid))
			continue
		}

		var answerData interface{}
		if question.Answer != nil {
			_ = json.Unmarshal(question.Answer, &answerData)
		}
		questionType, questionOptions, config, description := parseQuestionMetadata(question)
		fmtCtx := AnswerFormatContext{
			Description:       description,
			Config:            config,
			OptionSupplements: l.buildOptionSupplementMap(ctx, sessionID, qid),
		}
		summary := formatAnswerData(qid, answerData, questionType, questionOptions, fmtCtx)
		// image/file 类型：追加 OTP 下载令牌（P-11）
		summary = l.enhanceMediaAnswerWithOTP(ctx, questionType, sessionID, summary, answerData)
		if question.Status == "answered" {
			sb.WriteString(fmt.Sprintf("--- question:%s ---\n[ANSWER] %s\n", qid, summary))
		} else {
			sb.WriteString(fmt.Sprintf("--- question:%s ---\n[PENDING] 该问题用户暂未回答，请使用 qa_get_answer 来获取用户最近的一次回答。\n", qid))
		}
	}

	return sb.String(), nil
}

// GetSessionMCP 获取会话信息（MCP工具，返回人类可读字符串）
func (l *QaLogic) GetSessionMCP(ctx context.Context, sessionID string) (string, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("GetSessionMCP - 获取会话信息 [%s]", sessionID))

	// 解析会话ID
	parsedSID, err := xSnowflake.ParseSnowflakeID(sessionID)
	if err != nil {
		return "", xError.NewError(ctx, xError.BusinessError, "无效的会话ID", false, nil)
	}

	// 查询会话及问题
	session, questions, xErr := l.repo.session.GetByIDWithQuestions(ctx, parsedSID)
	if xErr != nil {
		return "", xErr
	}

	// 构建人类可读字符串
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("会话ID: %s\n", session.ID.String()))
	sb.WriteString(fmt.Sprintf("标题: %s\n", session.Title))
	sb.WriteString(fmt.Sprintf("Agent: %s\n", session.Agent))
	sb.WriteString(fmt.Sprintf("类型: %s\n", session.Type))
	sb.WriteString(fmt.Sprintf("状态: %s\n", session.Status))
	sb.WriteString(fmt.Sprintf("关联项目: %d\n", session.ProjectID.Int64()))
	sb.WriteString(fmt.Sprintf("在线设备: %d\n", session.OnlineDevices))
	if session.ExpiresAt != nil {
		sb.WriteString(fmt.Sprintf("过期时间: %s\n", session.ExpiresAt.Format(time.RFC3339)))
	}
	sb.WriteString(fmt.Sprintf("创建时间: %s\n", session.CreatedAt.Format(time.RFC3339)))
	sb.WriteString("\n问题列表:\n")

	if len(questions) == 0 {
		sb.WriteString("  （暂无问题）\n")
	} else {
		for i, q := range questions {
			sb.WriteString(fmt.Sprintf("  %d. [%s] %s - %s\n", i+1, q.Type, q.Title, q.Status))
		}
	}

	return sb.String(), nil
}

// DeleteSessionMCP 删除会话（MCP工具）
func (l *QaLogic) DeleteSessionMCP(ctx context.Context, sessionID string) *xError.Error {
	return l.DeleteSession(ctx, sessionID)
}
