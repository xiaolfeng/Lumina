package middleware

import (
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

// 沙盒 iframe（sandbox="allow-scripts" 且无 allow-same-origin）内的文档处于
// unique / opaque origin。相对引用的 CSS/JS 由该 origin 发起，SameSite=Lax
// Cookie 不会随子资源请求发送，若仍强制登录态会 401 导致原型无法渲染。
//
// 会话 hash / 页面路径本身即能力凭证；仅对非文档静态子资源放行，HTML/MD
// 顶层与 iframe 文档导航仍走原有鉴权。

var sandboxStaticExt = map[string]struct{}{
	// 现代前端组件与源码
	".js":     {},
	".mjs":    {},
	".cjs":    {},
	".ts":     {},
	".mts":    {},
	".cts":    {},
	".jsx":    {},
	".tsx":    {},
	".vue":    {},
	".svelte": {},

	// 样式与样式预处理
	".css":     {},
	".scss":    {},
	".sass":    {},
	".less":    {},
	".styl":    {},
	".postcss": {},

	// 结构化文档、数据与配置
	".json":  {},
	".jsonc": {},
	".json5": {},
	".lpw":   {},
	".xml":   {},
	".yaml":  {},
	".yml":   {},
	".toml":  {},
	".ini":   {},
	".csv":   {},
	".tsv":   {},
	".txt":   {},
	".map":   {},

	// 图像与图标资源
	".png":  {},
	".jpg":  {},
	".jpeg": {},
	".gif":  {},
	".webp": {},
	".avif": {},
	".apng": {},
	".bmp":  {},
	".ico":  {},
	".svg":  {},
	".tiff": {},
	".tif":  {},

	// 字体与排版
	".woff":  {},
	".woff2": {},
	".ttf":   {},
	".otf":   {},
	".eot":   {},

	// 音频与视频媒体
	".mp3":  {},
	".wav":  {},
	".ogg":  {},
	".flac": {},
	".aac":  {},
	".m4a":  {},
	".mp4":  {},
	".webm": {},
	".ogv":  {},
	".mov":  {},

	// 3D 渲染与模型
	".gltf": {},
	".glb":  {},
	".obj":  {},
	".mtl":  {},
	".hdr":  {},

	// WebAssembly 与通用二进制
	".wasm": {},
	".bin":  {},
}

// IsSandboxStaticSubresource 判断当前请求是否为沙盒内相对引用的静态子资源。
func IsSandboxStaticSubresource(ctx *gin.Context) bool {
	if isTopLevelDocument(ctx) {
		return false
	}
	raw := strings.TrimPrefix(ctx.Param("filepath"), "/")
	if raw == "" {
		return false
	}
	ext := strings.ToLower(path.Ext(path.Base(raw)))
	_, ok := sandboxStaticExt[ext]
	return ok
}
