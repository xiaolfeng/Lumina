package route

import (
	"github.com/gin-gonic/gin"
	"github.com/xiaolfeng/Lumina/internal/app/middleware"
	"github.com/xiaolfeng/Lumina/internal/handler"
)

func (r *route) workspaceRouter(route gin.IRouter) {
	workspaceHandler := handler.NewHandler[handler.WorkspaceHandler](r.context, "WorkspaceHandler")

	workspaceGroup := route.Group("/workspace")
	workspaceGroup.Use(middleware.Auth(r.context))
	workspaceGroup.POST("", workspaceHandler.CreateWorkspace)
	workspaceGroup.GET("", workspaceHandler.ListWorkspaces)
	workspaceGroup.GET("/:id", workspaceHandler.GetWorkspace)
	workspaceGroup.PUT("/:id", workspaceHandler.UpdateWorkspace)
	workspaceGroup.DELETE("/:id", workspaceHandler.DeleteWorkspace)
}
