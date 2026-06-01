package middleware

import (
	"context"
	"strings"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
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
// 参数:
//   - ctx: 主上下文，用于依赖注入和日志
func Auth(ctx context.Context) gin.HandlerFunc {
	log := xLog.WithName(xLog.NamedMIDE, "Auth")
	authLogic := logic.NewAuthLogic(ctx)

	return func(c *gin.Context) {
		log.Info(c, "Auth - 验证访问令牌")

		// 提取 Authorization 头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			xResult.AbortError(c, xError.TokenMissing, "未提供访问令牌", false)
			return
		}

		// 解析 Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			xResult.AbortError(c, xError.FormatError, "Authorization 头格式错误，应为 Bearer <token>", false)
			return
		}
		accessToken := parts[1]

		// 验证访问令牌
		found, xErr := authLogic.ValidateAccessToken(c, accessToken)
		if xErr != nil {
			xResult.AbortError(c, xErr.ErrorCode, xErr.ErrorMessage, false)
			return
		}

		// 注入认证标记到上下文
		newCtx := context.WithValue(c.Request.Context(), bConst.CtxOwnerKey, found)
		c.Request = c.Request.WithContext(newCtx)

		c.Next()
	}
}
