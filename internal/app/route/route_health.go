package route

import (
	"github.com/xiaolfeng/Lumina/internal/handler"
	"github.com/gin-gonic/gin"
)

func (r *route) healthRouter(route gin.IRouter) {
	healthHandler := handler.NewHandler[handler.HealthHandler](r.context, "HealthHandler")

	healthGroup := route.Group("/health")
	healthGroup.GET("/ping", healthHandler.Ping)
}
