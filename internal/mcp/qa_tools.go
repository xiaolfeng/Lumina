package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/xiaolfeng/Lumina/internal/logic"
)

// qaLogic 保存 QaLogic 实例，供 MCP 工具处理器使用。
var qaLogic *logic.QaLogic

// SetQaLogic 设置 QaLogic 实例，供 MCP 工具处理器使用。
func SetQaLogic(l *logic.QaLogic) {
	qaLogic = l
}

// qaToolDefs 定义 Q&A 模块的全部 MCP 工具。
// 每个条目包含工具名、描述、JSON Schema 输入参数定义。
var qaToolDefs = []struct {
	name        string
	description string
	inputSchema map[string]any
}{
	{
		name: "qa_session_create",
		description: `创建一个 Q&A 问答会话。Agent 通过此工具与用户进行富交互式问答。

创建成功后返回 session_id 和浏览器访问链接，用户在浏览器中实时接收问题并回答。

支持两种会话类型：
- temporary（默认）：48 小时后过期，适合临时问答场景
- permanent：永不过期，支持跨设备协同（手机、电脑同时在线）

使用流程：创建会话 → qa_push_question 推送问题 → get_answer 阻塞等待回答。

注意：如果已有活跃的临时会话可直接复用，请先用 qa_session_get 检查，避免创建过多会话。`,
		inputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"session_type": map[string]any{
					"type":        "string",
					"description": "Session type: 'temporary' (48h TTL) or 'permanent'",
					"enum":        []string{"temporary", "permanent"},
				},
				"title": map[string]any{
					"type":        "string",
					"description": "Optional session title for display",
				},
			},
		},
	},
	{
		name: "qa_session_get",
		description: `获取 Q&A 会话的状态信息，包含会话元数据和所有问题列表。

返回内容：会话标题、状态（active/expired/deleted）、Agent 名称、在线设备数、进度（已答/总数）、过期时间、所有问题及其状态。

使用场景：
- 推送问题前检查会话是否仍然活跃
- 查看用户的回答进度
- 检查是否存在可复用的活跃会话

如果会话不存在或已删除，将返回引导提示帮助定位问题。`,
		inputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"session_id": map[string]any{
					"type":        "string",
					"description": "The session ID to query",
				},
			},
			"required": []string{"session_id"},
		},
	},
	{
		name: "qa_session_delete",
		description: `删除一个 Q&A 会话。删除后所有问题和回答数据将不可恢复。

仅在确认不再需要该会话时使用。活跃会话删除后，用户浏览器会收到通知并退出当前问答界面。

如果需要保留历史记录，请勿删除——过期的会话会自动归档为只读状态，仍然可以查询。

注意：删除前建议确认用户已不需要该会话的内容。`,
		inputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"session_id": map[string]any{
					"type":        "string",
					"description": "The session ID to delete",
				},
			},
			"required": []string{"session_id"},
		},
	},
	{
		name: "qa_push_question",
		description: `向指定会话推送一个交互式问题。推送后问题通过 WebSocket 实时送达用户浏览器。

支持 14 种交互类型，每种类型有特定的参数要求。传入 need_help=true 可获取所有类型的详细用法说明，传入 need_help_type="select,diff" 可获取指定类型的参数格式和示例。

【快速参考 — 14 种类型】
选择类：select（单选）、multi-select（多选）
输入类：text（文本）、boolean（是/否）、code（代码）、image（图片）、file（文件）
展示类：diff（代码差异）、plan（方案展示）、options（带优劣选项）、review（内容审阅）
评分类：slider（滑动条）、rank（排序）、rate（评分）

推送成功后返回 question_id 和每个选项的 option_id（若有选项）。
这些 ID 用于后续操作：get_answer 等待回答、qa_push_supplement 推送补充内容。

选项描述应简洁，如需详细说明可通过 qa_push_supplement 单独推送 Markdown/HTML 补充内容，右侧面板即时渲染。`,
		inputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"session_id": map[string]any{
					"type":        "string",
					"description": "Target session ID",
				},
				"question_type": map[string]any{
					"type":        "string",
					"description": "Question type identifier (e.g. text_input, single_choice, multi_choice, etc.)",
				},
				"content": map[string]any{
					"type":        "string",
					"description": "Question content in Markdown format",
				},
				"options": map[string]any{
					"type":        "array",
					"description": "Options for choice-type questions",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"label":       map[string]any{"type": "string", "description": "Option label"},
							"description": map[string]any{"type": "string", "description": "Optional option description"},
						},
					},
				},
				"need_help": map[string]any{
					"type":        "boolean",
					"description": "If true, returns a usage guide instead of creating a question",
				},
				"need_help_type": map[string]any{
					"type":        "string",
					"description": "Specific question type to get help for (used with need_help=true)",
				},
			},
			"required": []string{"session_id", "question_type", "content"},
		},
	},
	{
		name: "qa_push_supplement",
		description: `为已推送的问题或选项补充详细内容。补充内容在用户浏览器右侧面板即时渲染。

支持 Markdown 和 HTML 两种格式：
- Markdown：适合文本说明、代码块、表格、Mermaid 流程图
- HTML：适合交互式预览、自定义布局、动画演示

关联方式：
- target_type="question" + target_id=question_id → 为整个问题补充背景信息
- target_type="option" + target_id=option_id → 为特定选项补充详细说明

重要：每个目标是 1:1 映射，重复推送会覆盖之前的内容。

使用场景：
- 选项的 description 不够详细，需要展开说明
- 展示架构图、流程图等可视化内容
- 提供代码示例或配置模板
- 用户主动请求补全时（get_answer 返回 [NEED_SUPPLEMENT]）

推送后用户仍停留在当前问题，可继续回答。`,
		inputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"session_id": map[string]any{
					"type":        "string",
					"description": "Target session ID",
				},
				"question_id": map[string]any{
					"type":        "string",
					"description": "Question ID to supplement",
				},
				"option_id": map[string]any{
					"type":        "string",
					"description": "Optional option ID for option-level supplement",
				},
				"content": map[string]any{
					"type":        "string",
					"description": "Supplement content in Markdown or HTML format",
				},
			},
			"required": []string{"session_id", "question_id", "content"},
		},
	},
	{
		name: "get_answer",
		description: `阻塞等待用户对当前会话的回答。调用后将挂起直到用户提交回答、跳过、或发送补全请求。

用户可以自由选择优先回答哪个问题，因此不要求传入 question_id——返回的是当前所有待消费的回答，按用户实际回答顺序排列。

返回格式为多行文本（非 JSON），多个回答之间用 "--- question:<id> ---" 分隔。
每个回答使用前缀标识不同类型的响应：
- [ANSWERED] 用户已回答，后接回答内容摘要
- [SKIPPED] 用户跳过了此问题
- [NEED_SUPPLEMENT] 用户请求补充内容，附目标 ID 和提示
- [TIP] 操作引导（如如何推送补充内容）
- [REF] 选项关联的参考内容（帮助理解用户选择的上下文）
- [IMAGE_URL] / [IMAGE_BASE64] 多媒体附件

如果返回 [NEED_SUPPLEMENT]，请使用 qa_push_supplement 推送补充内容，然后再次调用 get_answer 继续等待回答。

行为说明：
- 队列中有待消费回答：立即一次性消费完毕，按用户回答顺序返回
- 队列为空：阻塞等待直到用户回答任意一个问题
- 消费后队列清空

注意：此工具是阻塞调用。如需非阻塞获取已有答案，请使用 reget_answer。`,
		inputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"session_id": map[string]any{
					"type":        "string",
					"description": "Target session ID",
				},
				"question_ids": map[string]any{
					"type":        "array",
					"description": "Question IDs to wait for answers",
					"items":       map[string]any{"type": "string"},
				},
				"timeout": map[string]any{
					"type":        "integer",
					"description": "Maximum wait time in seconds",
				},
			},
			"required": []string{"session_id", "question_ids"},
		},
	},
	{
		name: "reget_answer",
		description: `非阻塞获取已回答问题的内容。立即返回，不等待。

与 get_answer 的区别：
- get_answer：阻塞等待首次回答（用于问题推送后的等待循环）
- reget_answer：立即返回已有答案（用于重取、获取多媒体 base64 数据）

使用场景：
- get_answer 返回了图片 URL 但 Agent 无法访问，传入 base64=true 获取原始数据
- 需要重新查看已回答问题的内容
- 获取回答中的多媒体附件

如果问题尚未回答，将返回引导提示（提示使用 get_answer 阻塞等待），不会报错。`,
		inputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"session_id": map[string]any{
					"type":        "string",
					"description": "Target session ID",
				},
				"question_ids": map[string]any{
					"type":        "array",
					"description": "Question IDs to retrieve answers for",
					"items":       map[string]any{"type": "string"},
				},
				"base64": map[string]any{
					"type":        "boolean",
					"description": "If true, return media data as base64 encoded strings",
				},
			},
			"required": []string{"session_id", "question_ids"},
		},
	},
}

