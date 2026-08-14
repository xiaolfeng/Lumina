package route

import (
	"github.com/gin-gonic/gin"
	"github.com/xiaolfeng/Lumina/internal/app/middleware"
	"github.com/xiaolfeng/Lumina/internal/handler"
)

// previewRouter 注册预览模块公开读取路由与管理路由。
func (r *route) previewRouter(route gin.IRouter) {
	previewHandler := handler.NewHandler[handler.PreviewHandler](r.context, "PreviewHandler")

	// 公开读取路由（hash 鉴权，无需登录）
	publicGroup := route.Group("/preview")
	publicGroup.GET("/sessions/:hash", previewHandler.GetSession)
	publicGroup.GET("/sessions/:hash/files/:filename", previewHandler.GetFile)
	publicGroup.GET("/files/:id", previewHandler.GetFileByID)

	// 管理路由（Bearer Token 认证）
	adminGroup := route.Group("/preview")
	adminGroup.Use(middleware.Auth(r.context))
	adminGroup.POST("/sessions", previewHandler.CreateSession)
	adminGroup.GET("/sessions", previewHandler.ListSessions)
	adminGroup.DELETE("/sessions/:id", previewHandler.DeleteSession)
	adminGroup.DELETE("/files/:id", previewHandler.DeleteFile)
}
