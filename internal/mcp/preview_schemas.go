package mcp

// preview_schemas.go 定义 Preview 模块 MCP 工具的输出 Schema 构建器。
//
// Schema 与 handler 实际返回的字段一一对应；新增字段时须同步维护此处与 preview_handlers.go
// 的数据组装逻辑，避免结构化结果与声明漂移。

// previewSessionSchema 会话对象 Schema（含访问地址与 Fork 来源信息）
func previewSessionSchema() map[string]any {
	return objectSchema(map[string]any{
		"id":                map[string]any{"type": "string", "description": "Preview 会话雪花 ID。"},
		"project_id":        map[string]any{"type": "string", "description": "关联项目雪花 ID。"},
		"title":             map[string]any{"type": "string", "description": "会话标题。"},
		"hash":              map[string]any{"type": "string", "description": "网页访问哈希标识；不可替代 Q&A file_id 引用。"},
		"status":            map[string]any{"type": "string", "description": "会话状态。"},
		"file_count":        map[string]any{"type": "integer", "minimum": 0, "description": "会话内文件数量（批量统计，避免 Agent 逐会话轮询）。"},
		"expires_at":        map[string]any{"type": []string{"string", "null"}, "description": "会话过期时间。"},
		"source_page_id":    map[string]any{"type": []string{"string", "null"}, "description": "Fork 来源 Page ID。"},
		"source_page_slug":  map[string]any{"type": []string{"string", "null"}, "description": "Fork 来源 Page slug。"},
		"source_version_id": map[string]any{"type": []string{"string", "null"}, "description": "Fork 基准版本 ID。"},
		"created_at":        map[string]any{"type": "string", "description": "RFC 3339 创建时间。"},
		"updated_at":        map[string]any{"type": "string", "description": "RFC 3339 更新时间。"},
		"preview_url":       map[string]any{"type": "string", "format": "uri", "description": "路径式绝对预览页 URL，如 /preview/<hash>/<filename>。"},
	}, "id", "project_id", "title", "hash", "status", "file_count", "created_at", "updated_at", "preview_url")
}

// previewFileSchema 文件元数据 Schema（不含源码正文）
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

// previewFileWithLinesSchema 文件元数据 Schema 附加总行数（编辑/行级读取场景）
func previewFileWithLinesSchema() map[string]any {
	schema := previewFileSchema()
	props, _ := schema["properties"].(map[string]any)
	props["total_lines"] = map[string]any{"type": "integer", "minimum": 0, "description": "编辑后文件总行数（LF 换行语义）。"}
	schema["required"] = append(schema["required"].([]string), "total_lines")
	return schema
}

// previewWorkflowSchema 工作流指引 Schema
func previewWorkflowSchema() map[string]any {
	return objectSchema(map[string]any{
		"state":        map[string]any{"type": "string", "description": "当前工作流状态。"},
		"next_tool":    map[string]any{"type": "string", "description": "建议下一 MCP 工具；空字符串表示无需固定的下一工具。"},
		"instructions": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Agent 必须按顺序判断的后续动作。"},
	}, "state", "next_tool", "instructions")
}

// previewSupplementSchema Q&A preview 引用 Schema
func previewSupplementSchema() map[string]any {
	return objectSchema(map[string]any{
		"content_type": map[string]any{"type": "string", "const": "preview"},
		"content":      map[string]any{"type": "string", "description": "存在 HTML 入口时为可原样传给 qa_push_supplement.content 的 JSON 字符串；入口未就绪时为空。不可追加代码围栏或说明文字。"},
	}, "content_type", "content")
}

// previewSessionCreateOutputSchema 创建会话输出 Schema
func previewSessionCreateOutputSchema() map[string]any {
	return objectSchema(map[string]any{
		"status":      map[string]any{"type": "string", "const": "success"},
		"message":     map[string]any{"type": "string"},
		"session":     previewSessionSchema(),
		"preview_url": map[string]any{"type": "string", "format": "uri"},
		"workflow":    previewWorkflowSchema(),
	}, "status", "message", "session", "preview_url", "workflow")
}

// previewSessionListOutputSchema 会话列表输出 Schema
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

// previewFileUploadOutputSchema 上传/覆写文件输出 Schema
func previewFileUploadOutputSchema() map[string]any {
	return objectSchema(map[string]any{
		"status":        map[string]any{"type": "string", "const": "success"},
		"message":       map[string]any{"type": "string"},
		"session":       previewSessionSchema(),
		"file":          previewFileSchema(),
		"entry_file":    map[string]any{"type": "string", "description": "当前自动选择的入口文件（优先 index.lpw，回退 HTML）；无入口时为空。"},
		"preview_url":   map[string]any{"type": "string", "format": "uri"},
		"qa_supplement": previewSupplementSchema(),
		"workflow":      previewWorkflowSchema(),
	}, "status", "message", "session", "file", "entry_file", "preview_url", "qa_supplement", "workflow")
}

