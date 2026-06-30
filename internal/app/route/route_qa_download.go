package route

import (
	"github.com/gin-gonic/gin"
	"github.com/xiaolfeng/Lumina/internal/handler"
)

// qaDownloadRouter 注册 Q&A 文件下载路由（无认证，令牌本身即凭据）
func (r *route) qaDownloadRouter(route gin.IRouter) {
	qaHandler := handler.NewHandler[handler.QaHandler](r.context, "QaDownloadHandler")
	route.GET("/qa/download/:token", qaHandler.Download)
}
