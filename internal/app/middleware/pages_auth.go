package middleware

import (
	"errors"
	"fmt"
	"strings"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xResult "github.com/bamboo-services/bamboo-base-go/major/result"
	"github.com/gin-gonic/gin"
	"github.com/xiaolfeng/Lumina/internal/service"
)

// PagePasswordGetter 查询页面密码哈希；空串表示公开访问。
type PagePasswordGetter func(ctx *gin.Context, projectName, slug string) (pageID int64, passwordHash string, err error)

// PagesAuth Pages 密码门中间件。
//
// public / 空哈希直接放行。password 模式下校验 Cookie。
// 顶层 document 未解锁时仍放行，交由 SPA 渲染密码门；iframe/资源请求返回 401。
func PagesAuth(authToken *service.PageAuthTokenService, getter PagePasswordGetter) gin.HandlerFunc {
	log := xLog.WithName(xLog.NamedMIDE, "PagesAuth")

	return func(ctx *gin.Context) {
		projectName := ctx.Param("project_name")
		slug := ctx.Param("slug")
		if projectName == "" || slug == "" {
			xResult.AbortError(ctx, xError.BadRequest, "无效的页面路径", false)
			return
		}

		pageID, passwordHash, err := getter(ctx, projectName, slug)
		if err != nil {
			var xErr *xError.Error
			if errors.As(err, &xErr) && xErr.GetErrorCode() == xError.NotFound {
				xResult.AbortError(ctx, xError.NotFound, "页面不存在", false)
				return
			}
			log.Error(ctx, fmt.Sprintf("PagesAuth - 查询页面失败 [%s/%s]: %v", projectName, slug, err))
			xResult.AbortError(ctx, xError.ServerInternalError, "internal server error", false)
			return
		}

		ctx.Set("pageID", pageID)
		if passwordHash == "" {
			ctx.Next()
			return
		}

		cookieName := service.CookieName(pageID)
		cookieValue, cookieErr := ctx.Cookie(cookieName)
		if cookieErr == nil && authToken.ValidateToken(cookieValue, pageID) {
			ctx.Next()
			return
		}

		if isTopLevelDocument(ctx) {
			ctx.Next()
			return
		}

		// 与 Preview 相同：沙盒 iframe 内相对 CSS/JS 不携带 Lax Cookie，
		// 密码门仍挡住 HTML 文档；静态子资源凭路径直出才能完成渲染。
		if IsSandboxStaticSubresource(ctx) {
			ctx.Next()
			return
		}

		log.Info(ctx, fmt.Sprintf("PagesAuth - Cookie 校验失败 [%d]", pageID))
		xResult.AbortError(ctx, xError.Unauthorized, "page authentication required", false)
	}
}

func isTopLevelDocument(ctx *gin.Context) bool {
	// 前端预览 iframe 会追加 lumina_frame 查询参数：带参请求一律按非顶层文档处理，
	// 密码未解锁时返回 401 而非放行渲染 SPA（老浏览器缺少 Sec-Fetch-Dest 的兜底）
	if ctx.Query("lumina_frame") != "" {
		return false
	}
	dest := strings.ToLower(strings.TrimSpace(ctx.GetHeader("Sec-Fetch-Dest")))
	switch dest {
	case "document":
		return true
	case "":
		accept := strings.ToLower(ctx.GetHeader("Accept"))
		return strings.Contains(accept, "text/html")
	default:
		return false
	}
}
