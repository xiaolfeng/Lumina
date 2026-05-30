package route

import (
	"context"

	xEnv "github.com/bamboo-services/bamboo-base-go/defined/env"
	xMiddle "github.com/bamboo-services/bamboo-base-go/major/middleware"
	xReg "github.com/bamboo-services/bamboo-base-go/major/register"
	xRoute "github.com/bamboo-services/bamboo-base-go/major/route"
	"github.com/gin-gonic/gin"
)

type route struct {
	engine  *gin.Engine
	context context.Context
}

func NewRoute(reg *xReg.Reg) {
	r := &route{
		engine:  reg.Serve,
		context: reg.Init.Ctx,
	}

	r.engine.NoMethod(xRoute.NoMethod)
	r.engine.NoRoute(xRoute.NoRoute)

	r.engine.Use(xMiddle.ResponseMiddleware)
	r.engine.Use(xMiddle.ReleaseAllCors)
	r.engine.Use(xMiddle.AllowOption)

	if xEnv.GetEnvBool(xEnv.Debug, false) {
		swaggerRegister(r.engine)
	}

	apiRouter := r.engine.Group("/api/v1")
	r.healthRouter(apiRouter)
}
