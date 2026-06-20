package route

import (
	"github.com/gin-gonic/gin"
	"github.com/xiaolfeng/Lumina/internal/app/middleware"
	"github.com/xiaolfeng/Lumina/internal/handler"
)

func (r *route) pinRouter(route gin.IRouter) {
	pinHandler := handler.NewHandler[handler.PinHandler](r.context, "PinHandler")

	pinGroup := route.Group("/pin")
	pinGroup.Use(middleware.Auth(r.context))
	pinGroup.POST("", pinHandler.CreatePin)
	pinGroup.GET("", pinHandler.ListPins)
	pinGroup.GET("/:id", pinHandler.GetPin)
	pinGroup.PUT("/:id", pinHandler.UpdatePin)
	pinGroup.DELETE("/:id", pinHandler.DeletePin)
}
