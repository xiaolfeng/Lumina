package mcp

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/xiaolfeng/Lumina/internal/logic"
)

// qaLogic 保存 QaLogic 实例，供 MCP 工具处理器使用。
var qaLogic *logic.QaLogic

// SetQaLogic 设置 QaLogic 实例，供 MCP 工具处理器使用。
func SetQaLogic(l *logic.QaLogic) { qaLogic = l }

const (
	// fallbackPollSlice 是 qa_get_answer 单次阻塞上限的兜底值。
	// 当 DB 配置（qa.get-answer.poll-slice）读取失败时使用。
	// 必须小于 MCP 客户端的 tool 执行超时（如 ZCode 约 30s）。
	fallbackPollSlice = 25 * time.Second

	// noteFullRetries 是 qa_get_answer 返回完整 [NOTE] 提示的次数上限。
	// 超过此值后改为简写（[NOTE] 同前述规则），降低多次调用的上下文开销。
	noteFullRetries = 5
)

// qaKnownTypes 是全部受支持的题型集合，供 qa_push_question 校验与白名单兜底共用。
// 必须与 qaTypeDetails（qa_type_details.go）的 key 保持一致；新增题型时两处同步更新。
var qaKnownTypes = map[string]bool{
	"select":       true,
	"multi-select": true,
	"text":         true,
	"boolean":      true,
	"code":         true,
	"image":        true,
	"file":         true,
	"diff":         true,
	"plan":         true,
	"options":      true,
	"review":       true,
	"slider":       true,
	"rank":         true,
	"rate":         true,
}

