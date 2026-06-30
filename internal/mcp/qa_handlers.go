package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	apiQa "github.com/xiaolfeng/Lumina/api/qa"
)

// ─── Tool Handlers ──────────────────────────────────────────────────────

// handleQaSessionCreate 创建 QA 会话
func handleQaSessionCreate(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if qaLogic == nil {
		return textResult("QaLogic 未初始化，请联系管理员"), nil
	}

	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return textResult(errMsg), nil
	}
	projectID, _ := args["project_id"].(string)
	if projectID == "" {
		return textResult("缺少必填参数: project_id"), nil
	}

	title, _ := args["title"].(string)
	sessionType, _ := args["session_type"].(string)
	agentName, _ := args["agent_name"].(string)
	if agentName == "" {
		agentName = "mcp-agent"
	}

	id, link, xErr := qaLogic.CreateSession(ctx, title, agentName, sessionType, projectID)
	if xErr != nil {
		return textResult(fmt.Sprintf("创建会话失败: %s", xErr.Error())), nil
	}

	return textResult(fmt.Sprintf(
		"[RESPONSE] 会话创建成功（ID: %s）\n[URL] %s\n[TIP] 会话已创建，请在可视化桌面环境下使用命令打开浏览器引导用户进入会话页：\n"+
			"  - macOS:   open \"%s\"\n"+
			"  - Linux:   xdg-open \"%s\"（需 X11/Wayland 桌面环境）\n"+
			"  - Windows: start \"\" \"%s\"\n"+
			"  若当前为无头环境（如远程 SSH/容器），请直接将 URL 输出给用户手动访问。",
		id, link, link, link, link,
	)), nil
}

// handleQaSessionList 获取会话列表
func handleQaSessionList(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if qaLogic == nil {
		return textResult("QaLogic 未初始化，请联系管理员"), nil
	}

	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return textResult(errMsg), nil
	}
	page := 1
	size := 20
	if p, ok := args["page"].(float64); ok && p > 0 {
		page = int(p)
	}
	if s, ok := args["size"].(float64); ok && s > 0 && s <= 100 {
		size = int(s)
	}
	statusFilter, _ := args["status"].(string)
	typeFilter, _ := args["session_type"].(string)

	listReq := &apiQa.ListSessionRequest{
		Page:   page,
		Size:   size,
		Status: statusFilter,
		Type:   typeFilter,
	}

	resp, xErr := qaLogic.ListSessions(ctx, listReq)
	if xErr != nil {
		return textResult(fmt.Sprintf("获取会话列表失败: %s", xErr.Error())), nil
	}

	totalPages := (resp.Total + int64(size) - 1) / int64(size)
	result := fmt.Sprintf("[RESPONSE] 会话列表（共 %d 个，第 %d/%d 页）：\n\n", resp.Total, page, totalPages)
	for i, s := range resp.Items {
		result += fmt.Sprintf("%d. [%s] %s", i+1, s.ID, s.Title)
		if s.ProjectName != "" {
			result += fmt.Sprintf("（%s）", s.ProjectName)
		}
		result += fmt.Sprintf(" | 类型: %s | 状态: %s | 在线: %d\n", s.Type, s.Status, s.OnlineDevices)
	}
	if len(resp.Items) == 0 {
		result += "（暂无会话）\n"
	}
	return textResult(result), nil
}

// handleQaSessionGet 获取会话信息
func handleQaSessionGet(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if qaLogic == nil {
		return textResult("QaLogic 未初始化，请联系管理员"), nil
	}

	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return textResult(errMsg), nil
	}
	sessionID, _ := args["session_id"].(string)
	if sessionID == "" {
		return textResult("缺少必填参数: session_id"), nil
	}

	result, xErr := qaLogic.GetSessionMCP(ctx, sessionID)
	if xErr != nil {
		return textResult(fmt.Sprintf("获取会话失败: %s", xErr.Error())), nil
	}

	return textResult(result), nil
}

