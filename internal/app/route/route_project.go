package route

import (
	"github.com/gin-gonic/gin"
	"github.com/xiaolfeng/Lumina/internal/app/middleware"
	"github.com/xiaolfeng/Lumina/internal/handler"
)

func (r *route) projectRouter(route gin.IRouter) {
	projectHandler := handler.NewHandler[handler.ProjectHandler](r.context, "ProjectHandler")

	projectGroup := route.Group("/project")
	projectGroup.Use(middleware.Auth(r.context))
	projectGroup.POST("", projectHandler.CreateProject)
	projectGroup.GET("", projectHandler.ListProjects)
	projectGroup.GET("/:id", projectHandler.GetProject)
	projectGroup.PUT("/:id", projectHandler.UpdateProject)
	projectGroup.DELETE("/:id", projectHandler.DeleteProject)
}
