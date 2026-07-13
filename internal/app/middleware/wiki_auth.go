// Package middleware Wiki 阅读器 Cookie 鉴权中间件。
//
// 与 Bearer Token（auth.go）和 API Key（apikey.go）不同，Wiki Reader
// 使用独立的 HMAC Cookie 授权体系：每个 Wiki 配置可设置访问密码，
// 用户通过密码验证后获得短期 HMAC Cookie（有效期 2h），后续请求
// 携带 Cookie 即可访问受保护的 Wiki 内容。
//
// 未设置密码的 Wiki 直接放行，无需 Cookie。

package middleware

import (
	"errors"
	"fmt"
	"strconv"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xResult "github.com/bamboo-services/bamboo-base-go/major/result"
	"github.com/gin-gonic/gin"
	"github.com/xiaolfeng/Lumina/internal/service"
)

// WikiConfigGetter Wiki 密码哈希查询回调
//
// 由调用方（route 注册层）注入，避免中间件直接操作 DB。
// 返回空字符串表示该 Wiki 未设置密码保护。
// 接收 gin.Context 以使用请求级别的 context。
type WikiConfigGetter func(ctx *gin.Context, wikiID int64) (passwordHash string, err error)

// WikiAuth Wiki 阅读器 Cookie 鉴权中间件
//
// 鉴权流程：
//  1. 从 URL path 参数 `:id` 解析 wikiID（即 RepoWikiConfig ID）
//  2. 通过 configGetter 查询该 Wiki 的密码哈希
//  3. 密码哈希为空 → 未设密码保护 → 直接放行
//  4. 密码哈希非空 → 检查 Cookie `wiki_auth_{wikiID}` 的 HMAC 签名
//  5. Cookie 有效 → 注入 wikiID 到 context 并放行
//  6. Cookie 缺失/无效 → 401 Unauthorized
//
// 参数:
//   - authTokenService: HMAC Cookie 签名/校验服务
//   - configGetter:     Wiki 配置查询回调（返回密码哈希，空串=无密码）
func WikiAuth(authTokenService *service.WikiAuthTokenService, configGetter WikiConfigGetter) gin.HandlerFunc {
	log := xLog.WithName(xLog.NamedMIDE, "WikiAuth")

	return func(ctx *gin.Context) {
		wikiIDStr := ctx.Param("id")
		wikiID, err := strconv.ParseInt(wikiIDStr, 10, 64)
		if err != nil {
			log.Info(ctx, fmt.Sprintf("WikiAuth - 无效的 Wiki ID [%s]", wikiIDStr))
			xResult.AbortError(ctx, xError.BadRequest, "invalid wiki ID", false)
			return
		}

		// 查询 Wiki 密码哈希（通过回调避免直接 DB 操作）
		passwordHash, err := configGetter(ctx, wikiID)
		if err != nil {
			// 按 xError 错误类型分发：业务 NotFound → 404，其他 → 500
			var xErr *xError.Error
			if errors.As(err, &xErr) && xErr.GetErrorCode() == xError.NotFound {
				log.Info(ctx, fmt.Sprintf("WikiAuth - Wiki 配置不存在 [%d]", wikiID))
				xResult.AbortError(ctx, xError.NotFound, "Wiki 配置不存在", false)
				return
			}
			log.Error(ctx, fmt.Sprintf("WikiAuth - 查询 Wiki 配置失败 [%d]: %v", wikiID, err))
			xResult.AbortError(ctx, xError.ServerInternalError, "internal server error", false)
			return
		}

		// 未设置密码 → 公开 Wiki → 直接放行
		if passwordHash == "" {
			ctx.Set("wikiID", wikiID)
			ctx.Next()
			return
		}

		// 受保护 Wiki → 校验 HMAC Cookie
		cookieName := fmt.Sprintf("wiki_auth_%d", wikiID)
		cookieValue, err := ctx.Cookie(cookieName)
		if err != nil || !authTokenService.ValidateToken(cookieValue, wikiID) {
			log.Info(ctx, fmt.Sprintf("WikiAuth - Cookie 校验失败 [%d]", wikiID))
			xResult.AbortError(ctx, xError.Unauthorized, "wiki authentication required", false)
			return
		}

		// Cookie 有效 → 注入 wikiID 并放行
		ctx.Set("wikiID", wikiID)
		ctx.Next()
	}
}
