package route

import (
	"net/http"
	"strings"

	xRoute "github.com/bamboo-services/bamboo-base-go/major/route"
	"github.com/gin-gonic/gin"
)

const wikiReaderURLPrefix = "/wiki/"

// frontendRouter 注册前端静态资源服务与 SPA fallback。
// 将 web/dist 与 web-wiki/dist 构建产物通过 go:embed 嵌入到 Go 二进制中，
// 使得单个可执行文件即可提供完整的前后端服务。
func (r *route) frontendRouter() {
	fileServer := http.FileServer(http.FS(r.frontendFS))
	wikiFileServer := http.FileServer(http.FS(r.wikiFrontendFS))

	r.engine.NoRoute(func(ctx *gin.Context) {
		path := ctx.Request.URL.Path

		if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/swagger") {
			xRoute.NoRoute(ctx)
			return
		}

		if strings.HasPrefix(path, wikiReaderURLPrefix) {
			r.serveWikiReaderSPA(ctx, wikiFileServer)
			return
		}

		cleanPath := strings.TrimPrefix(path, "/")
		if cleanPath != "" {
			if f, openErr := r.frontendFS.Open(cleanPath); openErr == nil {
				if stat, statErr := f.Stat(); statErr == nil && !stat.IsDir() {
					f.Close()
					fileServer.ServeHTTP(ctx.Writer, ctx.Request)
					return
				}
				f.Close()
			}
		}

		ctx.Request.URL.Path = "/"
		fileServer.ServeHTTP(ctx.Writer, ctx.Request)
	})
}

// serveWikiReaderSPA 服务 Wiki Reader SPA。
// 静态资源（CSS/JS/图片）直接返回；其余路径返回 index.html 作为 SPA fallback。
func (r *route) serveWikiReaderSPA(ctx *gin.Context, wikiFileServer http.Handler) {
	cleanPath := strings.TrimPrefix(ctx.Request.URL.Path, wikiReaderURLPrefix)
	if cleanPath == "" {
		cleanPath = "/"
	}

	if f, openErr := r.wikiFrontendFS.Open(cleanPath); openErr == nil {
		if stat, statErr := f.Stat(); statErr == nil && !stat.IsDir() {
			f.Close()
			ctx.Request.URL.Path = "/" + cleanPath
			wikiFileServer.ServeHTTP(ctx.Writer, ctx.Request)
			return
		}
		f.Close()
	}

	ctx.Request.URL.Path = "/"
	wikiFileServer.ServeHTTP(ctx.Writer, ctx.Request)
}
