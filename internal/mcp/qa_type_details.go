package mcp

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
		"  [NOTE] 此处的 qa_reget_answer 仅用于多媒体重取，不可用于等待用户回答（等待请用 qa_get_answer）。\n" +
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
