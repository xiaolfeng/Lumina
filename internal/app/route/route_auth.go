package route

import (
	"github.com/gin-gonic/gin"
	"github.com/xiaolfeng/Lumina/internal/app/middleware"
	"github.com/xiaolfeng/Lumina/internal/handler"
)

func (r *route) authPublicRouter(route gin.IRouter) {
	authHandler := handler.NewHandler[handler.AuthHandler](r.context, "AuthHandler")
	biometricHandler := handler.NewHandler[handler.BiometricHandler](r.context, "BiometricHandler")

	authGroup := route.Group("/auth")
	authGroup.POST("/initialize", authHandler.Initialize)
	authGroup.POST("/login", authHandler.Login)
	authGroup.POST("/refresh", authHandler.Refresh)
	authGroup.GET("/status", authHandler.Status)

	biometricGroup := authGroup.Group("/biometric")
	biometricGroup.GET("/availability", biometricHandler.Availability)
	biometricGroup.POST("/login/start", biometricHandler.LoginStart)
	biometricGroup.POST("/login/finish", biometricHandler.LoginFinish)
}

func (r *route) authProtectedRouter(route gin.IRouter) {
	authHandler := handler.NewHandler[handler.AuthHandler](r.context, "AuthHandler")
	biometricHandler := handler.NewHandler[handler.BiometricHandler](r.context, "BiometricHandler")

	authGroup := route.Group("/auth")
	authGroup.Use(middleware.Auth(r.context))
	authGroup.POST("/logout", authHandler.Logout)

	biometricGroup := authGroup.Group("/biometric")
	biometricGroup.POST("/register/start", biometricHandler.RegisterStart)
	biometricGroup.POST("/register/finish", biometricHandler.RegisterFinish)
}
