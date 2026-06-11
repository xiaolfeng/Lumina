package route

import (
	"github.com/gin-gonic/gin"
	"github.com/xiaolfeng/Lumina/internal/app/middleware"
	"github.com/xiaolfeng/Lumina/internal/handler"
)

func (r *route) qaRouter(route gin.IRouter) {
	qaHandler := handler.NewHandler[handler.QaHandler](r.context, "QaHandler")

	qaGroup := route.Group("/qa")
	qaGroup.Use(middleware.Auth(r.context))

	qaGroup.GET("/sessions", qaHandler.ListSessions)
	qaGroup.POST("/sessions", qaHandler.CreateSession)
	qaGroup.GET("/sessions/:id", qaHandler.GetSession)
	qaGroup.DELETE("/sessions/:id", qaHandler.DeleteSession)
	qaGroup.GET("/sessions/:id/questions/:qid", qaHandler.GetQuestion)
	qaGroup.GET("/config", qaHandler.GetQaConfig)
	qaGroup.PUT("/config", qaHandler.UpdateQaConfig)
}