// handleQaSessionArchive 归档会话
func handleQaSessionArchive(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if qaLogic == nil {
		return textResult("QaLogic 未初始化，请联系管理员"), nil
	}

	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return textResult(errMsg), nil
	}
	sessionID, _ := args["session_id"].(string)
	if sessionID == "" {
		return textResult("缺少必填参数: session_id"), nil
	}

	xErr := qaLogic.ArchiveSession(ctx, sessionID)
	if xErr != nil {
		return textResult(fmt.Sprintf("归档会话失败: %s", xErr.Error())), nil
	}

	return textResult(fmt.Sprintf("[RESPONSE] 会话已归档（ID: %s）", sessionID)), nil
}

// handleQaPushQuestion 推送问题到会话
func handleQaPushQuestion(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if qaLogic == nil {
		return textResult("QaLogic 未初始化，请联系管理员"), nil
	}

	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return textResult(errMsg), nil
	}

	// 提取必填参数
	sessionID, _ := args["session_id"].(string)
	if sessionID == "" {
		return textResult("缺少必填参数: session_id"), nil
	}
	qType, _ := args["question_type"].(string)
	if qType == "" {
		return textResult("缺少必填参数: question_type"), nil
	}
	if !qaKnownTypes[qType] {
		return textResult(fmt.Sprintf(
			"[ERROR] 未知的问题类型: %q。\n[RULE] 题型必须来自 qa_what_question 的返回，不允许猜测。\n请先调用 qa_what_question（传入你想要的 question_type）获取正确参数格式，或调用 qa_what_question（不传参）查看全部受支持的类型列表。\n\n受支持的类型：%s",
			qType, strings.Join(qaKnownTypeList, " / "),
		)), nil
	}
	content, _ := args["content"].(string)
	if content == "" {
		return textResult("缺少必填参数: content"), nil
	}

	// 可选参数
	description, _ := args["description"].(string)
	var options, config, batch any
	if v, ok := args["options"]; ok {
		options = v
	}
	if v, ok := args["config"]; ok {
		config = v
	}
	if v, ok := args["batch"]; ok {
		batch = v
	}
	groupLabel, _ := args["group_label"].(string)
	var supplement bool
	if s, ok := args["supplement"].(bool); ok {
		supplement = s
	}

	qID, optionIDMap, xErr := qaLogic.PushQuestion(ctx, sessionID, qType, content, description, options, config, batch, groupLabel, supplement)
	if xErr != nil {
		return textResult(fmt.Sprintf("推送问题失败: %s", xErr.Error())), nil
	}

	result := fmt.Sprintf("[RESPONSE] 问题已推送（ID: %s）\n", qID)
	if len(optionIDMap) > 0 {
		result += "[OPTIONS]\n"
		for label, id := range optionIDMap {
			result += fmt.Sprintf("    - %s → %s\n", label, id)
		}
	}
	if supplement {
		result += "[REQUIRE_SUPPLEMENT] 您已选中为此问题传递补充详情内容（supplement: true）。" +
			"请使用 qa_push_supplement 为该问题（或其各选项）推送 Markdown/HTML 详情后，" +
			"再调用 qa_get_answer 等待用户回答。否则前端将持续等待补充内容而阻塞用户操作。\n"
	}
	result += "[TIP] 使用 qa_get_answer 等待用户回答"

	return textResult(result), nil
}

