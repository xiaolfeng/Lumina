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
	userGroup.PUT("/profile", userHandler.UpdateProfile)
	userGroup.PUT("/password", userHandler.UpdatePassword)
	userGroup.GET("/biometric/credentials", userHandler.BiometricCredentials)
	userGroup.DELETE("/biometric/credentials/:id", userHandler.DeleteBiometricCredential)
}
