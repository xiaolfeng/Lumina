package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	apiPreview "github.com/xiaolfeng/Lumina/api/preview"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/logic"
)

// previewLogic 保存 PreviewLogic 实例，供 MCP 工具处理器使用。
var previewLogic *logic.PreviewLogic

// SetPreviewLogic 设置 PreviewLogic 实例，供 MCP 工具处理器使用。
func SetPreviewLogic(l *logic.PreviewLogic) {
	previewLogic = l
}

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
		description: `用途：为已注册项目创建一个独立的前端预览工作区；一个项目可以有多个会话。它只创建空会话，不会生成代码、上传文件、修改本地仓库，也不代表用户已确认设计。

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

何时调用：创建新会话前检查是否存在与当前任务匹配的 active 会话，或需要恢复既有预览工作区时调用。不要仅因同一项目存在会话就盲目复用；标题和任务不匹配时应创建新会话，避免覆盖无关方案。

该工具只读且可安全重试。选定会话后调用 preview_file_list 检查文件清单；只有确认是当前任务的会话后，才继续上传或覆写文件。`,
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
		description: `用途：向指定 Preview 会话写入一个文本型前端文件；同一 session_id 与 filename 已存在时会原位覆写。支持 HTML、CSS、JavaScript/MJS、JSON、SVG 和纯文本，多文件必须逐个调用。

何时调用：已经掌握完整文件内容，需要创建预览、补齐 HTML 的相对依赖，或将 preview_file_get 读取的文件迭代后写回时调用。HTML 内应使用同层相对路径引用 CSS/JS，例如 style.css 和 app.js。

限制：文件名只能是扁平单层名称，禁止 /、\\ 和 ..；最长 255 字符；content 按 UTF-8 字节计最大 256 KiB。它不支持目录、二进制附件、构建命令或 npm 依赖安装，也不会修改 Agent 当前工作区的真实源文件。

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
		name:  "preview_file_list",
		title: "核对预览文件清单",
		description: `用途：按文件名升序返回指定 Preview 会话的完整文件清单、文件 ID、自动选择的 HTML 入口、绝对预览深链和 Q&A preview 引用；不返回源码正文。

何时调用：复用会话前检查内容、上传过程中确认状态，以及全部文件上传后的最终核对。该工具只读且可安全重试。

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
		description: `用途：读取指定 Preview 会话中单个文件的完整源码。它只返回目标文件，不解析依赖、不批量读取，也不会把 Preview 内容自动写入本地项目。

何时调用：先通过 preview_file_list 确认准确 filename，然后在审查既有预览、继续迭代或提取已确认视觉规范时调用。不要把 Preview 中的代码默认视为已批准实现；只有用户明确确认后，才能将其作为真实项目修改的参考。

该工具只读且可安全重试。若要修改 Preview，编辑返回内容后使用相同 session_id 与 filename 调用 preview_file_upload 覆写，再用 preview_file_list 核对并重新打开预览。`,
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
			},
			"required": []string{"session_id", "filename"},
		},
		outputSchema: previewFileGetOutputSchema(),
		annotations:  previewToolAnnotations(true, false, true),
	},
}

// previewToolAnnotations 描述工具的只读、覆盖与幂等语义。Preview 仅操作 Lumina
// 内部工作区，因此 openWorldHint 固定为 false。
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

func previewSessionSchema() map[string]any {
	return objectSchema(map[string]any{
		"id":          map[string]any{"type": "string", "description": "Preview 会话雪花 ID。"},
		"project_id":  map[string]any{"type": "string", "description": "关联项目雪花 ID。"},
		"title":       map[string]any{"type": "string", "description": "会话标题。"},
		"hash":        map[string]any{"type": "string", "pattern": "^[0-9a-f]{32}$", "description": "网页访问哈希（32 位 hex）；不可替代 Q&A file_id 引用。"},
		"status":      map[string]any{"type": "string", "description": "会话状态。"},
		"file_count":  map[string]any{"type": "integer", "minimum": 0, "description": "会话内文件数量（批量统计，避免 Agent 逐会话轮询）。"},
		"created_at":  map[string]any{"type": "string", "description": "RFC 3339 创建时间。"},
		"updated_at":  map[string]any{"type": "string", "description": "RFC 3339 更新时间。"},
		"preview_url": map[string]any{"type": "string", "format": "uri", "description": "绝对预览页 URL。"},
	}, "id", "project_id", "title", "hash", "status", "file_count", "created_at", "updated_at", "preview_url")
}