// qaHelpText Q&A 问题类型使用帮助文本
const qaHelpText = `Q&A 支持以下 14 种问题类型：

📋 选择类：
  - single_choice    单选（从选项中选一个）
  - multi_choice     多选（从选项中选多个）
  - dropdown         下拉选择（从长列表中选一个）
  - cascading        级联选择（多级联动选择）

📝 文本类：
  - text_input       文本输入（单行/多行）
  - rich_text        富文本输入（支持 Markdown）
  - code_input       代码输入（语法高亮）
  - textarea         多行文本区域

✅ 确认类：
  - confirm          确认（是/否）
  - agree_terms      条款同意确认

⭐ 评分类：
  - rating           星级评分
  - slider           滑块评分

🖼️ 媒体类：
  - file_upload      文件上传
  - image_upload     图片上传

使用方式：设置 question_type 为上述类型之一，content 为问题内容（Markdown），
选项类型需额外传入 options 数组 [{label: "选项文本", description: "可选说明"}]。`

// parseArgs 将 json.RawMessage 格式的参数解析为 map[string]any。
func parseArgs(raw json.RawMessage) map[string]any {
	args := make(map[string]any)
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &args)
	}
	return args
}

// ─── Tool Handlers ──────────────────────────────────────────────────────

// handleQaSessionCreate 创建 QA 会话
func handleQaSessionCreate(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if qaLogic == nil {
		return textResult("QaLogic 未初始化，请联系管理员"), nil
	}

	args := parseArgs(req.Params.Arguments)
	title, _ := args["title"].(string)
	sessionType, _ := args["session_type"].(string)

	id, link, xErr := qaLogic.CreateSession(context.Background(), title, "open-code-agent", sessionType)
	if xErr != nil {
		return textResult(fmt.Sprintf("创建会话失败: %s", xErr.Error())), nil
	}

	return textResult(fmt.Sprintf(
		"会话创建成功！\n\n会话 ID: %s\n浏览器链接: %s\n\n请在浏览器中打开链接开始问答。",
		id, link,
	)), nil
}

