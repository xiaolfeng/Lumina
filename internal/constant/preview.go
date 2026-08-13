package bConst

// PreviewSession 状态常量
const (
	PreviewSessionStatusActive  = "active"  // 活跃
	PreviewSessionStatusDeleted = "deleted" // 已删除
)

// Preview 文件 MIME 类型常量
const (
	PreviewMimeHTML  = "text/html"              // HTML 文件
	PreviewMimeCSS   = "text/css"               // CSS 文件
	PreviewMimeJS    = "application/javascript" // JavaScript 文件
	PreviewMimeJSON  = "application/json"       // JSON 文件
	PreviewMimeSVG   = "image/svg+xml"          // SVG 文件
	PreviewMimePlain = "text/plain"             // 纯文本文件
)

// PreviewFileMaxSize 预览文件单文件大小上限（256KB），超出拒绝上传
const PreviewFileMaxSize = 256 * 1024
