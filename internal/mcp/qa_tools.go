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

触发场景：选项 label 过于简洁需要展开技术细节、展示架构图/流程图等可视化内容、或 qa_get_answer 返回 [NEED_SUPPLEMENT] 时响应用户的补充请求。建议在推送选择类问题时主动为每个选项补充详细说明，帮助用户做出知情决策。

Markdown 适合文本说明、代码块、表格、Mermaid 流程图；HTML 适合交互式预览、自定义布局。每个目标是 1:1 映射，重复推送会覆盖之前内容。`,
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

行为：队列有待消费回答时立即返回全部；队列为空时阻塞等待。消费后队列清空。

此工具是阻塞调用。如需非阻塞获取已有答案，请用 qa_reget_answer。`,
		inputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"session_id": map[string]any{
					"type":        "string",
					"description": "目标会话 ID",
				},
				"timeout": map[string]any{
					"type":        "integer",
					"description": "最大等待时间（秒），不传则使用默认超时",
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
}

// qaTypeListText 返回全部 14 种类型的概览文本。
const qaTypeListText = `Q&A 支持以下 14 种问题类型：

选择类
  - select         单选 — 从选项列表中选择一个（最常用）
  - multi-select   多选 — 从选项列表中选择多个

输入类
  - text           文本输入 — 单行或多行自由文本（最灵活）
  - boolean        布尔确认 — 是/否二选一
  - code           代码输入 — 带语法高亮的代码编辑器
  - image          图片上传 — 支持拖拽上传图片
  - file           文件上传 — 支持上传任意文件

展示类
  - diff           差异对比 — 展示修改前后的代码对比
  - plan           方案展示 — 分段展示计划供审阅
  - options        选项展示 — 展示带 pros/cons 的选项卡
  - review         内容审阅 — 展示内容供逐段批注

评分类
  - slider         滑块评分 — 在数值范围内滑动选择
  - rank           排序 — 拖拽排列选项优先级
  - rate           星级评分 — 1-5 星评价`

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