func previewFileSchema() map[string]any {
	return objectSchema(map[string]any{
		"id":         map[string]any{"type": "string", "description": "Preview 文件雪花 ID。"},
		"session_id": map[string]any{"type": "string", "description": "所属 Preview 会话雪花 ID。"},
		"filename":   map[string]any{"type": "string", "description": "扁平单层文件名。"},
		"mime_type":  map[string]any{"type": "string", "description": "由扩展名推断的 MIME 类型。"},
		"size":       map[string]any{"type": "integer", "minimum": 0, "description": "UTF-8 内容字节数。"},
		"created_at": map[string]any{"type": "string", "description": "RFC 3339 创建时间。"},
		"updated_at": map[string]any{"type": "string", "description": "RFC 3339 更新时间。"},
	}, "id", "session_id", "filename", "mime_type", "size", "created_at", "updated_at")
}

func previewWorkflowSchema() map[string]any {
	return objectSchema(map[string]any{
		"state":        map[string]any{"type": "string", "description": "当前工作流状态。"},
		"next_tool":    map[string]any{"type": "string", "description": "建议下一 MCP 工具；空字符串表示无需固定的下一工具。"},
		"instructions": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Agent 必须按顺序判断的后续动作。"},
	}, "state", "next_tool", "instructions")
}

func previewSupplementSchema() map[string]any {
	return objectSchema(map[string]any{
		"content_type": map[string]any{"type": "string", "const": "preview"},
		"content":      map[string]any{"type": "string", "description": "存在 HTML 入口时为可原样传给 qa_push_supplement.content 的 JSON 字符串；入口未就绪时为空。不可追加代码围栏或说明文字。"},
	}, "content_type", "content")
}

func previewSessionCreateOutputSchema() map[string]any {
	return objectSchema(map[string]any{
		"status":      map[string]any{"type": "string", "const": "success"},
		"message":     map[string]any{"type": "string"},
		"session":     previewSessionSchema(),
		"preview_url": map[string]any{"type": "string", "format": "uri"},
		"workflow":    previewWorkflowSchema(),
	}, "status", "message", "session", "preview_url", "workflow")
}

func previewSessionListOutputSchema() map[string]any {
	return objectSchema(map[string]any{
		"status":      map[string]any{"type": "string", "const": "success"},
		"message":     map[string]any{"type": "string"},
		"items":       map[string]any{"type": "array", "items": previewSessionSchema()},
		"total":       map[string]any{"type": "integer", "minimum": 0},
		"page":        map[string]any{"type": "integer", "minimum": 1},
		"size":        map[string]any{"type": "integer", "minimum": 1},
		"total_pages": map[string]any{"type": "integer", "minimum": 0},
		"workflow":    previewWorkflowSchema(),
	}, "status", "message", "items", "total", "page", "size", "total_pages", "workflow")
}

func previewFileUploadOutputSchema() map[string]any {
	return objectSchema(map[string]any{
		"status":        map[string]any{"type": "string", "const": "success"},
		"message":       map[string]any{"type": "string"},
		"session":       previewSessionSchema(),
		"file":          previewFileSchema(),
		"entry_file":    map[string]any{"type": "string", "description": "当前自动选择的 HTML 入口；无 HTML 时为空。"},
		"preview_url":   map[string]any{"type": "string", "format": "uri"},
		"qa_supplement": previewSupplementSchema(),
		"workflow":      previewWorkflowSchema(),
	}, "status", "message", "session", "file", "entry_file", "preview_url", "qa_supplement", "workflow")
}

func previewFileListOutputSchema() map[string]any {
	return objectSchema(map[string]any{
		"status":        map[string]any{"type": "string", "const": "success"},
		"message":       map[string]any{"type": "string"},
		"session":       previewSessionSchema(),
		"files":         map[string]any{"type": "array", "items": previewFileSchema()},
		"entry_file":    map[string]any{"type": "string", "description": "当前自动选择的 HTML 入口；无 HTML 时为空。"},
		"preview_url":   map[string]any{"type": "string", "format": "uri"},
		"qa_supplement": previewSupplementSchema(),
		"workflow":      previewWorkflowSchema(),
	}, "status", "message", "session", "files", "entry_file", "preview_url", "qa_supplement", "workflow")
}

