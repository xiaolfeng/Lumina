package route

import (
	"net/http"
	"strings"

	xRoute "github.com/bamboo-services/bamboo-base-go/major/route"
	"github.com/gin-gonic/gin"
)

// frontendRouter 注册前端静态资源服务与 SPA fallback。
// 将 web/dist 构建产物通过 go:embed 嵌入到 Go 二进制中，
// 使得单个可执行文件即可提供完整的前后端服务。
func (r *route) frontendRouter() {
	fileServer := http.FileServer(http.FS(r.frontendFS))

	// SPA fallback：所有未匹配已注册路由的请求，
	// 若为静态资源则直接返回文件，否则返回 index.html 让前端路由接管
	r.engine.NoRoute(func(ctx *gin.Context) {
		path := ctx.Request.URL.Path

		// API 与 Swagger 路径交给框架默认 404 处理
		if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/swagger") {
			xRoute.NoRoute(ctx)
			return
		}

		// 尝试打开请求路径对应的静态文件
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

		// SPA fallback：重写为 index.html，让前端路由处理页面路径
		ctx.Request.URL.Path = "/"
		fileServer.ServeHTTP(ctx.Writer, ctx.Request)
	})
}
