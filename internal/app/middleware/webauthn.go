package middleware

import (
	"context"
	"log/slog"
	"net/url"
	"strings"

	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	"github.com/gin-gonic/gin"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

// WebAuthnOrigin 解析浏览器真实 Origin 并注入请求上下文
//
// WebAuthn RP 配置（RPID / RPOrigins）必须与浏览器实际 Origin 一致，
// 否则浏览器在 navigator.credentials.create()/get() 阶段即拒绝（典型错误：
// 「The relying party ID is not a registrable domain suffix of, nor equal to
// the current domain」）。启动期静态配置无法感知部署域名，该中间件从请求中
// 提取浏览器 Origin 注入 context，供 BiometricLogic 按请求动态推导 RP 配置
// （见 BiometricLogic.resolveWebAuthn）。
//
// 解析优先级:
//  1. Origin 请求头（浏览器 POST 请求均携带，最可靠）
//  2. Referer 请求头（Origin 缺失时回退）
//  3. Host 请求头 + scheme 推导（X-Forwarded-Proto / TLS）
//
// 解析失败时不阻断请求，仅记录日志；Logic 层会回退到启动期静态配置。
func WebAuthnOrigin() gin.HandlerFunc {
	log := xLog.WithName(xLog.NamedMIDE, "WebAuthnOrigin")

	return func(c *gin.Context) {
		origin := resolveWebAuthnOrigin(c)

		if origin != "" {
			log.Info(c, "WebAuthnOrigin - 解析浏览器 Origin", slog.String("origin", origin))
			newCtx := context.WithValue(c.Request.Context(), bConst.WebAuthnOriginContextKey, origin)
			c.Request = c.Request.WithContext(newCtx)
		}

		c.Next()
	}
}

// resolveWebAuthnOrigin 从请求中解析浏览器 Origin。
func resolveWebAuthnOrigin(c *gin.Context) string {
	// 1. Origin 请求头（最可靠，形如 http://localhost:3000）
	if origin := c.GetHeader("Origin"); origin != "" {
		if u, err := url.Parse(origin); err == nil && u.Scheme != "" && u.Host != "" {
			return origin
		}
	}

	// 2. Referer 请求头（回退，取 scheme://host）
	if referer := c.GetHeader("Referer"); referer != "" {
		if u, err := url.Parse(referer); err == nil && u.Scheme != "" && u.Host != "" {
			return u.Scheme + "://" + u.Host
		}
	}

	// 3. Host 请求头 + scheme 推导（反向代理场景依赖 X-Forwarded-Proto）
	host := c.Request.Host
	if host == "" {
		return ""
	}
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	} else if proto := c.GetHeader("X-Forwarded-Proto"); proto != "" {
		scheme = strings.ToLower(strings.TrimSpace(strings.Split(proto, ",")[0]))
	}
	return scheme + "://" + host
}
