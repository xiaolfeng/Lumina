package route

import (
	"github.com/gin-gonic/gin"
	"github.com/xiaolfeng/Lumina/internal/app/middleware"
	"github.com/xiaolfeng/Lumina/internal/handler"
)

// dashboardRouter 注册看板概览统计路由。
func (r *route) dashboardRouter(route gin.IRouter) {
	dashboardHandler := handler.NewHandler[handler.DashboardHandler](r.context, "DashboardHandler")

	dashboardGroup := route.Group("/dashboard")
	dashboardGroup.Use(middleware.Auth(r.context))
	dashboardGroup.GET("/overview", dashboardHandler.GetDashboardOverview)
}
