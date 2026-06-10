package middleware

import (
	"context"
	"strings"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xHttp "github.com/bamboo-services/bamboo-base-go/defined/http"
	xResult "github.com/bamboo-services/bamboo-base-go/major/result"
	"github.com/gin-gonic/gin"
	"github.com/xiaolfeng/Lumina/internal/logic"
)

// ApikeyAuth API Key 认证中间件，验证 Bearer Token 中的 API Key
func ApikeyAuth(ctx context.Context) gin.HandlerFunc {
	log := xLog.WithName(xLog.NamedMIDE, "ApikeyAuth")
	apikeyLogic := logic.NewApikeyLogic(ctx)

	return func(c *gin.Context) {
		log.Info(c, "ApikeyAuth - 验证 API Key")

		apiKey, err := xHttp.GetAuthorization(c)
		if err != nil {
			xResult.AbortError(c, xError.TokenMissing, "未提供 API Key（使用 Bearer <key> 格式）", false)
			return
		}

		if !strings.HasPrefix(apiKey, "lumi_") {
			xResult.AbortError(c, xError.TokenInvalid, "API Key 格式无效", false)
			return
		}

		keyPrefix := apiKey[:8]

		errCode, errMsg := apikeyLogic.ValidateAPIKey(c.Request.Context(), keyPrefix, apiKey)
		if errCode != nil {
			xResult.AbortError(c, errCode, xError.ErrMessage(errMsg), false)
			return
		}

		c.Next()
	}
}
