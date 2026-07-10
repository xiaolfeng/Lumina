package route

import (
	"github.com/gin-gonic/gin"
	"github.com/xiaolfeng/Lumina/internal/app/middleware"
	"github.com/xiaolfeng/Lumina/internal/handler"
)

func (r *route) settingsRouter(route gin.IRouter) {
	settingsHandler := handler.NewHandler[handler.SettingsHandler](r.context, "SettingsHandler")

	settingsGroup := route.Group("/settings")
	settingsGroup.Use(middleware.Auth(r.context))

	settingsGroup.GET("/:category", settingsHandler.GetSettings)
	settingsGroup.PUT("/:category", settingsHandler.UpdateSettings)
}
