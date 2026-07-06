package route

import (
	"github.com/gin-gonic/gin"
	"github.com/xiaolfeng/Lumina/internal/app/middleware"
	"github.com/xiaolfeng/Lumina/internal/handler"
)

func (r *route) llmRouter(route gin.IRouter) {
	llmHandler := handler.NewHandler[handler.LlmHandler](r.context, "LlmHandler")

	llmGroup := route.Group("/llm")
	llmGroup.Use(middleware.Auth(r.context))

	providerGroup := llmGroup.Group("/provider")
	{
		providerGroup.POST("", llmHandler.CreateProvider)
		providerGroup.GET("", llmHandler.ListProviders)
		providerGroup.GET("/:id", llmHandler.GetProvider)
		providerGroup.PUT("/:id", llmHandler.UpdateProvider)
		providerGroup.DELETE("/:id", llmHandler.DeleteProvider)
	}

	modelGroup := llmGroup.Group("/model")
	{
		modelGroup.POST("", llmHandler.CreateModel)
		modelGroup.GET("", llmHandler.ListModels)
		modelGroup.GET("/:id", llmHandler.GetModel)
		modelGroup.PUT("/:id", llmHandler.UpdateModel)
		modelGroup.DELETE("/:id", llmHandler.DeleteModel)
	}

	agentGroup := llmGroup.Group("/agent")
	{
		agentGroup.GET("/:role/model", llmHandler.GetAgentModel)
		agentGroup.PUT("/:role/model", llmHandler.UpdateAgentModel)
	}
}
