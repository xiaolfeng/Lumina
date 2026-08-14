package bConst

// PreviewSession 状态常量
const (
	PreviewSessionStatusActive  = "active"  // 活跃
	PreviewSessionStatusDeleted = "deleted" // 已删除
)

// Preview 文件 MIME 类型常量
const (
	PreviewMimeHTML  = "text/html; charset=utf-8"              // HTML 文件
	PreviewMimeCSS   = "text/css; charset=utf-8"               // CSS 文件
	PreviewMimeJS    = "application/javascript; charset=utf-8" // JavaScript 文件
	PreviewMimeJSON  = "application/json; charset=utf-8"       // JSON 文件
	PreviewMimeSVG   = "image/svg+xml"                         // SVG 文件（SVG 自声明编码，不加 charset）
	PreviewMimePlain = "text/plain; charset=utf-8"             // 纯文本文件
)

// PreviewFileMaxSize 预览文件单文件大小上限（256KB），超出拒绝上传
const PreviewFileMaxSize = 256 * 1024
