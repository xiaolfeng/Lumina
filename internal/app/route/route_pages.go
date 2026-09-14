package route

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiaolfeng/Lumina/internal/app/middleware"
	"github.com/xiaolfeng/Lumina/internal/handler"
	"github.com/xiaolfeng/Lumina/internal/logic"
	"github.com/xiaolfeng/Lumina/internal/service"
)

func (r *route) pagesRouter(route gin.IRouter) {
	pagesHandler := handler.NewHandler[handler.PagesHandler](r.context, "PagesHandler")

	public := route.Group("/pages")
	public.GET("/:project_name/:slug/auth-check", pagesHandler.CheckPageAuth)
	public.POST("/:project_name/:slug/unlock", pagesHandler.UnlockPage)
	public.GET("/:project_name/:slug/meta", pagesHandler.GetPageMeta)

	admin := route.Group("/pages")
	admin.Use(middleware.Auth(r.context))
	admin.GET("", pagesHandler.ListPages)
	admin.GET("/:id", pagesHandler.GetPage)
	admin.GET("/:id/versions", pagesHandler.ListPageVersions)
	admin.POST("/:id/versions/:version_id/switch-active", pagesHandler.SwitchActiveVersion)
	admin.POST("/:id/fork", pagesHandler.ForkPage)
	admin.POST("/:id/archive", pagesHandler.ArchivePage)
	admin.PUT("/:id/access-policy", pagesHandler.UpdateAccessPolicy)
}

func (r *route) pagesPathRouter() {
	pagesHandler := handler.NewHandler[handler.PagesHandler](r.context, "PagesPathHandler")
	pagesLogic := logic.NewPagesLogic(r.context)
	authToken := service.NewPageAuthTokenService()

	g := r.engine.Group("/pages")
	g.Use(middleware.PagesAuth(authToken, func(ctx *gin.Context, projectName, slug string) (int64, string, error) {
		return pagesLogic.LookupAuth(ctx.Request.Context(), projectName, slug)
	}))
	g.GET("/:project_name/:slug", r.servePagesOrSPA(pagesHandler))
	g.GET("/:project_name/:slug/*filepath", r.servePagesOrSPA(pagesHandler))
}

func (r *route) servePagesOrSPA(pagesHandler *handler.PagesHandler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if isDocumentRequest(ctx) {
			r.serveConsoleIndex(ctx)
			return
		}
		pagesHandler.ServePagesPath(ctx)
	}
}

func (r *route) previewPathRouter() {
	previewHandler := handler.NewHandler[handler.PreviewHandler](r.context, "PreviewPathHandler")
	g := r.engine.Group("/preview")
	g.Use(middleware.AuthOrRedirectLogin(r.context))
	g.GET("/:session_hash", r.servePreviewOrSPA(previewHandler))
	g.GET("/:session_hash/*filepath", r.servePreviewOrSPA(previewHandler))
}

func (r *route) servePreviewOrSPA(previewHandler *handler.PreviewHandler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if isDocumentRequest(ctx) {
			r.serveConsoleIndex(ctx)
			return
		}
		previewHandler.ServePreviewPath(ctx)
	}
}

func (r *route) serveConsoleIndex(ctx *gin.Context) {
	if r.frontendFS == nil {
		ctx.Status(http.StatusNotFound)
		return
	}
	ctx.Request.URL.Path = "/"
	http.FileServer(http.FS(r.frontendFS)).ServeHTTP(ctx.Writer, ctx.Request)
}

func isDocumentRequest(ctx *gin.Context) bool {
	return handler.IsDocumentRequest(ctx)
}