// handleQaSessionGet 获取会话信息
func handleQaSessionGet(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if qaLogic == nil {
		return textResult("QaLogic 未初始化，请联系管理员"), nil
	}

	args := parseArgs(req.Params.Arguments)
	sessionID, _ := args["session_id"].(string)
	if sessionID == "" {
		return textResult("缺少必填参数: session_id"), nil
	}

	result, xErr := qaLogic.GetSessionMCP(context.Background(), sessionID)
	if xErr != nil {
		return textResult(fmt.Sprintf("获取会话失败: %s", xErr.Error())), nil
	}

	return textResult(result), nil
}

// handleQaSessionDelete 删除会话
func handleQaSessionDelete(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if qaLogic == nil {
		return textResult("QaLogic 未初始化，请联系管理员"), nil
	}

	args := parseArgs(req.Params.Arguments)
	sessionID, _ := args["session_id"].(string)
	if sessionID == "" {
		return textResult("缺少必填参数: session_id"), nil
	}

	xErr := qaLogic.DeleteSessionMCP(context.Background(), sessionID)
	if xErr != nil {
		return textResult(fmt.Sprintf("删除会话失败: %s", xErr.Error())), nil
	}

	return textResult(fmt.Sprintf("会话 %s 已删除。", sessionID)), nil
}

