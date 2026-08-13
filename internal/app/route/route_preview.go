package route

import (
	"github.com/gin-gonic/gin"
	"github.com/xiaolfeng/Lumina/internal/handler"
)

// previewRouter 注册预览模块公开读取路由（hash 鉴权，无需登录）。
func (r *route) previewRouter(route gin.IRouter) {
	previewHandler := handler.NewHandler[handler.PreviewHandler](r.context, "PreviewHandler")

	previewGroup := route.Group("/preview")
	previewGroup.GET("/sessions/:hash", previewHandler.GetSession)
	previewGroup.GET("/sessions/:hash/files/:filename", previewHandler.GetFile)
	previewGroup.GET("/files/:id", previewHandler.GetFileByID)
}
