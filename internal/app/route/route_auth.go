package route

import (
	"github.com/gin-gonic/gin"
	"github.com/xiaolfeng/Lumina/internal/app/middleware"
	"github.com/xiaolfeng/Lumina/internal/handler"
)

func (r *route) authPublicRouter(route gin.IRouter) {
	authHandler := handler.NewHandler[handler.AuthHandler](r.context, "AuthHandler")

	authGroup := route.Group("/auth")
	authGroup.POST("/initialize", authHandler.Initialize)
	authGroup.POST("/login", authHandler.Login)
	authGroup.POST("/refresh", authHandler.Refresh)
	authGroup.GET("/status", authHandler.Status)
}

func (r *route) authProtectedRouter(route gin.IRouter) {
	authHandler := handler.NewHandler[handler.AuthHandler](r.context, "AuthHandler")

	authGroup := route.Group("/auth")
	authGroup.Use(middleware.Auth(r.context))
	authGroup.POST("/logout", authHandler.Logout)
}
