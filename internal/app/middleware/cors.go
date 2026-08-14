package middleware

import (
	"strings"

	xEnv "github.com/bamboo-services/bamboo-base-go/defined/env"
	"github.com/gin-gonic/gin"
)

// Cors 白名单 CORS 中间件，反射白名单内的 Origin，替代全局 Access-Control-Allow-Origin:*。
//
// 白名单来自 XLF_ALLOWED_ORIGINS（逗号分隔），默认覆盖前端开发服务器 localhost:3000。
// 生产部署为前后端同源，无需 CORS；白名单模式避免「任意站点跨源读取 API 响应」。
func Cors() gin.HandlerFunc {
	raw := xEnv.GetEnvString("XLF_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:8080")
	allowed := make([]string, 0, 4)
	for _, o := range strings.Split(raw, ",") {
		if o = strings.TrimSpace(o); o != "" {
			allowed = append(allowed, o)
		}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && containsOrigin(allowed, origin) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
			c.Header("Access-Control-Max-Age", "86400")
		}
		c.Next()
	}
}

// containsOrigin 判断 origin 是否在允许白名单内
func containsOrigin(allowed []string, origin string) bool {
	for _, a := range allowed {
		if a == origin || a == "*" {
			return true
		}
	}
	return false
}