// handleQaPushSupplement 推送补充内容
func handleQaPushSupplement(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if qaLogic == nil {
		return textResult("QaLogic 未初始化，请联系管理员"), nil
	}

	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return textResult(errMsg), nil
	}
	sessionID, _ := args["session_id"].(string)
	if sessionID == "" {
		return textResult("缺少必填参数: session_id"), nil
	}
	questionID, _ := args["question_id"].(string)
	if questionID == "" {
		return textResult("缺少必填参数: question_id"), nil
	}
	content, _ := args["content"].(string)
	if content == "" {
		return textResult("缺少必填参数: content"), nil
	}

	// 确定目标类型和目标 ID
	targetType := "question"
	targetIDStr := questionID

	// 如果提供了 option_id，则目标类型为 option
	if optID, _ := args["option_id"].(string); optID != "" {
		targetType = "option"
		targetIDStr = optID
	}

	// 解析 targetID 为 SnowflakeID
	var targetID xSnowflake.SnowflakeID
	if parsedTID, err := xSnowflake.ParseSnowflakeID(targetIDStr); err != nil {
		return textResult(fmt.Sprintf("无效的 ID 格式: %s", targetIDStr)), nil
	} else {
		targetID = parsedTID
	}

	// 内容类型
	contentType := "markdown"
	if ct, _ := args["content_type"].(string); ct == "html" {
		contentType = "html"
	}

	xErr := qaLogic.PushSupplement(ctx, sessionID, targetType, targetID, contentType, content)
	if xErr != nil {
		return textResult(fmt.Sprintf("推送补充内容失败: %s", xErr.Error())), nil
	}

	// 查询关联问题的选项列表，静默降级
	var optionLines []string
	if parsedQID, err := xSnowflake.ParseSnowflakeID(questionID); err == nil {
		if q, qErr := qaLogic.GetQuestionByID(ctx, parsedQID); qErr == nil && len(q.Options) > 0 {
			var options []map[string]interface{}
			if json.Unmarshal(q.Options, &options) == nil {
				for _, opt := range options {
					if label, _ := opt["label"].(string); label != "" {
						if id, _ := opt["id"].(string); id != "" {
							optionLines = append(optionLines, fmt.Sprintf("  - %s → %s", label, id))
						}
					}
				}
			}
		}
	}

	result := "[RESPONSE] 补充内容已推送"
	if len(optionLines) > 0 {
		result += "\n\n[TIP] 具体选项可以单独提交详细信息，使用 qa_push_supplement 并传入 option_id 参数\n\n关联选项:\n"
		for _, line := range optionLines {
			result += line + "\n"
		}
	}

	return textResult(result), nil
}

// handleQaWhatQuestion 返回问题类型帮助信息
func handleQaWhatQuestion(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return textResult(errMsg), nil
	}

	// 如果指定了具体类型，返回该类型的详细帮助
	if qType, _ := args["question_type"].(string); qType != "" {
		if detail, ok := qaTypeDetails[qType]; ok {
			return textResult(detail), nil
		}
		return textResult(fmt.Sprintf("未知的问题类型: %s\n\n%s", qType, qaTypeListText)), nil
	}

	// 未指定类型，返回概览
	return textResult(qaHelpText), nil
}

// handleGetAnswer 阻塞获取回答
func handleGetAnswer(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if qaLogic == nil {
		return textResult("QaLogic 未初始化，请联系管理员"), nil
	}

	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return textResult(errMsg), nil
	}
	sessionID, _ := args["session_id"].(string)
	if sessionID == "" {
		return textResult("缺少必填参数: session_id"), nil
	}

	// 读取配置：poll_slice 和 max_retries
	cfg, xErr := qaLogic.GetQaConfig(ctx)
	if xErr != nil {
		return textResult(fmt.Sprintf("读取配置失败: %s", xErr.Error())), nil
	}

	actualTimeout := fallbackPollSlice
	if cfg.PollSlice > 0 {
		actualTimeout = time.Duration(cfg.PollSlice) * time.Second
	}
	maxRetries := cfg.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 36
	}

	result, xErr := qaLogic.GetAnswer(ctx, sessionID, actualTimeout)
	if xErr != nil {
		return textResult(fmt.Sprintf("获取回答失败: %s", xErr.Error())), nil
	}

	if result == "" {
		// 增加重试计数器
		retryCount, incrErr := qaLogic.IncrementGetAnswerRetry(ctx, sessionID, cfg.SessionTTL)
		if incrErr != nil {
			// 计数器失败时降级为完整 PENDING（拿不到计数，保守返回完整规则，确保禁 sleep 约束不丢失）
			return textResult(
				"[PENDING] 用户暂未回答，立即重新调用 qa_get_answer 继续等待。\n" +
					"[RULE] 禁止 sleep、禁止改用 qa_reget_answer 轮询。",
			), nil
		}

		if retryCount >= maxRetries {
			// 终止态：始终完整提示（AI 需明确指令切换为自然语言询问）
			return textResult(
				"[STOPPED] 已等待较长时间仍未收到用户回答。\n" +
					"[RULE] 禁止使用 bash sleep 或任何休眠手段，禁止在此期间调用 qa_reget_answer 轮询。\n" +
					"请立即停止调用任何工具，直接用自然语言告知用户：「如果您已回答，请回复\"继续\"，我将获取您的回答。」用户回复\"继续\"后，再次调用 qa_get_answer。",
			), nil
		}

		// 分级返回：前 noteFullRetries 次完整提示，之后简写以降低上下文开销
		if retryCount <= noteFullRetries {
			return textResult(fmt.Sprintf(
				"[PENDING] 用户暂未回答，立即重新调用 qa_get_answer 继续阻塞等待。\n"+
					"[RETRY] %d/%d（未达上限，按规则重新调用）\n"+
					"[NOTE] 禁止使用 bash sleep 或任何休眠手段等待；禁止改用 qa_reget_answer 轮询。qa_get_answer 自身已阻塞，多次调用的 token 开销是预期且可接受的。",
				retryCount, maxRetries,
			)), nil
		}
		return textResult(fmt.Sprintf(
			"[PENDING] 用户暂未回答，立即重新调用 qa_get_answer。\n[RETRY] %d/%d\n[NOTE] 同前述规则。",
			retryCount, maxRetries,
		)), nil
	}

	return textResult(result), nil
}

