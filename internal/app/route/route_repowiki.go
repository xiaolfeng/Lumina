package route

import (
	"github.com/gin-gonic/gin"

	"github.com/xiaolfeng/Lumina/internal/app/middleware"
	"github.com/xiaolfeng/Lumina/internal/handler"
	wikiService "github.com/xiaolfeng/Lumina/internal/service"
)

// repowikiRouter 注册 RepoWiki 管理 API 路由（受 Auth 中间件保护）
func (r *route) repowikiRouter(route gin.IRouter) {
	h := handler.NewHandler[handler.RepoWikiHandler](r.context, "RepoWikiHandler")

	g := route.Group("/repowiki")
	g.Use(middleware.Auth(r.context))

	g.POST("/configs", h.CreateConfig)
	g.GET("/configs", h.ListConfigs)
	g.GET("/configs/:id", h.GetConfig)
	g.PUT("/configs/:id", h.UpdateConfig)
	g.DELETE("/configs/:id", h.DeleteConfig)

	g.POST("/configs/:id/analyze", h.Analyze)
	g.PUT("/configs/:id/update", h.Update)

	g.GET("/configs/:id/versions", h.ListVersions)
	g.GET("/versions/:id", h.GetVersionDetail)
	g.GET("/versions/:id/status", h.GetVersionStatus)
}

// wikiReaderRouter 注册 Wiki Reader 公开 API 路由（HMAC Cookie 鉴权）
func (r *route) wikiReaderRouter(route gin.IRouter) {
	h := handler.NewWikiReaderHandler(r.context)
	authTokenService := wikiService.NewWikiAuthTokenService()

	wiki := route.Group("/wiki")
	{
		// 公开端点：密码验证 + 授权检查（无需 Cookie）
		wiki.POST("/:id/auth", h.WikiAuth)
		wiki.GET("/:id/auth-check", h.CheckWikiAuth)

		// 受保护端点：manifest + page 需要 WikiAuth 中间件校验
		protected := wiki.Group("/:id")
		protected.Use(middleware.WikiAuth(authTokenService, func(ctx *gin.Context, wikiID int64) (string, error) {
			return h.GetConfigPasswordHash(ctx.Request.Context(), wikiID)
		}))
		{
			protected.GET("/manifest", h.GetWikiManifest)
			protected.GET("/page/*path", h.GetWikiPage)
		}
	}
}
