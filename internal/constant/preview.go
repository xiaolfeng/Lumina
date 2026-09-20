package bConst

// PreviewSession 状态常量
const (
	PreviewSessionStatusActive  = "active"  // 活跃
	PreviewSessionStatusDeleted = "deleted" // 已删除
)

// Preview 文件 MIME 类型常量
const (
	PreviewMimeHTML     = "text/html; charset=utf-8"              // HTML 文件
	PreviewMimeCSS      = "text/css; charset=utf-8"               // CSS 文件
	PreviewMimeJS       = "application/javascript; charset=utf-8" // JavaScript 文件
	PreviewMimeJSON     = "application/json; charset=utf-8"       // JSON 文件
	PreviewMimeMarkdown = "text/markdown; charset=utf-8"          // Markdown 文件
	PreviewMimeSVG      = "image/svg+xml"                         // SVG 文件（SVG 自声明编码，不加 charset）
	PreviewMimeTSX      = "text/typescript-jsx; charset=utf-8"   // TSX 组件文件
	PreviewMimeJSX      = "text/jsx; charset=utf-8"              // JSX 组件文件
	PreviewMimePlain    = "text/plain; charset=utf-8"             // 纯文本文件
	PreviewMimeLPW      = "application/vnd.lumina.preview+json; charset=utf-8" // LPW 预览文档
)

// PreviewFileMaxSize 预览文件单文件大小上限（256KB），超出拒绝上传
const PreviewFileMaxSize = 256 * 1024

// Preview 行级编辑操作常量（MCP preview_file_edit 的 operation 取值）
const (
	PreviewEditOperationInsert  = "insert"  // 在 start_line 前插入内容行
	PreviewEditOperationReplace = "replace" // 替换 [start_line, end_line] 闭区间
	PreviewEditOperationDelete  = "delete"  // 删除 [start_line, end_line] 闭区间
)
