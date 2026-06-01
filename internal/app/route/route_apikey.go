package route

import (
	"github.com/gin-gonic/gin"
	"github.com/xiaolfeng/Lumina/internal/app/middleware"
	"github.com/xiaolfeng/Lumina/internal/handler"
)

func (r *route) apikeyRouter(route gin.IRouter) {
	apikeyHandler := handler.NewHandler[handler.ApikeyHandler](r.context, "ApikeyHandler")

	apikeyGroup := route.Group("/apikey")
	apikeyGroup.Use(middleware.Auth(r.context))
	apikeyGroup.POST("", apikeyHandler.Create)
	apikeyGroup.GET("", apikeyHandler.List)
	apikeyGroup.GET("/:id", apikeyHandler.GetByID)
	apikeyGroup.PUT("/:id", apikeyHandler.Update)
	apikeyGroup.DELETE("/:id", apikeyHandler.Delete)
	apikeyGroup.POST("/:id/reset", apikeyHandler.Reset)
}