// qaKnownTypeList 用于错误提示中列举合法题型名（固定顺序，便于阅读）。
var qaKnownTypeList = []string{
	"select", "multi-select", "text", "boolean", "code", "image", "file",
	"diff", "plan", "options", "review", "slider", "rank", "rate",
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
		description: `创建一个 Q&A 问答会话，用于与用户进行富交互式问答。会话创建后需通过 qa_push_question 推送问题、qa_get_answer 等待回答。

触发场景：Agent 需要向用户提问并等待回复时使用。适合需要用户确认决策、选择方案、补充信息等交互场景，优于直接猜测用户意图。

创建前应先用 qa_session_list 检查是否已有可复用的活跃临时会话，避免创建过多会话。创建会话需要关联一个已有的 project_id，若不确定项目 ID 可先通过 project_get 或 project_list 查询。

推荐使用流程：
1. 用 qa_session_list 检查已有会话（可选，避免重复创建）
2. 创建会话
3. 用 qa_push_question 推送问题
4. 用 qa_get_answer 阻塞等待用户回答`,
		inputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id": map[string]any{
					"type":        "string",
					"description": "关联的项目 ID（必填），会话将绑定到该项目",
				},
				"session_type": map[string]any{
					"type":        "string",
					"description": "会话类型：temporary（48 小时过期）或 permanent（永久）",
					"enum":        []string{"temporary", "permanent"},
				},
				"title": map[string]any{
					"type":        "string",
					"description": "会话标题，用于浏览器展示，可选",
				},
				"agent_name": map[string]any{
					"type":        "string",
					"description": "Agent 名称标识，用于区分不同 Agent 创建的会话，可选（默认 mcp-agent）",
				},
			},
			"required": []string{"project_id"},
		},
	},
	{
		name: "qa_session_list",
		description: `获取 Q&A 会话列表，支持按状态和类型过滤。

触发场景：创建新会话前先调用此工具检查是否有可复用的 active 临时会话，避免创建过多会话。也可以用于查看当前有多少活跃会话、或归档/清理过期会话前先列出。

使用建议：如果返回结果中有 status=active 且 type=temporary 的会话，可直接复用该会话推送新问题，无需创建新会话。若会话已不活跃或已过期，再使用 qa_session_create 创建新会话。`,
		inputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"status": map[string]any{
					"type":        "string",
					"description": "按状态过滤：active、expired、deleted。不传则返回全部",
				},
				"session_type": map[string]any{
					"type":        "string",
					"description": "按类型过滤：temporary 或 permanent。不传则返回全部",
				},
				"page": map[string]any{
					"type":        "integer",
					"description": "页码（从 1 开始，默认 1）",
				},
				"size": map[string]any{
					"type":        "integer",
					"description": "每页数量（默认 20，最大 100）",
				},
			},
		},
	},
	{
		name: "qa_session_get",
		description: `获取 Q&A 会话的状态信息，包含会话元数据和所有问题列表。

决策引导：
- 确认会话仍然活跃：如果 status 不是 active，不要再推送问题，应先创建新会话
- 查看回答进度：对比已回答/总数，判断是否还有待回答的问题
- 确认用户在线：如果 qa_get_answer 返回 [STOPPED]，可检查 online_devices 是否为 0（用户已断开）。
  若用户在线仍无响应，禁止使用 bash sleep 或任何休眠手段，应直接用自然语言催促用户；待用户回复"继续"后再次 qa_get_answer。

如果会话不存在或已删除，将返回错误提示。`,
		inputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"session_id": map[string]any{
					"type":        "string",
					"description": "要查询的会话 ID",
				},
			},
			"required": []string{"session_id"},
		},
	},
	{
		name: "qa_session_archive",
		description: `归档一个 Q&A 会话，将其转为只读状态。归档后会话数据保留，但无法再推送新问题或等待回答。

触发场景：Agent 确认当前会话的所有问答已完成、用户主动要求结束会话、或 temporary 会话即将过期时主动归档。归档是安全的——数据不丢失，仍可通过 qa_session_get 查看历史。`,
		inputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"session_id": map[string]any{
					"type":        "string",
					"description": "要归档的会话 ID",
				},
			},
			"required": []string{"session_id"},
		},
	},
	{
		name: "qa_push_question",
		description: `向指定会话推送一个交互式问题，推送后实时送达用户浏览器。

[RULE] 题型必须来自 qa_what_question 的返回，不允许凭记忆或猜测构造：
  1. 首次使用某题型前，必须先调用 qa_what_question（传 question_type）获取该类型的参数格式与 JSON 示例。
  2. 严格按返回示例构造 options/config/batch/supplement 等字段，不得自行发明字段或题型名。
  3. 后端仅接受下列 14 种标准题型名，传入其它名称会被拒绝：
     select / multi-select / text / boolean / code / image / file / diff / plan / options / review / slider / rank / rate

14 种交互类型速查：
  选择类：select、multi-select
  输入类：text、boolean、code、image、file
  展示类：diff、plan、options、review
  评分类：slider、rank、rate

不确定如何构造某种类型的参数时，先调用 qa_what_question 获取完整用法说明和示例（即便自认为熟悉也建议查一次，题型参数会迭代）。

推荐实践：
- 选项的 label 应简洁（1-5 词），详细说明通过 qa_push_supplement 补充到右侧面板
- 推送后立即调用 qa_get_answer 等待回答，避免推送后忘记等待
- 连续推送多个问题时，最后一次性 qa_get_answer 即可收取所有回答`,
		inputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"session_id": map[string]any{
					"type":        "string",
					"description": "目标会话 ID",
				},
				"question_type": map[string]any{
					"type":        "string",
					"description": "问题类型标识符。必须为 qa_what_question 返回的 14 种标准题型之一，严禁猜测或自创题型名。",
					"enum":        qaKnownTypeList,
				},
				"content": map[string]any{
					"type":        "string",
					"description": "问题内容，支持 Markdown 格式",
				},
				"options": map[string]any{
					"type":        "array",
					"description": "选项列表，选择类问题必填",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"label":       map[string]any{"type": "string", "description": "选项标签"},
							"description": map[string]any{"type": "string", "description": "选项说明，可选"},
						},
					},
				},
				"description": map[string]any{
					"type":        "string",
					"description": "问题的补充描述，可选",
				},
				"config": map[string]any{
					"type":        "object",
					"description": "问题配置，如 min/max/step 等，可选",
				},
				"batch": map[string]any{
					"type":        "object",
					"description": "批量推送配置，可选",
				},
				"group_label": map[string]any{
					"type":        "string",
					"description": "问题分组标签，可选",
				},
				"supplement": map[string]any{
					"type":        "boolean",
					"description": "是否携带补充内容，若为 true 前端将等待补充内容推送后才允许用户操作",
				},
			},
			"required": []string{"session_id", "question_type", "content"},
		},
	},
	{
		name: "qa_push_supplement",
		description: `用途：为已推送的问题或某个选项写入一份右侧详情面板内容。每个目标只保留一份 supplement；对同一 question_id/option_id 再次调用会覆盖旧内容。

何时调用：qa_push_question 使用 supplement=true、选项需要展开说明，或 qa_get_answer 返回 [NEED_SUPPLEMENT] 时调用。它不会创建问题、Preview 文件或浏览器页面；目标问题必须已存在。

content_type 选择：
- markdown（默认）：技术说明、约束、代码、表格、Mermaid、KaTeX；浏览器与 Agent 都可读。凡是后续推理必须保留的信息都应使用 markdown。
- html：一个自包含的展示片段或轻量交互；仅浏览器渲染，不返回给 Agent。不要用它承载 Agent 后续必须读取的约束。
- preview：多文件或可复用前端预览；仅浏览器渲染。必须先完成 preview_session_create/list → preview_file_upload → preview_file_list，并将 Preview 返回的 qa_supplement.content 原样作为 content。

preview 的 content 必须是且只能是 JSON 字符串 {"session_id":"...","file_id":"..."}；file_id 决定实际渲染入口。不要传 preview hash、preview_url、HTML 源码，不要添加 Markdown 围栏或解释文字。

下一步：问题/各选项需要的 supplement 全部推送后，确保用户已进入 qa_session_create 返回的交互 URL，再调用 qa_get_answer 等待回答。`,
		inputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"session_id": map[string]any{
					"type":        "string",
					"description": "目标会话 ID",
				},
				"question_id": map[string]any{
					"type":        "string",
					"description": "要补充内容的问题 ID",
				},
				"option_id": map[string]any{
					"type":        "string",
					"description": "选项 ID，提供时关联到特定选项而非整个问题，可选",
				},
				"content": map[string]any{
					"type":        "string",
					"description": "补充内容。markdown/html 时为正文；preview 时必须原样传入 Preview 工具返回的 qa_supplement.content JSON 字符串。",
				},
				"content_type": map[string]any{
					"type":        "string",
					"description": "内容类型：markdown / html / preview，默认 markdown",
					"enum":        []string{"markdown", "html", "preview"},
				},
			},
			"required": []string{"session_id", "question_id", "content"},
		},
	},
	{
		name: "qa_what_question",
		description: `查询 Q&A 模块支持的问题类型及详细用法。

不传参数时返回全部 14 种类型的概览列表；传入具体类型时返回该类型的完整用法说明，包括参数格式、JSON 示例和注意事项。

[RULE] 这是构造 qa_push_question 参数的权威来源。在使用 qa_push_question 推送任一题型前，必须先调用本工具（传入该 question_type）获取参数格式与 JSON 示例，严格按返回结构构造，不允许凭记忆猜测。qa_push_question 仅接受本工具返回的 14 种标准题型名。`,
		inputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"question_type": map[string]any{
					"type":        "string",
					"description": "要查询的具体类型名称，如 select、text、code 等。不传则返回全部类型概览。",
				},
			},
		},
	},
	{
		name: "qa_get_answer",
		description: `阻塞等待用户对当前会话的回答。调用后挂起直到用户提交回答、跳过、或发送补全请求。

用户可以自由选择优先回答哪个问题，因此返回的是当前所有待消费的回答，按用户实际回答顺序排列。

返回文本格式说明：
- [ANSWERED] 用户已回答
- [SKIPPED] 用户跳过了此问题
- [NEED_SUPPLEMENT] 用户请求补充内容 → 用 qa_push_supplement 推送后再次 qa_get_answer
- [IMAGE_URL] / [IMAGE_BASE64] 多媒体附件（如需原始 base64 数据用 qa_reget_answer）
- [PENDING] 用户暂未回答，立即重新调用 qa_get_answer 继续阻塞等待
- [STOPPED] 已等待较长时间仍未收到用户回答。停止调用任何工具，直接用自然语言告知用户：「如果您已回答，请回复"继续"，我将获取您的回答。」用户回复"继续"后，再次调用 qa_get_answer。

行为：队列有待消费回答时立即返回全部；队列为空时阻塞等待，单次最多阻塞约 25 秒。
若到时仍无回答，按重试次数返回 [PENDING] 或 [STOPPED]。消费后队列清空。

[RULE] 必须遵守：
  1. 禁止使用 bash sleep / 任何休眠命令来"等待"。本工具自身已实现阻塞等待，无需也不允许额外休眠。
  2. 禁止在等待回答期间用 qa_reget_answer 轮询——qa_reget_answer 仅用于"重新获取已回答内容或多媒体附件"，不能替代本工具。
  3. 收到 [PENDING] 时立即重新调用本工具，多次调用产生的 token 开销是预期且可接受的。`,
		inputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"session_id": map[string]any{
					"type":        "string",
					"description": "目标会话 ID",
				},
			},
			"required": []string{"session_id"},
		},
	},
	{
		name: "qa_reget_answer",
		description: `非阻塞批量获取已回答问题的内容，立即返回不等待。

与 qa_get_answer 的区别：qa_get_answer 是阻塞等待用户回答的唯一方式；qa_reget_answer 是非阻塞"重新获取"，仅用于已回答内容的重取。

典型使用场景：
- qa_get_answer 返回了图片 URL 但 Agent 无法访问，使用 qa_reget_answer 获取详细数据
- 需要重新查看已回答问题的内容
- 获取回答中的多媒体附件
- 一次性下载令牌失效后，重新获取新令牌

如果问题尚未回答，返回 [PENDING] 标记而非报错。

[RULE] 仅在上述"重新获取已回答内容"场景使用。严禁用于等待用户回答期间的轮询——等待回答必须且只能用 qa_get_answer。`,
		inputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"session_id": map[string]any{
					"type":        "string",
					"description": "目标会话 ID",
				},
				"question_ids": map[string]any{
					"type":        "array",
					"description": "要获取回答的问题 ID 列表",
					"items":       map[string]any{"type": "string"},
				},
			},
			"required": []string{"session_id", "question_ids"},
		},
	},
	{
		name: "qa_cancel_question",
		description: `取消指定会话中等待回答的问题，将问题状态标记为已取消。

支持两种取消模式：
  - 指定 question_id：取消单个特定问题
  - type=all：取消该会话的全部待回答问题（包括队列中的回答）

取消操作会：
  1. 将问题状态从 pending 更新为 cancelled
  2. 清空该会话回答队列中的所有待消费回答
  3. 通过 WebSocket 通知在线设备问题已取消

触发场景：
  - Agent 推送了错误的问题需要撤回
  - 用户明确表示不想回答某些问题
  - 会话流程需要中止当前等待，重新推送新问题
  - Agent 决策变更，不再需要之前的问题答案

注意：已回答（answered）的问题无法取消，仅 pending 状态的问题可取消。`,
		inputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"session_id": map[string]any{
					"type":        "string",
					"description": "目标会话 ID",
				},
				"question_id": map[string]any{
					"type":        "string",
					"description": "要取消的问题 ID。与 type 二选一，同时传入时以 question_id 优先",
				},
				"type": map[string]any{
					"type":        "string",
					"description": "取消模式：传 \"all\" 表示取消该会话全部待回答问题",
					"enum":        []string{"all"},
				},
			},
			"required": []string{"session_id"},
		},
	},
}