func previewFileGetOutputSchema() map[string]any {
	return objectSchema(map[string]any{
		"status":  map[string]any{"type": "string", "const": "success"},
		"message": map[string]any{"type": "string"},
		"file": objectSchema(map[string]any{
			"session_id": map[string]any{"type": "string"},
			"filename":   map[string]any{"type": "string"},
			"mime_type":  map[string]any{"type": "string"},
			"size":       map[string]any{"type": "integer", "minimum": 0},
			"content":    map[string]any{"type": "string", "description": "文件完整源码；为控制结构化结果体积，较大内容同时以响应的第二段文本消息返回。"},
		}, "session_id", "filename", "mime_type", "size", "content"),
		"workflow": previewWorkflowSchema(),
	}, "status", "message", "file", "workflow")
}

func objectSchema(properties map[string]any, required ...string) map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties":           properties,
		"required":             required,
	}
}

// ─── Tool Handlers ──────────────────────────────────────────────────────

// handlePreviewSessionCreate 创建预览会话
func handlePreviewSessionCreate(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if previewLogic == nil {
		return previewErrorResult("PreviewLogic 未初始化，请联系管理员"), nil
	}
	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return previewErrorResult(errMsg), nil
	}

	projectIDStr, _ := args["project_id"].(string)
	if projectIDStr == "" {
		return previewErrorResult("缺少必填参数: project_id"), nil
	}
	projectID, err := xSnowflake.ParseSnowflakeID(projectIDStr)
	if err != nil {
		return previewErrorResult(fmt.Sprintf("无效的 project_id: %s", projectIDStr)), nil
	}

	title, _ := args["title"].(string)
	title = strings.TrimSpace(title)
	if len([]rune(title)) > 255 {
		return previewErrorResult("title 长度不能超过 255 个字符"), nil
	}

	resp, xErr := previewLogic.CreateSession(ctx, projectID, title)
	if xErr != nil {
		return previewErrorResult(fmt.Sprintf("创建预览会话失败: %s", xErr.Error())), nil
	}

	previewURL := previewLogic.BuildSessionURL(ctx, resp.Hash, "")
	result := map[string]any{
		"status":      "success",
		"message":     "预览会话已创建；当前为空会话，尚不可交付评审。",
		"session":     previewSessionData(resp, previewURL),
		"preview_url": previewURL,
		"workflow": map[string]any{
			"state":     "awaiting_files",
			"next_tool": "preview_file_upload",
			"instructions": []string{
				"先上传 HTML 入口及其引用的 CSS/JavaScript 文件。",
				"不要打开或向用户交付空的 preview_url。",
				"全部文件上传后调用 preview_file_list 做最终核对。",
			},
		},
	}
	return previewStructuredResult(result), nil
}

// handlePreviewSessionList 分页获取预览会话列表
func handlePreviewSessionList(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if previewLogic == nil {
		return previewErrorResult("PreviewLogic 未初始化，请联系管理员"), nil
	}
	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return previewErrorResult(errMsg), nil
	}

	projectIDStr, _ := args["project_id"].(string)
	if projectIDStr == "" {
		return previewErrorResult("缺少必填参数: project_id"), nil
	}
	projectID, err := xSnowflake.ParseSnowflakeID(projectIDStr)
	if err != nil {
		return previewErrorResult(fmt.Sprintf("无效的 project_id: %s", projectIDStr)), nil
	}

	page := 1
	size := 10
	if p, ok := args["page"].(float64); ok && p > 0 {
		page = int(p)
	}
	if s, ok := args["size"].(float64); ok && s > 0 && s <= 100 {
		size = int(s)
	}

	resp, xErr := previewLogic.ListSessions(ctx, projectID, page, size)
	if xErr != nil {
		return previewErrorResult(fmt.Sprintf("获取预览会话列表失败: %s", xErr.Error())), nil
	}

	totalPages := (resp.Total + int64(size) - 1) / int64(size)
	items := make([]map[string]any, 0, len(resp.Items))
	for i := range resp.Items {
		session := &resp.Items[i]
		previewURL := previewLogic.BuildSessionURL(ctx, session.Hash, "")
		items = append(items, previewSessionData(session, previewURL))
	}

	state := "session_available"
	nextTool := "preview_file_list"
	instructions := []string{
		"按标题和任务语义选择匹配的 active 会话，不要覆盖无关任务。",
		"选定后调用 preview_file_list 核对现有文件，再决定复用或新建。",
	}
	message := fmt.Sprintf("找到 %d 个预览会话。", len(items))
	if len(items) == 0 {
		state = "no_session"
		nextTool = "preview_session_create"
		message = "当前项目没有可复用的预览会话。"
		instructions = []string{"需要可视化预览时调用 preview_session_create 创建新会话。"}
	}

	return previewStructuredResult(map[string]any{
		"status":      "success",
		"message":     message,
		"items":       items,
		"total":       resp.Total,
		"page":        page,
		"size":        size,
		"total_pages": totalPages,
		"workflow": map[string]any{
			"state":        state,
			"next_tool":    nextTool,
			"instructions": instructions,
		},
	}), nil
}

