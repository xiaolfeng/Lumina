package route

import (
	"github.com/gin-gonic/gin"

	"github.com/xiaolfeng/Lumina/internal/app/middleware"
	"github.com/xiaolfeng/Lumina/internal/handler"
)

// sshRouter 注册 SSH 密钥管理 API 路由（受 Auth 中间件保护）
func (r *route) sshRouter(route gin.IRouter) {
	h := handler.NewHandler[handler.SshKeyHandler](r.context, "SshKeyHandler")

	g := route.Group("/ssh")
	g.Use(middleware.Auth(r.context))

	g.POST("", h.CreateSshKey)
	g.GET("", h.ListSshKeys)
	g.GET("/:id", h.GetSshKey)
	g.PUT("/:id", h.UpdateSshKey)
	g.DELETE("/:id", h.DeleteSshKey)
	g.GET("/:id/public-key", h.GetPublicKey)
}
