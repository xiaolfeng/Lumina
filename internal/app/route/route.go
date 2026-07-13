package route

import (
	"context"
	"io/fs"
	"reflect"

	xMiddle "github.com/bamboo-services/bamboo-base-go/major/middleware"
	xReg "github.com/bamboo-services/bamboo-base-go/major/register"
	xRoute "github.com/bamboo-services/bamboo-base-go/major/route"
	xVaild "github.com/bamboo-services/bamboo-base-go/common/validator"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/gin-gonic/gin"
)

type route struct {
	engine         *gin.Engine
	context        context.Context
	frontendFS     fs.FS
	wikiFrontendFS fs.FS
}

// NewRoute 注册所有后端 API 路由（不含前端静态资源）。
func NewRoute(reg *xReg.Reg) {
	NewRouteWithFrontend(nil, nil)(reg)
}

// NewRouteWithFrontend 注册全部路由，包括前端 SPA 静态资源服务。
// frontendFS 或 wikiFrontendFS 为 nil 时跳过对应前端路由注册。
func NewRouteWithFrontend(frontendFS fs.FS, wikiFrontendFS fs.FS) func(reg *xReg.Reg) {
	return func(reg *xReg.Reg) {
		if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
			_ = xVaild.RegisterCustomValidators(v)
			_ = xVaild.RegisterTranslator(v)
			v.RegisterTagNameFunc(func(fld reflect.StructField) string {
				name := fld.Tag.Get("label")
				if name == "" {
					name = fld.Tag.Get("json")
				}
				return name
			})
		}

		r := &route{
			engine:         reg.Serve,
			context:        reg.Init.Ctx,
			frontendFS:     frontendFS,
			wikiFrontendFS: wikiFrontendFS,
		}

		r.engine.NoMethod(xRoute.NoMethod)

		// MCP 路由必须在 engine.Use() 之前注册，以绕开 ResponseMiddleware
		// Gin 的 engine.Group() 在创建时复制当前 engine.Handlers，
		// 因此在此之前创建的 group 不会包含后续注册的全局中间件
		r.mcpRouter(r.engine.Group("/api/v1"))

		// Webhook receiver must also be registered before engine.Use() for raw body access
		r.webhookRouter(r.engine)

		r.engine.Use(xMiddle.ResponseMiddleware)
		r.engine.Use(xMiddle.ReleaseAllCors)
		r.engine.Use(xMiddle.AllowOption)

		swaggerRegister(r.engine)

		apiRouter := r.engine.Group("/api/v1")
		r.healthRouter(apiRouter)
		r.authPublicRouter(apiRouter)
		r.authProtectedRouter(apiRouter)
		r.apikeyRouter(apiRouter)
		r.llmRouter(apiRouter)
		r.settingsRouter(apiRouter)
		r.projectRouter(apiRouter)
		r.pinRouter(apiRouter)
		r.qaRouter(apiRouter)
		r.qaDownloadRouter(apiRouter)
		r.userProtectedRouter(apiRouter)
		r.repowikiRouter(apiRouter)
		r.sshRouter(apiRouter)
		r.wikiReaderRouter(apiRouter)
		r.wsRouter(apiRouter)

		if r.frontendFS != nil {
			r.frontendRouter()
		} else {
			r.engine.NoRoute(xRoute.NoRoute)
		}
	}
}
