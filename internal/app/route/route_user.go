package route

import (
	"github.com/gin-gonic/gin"
	"github.com/xiaolfeng/Lumina/internal/app/middleware"
	"github.com/xiaolfeng/Lumina/internal/handler"
)

func (r *route) userProtectedRouter(route gin.IRouter) {
	userHandler := handler.NewHandler[handler.UserHandler](r.context, "UserHandler")

	userGroup := route.Group("/user")
	userGroup.Use(middleware.Auth(r.context))
	userGroup.GET("/current", userHandler.Current)
}
