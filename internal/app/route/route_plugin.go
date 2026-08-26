package route

import (
	"github.com/gin-gonic/gin"
	"github.com/xiaolfeng/Lumina/internal/handler"
)

// pluginRouter 注册 AI 插件公开分发路由（免鉴权，供终端命令直接下载）。
func (r *route) pluginRouter(api gin.IRouter) {
	h := handler.NewHandler[handler.AIPluginHandler](r.context, "AIPluginHandler")

	api.GET("/plugins/marketplace.json", h.GetMarketplace)
	api.GET("/plugins/lumina.zip", h.DownloadZip)

	r.engine.GET("/.well-known/skills/index.json", h.GetWellKnownSkills)
	r.engine.GET("/.well-known/skills/:name/SKILL.md", h.GetWellKnownSkillFile)
}
