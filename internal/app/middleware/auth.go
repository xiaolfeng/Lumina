package middleware

import (
	"context"
	"net/url"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xHttp "github.com/bamboo-services/bamboo-base-go/defined/http"
	xResult "github.com/bamboo-services/bamboo-base-go/major/result"
	"github.com/gin-gonic/gin"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/logic"
)

// Auth 授权中间件，验证 Bearer AccessToken 并注入认证标记到上下文
//
// 从 Authorization: Bearer <token> 头中提取 AT，通过 AuthLogic 验证后
// 将认证标记注入到 context.Context 中，后续 Handler 可通过 CtxOwnerKey 获取。
//
// iframe / WebSocket 无法设置 Header，回退 Cookie；WebSocket 仍保留 query token 兜底。
func Auth(ctx context.Context) gin.HandlerFunc {
	log := xLog.WithName(xLog.NamedMIDE, "Auth")
	authLogic := logic.NewAuthLogic(ctx)

	return func(c *gin.Context) {
		log.Info(c, "Auth - 验证访问令牌")

		accessToken, err := xHttp.GetAuthorization(c)
		if err != nil {
			if cookie, cErr := c.Cookie("access_token"); cErr == nil && cookie != "" {
				accessToken = cookie
			} else if c.IsWebsocket() {
				accessToken = c.Query("token")
			}
			if accessToken == "" {
				xResult.AbortError(c, xError.TokenMissing, "未提供访问令牌", false)
				return
			}
		}

		found, xErr := authLogic.ValidateAccessToken(c, accessToken)
		if xErr != nil {
			xResult.AbortError(c, xErr.ErrorCode, xErr.ErrorMessage, false)
			return
		}

		newCtx := context.WithValue(c.Request.Context(), bConst.CtxOwnerKey, found)
		c.Request = c.Request.WithContext(newCtx)
		c.Next()
	}
}

// AuthOrRedirectLogin 与 Auth 相同，但顶层 document 未登录时 302 到登录页。
func AuthOrRedirectLogin(ctx context.Context) gin.HandlerFunc {
	log := xLog.WithName(xLog.NamedMIDE, "AuthOrRedirectLogin")
	authLogic := logic.NewAuthLogic(ctx)

	return func(c *gin.Context) {
		accessToken, err := xHttp.GetAuthorization(c)
		if err != nil {
			if cookie, cErr := c.Cookie("access_token"); cErr == nil && cookie != "" {
				accessToken = cookie
			}
		}
		if accessToken == "" {
			abortPreviewAuth(c)
			return
		}
		found, xErr := authLogic.ValidateAccessToken(c, accessToken)
		if xErr != nil {
			log.Info(c, "AuthOrRedirectLogin - 令牌无效")
			abortPreviewAuth(c)
			return
		}
		newCtx := context.WithValue(c.Request.Context(), bConst.CtxOwnerKey, found)
		c.Request = c.Request.WithContext(newCtx)
		c.Next()
	}
}

func abortPreviewAuth(c *gin.Context) {
	if isTopLevelDocument(c) {
		redirect := url.QueryEscape(c.Request.URL.RequestURI())
		c.Redirect(302, "/auth/login?redirect="+redirect)
		c.Abort()
		return
	}
	xResult.AbortError(c, xError.TokenMissing, "未提供访问令牌", false)
}
