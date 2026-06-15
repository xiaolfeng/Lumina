package route

import (
	"fmt"
	"io"

	xCtxUtil "github.com/bamboo-services/bamboo-base-go/common/utility/context"
	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xResult "github.com/bamboo-services/bamboo-base-go/major/result"
	"github.com/gin-gonic/gin"
	"github.com/xiaolfeng/Lumina/internal/service"
)

// qaDownloadRouter 注册 Q&A 文件下载路由（无认证，令牌本身即凭据）
//
// 下载流程：ConsumeToken（一次性消费）→ ReadFile → 流式写入响应体。
// 令牌通过 Redis Lua 脚本原子 GET+DEL，保证一次性使用。
func (r *route) qaDownloadRouter(route gin.IRouter) {
	rdb := xCtxUtil.MustGetRDB(r.context)
	tokenSvc := service.NewDownloadTokenService(rdb)
	fileCacheSvc := service.NewFileCacheService()

	route.GET("/qa/download/:token", func(c *gin.Context) {
		token := c.Param("token")

		// 消费令牌（原子 GET+DEL，一次性使用）
		info, err := tokenSvc.ConsumeToken(c.Request.Context(), token)
		if err != nil {
			xResult.AbortError(c, xError.NotExist, "下载链接无效或已过期，请重新获取", false)
			return
		}

		// 读取缓存文件
		reader, err := fileCacheSvc.ReadFile(c.Request.Context(), info.FilePath)
		if err != nil {
			xResult.AbortError(c, xError.FileNotFound, "文件不存在或已被清理", false)
			return
		}
		defer reader.Close()

		// 设置响应头
		c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, info.Filename))
		if info.MimeType != "" {
			c.Header("Content-Type", info.MimeType)
		}
		c.Status(200)

		// 流式写入响应体
		if _, err := io.Copy(c.Writer, reader); err != nil {
			// 客户端断开等 IO 错误：日志记录即可，响应头已发送无法回退
			return
		}
	})
}
