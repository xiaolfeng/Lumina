package middleware

import (
	"context"
	"strings"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xHttp "github.com/bamboo-services/bamboo-base-go/defined/http"
	xResult "github.com/bamboo-services/bamboo-base-go/major/result"
	"github.com/gin-gonic/gin"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/logic"
)

// McpAuth MCP 端点认证中间件：OAuth 2.1 访问令牌优先，API Key 回退。
//
// Bearer 令牌带 lum_at_ 前缀时走 OAuth 校验（含 RFC 8707 资源绑定），
// 其余 Bearer 凭证沿用 API Key 校验；认证失败时按 MCP 授权规范返回
// 401 并携带 WWW-Authenticate: Bearer resource_metadata=... 引导客户端
// 发起 OAuth 流程。
func McpAuth(ctx context.Context) gin.HandlerFunc {
	log := xLog.WithName(xLog.NamedMIDE, "McpAuth")
	oauthLogic := logic.NewOAuthLogic(ctx)
	apikeyLogic := logic.NewApikeyLogic(ctx)

	return func(c *gin.Context) {
		token, err := xHttp.GetAuthorization(c)
		if err != nil || token == "" {
			mcpAbortUnauthorized(c, log, "未提供访问凭证（Bearer <token>）")
			return
		}

		if strings.HasPrefix(token, bConst.OAuthAccessTokenPrefix) {
			if oauthLogic.ValidateAccessToken(c.Request.Context(), token, mcpResourceURL(c)) {
				c.Next()
				return
			}
			mcpAbortUnauthorized(c, log, "OAuth 访问令牌无效或已过期")
			return
		}

		// API Key 回退（与 ApikeyAuth 相同校验路径）
		if len(token) < 8 {
			mcpAbortUnauthorized(c, log, "无效的 API Key")
			return
		}
		keyPrefix := token[:8]
		errCode, errMsg := apikeyLogic.ValidateAPIKey(c.Request.Context(), keyPrefix, token)
		if errCode != nil {
			mcpAbortUnauthorized(c, log, errMsg)
			return
		}
		c.Next()
	}
}

// mcpAbortUnauthorized 按 MCP 授权规范返回 401 并附资源元数据指引
func mcpAbortUnauthorized(c *gin.Context, log *xLog.LogNamedLogger, message string) {
	c.Header("WWW-Authenticate", `Bearer resource_metadata="`+mcpBaseURL(c)+bConst.OAuthWellKnownProtectedResource+`"`)
	log.Info(c, "McpAuth - "+message)
	xResult.AbortError(c, xError.Unauthorized, xError.ErrMessage(message), false)
}

// mcpFirstForwardedValue 取多值转发头的第一段（小写、去空白）
func mcpFirstForwardedValue(raw string) string {
	if raw == "" {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(strings.Split(raw, ",")[0]))
}

// mcpBaseURL 从请求推导站点根地址（与 handler.requestBaseURL 同规则）
func mcpBaseURL(c *gin.Context) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	if proto := mcpFirstForwardedValue(c.GetHeader("X-Forwarded-Proto")); proto != "" {
		scheme = proto
	}
	host := c.Request.Host
	if forwarded := mcpFirstForwardedValue(c.GetHeader("X-Forwarded-Host")); forwarded != "" {
		host = forwarded
	}
	if host == "" {
		return ""
	}
	return scheme + "://" + host
}

// mcpResourceURL 当前请求对应的 MCP 资源标识（RFC 8707 audience）
func mcpResourceURL(c *gin.Context) string {
	return mcpBaseURL(c) + bConst.AIPluginMCPPath
}
