package mcp

import (
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// previewToolDef 定义 Preview MCP 工具的模型提示、Schema 与行为元数据。
type previewToolDef struct {
	name         string
	title        string
	description  string
	inputSchema  map[string]any
	outputSchema map[string]any
	annotations  *mcp.ToolAnnotations
}

// previewToolDefs 定义 Preview 模块的全部 MCP 工具。
var previewToolDefs = []previewToolDef{
	{
		name:  "preview_session_create",
		title: "创建前端预览会话",
		description: `用途：为已注册项目创建一个独立的前端预览会话；一个项目可以有多个会话。它只创建空会话，不会生成代码、上传文件、修改本地仓库，也不代表用户已确认设计。

何时调用：用户要求可视化查看 HTML/CSS/JavaScript 原型，且没有合适的现有 Preview 会话时调用。调用前先用 project_get/project_list 确定 project_id，并优先用 preview_session_list 检查能否复用当前任务的会话。

不要调用：仅需展示 Markdown、代码片段或单个静态说明时无需创建 Preview；用户要求实现现有产品功能时，Preview 只能用于评审，不能替代对真实项目文件的修改和验证。

副作用：每次调用都会新建会话，重复调用不幂等。返回 session_id、访问 hash 和绝对 preview_url；hash 只用于网页访问，不是 Q&A preview supplement 的引用字段。

下一步：不要立即打开空页面。先逐个调用 preview_file_upload 上传文件；文件齐全后调用 preview_file_list 做最终核对，再按返回指引打开网页或挂载到 Q&A。`,
		inputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"project_id": map[string]any{
					"type":        "string",
					"pattern":     "^[0-9]+$",
					"description": "关联的 Lumina 项目雪花 ID。不得传项目名；未知时先调用 project_get 或 project_list。",
				},
				"title": map[string]any{
					"type":        "string",
					"maxLength":   255,
					"description": "便于用户识别任务的会话标题；省略或留空时为「未命名预览」。建议包含功能或方案名称。",
				},
			},
			"required": []string{"project_id"},
		},
		outputSchema: previewSessionCreateOutputSchema(),
		annotations:  previewToolAnnotations(false, false, false),
	},
	{
		name:  "preview_session_list",
		title: "列出项目预览会话",
		description: `用途：分页列出指定项目的 Preview 会话，返回会话 ID、标题、状态、访问 hash 和绝对预览 URL；不返回文件内容。

何时调用：创建新会话前检查是否存在与当前任务匹配的 active 会话，或需要恢复既有预览会话时调用。不要仅因同一项目存在会话就盲目复用；标题和任务不匹配时应创建新会话，避免覆盖无关方案。

该工具只读且可安全重试。选定会话后调用 preview_file_list 检查文件清单；只有确认是当前任务的会话后，才继续上传、编辑或删除文件。`,
		inputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"project_id": map[string]any{
					"type":        "string",
					"pattern":     "^[0-9]+$",
					"description": "要查询的 Lumina 项目雪花 ID。",
				},
				"page": map[string]any{
					"type":        "integer",
					"minimum":     1,
					"default":     1,
					"description": "页码，从 1 开始；默认 1。",
				},
				"size": map[string]any{
					"type":        "integer",
					"minimum":     1,
					"maximum":     100,
					"default":     10,
					"description": "每页数量，默认 10，最大 100。",
				},
			},
			"required": []string{"project_id"},
		},
		outputSchema: previewSessionListOutputSchema(),
		annotations:  previewToolAnnotations(true, false, true),
	},
	{
		name:  "preview_file_upload",
		title: "上传或覆写预览文件",
		description: `用途：向指定 Preview 会话整体写入一个文本型前端文件；同一 session_id 与 filename 已存在时会原位覆写。支持 HTML、CSS、JavaScript/MJS、JSON、SVG 和纯文本，多文件必须逐个调用。

何时调用：已经掌握完整文件内容，需要创建预览、补齐 HTML 的相对依赖，或将整文件重写时调用。只修改局部内容时不要全量重传，改用 preview_file_get 读取目标行区间后调用 preview_file_edit 增量编辑。HTML 内应使用同层相对路径引用 CSS/JS，例如 style.css 和 app.js。

限制：文件名只能是扁平单层名称，禁止 /、\\ 和 ..；最长 255 字符；content 按 UTF-8 字节计最大 256 KiB。它不支持目录、二进制附件、构建命令或 npm 依赖安装，也不会修改 Agent 当前项目路径下的真实源文件。

副作用：可能创建新文件，也可能覆盖既有内容；同参数重试后的最终文件内容相同，但覆盖前应确认会话属于当前任务。返回 file_id；preview_url 指向当前可预览文件——会话已有 HTML 入口时指向入口文件，否则指向本次上传的文件；仅当存在 HTML 入口时，qa_supplement.content 才会返回可直接传给 qa_push_supplement 的非空引用 JSON。

下一步：仍有依赖文件时继续上传；全部上传后必须调用 preview_file_list 核对。只有最终清单包含可渲染 HTML 入口且依赖齐全时，才打开/分享 preview_url 或挂载到 Q&A。`,
		inputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"session_id": map[string]any{
					"type":        "string",
					"pattern":     "^[0-9]+$",
					"description": "目标 Preview 会话雪花 ID；应来自 preview_session_create/list。",
				},
				"filename": map[string]any{
					"type":        "string",
					"minLength":   1,
					"maxLength":   255,
					"description": "扁平单层文件名，例如 index.html、style.css、app.js；不得包含路径分隔符或 ..。",
				},
				"content": map[string]any{
					"type":        "string",
					"description": "完整文件文本，允许空文件；UTF-8 编码后的大小不得超过 256 KiB。不要包裹 Markdown 代码围栏。",
				},
			},
			"required": []string{"session_id", "filename", "content"},
		},
		outputSchema: previewFileUploadOutputSchema(),
		annotations:  previewToolAnnotations(false, true, true),
	},
	{
		name:  "preview_file_edit",
		title: "行级编辑预览文件",
		description: `用途：对既有预览文件做行级增量编辑（insert/replace/delete），只传输变更片段，无需整文件重传。行号 1 起始、闭区间，与 preview_file_get 传 start_line/end_line 后返回的行号一一对应。

何时调用：需要修改已上传预览文件的一部分（改几行、插入一段、删除几行）时调用。推荐流程：先 preview_file_get 传行区间拿到带行号源码并确定目标区间，再调用本工具，最后核对返回的 edited_region；多处修改可重复 get→edit 循环。

不要调用：目标文件尚不存在时不要用本工具创建（改用 preview_file_upload 整体写入）；不确定目标行号时必须先读取再编辑，禁止盲猜行号；要删除整个文件时改用 preview_file_delete，不要用 delete 全区间代删。

限制：仅支持文本行编辑；换行统一按 LF 处理（原文件的 CRLF 会被归一化），原文件结尾换行符原样保留；编辑后文件大小仍受上传上限约束；insert 的 content 不能为空，replace 允许空 content（等价删除该区间），delete 不接受 content；start_line/end_line 对 replace/delete 必填，insert 可省略 start_line（追加到末尾）。

副作用：原位覆写目标文件内容（保留文件 ID 与创建时间），并通过 WebSocket 同步预览页；同参数重放结果一致。返回编辑后总行数与编辑落点区域（变更主体 ±3 行上下文、上限 40 行，带行号）。

下一步：核对 edited_region 确认编辑落点正确；需要继续修改则重复 get→edit 循环；全部改完后调用 preview_file_list 做最终核对，再重新打开预览或挂载到 Q&A。`,
		inputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"session_id": map[string]any{
					"type":        "string",
					"pattern":     "^[0-9]+$",
					"description": "目标 Preview 会话雪花 ID；应来自 preview_session_create/list。",
				},
				"filename": map[string]any{
					"type":        "string",
					"minLength":   1,
					"maxLength":   255,
					"description": "preview_file_list 返回的精确文件名；区分大小写。",
				},
				"operation": map[string]any{
					"type":        "string",
					"enum":        []string{"insert", "replace", "delete"},
					"description": "行级编辑操作：insert 把 content 各行插入到 start_line 之前；replace 把 [start_line, end_line] 闭区间替换为 content 各行；delete 删除该闭区间。",
				},
				"start_line": map[string]any{
					"type":        "integer",
					"minimum":     1,
					"description": "起始行号（1 起始，闭区间）。insert 省略时追加到文件末尾（等价 total_lines+1）；replace/delete 必填。",
				},
				"end_line": map[string]any{
					"type":        "integer",
					"minimum":     1,
					"description": "结束行号（闭区间，须 ≥ start_line，超出总行数会被拒绝）。replace/delete 必填；insert 忽略。",
				},
				"content": map[string]any{
					"type":        "string",
					"description": "insert/replace 的新内容片段（可多行，CRLF 自动归一为 LF，不要包裹 Markdown 代码围栏）。insert 不能为空；replace 允许空串（等价删除该区间）；delete 必须省略。",
				},
			},
			"required": []string{"session_id", "filename", "operation"},
		},
		outputSchema: previewFileEditOutputSchema(),
		annotations:  previewToolAnnotations(false, true, true),
	},
	{
		name:  "preview_file_delete",
		title: "删除预览文件",
		description: `用途：从指定 Preview 会话中按文件名物理删除单个文件，并返回剩余文件清单与 HTML 入口状态，供继续编排预览内容。

何时调用：会话中存在不再需要的文件（废弃的样式、脚本或旧入口）时调用。删除前应先 preview_file_list 核对文件名，并与用户确认删除意图。

不要调用：文件名不确定时不要盲删；本工具只删单个文件且不可恢复，会话内文件较多时逐个确认后删除；用户要求清理整个会话时告知其到控制台删除会话，而不是逐个代删全部文件。

副作用：物理删除该文件且不可恢复；若删除的是 HTML 入口，会话立即失去可评审入口（返回的 entry_file 变为空并提示补齐）；通过 WebSocket 同步预览页。删除不存在的文件会返回错误。

下一步：根据返回的 entry_file 与 files 判断是否需要重新上传 HTML 入口；删除入口后必须先补入口，再调用 preview_file_list 核对后才可交付评审。`,
		inputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"session_id": map[string]any{
					"type":        "string",
					"pattern":     "^[0-9]+$",
					"description": "目标 Preview 会话雪花 ID；应来自 preview_session_create/list。",
				},
				"filename": map[string]any{
					"type":        "string",
					"minLength":   1,
					"maxLength":   255,
					"description": "preview_file_list 返回的精确文件名；区分大小写。",
				},
			},
			"required": []string{"session_id", "filename"},
		},
		outputSchema: previewFileDeleteOutputSchema(),
		annotations:  previewToolAnnotations(false, true, true),
	},
	{
		name:  "preview_file_list",
		title: "核对预览文件清单",
		description: `用途：按文件名升序返回指定 Preview 会话的完整文件清单、文件 ID、自动选择的 HTML 入口、绝对预览深链和 Q&A preview 引用；不返回源码正文。

何时调用：复用会话前检查内容、上传/编辑/删除过程中确认状态，以及全部变更完成后的最终核对。该工具只读且可安全重试。

返回后的分支必须遵守：没有文件时继续 preview_file_upload；没有 HTML 入口时先上传 HTML；存在 HTML 入口且依赖齐全时，若用户请求视觉评审，优先用客户端原生浏览器/打开链接能力访问 preview_url，能力不可用则把可点击 URL 交给用户。若预览属于 Q&A 问题或选项，则调用 qa_push_supplement，content_type=preview，content 严格使用返回的 qa_supplement.content，然后再 qa_get_answer。`,
		inputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"session_id": map[string]any{
					"type":        "string",
					"pattern":     "^[0-9]+$",
					"description": "目标 Preview 会话雪花 ID。",
				},
			},
			"required": []string{"session_id"},
		},
		outputSchema: previewFileListOutputSchema(),
		annotations:  previewToolAnnotations(true, false, true),
	},
	{
		name:  "preview_file_get",
		title: "读取预览文件源码",
		description: `用途：读取指定 Preview 会话中单个文件的源码；可选 start_line/end_line 行区间（1 起始闭区间），带区间时内容按「行号| 文本」格式返回，行号可直接用于 preview_file_edit 定位。不传区间时返回原始全量源码。它只返回目标文件，不解析依赖、不批量读取，也不会把 Preview 内容自动写入本地项目。

何时调用：准备行级编辑前先定位目标行、审查既有预览、继续迭代或提取已确认视觉规范时调用。大文件建议传行区间（如末尾 50 行）控制上下文体积；该工具只读且可安全重试。

限制：空文件传行区间会报错；start_line 必须落在 1..total_lines 内；end_line 超出总行数时自动钳制到末行（便于读到文件尾）；指定 end_line 时必须同时指定 start_line。不要把 Preview 中的代码默认视为已批准实现；只有用户明确确认后，才能将其作为真实项目修改的参考。

修改路径：局部修改优先调用 preview_file_edit 行级编辑；仅整文件重写才使用 preview_file_upload 覆写；改动后调用 preview_file_list 核对并重新打开预览。`,
		inputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"session_id": map[string]any{
					"type":        "string",
					"pattern":     "^[0-9]+$",
					"description": "目标 Preview 会话雪花 ID。",
				},
				"filename": map[string]any{
					"type":        "string",
					"minLength":   1,
					"maxLength":   255,
					"description": "preview_file_list 返回的精确文件名；区分大小写。",
				},
				"start_line": map[string]any{
					"type":        "integer",
					"minimum":     1,
					"description": "可选起始行号（1 起始，闭区间）；传入后内容按「行号| 文本」格式返回。",
				},
				"end_line": map[string]any{
					"type":        "integer",
					"minimum":     1,
					"description": "可选结束行号（闭区间，须 ≥ start_line）；省略或超出总行数时读到最后一行。",
				},
			},
			"required": []string{"session_id", "filename"},
		},
		outputSchema: previewFileGetOutputSchema(),
		annotations:  previewToolAnnotations(true, false, true),
	},
}

