package route

import (
	"net/http"

	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	"github.com/gin-gonic/gin"
	"github.com/xiaolfeng/Lumina/internal/app/middleware"
	"github.com/xiaolfeng/Lumina/internal/app/startup"
)

// mcpRouter 注册 MCP Streamable HTTP 端点（API Key 认证）
func (r *route) mcpRouter(route gin.IRouter) {
	handler, ok := r.context.Value(startup.MCPHandlerKey).(http.Handler)
	if !ok {
		log := xLog.WithName(xLog.NamedINIT)
		log.Warn(r.context, "MCP handler not found in context, skipping MCP route registration")
		return
	}

	mcpGroup := route.Group("/mcp")
	mcpGroup.Use(middleware.MCPCompat) // 兼容性处理：补全客户端缺失的 Accept 头
	mcpGroup.Use(middleware.ApikeyAuth(r.context))
	mcpGroup.Any("", gin.WrapH(handler))
	mcpGroup.Any("/*path", gin.WrapH(handler))
}