// handlePreviewFileUpload 上传预览文件
func handlePreviewFileUpload(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if previewLogic == nil {
		return previewErrorResult("PreviewLogic 未初始化，请联系管理员"), nil
	}
	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return previewErrorResult(errMsg), nil
	}

	sessionIDStr, _ := args["session_id"].(string)
	if sessionIDStr == "" {
		return previewErrorResult("缺少必填参数: session_id"), nil
	}
	sessionID, err := xSnowflake.ParseSnowflakeID(sessionIDStr)
	if err != nil {
		return previewErrorResult(fmt.Sprintf("无效的 session_id: %s", sessionIDStr)), nil
	}

	filename, _ := args["filename"].(string)
	if filename == "" {
		return previewErrorResult("缺少必填参数: filename"), nil
	}
	content, contentOK := args["content"].(string)
	if !contentOK {
		return previewErrorResult("缺少必填参数或参数类型错误: content"), nil
	}

	resp, xErr := previewLogic.UploadFile(ctx, sessionID, filename, content)
	if xErr != nil {
		return previewErrorResult(fmt.Sprintf("上传预览文件失败: %s", xErr.Error())), nil
	}

	session, xErr := previewLogic.GetSessionByID(ctx, sessionID)
	if xErr != nil {
		return previewErrorResult(fmt.Sprintf("读取预览会话失败: %s", xErr.Error())), nil
	}
	files, xErr := previewLogic.ListFiles(ctx, sessionID)
	if xErr != nil {
		return previewErrorResult(fmt.Sprintf("核对预览文件失败: %s", xErr.Error())), nil
	}

	entry := findPreviewEntry(files)
	entryFilename := ""
	entryID := ""
	previewFilename := resp.Filename
	if entry != nil {
		entryFilename = entry.Filename
		entryID = entry.ID.String()
		previewFilename = entry.Filename
	}
	previewURL := previewLogic.BuildSessionURL(ctx, session.Hash, previewFilename)

	state := "awaiting_html_entry"
	nextTool := "preview_file_upload"
	message := "文件已写入，但会话还没有 HTML 入口，暂不可作为前端页面交付评审。"
	instructions := []string{
		"继续上传 HTML 入口文件，并用相对路径引用同会话中的 CSS/JavaScript。",
		"文件齐全后调用 preview_file_list 做最终核对。",
	}
	if entry != nil {
		state = "reviewable_unverified"
		nextTool = "preview_file_list"
		message = "文件已写入，会话已有 HTML 入口；完成其余依赖上传后仍需最终核对。"
		instructions = []string{
			"若 HTML 仍引用未上传的依赖，继续调用 preview_file_upload。",
			"全部文件写入后调用 preview_file_list；不要跳过最终核对。",
			"核对通过后再打开/分享 preview_url，或使用 qa_supplement 挂载到 Q&A。",
		}
	}

	return previewStructuredResult(map[string]any{
		"status":        "success",
		"message":       message,
		"session":       previewSessionData(session, previewURL),
		"file":          previewFileData(resp),
		"entry_file":    entryFilename,
		"preview_url":   previewURL,
		"qa_supplement": previewSupplementData(sessionIDStr, entryID),
		"workflow": map[string]any{
			"state":        state,
			"next_tool":    nextTool,
			"instructions": instructions,
		},
	}), nil
}

