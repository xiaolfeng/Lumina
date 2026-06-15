package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	apiQa "github.com/xiaolfeng/Lumina/api/qa"
	"github.com/xiaolfeng/Lumina/internal/logic"
)

// qaLogic 保存 QaLogic 实例，供 MCP 工具处理器使用。
var qaLogic *logic.QaLogic

// SetQaLogic 设置 QaLogic 实例，供 MCP 工具处理器使用。
func SetQaLogic(l *logic.QaLogic) { qaLogic = l }

const (
	// fallbackPollSlice 是 qa_get_answer 单次阻塞上限的兜底值。
	// 当 DB 配置（qa.get_answer.poll_slice）读取失败时使用。
	// 必须小于 MCP 客户端的 tool 执行超时（如 ZCode 约 30s）。
	fallbackPollSlice = 25 * time.Second
)

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
- 确认用户在线：如果 qa_get_answer 长时间无响应，检查 online_devices 是否为 0（用户已断开）

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

14 种交互类型速查：
  选择类：select、multi-select
  输入类：text、boolean、code、image、file
  展示类：diff、plan、options、review
  评分类：slider、rank、rate

不确定如何构造某种类型的参数时，先调用 qa_what_question 获取完整用法说明和示例。

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
					"description": "问题类型标识符，如 text、select、multi-select 等",
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
		description: `为已推送的问题或选项补充详细内容，补充内容在用户浏览器右侧面板即时渲染。

⚠️ content_type 约束（重要）：
  - markdown（默认）：双通道 — 浏览器渲染 + AI 可读（作为约束/上下文返回给 Agent）
  - html：单通道 — 仅浏览器渲染（提升用户可读性），不会返回给 Agent

选择建议：
  - 技术说明、决策矩阵、约束条件 → 使用 markdown（AI 需要读取这些信息）
  - 交互式预览、动画演示、复杂布局 → 使用 html（仅展示给用户看）

触发场景：选项 label 过于简洁需要展开技术细节、展示架构图/流程图等可视化内容、
或 qa_get_answer 返回 [NEED_SUPPLEMENT] 时响应用户的补充请求。

Markdown 适合文本说明、代码渲染、表格、Mermaid 流程图、KaTeX 数学公式；
HTML 适合交互式预览、自定义布局。每个目标是 1:1 映射，重复推送会覆盖之前内容。`,
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
					"description": "补充内容，支持 Markdown 或 HTML 格式",
				},
				"content_type": map[string]any{
					"type":        "string",
					"description": "内容类型：markdown 或 html，默认 markdown",
					"enum":        []string{"markdown", "html"},
				},
			},
			"required": []string{"session_id", "question_id", "content"},
		},
	},
	{
		name: "qa_what_question",
		description: `查询 Q&A 模块支持的问题类型及详细用法。

不传参数时返回全部 14 种类型的概览列表；传入具体类型时返回该类型的完整用法说明，包括参数格式、JSON 示例和注意事项。

Agent 在使用 qa_push_question 前如对某种类型的参数结构不确定，应先调用此工具获取帮助。`,
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
- [PENDING] 用户暂未回答，请直接重新调用 qa_get_answer 继续等待，无需额外等待间隔
- [STOPPED] 已等待较长时间仍未收到用户回答。请暂停轮询，并告知用户：「如果您已回答，请告诉我"继续"以便我获取您的回答。」当用户说"继续"后，请再次调用 qa_get_answer 获取回答。

行为：队列有待消费回答时立即返回全部；队列为空时阻塞等待，单次最多阻塞约 25 秒。
若到时仍无回答，按重试次数返回 [PENDING] 或 [STOPPED]。消费后队列清空。

此工具是阻塞调用。如需非阻塞获取已有答案，请用 qa_reget_answer。`,
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

与 qa_get_answer 的区别：qa_get_answer 是阻塞等待首次回答（用于推送后的等待循环），qa_reget_answer 是立即返回已有答案（用于重取和获取多媒体数据）。

典型使用场景：
- qa_get_answer 返回了图片 URL 但 Agent 无法访问，使用 qa_reget_answer 获取详细数据
- 需要重新查看已回答问题的内容
- 获取回答中的多媒体附件

如果问题尚未回答，返回 [PENDING] 标记而非报错。`,
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

// qaTypeListText 返回全部 14 种类型的概览文本。
const qaTypeListText = `Q&A 支持以下 14 种问题类型：

选择类（需 options 参数）
  - select         单选 — 从选项列表中选择一个（最常用）
  - multi-select   多选 — 从选项列表中选择多个

输入类
  - text           文本输入 — 单行或多行自由文本（最灵活）
  - boolean        布尔确认 — 是/否二选一
  - code           代码输入 — 带语法高亮的代码编辑器，返回含语言标记
  - image          图片上传 — 支持拖拽上传，一次性令牌下载
  - file           文件上传 — 支持上传任意文件，一次性令牌下载

展示类（供审阅决策）
  - diff           差异对比 — 展示代码修改前后对比，返回最终代码
  - plan           方案展示 — 分段展示计划，返回审批结果和详情
  - options        差异化对比 — 带优缺点的方案对比（pros/cons 必填）
  - review         内容审阅 — 展示内容供逐段审阅

评分类
  - slider         滑块评分 — 在数值范围内滑动选择
  - rank           排序 — 拖拽排列选项优先级（需 options）
  - rate           星级评分 — 多维度评分（需 options，每项独立打分）`

// qaHelpText Q&A 问题类型使用帮助文本（概览版本）
const qaHelpText = qaTypeListText + `

使用方式：设置 question_type 为上述类型之一，content 为问题内容（Markdown），
选项类型需额外传入 options 数组 [{label: "选项文本", description: "可选说明"}]。

如需某类型的详细参数格式和示例，请调用 qa_what_question 并传入 question_type 参数。`

// jsonBlockStart 和 jsonBlockEnd 用于在 raw string 中包裹 JSON 代码块标记。
// Go raw string 无法包含反引号，因此通过拼接实现 ```json ... ``` 效果。
const (
	jsonBlockStart = "\n```json\n"
	jsonBlockEnd   = "\n```\n"
)

// qaTypeDetails 每种类型的详细用法说明，包含用途、适用场景、参数格式、返回格式标记、JSON 示例和注意事项。
var qaTypeDetails = map[string]string{
	"select": "select — 单选问题\n\n" +
		"【用途】用户从选项列表中选择一个选项。最常用的交互类型。\n" +
		"【适用场景】环境选择、方案确认、优先项选择等需要明确二选一/多选一的场景。\n" +
		"【特殊能力】支持 supplement 机制 — 设置 supplement: true 后，可使用 qa_push_supplement 为每个选项单独推送详细说明。\n\n" +
		"参数格式：\n" +
		"  - question_type: \"select\"\n" +
		"  - content: 问题内容（Markdown）\n" +
		"  - options: [{label: \"选项标签\", description: \"选项说明\"}, ...]（必填）\n" +
		"  - supplement: true/false（可选，建议选择类设为 true）\n\n" +
		"返回格式：\n" +
		"  [ANSWER] 用户选择：<选中选项的 label>\n" +
		"  [DESCRIPTION] <问题级描述（question.description，为空则不输出）>\n" +
		"  [OPTION_DESCRIPTION] <选中选项的 description（为空则不输出）>\n" +
		"  [SUPPLEMENT] <该选项的 supplement 内容（Agent 通过 qa_push_supplement 推送的 markdown，为空则不输出）>\n\n" +
		"JSON 示例：" + jsonBlockStart + `{
  "session_id": "123456",
  "question_type": "select",
  "content": "请选择部署环境",
  "supplement": true,
  "options": [
    {"label": "开发环境", "description": "用于本地开发和调试"},
    {"label": "测试环境", "description": "用于集成测试和 QA"},
    {"label": "生产环境", "description": "面向用户的正式环境"}
  ]
}` + jsonBlockEnd + "\n" +
		"注意事项：\n" +
		"  - options 的 label 应简洁（1-5 词），详细技术说明通过 qa_push_supplement 补充\n" +
		"  - 若设 supplement: true，建议在推送问题后立即为每个 option 推送 supplement\n" +
		"  - [DESCRIPTION] 是问题级描述，[OPTION_DESCRIPTION] 是选项级描述，二者来源不同\n" +
		"  - [SUPPLEMENT] 是选中选项的附属内容（Agent 为该选项推送的 markdown supplement）",

	"multi-select": "multi-select — 多选问题\n\n" +
		"【用途】用户可以从选项列表中选择多个选项。\n" +
		"【适用场景】功能选择、多维度需求确认等允许多选的场景。\n" +
		"【特殊能力】同 select，支持 supplement 机制。\n\n" +
		"参数格式：\n" +
		"  - question_type: \"multi-select\"\n" +
		"  - content: 问题内容（Markdown）\n" +
		"  - options: [{label, description}, ...]（必填）\n" +
		"  - config: {min: 1, max: 5}（可选，限制选择数量）\n" +
		"  - supplement: true/false（可选）\n\n" +
		"返回格式（多选项以 --- 分隔，避免描述杂糅）：\n" +
		"  [ANSWER] 用户选择 N 项\n" +
		"  [DESCRIPTION] <问题级描述（为空则不输出）>\n" +
		"  ---\n" +
		"  [OPTION] <选项1 label>\n" +
		"  [OPTION_DESCRIPTION] <选项1 description（为空则不输出）>\n" +
		"  [SUPPLEMENT] <选项1 的 supplement（Agent 推送的 markdown，为空则不输出）>\n" +
		"  ---\n" +
		"  [OPTION] <选项2 label>\n" +
		"  [OPTION_DESCRIPTION] <选项2 description>\n" +
		"  ---\n" +
		"  [OPTION] __other__\n" +
		"  [SUPPLEMENT] <用户自定义输入（\"其他\"选项，无则不输出此段）>\n\n" +
		"JSON 示例：" + jsonBlockStart + `{
  "session_id": "123456",
  "question_type": "multi-select",
  "content": "请选择需要启用的功能模块",
  "supplement": true,
  "options": [
    {"label": "用户认证", "description": "JWT + Session 双模式"},
    {"label": "数据导出"},
    {"label": "实时通知"},
    {"label": "审计日志"}
  ]
}` + jsonBlockEnd + "\n" +
		"注意事项：\n" +
		"  - [DESCRIPTION] 是问题级描述（全局唯一），[OPTION_DESCRIPTION] 是选项级描述（每个 [OPTION] 下一条）\n" +
		"  - [SUPPLEMENT] 是每个选项的可选附属行（Agent 为该选项推送的 markdown supplement，或\"其他\"自定义输入）\n" +
		"  - 返回的多选项用 --- 分隔，注意解析",

	"options": "options — 选项展示（差异化对比专用）\n\n" +
		"【用途】展示带优缺点（pros/cons）对比的选项卡片，供用户做出知情决策。\n" +
		"【适用场景】技术选型、架构方案对比、工具选择等需要权衡利弊的决策。\n" +
		"【重要约束】此类型专为差异化对比设计，每个选项的 pros 和 cons 为必填项。\n\n" +
		"参数格式：\n" +
		"  - question_type: \"options\"\n" +
		"  - content: 选择提示（Markdown）\n" +
		"  - options: [{label, description, pros: [...], cons: [...]}]（必填，pros/cons 必填）\n" +
		"  - supplement: true/false（可选）\n\n" +
		"返回格式：\n" +
		"  [ANSWER] <选中选项的 label>\n" +
		"  [DESCRIPTION] <选中选项的 description>\n" +
		"  [SUPPLEMENT] <用户的选择理由（feedback，如有）>\n\n" +
		"JSON 示例：" + jsonBlockStart + `{
  "session_id": "123456",
  "question_type": "options",
  "content": "请选择缓存策略",
  "supplement": true,
  "options": [
    {
      "label": "Redis",
      "description": "高性能内存缓存",
      "pros": ["速度快", "支持多种数据结构", "支持持久化"],
      "cons": ["需要额外部署", "内存成本高"]
    },
    {
      "label": "Memcached",
      "description": "轻量级缓存方案",
      "pros": ["简单易用", "多线程"],
      "cons": ["仅支持字符串", "功能单一"]
    }
  ]
}` + jsonBlockEnd + "\n" +
		"注意事项：\n" +
		"  - ⚠️ pros 和 cons 为必填项，缺失会报错\n" +
		"  - 当你需要对比多个方案的利弊时，使用此类型而非 select\n" +
		"  - select 不要求 pros/cons；options 要求且强调差异化对比",

	"code": "code — 代码输入\n\n" +
		"【用途】用户在带语法高亮的代码编辑器中输入代码。\n" +
		"【适用场景】正则表达式、配置片段、算法实现等需要代码格式的输入。\n\n" +
		"参数格式：\n" +
		"  - question_type: \"code\"\n" +
		"  - content: 问题内容（Markdown）\n" +
		"  - config: {language: \"语言标识\", placeholder: \"占位提示\"}\n\n" +
		"返回格式（含语言标记）：\n" +
		"  [ANSWER] <用户输入的代码>\n" +
		"  [LANGUAGE] <语言标识（与 config.language 一致）>\n\n" +
		"JSON 示例：" + jsonBlockStart + `{
  "session_id": "123456",
  "question_type": "code",
  "content": "请提供自定义的正则表达式",
  "config": {"language": "regex", "placeholder": "输入正则表达式..."}
}` + jsonBlockEnd + "\n" +
		"注意事项：\n" +
		"  - language 会原样在返回的 [LANGUAGE] 标记中输出\n" +
		"  - 支持的语言标识：javascript/js/typescript/ts/python/py/go/json/markdown/css/html/sql/rust/java/cpp/c/php/yaml/xml/regex/shell/bash 等",

	"image": "image — 图片上传（一次性令牌下载）\n\n" +
		"【用途】用户上传一张或多张图片。系统自动将 base64 转为文件存储，通过一次性令牌提供下载链接。\n" +
		"【适用场景】架构图、UI 设计稿、错误截图等需要图片输入的场景。\n\n" +
		"参数格式：\n" +
		"  - question_type: \"image\"\n" +
		"  - content: 问题内容（Markdown）\n" +
		"  - config: {maxImages: 5, maxSize: 10485760}（maxSize 单位为字节）\n\n" +
		"返回格式（一次性令牌下载）：\n" +
		"  [ANSWER] 用户已上传内容\n" +
		"  ---\n" +
		"  [FILE_NAME] <文件名>\n" +
		"  [DOWNLOAD_PATH] .lumina/cache/<session_id>/<文件名>\n" +
		"  [DOWNLOAD_URL]\n" +
		"      - http://<domain>/api/v1/qa/download/<一次性令牌>\n" +
		"  [IMPORTANT] 下载链接为一次性令牌，使用后即失效。需重新下载请调用 qa_reget_answer。\n" +
		"  [TIP] 使用 curl -o <path> <url> 下载后引用路径。\n" +
		"  [GIT_TIP] .lumina/cache/ 需加入 .gitignore。\n\n" +
		"JSON 示例：" + jsonBlockStart + `{
  "session_id": "123456",
  "question_type": "image",
  "content": "请上传架构设计图",
  "config": {"maxImages": 3}
}` + jsonBlockEnd + "\n" +
		"注意事项：\n" +
		"  - ⚠️ 图片数据不通过 MCP 直接返回（base64 会超 token 限制）\n" +
		"  - 必须通过 [DOWNLOAD_URL] 的一次性令牌下载文件内容\n" +
		"  - 令牌使用一次后失效，重新下载需调用 qa_reget_answer 获取新令牌\n" +
		"  - 令牌有效期 10 分钟",

	"file": "file — 文件上传（一次性令牌下载）\n\n" +
		"【用途】同 image，用户上传任意类型文件。通过一次性令牌提供下载。\n" +
		"【适用场景】配置文件、日志、导出数据等文件输入。\n\n" +
		"参数格式：\n" +
		"  - question_type: \"file\"\n" +
		"  - content: 问题内容（Markdown）\n" +
		"  - config: {accept: [\".json\", \".yaml\"], maxFiles: 5, maxSize: 5242880}\n\n" +
		"返回格式：同 image 类型（一次性令牌下载）。\n\n" +
		"JSON 示例：" + jsonBlockStart + `{
  "session_id": "123456",
  "question_type": "file",
  "content": "请上传当前项目的配置文件",
  "config": {"accept": [".json", ".yaml", ".toml"], "maxFiles": 3}
}` + jsonBlockEnd + "\n" +
		"注意事项：同 image 类型。",

	"diff": "diff — 差异对比（返回最终代码）\n\n" +
		"【用途】展示修改前后的代码差异，供用户审阅和确认。\n" +
		"【适用场景】代码重构、Bug 修复方案、配置变更等需要展示 before/after 的场景。\n\n" +
		"参数格式：\n" +
		"  - question_type: \"diff\"\n" +
		"  - content: 修改说明（Markdown）\n" +
		"  - config: {before: \"原始代码\", after: \"修改后代码\", language: \"go\"}\n\n" +
		"返回格式：\n" +
		"  approve（批准）:\n" +
		"    [ANSWER] 用户已批准该修改\n" +
		"    [FINAL]\n" +
		"    <修改后的完整代码（config.after）>\n" +
		"  reject（拒绝）:\n" +
		"    [ANSWER] 用户已拒绝该修改\n" +
		"    [FEEDBACK] <拒绝原因（如有）>\n" +
		"  edit（编辑后提交）:\n" +
		"    [ANSWER] 用户修改后提交\n" +
		"    [FINAL]\n" +
		"    <用户修改后的代码>\n\n" +
		"JSON 示例：" + jsonBlockStart + `{
  "session_id": "123456",
  "question_type": "diff",
  "content": "优化数据库查询性能，添加索引 hint",
  "config": {
    "before": "SELECT * FROM users WHERE name = ?",
    "after": "SELECT /*+ INDEX(users idx_name) */ * FROM users WHERE name = ?",
    "language": "sql"
  }
}` + jsonBlockEnd + "\n" +
		"注意事项：\n" +
		"  - approve 返回的 [FINAL] 是 config.after（Agent 可直接使用）\n" +
		"  - edit 返回的 [FINAL] 是用户在 diff 基础上修改后的代码\n" +
		"  - 只有 approve 和 edit 有 [FINAL]，reject 没有",

	"plan": "plan — 方案展示（返回审批结果+计划详情）\n\n" +
		"【用途】分段展示计划或方案，用户可以逐段审阅并做出整体决策。\n" +
		"【适用场景】项目计划、迁移方案、架构设计等需要分段展示和审批的场景。\n\n" +
		"参数格式：\n" +
		"  - question_type: \"plan\"\n" +
		"  - content: 方案总览（Markdown）\n" +
		"  - config: {sections: [{id, title, content}, ...]}\n\n" +
		"返回格式：\n" +
		"  approve（批准）:\n" +
		"    [ANSWER] 用户已批准该计划\n" +
		"    [PLAN_DETAIL]\n" +
		"    1. <section1 title>\n" +
		"       <section1 content>\n" +
		"    2. <section2 title>\n" +
		"       <section2 content>\n" +
		"  reject（拒绝）:\n" +
		"    [ANSWER] 用户已拒绝该计划\n" +
		"    [FEEDBACK] <拒绝原因>\n" +
		"  revise（需修订）:\n" +
		"    [ANSWER] 用户要求修改该计划\n" +
		"    [REVISIONS]\n" +
		"    1. [<sectionId>] <修订意见>\n" +
		"    [FEEDBACK] <整体反馈（如有）>\n\n" +
		"JSON 示例：" + jsonBlockStart + `{
  "session_id": "123456",
  "question_type": "plan",
  "content": "数据库迁移方案",
  "config": {
    "sections": [
      {"id": "backup", "title": "数据备份", "content": "全量备份当前数据库..."},
      {"id": "migrate", "title": "执行迁移", "content": "运行迁移脚本..."},
      {"id": "verify", "title": "验证结果", "content": "校验数据完整性..."}
    ]
  }
}` + jsonBlockEnd + "\n" +
		"注意事项：\n" +
		"  - approve 时会输出完整的 [PLAN_DETAIL]，Agent 可直接获取用户批准的完整计划\n" +
		"  - revise 时每个 section 的修订意见通过 [REVISIONS] 输出",

	"review": "review — 内容审阅\n\n" +
		"【用途】展示内容供用户逐段审阅和反馈。\n" +
		"【适用场景】API 文档审阅、代码规范审查、设计文档评审等。\n\n" +
		"参数格式：\n" +
		"  - question_type: \"review\"\n" +
		"  - content: 审阅内容（Markdown）\n" +
		"  - config: {sections: [{id, title, content}, ...]}\n\n" +
		"返回格式：\n" +
		"  approve（批准）:\n" +
		"    [ANSWER] 用户批准了该修改\n" +
		"  revise（需修改）:\n" +
		"    [ANSWER] 用户要求修改\n" +
		"    [FEEDBACK] <修改意见>\n\n" +
		"JSON 示例：" + jsonBlockStart + `{
  "session_id": "123456",
  "question_type": "review",
  "content": "请审阅以下 API 设计文档",
  "config": {
    "sections": [
      {"id": "auth", "title": "认证模块", "content": "## 认证方式\n使用 Bearer Token..."},
      {"id": "api", "title": "接口设计", "content": "## RESTful API\n遵循 OpenAPI 3.0 规范..."}
    ]
  }
}` + jsonBlockEnd + "\n" +
		"注意事项：\n" +
		"  - review 只有 approve/revise 两种决策（无 reject）\n" +
		"  - approve 返回简洁，不重复标记",

	"rank": "rank — 排序（需 options）\n\n" +
		"【用途】用户通过拖拽排列选项的优先级顺序。\n" +
		"【适用场景】需求优先级排序、任务排序、特性重要性排列等。\n" +
		"【重要约束】必须提供 options 参数。\n\n" +
		"参数格式：\n" +
		"  - question_type: \"rank\"\n" +
		"  - content: 排序提示（Markdown）\n" +
		"  - options: [{label: \"选项A\"}, {label: \"选项B\"}, ...]（必填）\n\n" +
		"返回格式（按优先级从高到低）：\n" +
		"  [ANSWER] 1. <选项1 label> → 2. <选项2 label> → 3. <选项3 label>\n\n" +
		"JSON 示例：" + jsonBlockStart + `{
  "session_id": "123456",
  "question_type": "rank",
  "content": "请按优先级排列以下需求",
  "options": [
    {"label": "性能优化"},
    {"label": "安全加固"},
    {"label": "UI 改进"},
    {"label": "新功能开发"}
  ]
}` + jsonBlockEnd + "\n" +
		"注意事项：\n" +
		"  - options 为必填参数\n" +
		"  - 返回的是 label（反查后的可读名称），非选项 ID",

	"rate": "rate — 星级评分（需 options，每项独立打分）\n\n" +
		"【用途】用户通过 1-N 星为每个选项独立评分。\n" +
		"【适用场景】体验评分、满意度评价、多维度质量打分等。\n" +
		"【重要约束】必须提供 options 参数，每个 option 独立打分。\n\n" +
		"参数格式：\n" +
		"  - question_type: \"rate\"\n" +
		"  - content: 评价提示（Markdown）\n" +
		"  - options: [{label, description}, ...]（必填）\n" +
		"  - config: {max: 5, step: 1}\n\n" +
		"返回格式（每个选项独立评分）：\n" +
		"  [ANSWER] <选项1 label>: <分数>, <选项2 label>: <分数>\n\n" +
		"JSON 示例：" + jsonBlockStart + `{
  "session_id": "123456",
  "question_type": "rate",
  "content": "请对以下维度进行评分",
  "options": [
    {"label": "易用性", "description": "操作是否简单直观"},
    {"label": "性能", "description": "响应速度和资源占用"},
    {"label": "稳定性", "description": "是否经常出错"}
  ],
  "config": {"max": 5}
}` + jsonBlockEnd + "\n" +
		"注意事项：\n" +
		"  - ⚠️ options 为必填参数，每个 option 独立打分\n" +
		"  - 不提供 options 将导致空评分错误",

	"text": "text — 文本输入\n\n" +
		"【用途】用户自由输入文本内容，支持单行和多行。最灵活的交互类型。\n" +
		"【适用场景】补充需求、问题描述、自由回答等。\n\n" +
		"参数格式：\n" +
		"  - question_type: \"text\"\n" +
		"  - content: 问题内容（Markdown）\n" +
		"  - config: {multiline: true/false, placeholder: \"提示文本\", maxLength: 500}\n\n" +
		"返回格式：\n" +
		"  [ANSWER] <用户输入的文本>\n\n" +
		"JSON 示例：" + jsonBlockStart + `{
  "session_id": "123456",
  "question_type": "text",
  "content": "请描述你遇到的问题",
  "config": {"multiline": true, "placeholder": "详细描述问题现象..."}
}` + jsonBlockEnd,

	"boolean": "boolean — 布尔确认\n\n" +
		"【用途】用户选择是或否。适合需要明确确认的场景。\n" +
		"【适用场景】删除确认、启用/停用、执行/取消等二选一场景。\n\n" +
		"参数格式：\n" +
		"  - question_type: \"boolean\"\n" +
		"  - content: 确认内容（Markdown）\n\n" +
		"返回格式：\n" +
		"  [ANSWER] 是  或  [ANSWER] 否\n\n" +
		"JSON 示例：" + jsonBlockStart + `{
  "session_id": "123456",
  "question_type": "boolean",
  "content": "确认删除所有缓存数据？此操作不可恢复。"
}` + jsonBlockEnd,

	"slider": "slider — 滑块评分\n\n" +
		"【用途】用户通过拖动滑块在一个数值范围内选择值。\n" +
		"【适用场景】满意度评分、百分比配置、优先级数值等。\n\n" +
		"参数格式：\n" +
		"  - question_type: \"slider\"\n" +
		"  - content: 问题描述（Markdown）\n" +
		"  - config: {min: 0, max: 100, step: 1, defaultValue: 50}\n\n" +
		"返回格式：\n" +
		"  [ANSWER] <用户选择的数值>\n\n" +
		"JSON 示例：" + jsonBlockStart + `{
  "session_id": "123456",
  "question_type": "slider",
  "content": "你对当前系统性能的满意度是多少？",
  "config": {"min": 0, "max": 10, "step": 1, "defaultValue": 5}
}` + jsonBlockEnd,
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
			// 计数器失败时降级为 PENDING，不阻塞正常流程
			return textResult("[PENDING] 用户暂未回答，请重新调用 qa_get_answer 继续等待用户回答。"), nil
		}

		if retryCount >= maxRetries {
			// 达到最大重试次数，返回 STOPPED 提示用户主动触发
			return textResult("[STOPPED] 已等待较长时间仍未收到用户回答。请暂停轮询，并告知用户：「如果您已回答，请告诉我\"继续\"以便我获取您的回答。」当用户说\"继续\"后，请再次调用 qa_get_answer 获取回答。"), nil
		}

		// 未达上限，返回 PENDING + RETRY 进度
		return textResult(fmt.Sprintf(
			"[PENDING] 用户暂未回答，请重新调用 qa_get_answer 继续等待用户回答。\n[RETRY] %d/%d（当前未达到最大重试次数，按照系统设计规则请重新调用）",
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