// handleQaPushQuestion 推送问题到会话
func handleQaPushQuestion(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if qaLogic == nil {
		return textResult("QaLogic 未初始化，请联系管理员"), nil
	}

	args := parseArgs(req.Params.Arguments)

	// 处理 need_help 请求
	if needHelp, _ := args["need_help"].(bool); needHelp {
		if helpType, _ := args["need_help_type"].(string); helpType != "" {
			return textResult(fmt.Sprintf("关于 %s 类型的帮助：\n\n请参考完整帮助文档，设置 question_type 为该类型即可。选项类型需要传入 options 数组。", helpType)), nil
		}
		return textResult(qaHelpText), nil
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

	qID, xErr := qaLogic.PushQuestion(context.Background(), sessionID, qType, content, description, options, config, batch, groupLabel)
	if xErr != nil {
		return textResult(fmt.Sprintf("推送问题失败: %s", xErr.Error())), nil
	}

	return textResult(fmt.Sprintf(
		"问题已推送！\n问题 ID: %s\n\n使用 get_answer 等待用户回答。",
		qID,
	)), nil
}

// handleQaPushSupplement 推送补充内容
func handleQaPushSupplement(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if qaLogic == nil {
		return textResult("QaLogic 未初始化，请联系管理员"), nil
	}

	args := parseArgs(req.Params.Arguments)
	sessionID, _ := args["session_id"].(string)
	if sessionID == "" {
		return textResult("缺少必填参数: session_id"), nil
	}
	content, _ := args["content"].(string)
	if content == "" {
		return textResult("缺少必填参数: content"), nil
	}

	// 从 question_id / option_id 映射到 target_type / target_id
	targetType := "question"
	targetIDStr, _ := args["question_id"].(string)
	if targetIDStr == "" {
		return textResult("缺少必填参数: question_id"), nil
	}

	// 如果提供了 option_id，则目标类型为 option
	if optID, _ := args["option_id"].(string); optID != "" {
		targetType = "option"
		targetIDStr = optID
	}

	// 解析 targetID 为 int64（雪花 ID）
	var targetID int64
	if _, err := fmt.Sscanf(targetIDStr, "%d", &targetID); err != nil {
		return textResult(fmt.Sprintf("无效的 ID 格式: %s", targetIDStr)), nil
	}

	xErr := qaLogic.PushSupplement(context.Background(), sessionID, targetType, targetID, "markdown", content)
	if xErr != nil {
		return textResult(fmt.Sprintf("推送补充内容失败: %s", xErr.Error())), nil
	}

	return textResult("补充内容已推送成功。"), nil
}

// handleGetAnswer 阻塞获取回答
func handleGetAnswer(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if qaLogic == nil {
		return textResult("QaLogic 未初始化，请联系管理员"), nil
	}

	args := parseArgs(req.Params.Arguments)
	sessionID, _ := args["session_id"].(string)
	if sessionID == "" {
		return textResult("缺少必填参数: session_id"), nil
	}

	// 解析超时时间（JSON 数字默认解析为 float64）
	var timeout time.Duration
	if timeoutVal, ok := args["timeout"]; ok {
		if tv, ok := timeoutVal.(float64); ok {
			timeout = time.Duration(tv) * time.Second
		}
	}

	result, xErr := qaLogic.GetAnswer(context.Background(), sessionID, timeout)
	if xErr != nil {
		return textResult(fmt.Sprintf("获取回答失败: %s", xErr.Error())), nil
	}

	if result == "" {
		return textResult("尚未收到回答，用户可能正在思考...\n\n提示：如需确认用户是否在线，可使用 qa_session_get 查看会话状态。"), nil
	}

	return textResult(result), nil
}

// handleRegetAnswer 重新获取已回答的问题
func handleRegetAnswer(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if qaLogic == nil {
		return textResult("QaLogic 未初始化，请联系管理员"), nil
	}

	args := parseArgs(req.Params.Arguments)
	sessionID, _ := args["session_id"].(string)
	if sessionID == "" {
		return textResult("缺少必填参数: session_id"), nil
	}

	// 从 question_ids 数组中取第一个
	questionID := ""
	if ids, ok := args["question_ids"].([]interface{}); ok && len(ids) > 0 {
		questionID, _ = ids[0].(string)
	}
	if questionID == "" {
		return textResult("缺少必填参数: question_ids"), nil
	}

	base64Flag, _ := args["base64"].(bool)

	result, xErr := qaLogic.RegetAnswer(context.Background(), sessionID, questionID, base64Flag)
	if xErr != nil {
		return textResult(fmt.Sprintf("获取回答失败: %s", xErr.Error())), nil
	}

	return textResult(result), nil
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

// RegisterQATools 将 Q&A 模块的 7 个 MCP 工具注册到 Server。
func RegisterQATools(server *mcp.Server) {
	for _, def := range qaToolDefs {
		schemaBytes, _ := json.Marshal(def.inputSchema)
		tool := &mcp.Tool{
			Name:        def.name,
			Description: def.description,
			InputSchema: json.RawMessage(schemaBytes),
		}

		var handler mcp.ToolHandler
		switch def.name {
		case "qa_session_create":
			handler = handleQaSessionCreate
		case "qa_session_get":
			handler = handleQaSessionGet
		case "qa_session_delete":
			handler = handleQaSessionDelete
		case "qa_push_question":
			handler = handleQaPushQuestion
		case "qa_push_supplement":
			handler = handleQaPushSupplement
		case "get_answer":
			handler = handleGetAnswer
		case "reget_answer":
			handler = handleRegetAnswer
		default:
			handler = stubToolHandler(def.name)
		}

		server.AddTool(tool, handler)
	}
}