// handlePreviewFileList 获取预览文件列表
func handlePreviewFileList(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if previewLogic == nil {
		return previewErrorResult("PreviewLogic 未初始化，请联系管理员"), nil
	}
	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return previewErrorResult(errMsg), nil
	}

	sessionIDStr, _ := args["session_id"].(string)
	if sessionIDStr == "" {
		return previewErrorResult("缺少必填参数: session_id"), nil
	}
	sessionID, err := xSnowflake.ParseSnowflakeID(sessionIDStr)
	if err != nil {
		return previewErrorResult(fmt.Sprintf("无效的 session_id: %s", sessionIDStr)), nil
	}

	session, xErr := previewLogic.GetSessionByID(ctx, sessionID)
	if xErr != nil {
		return previewErrorResult(fmt.Sprintf("获取预览会话失败: %s", xErr.Error())), nil
	}
	files, xErr := previewLogic.ListFiles(ctx, sessionID)
	if xErr != nil {
		return previewErrorResult(fmt.Sprintf("获取预览文件列表失败: %s", xErr.Error())), nil
	}

	fileItems := make([]map[string]any, 0, len(files))
	for i := range files {
		fileItems = append(fileItems, previewFileData(&files[i]))
	}

	entry := findPreviewEntry(files)
	entryFilename := ""
	entryID := ""
	state := "empty"
	nextTool := "preview_file_upload"
	message := "会话中没有文件。"
	previewURL := previewLogic.BuildSessionURL(ctx, session.Hash, "")
	instructions := []string{"调用 preview_file_upload 上传 HTML 入口及其依赖。"}

	if len(files) > 0 && entry == nil {
		state = "awaiting_html_entry"
		message = "文件清单已核对，但没有 HTML 入口，暂不可作为前端页面交付评审。"
		instructions = []string{
			"调用 preview_file_upload 添加 HTML 入口，并以相对路径引用已有文件。",
			"上传完成后再次调用 preview_file_list。",
		}
	}
	if entry != nil {
		entryFilename = entry.Filename
		entryID = entry.ID.String()
		state = "ready_for_review"
		nextTool = ""
		message = fmt.Sprintf("文件清单已核对，共 %d 个文件，HTML 入口为 %s。", len(files), entry.Filename)
		previewURL = previewLogic.BuildSessionURL(ctx, session.Hash, entry.Filename)
		instructions = []string{
			"若用户请求视觉评审，优先使用客户端原生浏览器/打开链接能力访问 preview_url；能力不可用时向用户提供可点击 URL。",
			"若这是 Q&A 的问题或选项补充，调用 qa_push_supplement，并将 content_type 设为 preview、content 原样设为 qa_supplement.content。",
			"挂载 Q&A 补充后调用 qa_get_answer；不要把 Preview 当作真实项目实现已完成的证据。",
		}
	}

	return previewStructuredResult(map[string]any{
		"status":        "success",
		"message":       message,
		"session":       previewSessionData(session, previewURL),
		"files":         fileItems,
		"entry_file":    entryFilename,
		"preview_url":   previewURL,
		"qa_supplement": previewSupplementData(sessionIDStr, entryID),
		"workflow": map[string]any{
			"state":        state,
			"next_tool":    nextTool,
			"instructions": instructions,
		},
	}), nil
}

