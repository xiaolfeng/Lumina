package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeaders 设置基础安全响应头。
//
// 缓解：
//   - 点击劫持：X-Frame-Options SAMEORIGIN + CSP frame-ancestors 'self'
//   - MIME 嗅探：X-Content-Type-Options nosniff
//   - Referrer 泄露：Referrer-Policy same-origin
//
// frame 相关头取 SAMEORIGIN/'self' 而非 DENY/'none'，以保留预览模块
// （preview）同源 iframe 内联渲染 HTML 原型的能力。
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "SAMEORIGIN")
		c.Header("Referrer-Policy", "same-origin")
		c.Header("Content-Security-Policy", "frame-ancestors 'self'")
		c.Next()
	}
}
