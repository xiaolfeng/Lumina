package logic

import (
	"context"
	"strings"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xCtxUtil "github.com/bamboo-services/bamboo-base-go/major/utility/context"

	"github.com/xiaolfeng/Lumina/internal/repository"
	"github.com/xiaolfeng/Lumina/internal/service"
)

// AIPluginLogic AI 插件动态打包与清单渲染编排。
//
// 负责把请求域名 / site.domain 收敛为站点根地址，再交给
// service.AIPluginService 产出 ZIP 与 JSON。不接触 HTTP 写出。
type AIPluginLogic struct {
	logic
	infoRepo *repository.InfoRepo
	plugin   *service.AIPluginService
}

// NewAIPluginLogic 创建 AI 插件业务逻辑层。
func NewAIPluginLogic(ctx context.Context) *AIPluginLogic {
	db := xCtxUtil.MustGetDB(ctx)
	return &AIPluginLogic{
		logic: logic{
			log: xLog.WithName(xLog.NamedLOGC, "AIPluginLogic"),
		},
		infoRepo: repository.NewInfoRepo(db),
		plugin:   service.NewAIPluginService(),
	}
}

// MarketplaceJSON 渲染带当前域名 ZIP 地址与真实 SHA-256 的 marketplace.json。
//
// userAgent 用于识别 ZCode 客户端并输出兼容的 url+zip 源形态。
func (l *AIPluginLogic) MarketplaceJSON(ctx context.Context, requestBaseURL, userAgent string) ([]byte, *xError.Error) {
	l.log.Info(ctx, "MarketplaceJSON - 渲染插件市场清单")
	baseURL := l.resolveBaseURL(ctx, requestBaseURL)
	data, _, err := l.plugin.MarketplaceJSON(baseURL, userAgent)
	if err != nil {
		return nil, xError.NewError(ctx, xError.ServerInternalError, "插件市场清单生成失败", false, err)
	}
	return data, nil
}

// MarketplaceZcodeJSON 渲染 ZCode 专用（url+zip）市场清单，地址确定性与 UA 无关
func (l *AIPluginLogic) MarketplaceZcodeJSON(ctx context.Context, requestBaseURL string) ([]byte, *xError.Error) {
	l.log.Info(ctx, "MarketplaceZcodeJSON - 渲染 ZCode 专用市场清单")
	baseURL := l.resolveBaseURL(ctx, requestBaseURL)
	data, _, err := l.plugin.MarketplaceJSONZcode(baseURL)
	if err != nil {
		return nil, xError.NewError(ctx, xError.ServerInternalError, "插件市场清单生成失败", false, err)
	}
	return data, nil
}

// Zip 返回带当前实例 MCP 地址的插件 ZIP 与其 SHA-256。
func (l *AIPluginLogic) Zip(ctx context.Context, requestBaseURL string) ([]byte, string, *xError.Error) {
	l.log.Info(ctx, "Zip - 打包 AI 插件")
	baseURL := l.resolveBaseURL(ctx, requestBaseURL)
	bundle, err := l.plugin.BundleForInstance(baseURL)
	if err != nil {
		return nil, "", xError.NewError(ctx, xError.ServerInternalError, "插件资源打包失败", false, err)
	}
	return bundle.Zip, bundle.SHA256, nil
}

// WellKnownJSON 渲染 Vercel / Agent Skills 域名探测清单。
func (l *AIPluginLogic) WellKnownJSON(ctx context.Context, requestBaseURL string) ([]byte, *xError.Error) {
	l.log.Info(ctx, "WellKnownJSON - 渲染技能发现清单")
	baseURL := l.resolveBaseURL(ctx, requestBaseURL)
	data, _, err := l.plugin.WellKnownJSON(baseURL)
	if err != nil {
		return nil, xError.NewError(ctx, xError.ServerInternalError, "技能发现清单生成失败", false, err)
	}
	return data, nil
}

// SkillMarkdown 读取指定技能的 SKILL.md 原文。
func (l *AIPluginLogic) SkillMarkdown(ctx context.Context, name string) ([]byte, *xError.Error) {
	l.log.Info(ctx, "SkillMarkdown - 读取技能 ["+name+"]")
	skill, ok, err := l.plugin.SkillByName(name)
	if err != nil {
		return nil, xError.NewError(ctx, xError.ServerInternalError, "读取技能失败", false, err)
	}
	if !ok {
		return nil, xError.NewError(ctx, xError.NotFound, "技能不存在", false, nil)
	}
	return skill.Content, nil
}

// resolveBaseURL 优先使用当前请求的站点根，未提供时回退到 site.domain / 监听地址。
func (l *AIPluginLogic) resolveBaseURL(ctx context.Context, requestBaseURL string) string {
	if base := strings.TrimRight(strings.TrimSpace(requestBaseURL), "/"); base != "" {
		return base
	}
	if l.infoRepo == nil {
		return ""
	}
	return resolveRuntimeDomain(ctx, l.infoRepo)
}
