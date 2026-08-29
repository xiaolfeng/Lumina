package route

import (
	"github.com/gin-gonic/gin"
	"github.com/xiaolfeng/Lumina/internal/app/middleware"
	"github.com/xiaolfeng/Lumina/internal/handler"
)

// oauthPublicRouter 注册 MCP OAuth 2.1 公开端点。
//
// 与 MCP 端点同理，必须在 engine.Use() 之前注册以绕开 ResponseMiddleware，
// 输出裸 JSON 供客户端与浏览器直接消费。
func (r *route) oauthPublicRouter(engine *gin.Engine) {
	h := handler.NewHandler[handler.OAuthHandler](r.context, "OAuthHandler")

	// RFC 9728：根路径与按资源路径插入两种发现形式
	engine.GET("/.well-known/oauth-protected-resource", h.GetProtectedResourceMetadata)
	engine.GET("/.well-known/oauth-protected-resource/api/v1/mcp", h.GetProtectedResourceMetadata)
	// RFC 8414：授权服务器元数据
	engine.GET("/.well-known/oauth-authorization-server", h.GetAuthorizationServerMetadata)

	engine.GET("/oauth/authorize", h.Authorize)
	engine.POST("/oauth/register", h.RegisterClient)
	engine.POST("/oauth/token", h.Token)
}

// oauthRouter 注册 OAuth 授权裁决接口（控制台登录态保护）。
func (r *route) oauthRouter(api gin.IRouter) {
	h := handler.NewHandler[handler.OAuthHandler](r.context, "OAuthHandler")

	consent := api.Group("/oauth")
	consent.Use(middleware.Auth(r.context))
	consent.GET("/consent", h.GetConsentDetail)
	consent.POST("/consent", h.PostConsent)
}