// handleRegetAnswer 批量非阻塞获取已回答问题的内容
func handleRegetAnswer(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if qaLogic == nil {
		return textResult("QaLogic 未初始化，请联系管理员"), nil
	}

	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return textResult(errMsg), nil
	}
	sessionID, _ := args["session_id"].(string)
	if sessionID == "" {
		return textResult("缺少必填参数: session_id"), nil
	}

	// 解析 question_ids 数组
	var questionIDs []string
	if ids, ok := args["question_ids"].([]interface{}); ok {
		for _, id := range ids {
			if s, ok := id.(string); ok {
				questionIDs = append(questionIDs, s)
			}
		}
	}
	if len(questionIDs) == 0 {
		return textResult("缺少必填参数: question_ids（至少需要一个问题 ID）"), nil
	}

	result, xErr := qaLogic.RegetAnswers(ctx, sessionID, questionIDs)
	if xErr != nil {
		return textResult(fmt.Sprintf("获取回答失败: %s", xErr.Error())), nil
	}

	return textResult(result), nil
}

// handleQaCancelQuestion 取消指定问题或全部待回答问题
func handleQaCancelQuestion(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if qaLogic == nil {
		return textResult("QaLogic 未初始化，请联系管理员"), nil
	}

	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return textResult(errMsg), nil
	}
	sessionID, _ := args["session_id"].(string)
	if sessionID == "" {
		return textResult("缺少必填参数: session_id"), nil
	}

	questionID, _ := args["question_id"].(string)
	cancelType, _ := args["type"].(string)

	// question_id 和 type 至少传一个
	if questionID == "" && cancelType != "all" {
		return textResult("请提供 question_id（取消单个问题）或 type=all（取消全部待回答问题）"), nil
	}

	cancelled, skipped, xErr := qaLogic.CancelQuestion(ctx, sessionID, questionID, cancelType == "all")
	if xErr != nil {
		return textResult(fmt.Sprintf("取消问题失败: %s", xErr.Error())), nil
	}

	if cancelType == "all" {
		return textResult(fmt.Sprintf(
			"[RESPONSE] 已取消会话 %s 的全部待回答问题\n[CANCELLED] %d 个问题已取消\n[SKIPPED] %d 个问题无法取消（非 pending 状态）",
			sessionID, cancelled, skipped,
		)), nil
	}

	if cancelled > 0 {
		return textResult(fmt.Sprintf(
			"[RESPONSE] 问题 %s 已取消（状态更新为 cancelled）",
			questionID,
		)), nil
	}

	return textResult(fmt.Sprintf(
		"[RESPONSE] 问题 %s 无法取消（可能已回答或不存在）",
		questionID,
	)), nil
}

// ─── Helpers ────────────────────────────────────────────────────────────

// textResult 快速构建纯文本 CallToolResult
func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: text},
		},
	}
}

// stubToolHandler 返回通用的「尚未实现」存根响应。
func stubToolHandler(toolName string) mcp.ToolHandler {
	return func(_ context.Context, _ *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return textResult(toolName + ": 尚未实现"), nil
	}
}
