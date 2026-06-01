package route

import (
	xEnv "github.com/bamboo-services/bamboo-base-go/defined/env"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/xiaolfeng/Lumina/docs"
)

func swaggerRegister(r gin.IRouter) {
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Title = "Bamboo Base Go Template"
	docs.SwaggerInfo.Description = "bamboo-base-go-template API 文档"
	docs.SwaggerInfo.Version = "v1.0.0"
	docs.SwaggerInfo.Host = xEnv.GetEnvString(xEnv.Host, "localhost") + ":" + xEnv.GetEnvString(xEnv.Port, "8080")
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	swaggerGroup := r.Group("/swagger")
	swaggerGroup.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
