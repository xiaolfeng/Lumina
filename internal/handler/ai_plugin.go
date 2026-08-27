package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	apiCommon "github.com/xiaolfeng/Lumina/api/common"
	apiPlugin "github.com/xiaolfeng/Lumina/api/plugin"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

var _ = apiCommon.BaseResponse{}
var _ = apiPlugin.Marketplace{}
var _ = apiPlugin.WellKnownIndex{}

// GetMarketplace 获取 Claude Code 插件市场清单
//
// @Summary     [公开] 插件市场清单
// @Description 按当前请求域名动态渲染 marketplace.json，内含 ZIP 下载地址与实时 SHA-256
// @Tags        插件接口
// @Produce     json
// @Success     200  {object}  apiPlugin.Marketplace  "市场清单"
// @Failure     500  {object}  apiCommon.BaseResponse  "插件资源打包失败"
// @Router      /api/v1/plugins/marketplace.json [GET]
func (h *AIPluginHandler) GetMarketplace(ctx *gin.Context) {
	h.log.Info(ctx, "GetMarketplace - 输出插件市场清单")

	data, xErr := h.service.aiPluginLogic.MarketplaceJSON(ctx.Request.Context(), requestBaseURL(ctx))
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	ctx.Header("Cache-Control", "no-store")
	ctx.Data(http.StatusOK, "application/json; charset=utf-8", data)
}

// DownloadZip 下载 AI 插件 ZIP
//
// @Summary     [公开] 下载插件 ZIP
// @Description 将内嵌的 AI 插件与技能在内存中打包为 zip 流式返回
// @Tags        插件接口
// @Produce     application/zip
// @Success     200  {file}  binary  "插件 ZIP"
// @Failure     500  {object}  apiCommon.BaseResponse  "插件资源打包失败"
// @Router      /api/v1/plugins/lumina.zip [GET]
func (h *AIPluginHandler) DownloadZip(ctx *gin.Context) {
	h.log.Info(ctx, "DownloadZip - 下载插件 ZIP")

	zipBytes, sha256sum, xErr := h.service.aiPluginLogic.Zip(ctx.Request.Context(), requestBaseURL(ctx))
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	ctx.Header("Content-Disposition", `attachment; filename="`+bConst.AIPluginZipFilename+`"`)
	ctx.Header("ETag", `"`+sha256sum+`"`)
	ctx.Header("Cache-Control", "public, max-age=0, must-revalidate")
	ctx.Header("Content-Length", strconv.Itoa(len(zipBytes)))
	ctx.Data(http.StatusOK, "application/zip", zipBytes)
}

// GetWellKnownSkills 获取 Agent Skills 域名探测清单
//
// @Summary     [公开] 技能发现清单
// @Description 输出 /.well-known/skills/index.json，供 npx skills add <origin> 探测本站技能
// @Tags        插件接口
// @Produce     json
// @Success     200  {object}  apiPlugin.WellKnownIndex  "技能发现清单"
// @Failure     500  {object}  apiCommon.BaseResponse    "技能发现清单生成失败"
// @Router      /.well-known/skills/index.json [GET]
func (h *AIPluginHandler) GetWellKnownSkills(ctx *gin.Context) {
	h.log.Info(ctx, "GetWellKnownSkills - 输出技能发现清单")

	data, xErr := h.service.aiPluginLogic.WellKnownJSON(ctx.Request.Context(), requestBaseURL(ctx))
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	ctx.Header("Cache-Control", "no-store")
	ctx.Data(http.StatusOK, "application/json; charset=utf-8", data)
}

// GetWellKnownSkillFile 获取单个技能的 SKILL.md
//
// @Summary     [公开] 读取技能原文
// @Description 按技能标识返回 /.well-known/skills/{name}/SKILL.md，供发现协议按 url 拉取
// @Tags        插件接口
// @Produce     text/markdown
// @Param       name  path  string  true  "技能标识"
// @Success     200  {string}  string  "SKILL.md"
// @Failure     404  {object}  apiCommon.BaseResponse  "技能不存在"
// @Failure     500  {object}  apiCommon.BaseResponse  "读取技能失败"
// @Router      /.well-known/skills/{name}/SKILL.md [GET]
func (h *AIPluginHandler) GetWellKnownSkillFile(ctx *gin.Context) {
	name := ctx.Param("name")
	h.log.Info(ctx, "GetWellKnownSkillFile - 读取技能 ["+name+"]")

	content, xErr := h.service.aiPluginLogic.SkillMarkdown(ctx.Request.Context(), name)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	ctx.Header("Cache-Control", "public, max-age=0, must-revalidate")
	ctx.Data(http.StatusOK, "text/markdown; charset=utf-8", content)
}

// requestBaseURL 从当前请求推导站点根地址（scheme://host[:port]）。
//
// 优先尊重反向代理注入的 X-Forwarded-Proto / X-Forwarded-Host，
// 使 marketplace.json 中的 ZIP 链接跟随用户实际访问地址。
func requestBaseURL(ctx *gin.Context) string {
	scheme := "http"
	if ctx.Request.TLS != nil {
		scheme = "https"
	}
	if proto := firstForwardedValue(ctx.GetHeader("X-Forwarded-Proto")); proto != "" {
		scheme = proto
	}

	host := ctx.Request.Host
	if forwarded := firstForwardedValue(ctx.GetHeader("X-Forwarded-Host")); forwarded != "" {
		host = forwarded
	}
	if host == "" {
		return ""
	}
	return scheme + "://" + host
}

func firstForwardedValue(raw string) string {
	if raw == "" {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(strings.Split(raw, ",")[0]))
}