// previewFileEditOutputSchema 行级编辑输出 Schema
func previewFileEditOutputSchema() map[string]any {
	return objectSchema(map[string]any{
		"status":  map[string]any{"type": "string", "const": "success"},
		"message": map[string]any{"type": "string"},
		"session": previewSessionSchema(),
		"file":    previewFileWithLinesSchema(),
		"edited_region": objectSchema(map[string]any{
			"start_line": map[string]any{"type": "integer", "minimum": 1, "description": "编辑落点区域起始行（含上下文，1 起始闭区间）。"},
			"end_line":   map[string]any{"type": "integer", "minimum": 1, "description": "编辑落点区域结束行（含上下文，闭区间）。"},
			"content":    map[string]any{"type": "string", "description": "编辑落点区域内容，按「行号| 文本」格式；编辑后文件为空时为空字符串。"},
		}, "start_line", "end_line", "content"),
		"entry_file":    map[string]any{"type": "string", "description": "当前自动选择的入口文件（优先 index.lpw，回退 HTML）；无入口时为空。"},
		"preview_url":   map[string]any{"type": "string", "format": "uri"},
		"qa_supplement": previewSupplementSchema(),
		"workflow":      previewWorkflowSchema(),
	}, "status", "message", "session", "file", "edited_region", "entry_file", "preview_url", "qa_supplement", "workflow")
}

// previewFileDeleteOutputSchema 按名删除输出 Schema
func previewFileDeleteOutputSchema() map[string]any {
	return objectSchema(map[string]any{
		"status":  map[string]any{"type": "string", "const": "success"},
		"message": map[string]any{"type": "string"},
		"deleted_file": objectSchema(map[string]any{
			"id":       map[string]any{"type": "string", "description": "被删除文件的雪花 ID。"},
			"filename": map[string]any{"type": "string", "description": "被删除的扁平单层文件名。"},
		}, "id", "filename"),
		"session":       previewSessionSchema(),
		"files":         map[string]any{"type": "array", "items": previewFileSchema(), "description": "删除后剩余的文件清单（按文件名升序）。"},
		"entry_file":    map[string]any{"type": "string", "description": "删除后自动选择的入口文件（优先 index.lpw，回退 HTML）；无入口时为空。"},
		"preview_url":   map[string]any{"type": "string", "format": "uri"},
		"qa_supplement": previewSupplementSchema(),
		"workflow":      previewWorkflowSchema(),
	}, "status", "message", "deleted_file", "session", "files", "entry_file", "preview_url", "qa_supplement", "workflow")
}

// previewFileListOutputSchema 文件清单输出 Schema
func previewFileListOutputSchema() map[string]any {
	return objectSchema(map[string]any{
		"status":        map[string]any{"type": "string", "const": "success"},
		"message":       map[string]any{"type": "string"},
		"session":       previewSessionSchema(),
		"files":         map[string]any{"type": "array", "items": previewFileSchema()},
		"entry_file":    map[string]any{"type": "string", "description": "当前自动选择的入口文件（优先 index.lpw，回退 HTML）；无入口时为空。"},
		"preview_url":   map[string]any{"type": "string", "format": "uri"},
		"qa_supplement": previewSupplementSchema(),
		"workflow":      previewWorkflowSchema(),
	}, "status", "message", "session", "files", "entry_file", "preview_url", "qa_supplement", "workflow")
}

// previewFileGetOutputSchema 行级读取输出 Schema
func previewFileGetOutputSchema() map[string]any {
	return objectSchema(map[string]any{
		"status":  map[string]any{"type": "string", "const": "success"},
		"message": map[string]any{"type": "string"},
		"file": objectSchema(map[string]any{
			"session_id":  map[string]any{"type": "string"},
			"filename":    map[string]any{"type": "string"},
			"mime_type":   map[string]any{"type": "string"},
			"size":        map[string]any{"type": "integer", "minimum": 0},
			"total_lines": map[string]any{"type": "integer", "minimum": 0, "description": "文件总行数（LF 换行语义）；空文件为 0。"},
			"start_line":  map[string]any{"type": "integer", "minimum": 0, "description": "返回内容起始行；空文件为 0。"},
			"end_line":    map[string]any{"type": "integer", "minimum": 0, "description": "返回内容结束行（闭区间）；空文件为 0。"},
			"content":     map[string]any{"type": "string", "description": "传行区间时按「行号| 文本」格式；未传区间时为原始全量源码。为控制结构化结果体积，较大内容同时以响应的第二段文本消息返回。"},
		}, "session_id", "filename", "mime_type", "size", "total_lines", "start_line", "end_line", "content"),
		"workflow": previewWorkflowSchema(),
	}, "status", "message", "file", "workflow")
}

// objectSchema 构建无附加属性的 object Schema（required 为必填字段列表）
func objectSchema(properties map[string]any, required ...string) map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties":           properties,
		"required":             required,
	}
}