// qaTypeDetails 每种类型的详细用法说明，包含参数格式和 JSON 示例。
var qaTypeDetails = map[string]string{
	"select": "select — 单选问题\n\n用户从选项列表中选择一个选项。最常用的交互类型。\n\n参数格式：\n  - question_type: \"select\"\n  - content: 问题内容（Markdown）\n  - options: [{label: \"选项1\", description: \"说明\"}, ...]\n\nJSON 示例：" + jsonBlockStart + `{
  "session_id": "123456",
  "question_type": "select",
  "content": "请选择部署环境",
  "supplement": true,
  "options": [
    {"label": "开发环境", "description": "用于本地开发和调试"},
    {"label": "测试环境", "description": "用于集成测试和 QA"},
    {"label": "生产环境", "description": "面向用户的正式环境"}
  ]
}` + jsonBlockEnd + "\n返回：用户选中的选项 label。",

	"multi-select": "multi-select — 多选问题\n\n用户可以从选项列表中选择多个选项。\n\n参数格式：\n  - question_type: \"multi-select\"\n  - content: 问题内容（Markdown）\n  - options: [{label: \"选项1\", description: \"说明\"}, ...]\n\nJSON 示例：" + jsonBlockStart + `{
  "session_id": "123456",
  "question_type": "multi-select",
  "content": "请选择需要启用的功能模块",
  "supplement": true,
  "options": [
    {"label": "用户认证"},
    {"label": "数据导出"},
    {"label": "实时通知"},
    {"label": "审计日志"}
  ]
}` + jsonBlockEnd + "\n返回：用户选中的选项 label 数组。",

	"text": "text — 文本输入\n\n用户自由输入文本内容，支持单行和多行。最灵活的交互类型。\n\n参数格式：\n  - question_type: \"text\"\n  - content: 问题内容（Markdown）\n  - config: {multiline: true/false, placeholder: \"提示文本\", maxLength: 500}\n\nJSON 示例：" + jsonBlockStart + `{
  "session_id": "123456",
  "question_type": "text",
  "content": "请描述你遇到的问题",
  "config": {"multiline": true, "placeholder": "详细描述问题现象..."}
}` + jsonBlockEnd + "\n返回：用户输入的文本字符串。",

	"boolean": "boolean — 布尔确认\n\n用户选择是或否。适合需要明确确认的场景。\n\n参数格式：\n  - question_type: \"boolean\"\n  - content: 确认内容（Markdown）\n\nJSON 示例：" + jsonBlockStart + `{
  "session_id": "123456",
  "question_type": "boolean",
  "content": "确认删除所有缓存数据？此操作不可恢复。"
}` + jsonBlockEnd + "\n返回：true 或 false。",

	"code": "code — 代码输入\n\n用户在带语法高亮的编辑器中输入代码。\n\n参数格式：\n  - question_type: \"code\"\n  - content: 问题内容（Markdown）\n  - config: {language: \"python\", placeholder: \"// 在此输入代码\"}\n\nJSON 示例：" + jsonBlockStart + `{
  "session_id": "123456",
  "question_type": "code",
  "content": "请提供自定义的正则表达式",
  "config": {"language": "regex", "placeholder": "输入正则表达式..."}
}` + jsonBlockEnd + "\n返回：用户输入的代码字符串。",

	"image": "image — 图片上传\n\n用户上传一张或多张图片，支持拖拽和粘贴。\n\n参数格式：\n  - question_type: \"image\"\n  - content: 问题内容（Markdown）\n  - config: {maxImages: 5, maxSize: 10485760}\n\nJSON 示例：" + jsonBlockStart + `{
  "session_id": "123456",
  "question_type": "image",
  "content": "请上传问题截图",
  "config": {"maxImages": 3}
}` + jsonBlockEnd + "\n返回：图片 URL 或 base64 数据。",

	"file": "file — 文件上传\n\n用户上传一个或多个文件。\n\n参数格式：\n  - question_type: \"file\"\n  - content: 问题内容（Markdown）\n  - config: {accept: [\".json\", \".yaml\"], maxFiles: 5, maxSize: 10485760}\n\nJSON 示例：" + jsonBlockStart + `{
  "session_id": "123456",
  "question_type": "file",
  "content": "请上传配置文件",
  "config": {"accept": [".json", ".yaml", ".toml"], "maxFiles": 3}
}` + jsonBlockEnd + "\n返回：文件 URL 或 base64 数据。",

	"diff": "diff — 差异对比\n\n展示修改前后的代码差异，供用户审阅和确认。\n\n参数格式：\n  - question_type: \"diff\"\n  - content: 修改说明（Markdown）\n  - config: {before: \"原始代码\", after: \"修改后代码\", language: \"go\"}\n\nJSON 示例：" + jsonBlockStart + `{
  "session_id": "123456",
  "question_type": "diff",
  "content": "优化数据库查询性能，添加索引 hint",
  "config": {
    "before": "SELECT * FROM users WHERE name = ?",
    "after": "SELECT /*+ INDEX(users idx_name) */ * FROM users WHERE name = ?",
    "language": "sql"
  }
}` + jsonBlockEnd + "\n返回：用户确认接受或拒绝。",

	"plan": "plan — 方案展示\n\n分段展示计划或方案，用户可以逐段审阅并批注。\n\n参数格式：\n  - question_type: \"plan\"\n  - content: 方案总览（Markdown）\n  - config: {sections: [{id: \"s1\", title: \"阶段一\", content: \"...\"}]}\n\nJSON 示例：" + jsonBlockStart + `{
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
}` + jsonBlockEnd + "\n返回：用户对每个 section 的批注。",

	"options": "options — 选项展示\n\n展示带优缺点对比的选项卡片，供用户做出知情选择。\n\n参数格式：\n  - question_type: \"options\"\n  - content: 选择提示（Markdown）\n  - options: [{label: \"方案A\", description: \"...\", pros: [\"...\"], cons: [\"...\"]}]\n\nJSON 示例：" + jsonBlockStart + `{
  "session_id": "123456",
  "question_type": "options",
  "content": "请选择缓存策略",
  "options": [
    {
      "label": "Redis",
      "description": "高性能内存缓存",
      "pros": ["速度快", "支持多种数据结构"],
      "cons": ["需要额外部署", "内存成本高"]
    },
    {
      "label": "Memcached",
      "description": "轻量级缓存方案",
      "pros": ["简单易用", "多线程"],
      "cons": ["仅支持字符串", "功能单一"]
    }
  ]
}` + jsonBlockEnd + "\n返回：用户选择的选项 label。",

	"review": "review — 内容审阅\n\n展示内容供用户逐段审阅和反馈。\n\n参数格式：\n  - question_type: \"review\"\n  - content: 审阅内容（Markdown）\n  - config: {sections: [{id: \"r1\", title: \"章节一\", content: \"...\"}]}\n\nJSON 示例：" + jsonBlockStart + `{
  "session_id": "123456",
  "question_type": "review",
  "content": "请审阅以下 API 设计文档",
  "config": {
    "sections": [
      {"id": "auth", "title": "认证模块", "content": "## 认证方式\n使用 Bearer Token..."},
      {"id": "api", "title": "接口设计", "content": "## RESTful API\n遵循 OpenAPI 3.0 规范..."}
    ]
  }
}` + jsonBlockEnd + "\n返回：用户对每个 section 的审阅反馈。",

	"slider": "slider — 滑块评分\n\n用户通过拖动滑块在一个数值范围内选择值。\n\n参数格式：\n  - question_type: \"slider\"\n  - content: 问题描述（Markdown）\n  - config: {min: 0, max: 100, step: 1, defaultValue: 50}\n\nJSON 示例：" + jsonBlockStart + `{
  "session_id": "123456",
  "question_type": "slider",
  "content": "你对当前系统性能的满意度是多少？",
  "config": {"min": 0, "max": 10, "step": 1, "defaultValue": 5}
}` + jsonBlockEnd + "\n返回：用户选择的数值。",

	"rank": "rank — 排序\n\n用户通过拖拽排列选项的优先级顺序。\n\n参数格式：\n  - question_type: \"rank\"\n  - content: 排序提示（Markdown）\n  - options: [{label: \"选项A\"}, {label: \"选项B\"}, ...]\n\nJSON 示例：" + jsonBlockStart + `{
  "session_id": "123456",
  "question_type": "rank",
  "content": "请按优先级排列以下需求",
  "options": [
    {"label": "性能优化"},
    {"label": "安全加固"},
    {"label": "UI 改进"},
    {"label": "新功能开发"}
  ]
}` + jsonBlockEnd + "\n返回：用户排列后的选项 label 数组（从高到低）。",

	"rate": "rate — 星级评分\n\n用户通过 1-5 星进行评价。\n\n参数格式：\n  - question_type: \"rate\"\n  - content: 评价提示（Markdown）\n  - config: {max: 5, step: 1}\n\nJSON 示例：" + jsonBlockStart + `{
  "session_id": "123456",
  "question_type": "rate",
  "content": "请对本次 AI 辅助体验进行评分",
  "config": {"max": 5}
}` + jsonBlockEnd + "\n返回：用户的评分数值。",
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
func handleQaSessionCreate(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	id, link, xErr := qaLogic.CreateSession(context.Background(), title, agentName, sessionType, projectID)
	if xErr != nil {
		return textResult(fmt.Sprintf("创建会话失败: %s", xErr.Error())), nil
	}

	return textResult(fmt.Sprintf(
		"会话创建成功！\n\n会话 ID: %s\n浏览器链接: %s\n\n请在浏览器中打开链接开始问答。",
		id, link,
	)), nil
}

// handleQaSessionList 获取会话列表
func handleQaSessionList(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	resp, xErr := qaLogic.ListSessions(context.Background(), listReq)
	if xErr != nil {
		return textResult(fmt.Sprintf("获取会话列表失败: %s", xErr.Error())), nil
	}

	totalPages := (resp.Total + int64(size) - 1) / int64(size)
	result := fmt.Sprintf(`Q&A 会话列表（共 %d 个，第 %d/%d 页）：

`, resp.Total, page, totalPages)
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
func handleQaSessionGet(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	result, xErr := qaLogic.GetSessionMCP(context.Background(), sessionID)
	if xErr != nil {
		return textResult(fmt.Sprintf("获取会话失败: %s", xErr.Error())), nil
	}

	return textResult(result), nil
}

// handleQaSessionArchive 归档会话
func handleQaSessionArchive(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	xErr := qaLogic.ArchiveSession(context.Background(), sessionID)
	if xErr != nil {
		return textResult(fmt.Sprintf("归档会话失败: %s", xErr.Error())), nil
	}

	return textResult(fmt.Sprintf("会话 %s 已归档。数据保留为只读状态，仍可通过 qa_session_get 查看历史。", sessionID)), nil
}

// handleQaPushQuestion 推送问题到会话
func handleQaPushQuestion(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	qID, optionIDMap, xErr := qaLogic.PushQuestion(context.Background(), sessionID, qType, content, description, options, config, batch, groupLabel, supplement)
	if xErr != nil {
		return textResult(fmt.Sprintf("推送问题失败: %s", xErr.Error())), nil
	}

	result := fmt.Sprintf("问题已推送！\n问题 ID: %s\n", qID)
	if len(optionIDMap) > 0 {
		result += "\n选项ID映射:\n"
		for label, id := range optionIDMap {
			result += fmt.Sprintf("  %s → %s\n", label, id)
		}
	}
	result += "\n使用 qa_get_answer 等待用户回答。"

	return textResult(result), nil
}

// handleQaPushSupplement 推送补充内容
func handleQaPushSupplement(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	xErr := qaLogic.PushSupplement(context.Background(), sessionID, targetType, targetID, contentType, content)
	if xErr != nil {
		return textResult(fmt.Sprintf("推送补充内容失败: %s", xErr.Error())), nil
	}

	// 查询关联问题的选项列表，静默降级
	var optionLines []string
	if parsedQID, err := xSnowflake.ParseSnowflakeID(questionID); err == nil {
		if q, qErr := qaLogic.GetQuestionByID(context.Background(), parsedQID); qErr == nil && len(q.Options) > 0 {
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

	result := "补充内容已推送成功。"
	if len(optionLines) > 0 {
		result += "\n\n[TIP] 具体选项可以单独提交详细信息，使用 qa_push_supplement 并传入 option_id 参数\n\n关联选项:\n"
		for _, line := range optionLines {
			result += line + "\n"
		}
	}

	return textResult(result), nil
}

// handleQaWhatQuestion 返回问题类型帮助信息
func handleQaWhatQuestion(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
func handleGetAnswer(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

// handleRegetAnswer 批量非阻塞获取已回答问题的内容
func handleRegetAnswer(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	result, xErr := qaLogic.RegetAnswers(context.Background(), sessionID, questionIDs)
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
		default:
			handler = stubToolHandler(def.name)
		}

		server.AddTool(tool, handler)
	}
}