// parseArgs 将 json.RawMessage 格式的参数解析为 map[string]any。
// 解析失败时返回包含 __parse_error 键的 map，供调用方判断并返回具体错误信息。
func parseArgs(raw json.RawMessage) map[string]any {
	args := make(map[string]any)
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &args); err != nil {
			args["__parse_error"] = err.Error()
		}
	}
	return args
}

// checkParseError 检查参数是否因 JSON 解析失败而包含错误信息。
func checkParseError(args map[string]any) string {
	if errMsg, ok := args["__parse_error"].(string); ok {
		return fmt.Sprintf("参数 JSON 解析失败: %s", errMsg)
	}
	return ""
}

// RegisterQATools 将 Q&A 模块的 9 个 MCP 工具注册到 Server。
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
		case "qa_session_list":
			handler = handleQaSessionList
		case "qa_session_get":
			handler = handleQaSessionGet
		case "qa_session_archive":
			handler = handleQaSessionArchive
		case "qa_push_question":
			handler = handleQaPushQuestion
		case "qa_push_supplement":
			handler = handleQaPushSupplement
		case "qa_what_question":
			handler = handleQaWhatQuestion
		case "qa_get_answer":
			handler = handleGetAnswer
		case "qa_reget_answer":
			handler = handleRegetAnswer
		case "qa_cancel_question":
			handler = handleQaCancelQuestion
		default:
			handler = stubToolHandler(def.name)
		}

		server.AddTool(tool, handler)
	}
}
