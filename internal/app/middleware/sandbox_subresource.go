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
	".css":   {},
	".js":    {},
	".mjs":   {},
	".cjs":   {},
	".json":  {},
	".map":   {},
	".txt":   {},
	".png":   {},
	".jpg":   {},
	".jpeg":  {},
	".gif":   {},
	".webp":  {},
	".ico":   {},
	".svg":   {},
	".woff":  {},
	".woff2": {},
	".ttf":   {},
	".otf":   {},
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
