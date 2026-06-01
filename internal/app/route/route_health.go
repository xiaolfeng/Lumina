package route

import (
	"github.com/gin-gonic/gin"
	"github.com/xiaolfeng/Lumina/internal/handler"
)

func (r *route) healthRouter(route gin.IRouter) {
	healthHandler := handler.NewHandler[handler.HealthHandler](r.context, "HealthHandler")

	healthGroup := route.Group("/health")
	healthGroup.GET("/ping", healthHandler.Ping)
}