// handlePreviewFileGet 获取预览文件完整内容
func handlePreviewFileGet(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if previewLogic == nil {
		return previewErrorResult("PreviewLogic 未初始化，请联系管理员"), nil
	}
	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return previewErrorResult(errMsg), nil
	}

	sessionIDStr, _ := args["session_id"].(string)
	if sessionIDStr == "" {
		return previewErrorResult("缺少必填参数: session_id"), nil
	}
	sessionID, err := xSnowflake.ParseSnowflakeID(sessionIDStr)
	if err != nil {
		return previewErrorResult(fmt.Sprintf("无效的 session_id: %s", sessionIDStr)), nil
	}

	filename, _ := args["filename"].(string)
	if filename == "" {
		return previewErrorResult("缺少必填参数: filename"), nil
	}

	resp, xErr := previewLogic.GetFileContentBySession(ctx, sessionID, filename)
	if xErr != nil {
		return previewErrorResult(fmt.Sprintf("获取预览文件内容失败: %s", xErr.Error())), nil
	}

	metadata := map[string]any{
		"status":  "success",
		"message": "已读取单个预览文件源码；该内容仍属于 Preview 工作区，不等同于真实项目文件。",
		"file": map[string]any{
			"session_id": sessionIDStr,
			"filename":   resp.Filename,
			"mime_type":  resp.MimeType,
			"size":       len(resp.Content),
			"content":    resp.Content,
		},
		"workflow": map[string]any{
			"state":     "source_loaded",
			"next_tool": "",
			"instructions": []string{
				"只在用户要求迭代或提取已确认规范时使用这份源码。",
				"修改 Preview 时，以相同 session_id 和 filename 调用 preview_file_upload 覆写。",
				"覆写后调用 preview_file_list 核对并重新打开预览。",
			},
		},
	}
	return previewFileContentResult(metadata, resp), nil
}

func previewSessionData(session *apiPreview.PreviewSessionResponse, previewURL string) map[string]any {
	return map[string]any{
		"id":          session.ID.String(),
		"project_id":  session.ProjectID.String(),
		"title":       session.Title,
		"hash":        session.Hash,
		"status":      session.Status,
		"file_count":  session.FileCount,
		"created_at":  session.CreatedAt,
		"updated_at":  session.UpdatedAt,
		"preview_url": previewURL,
	}
}

func previewFileData(file *apiPreview.PreviewFileResponse) map[string]any {
	return map[string]any{
		"id":         file.ID.String(),
		"session_id": file.SessionID.String(),
		"filename":   file.Filename,
		"mime_type":  file.MimeType,
		"size":       file.Size,
		"created_at": file.CreatedAt,
		"updated_at": file.UpdatedAt,
	}
}

// findPreviewEntry 从文件清单中选择 HTML 入口文件
//
// 判定依据是 MIME 类型等于 PreviewMimeHTML 常量（"text/html; charset=utf-8"），
// 与 logic 层 inferMimeType 存储的值完全一致，避免裸字符串漂移导致入口永远检测不到。
func findPreviewEntry(files []apiPreview.PreviewFileResponse) *apiPreview.PreviewFileResponse {
	for i := range files {
		if files[i].MimeType == bConst.PreviewMimeHTML {
			return &files[i]
		}
	}
	return nil
}

func previewSupplementData(sessionID, fileID string) map[string]any {
	content := ""
	if sessionID != "" && fileID != "" {
		reference := struct {
			SessionID string `json:"session_id"`
			FileID    string `json:"file_id"`
		}{
			SessionID: sessionID,
			FileID:    fileID,
		}
		if payload, err := json.Marshal(reference); err == nil {
			content = string(payload)
		}
	}
	return map[string]any{
		"content_type": "preview",
		"content":      content,
	}
}

// previewStructuredResult 同时返回结构化结果与其 JSON 文本，兼容尚未消费
// structuredContent 的 MCP 客户端。
func previewStructuredResult(structured map[string]any) *mcp.CallToolResult {
	payload, err := json.Marshal(structured)
	if err != nil {
		return previewErrorResult(fmt.Sprintf("序列化 Preview 工具结果失败: %s", err.Error()))
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(payload)},
		},
		StructuredContent: structured,
	}
}

// previewFileContentResult 将元数据作为结构化结果返回，并单独附加原始源码，避免
// structuredContent 与兼容 JSON 文本重复携带大段文件内容。
func previewFileContentResult(structured map[string]any, file *apiPreview.PreviewFileContentResponse) *mcp.CallToolResult {
	result := previewStructuredResult(structured)
	if result.IsError {
		return result
	}
	result.Content = append(result.Content, &mcp.TextContent{
		Text: fmt.Sprintf("=== %s（%s）===\n\n%s", file.Filename, file.MimeType, file.Content),
	})
	return result
}

func previewErrorResult(message string) *mcp.CallToolResult {
	return errorTextResult(message)
}

// RegisterPreviewTools 将 Preview 模块的 5 个 MCP 工具注册到 Server。
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