// previewToolAnnotations 描述工具的只读、覆盖与幂等语义。Preview 仅操作 Lumina
// 内部预览会话，因此 openWorldHint 固定为 false。
func previewToolAnnotations(readOnly, destructive, idempotent bool) *mcp.ToolAnnotations {
	return &mcp.ToolAnnotations{
		ReadOnlyHint:    readOnly,
		DestructiveHint: boolPtr(destructive),
		IdempotentHint:  idempotent,
		OpenWorldHint:   boolPtr(false),
	}
}

func boolPtr(value bool) *bool {
	return &value
}

// RegisterPreviewTools 将 Preview 模块的 7 个 MCP 工具注册到 Server。
func RegisterPreviewTools(server *mcp.Server) {
	for _, def := range previewToolDefs {
		inputSchemaBytes, _ := json.Marshal(def.inputSchema)
		outputSchemaBytes, _ := json.Marshal(def.outputSchema)
		tool := &mcp.Tool{
			Name:         def.name,
			Title:        def.title,
			Description:  def.description,
			InputSchema:  json.RawMessage(inputSchemaBytes),
			OutputSchema: json.RawMessage(outputSchemaBytes),
			Annotations:  def.annotations,
		}

		var handler mcp.ToolHandler
		switch def.name {
		case "preview_session_create":
			handler = handlePreviewSessionCreate
		case "preview_session_list":
			handler = handlePreviewSessionList
		case "preview_file_upload":
			handler = handlePreviewFileUpload
		case "preview_file_edit":
			handler = handlePreviewFileEdit
		case "preview_file_delete":
			handler = handlePreviewFileDelete
		case "preview_file_list":
			handler = handlePreviewFileList
		case "preview_file_get":
			handler = handlePreviewFileGet
		default:
			handler = stubToolHandler(def.name)
		}

		server.AddTool(tool, handler)
	}
}
