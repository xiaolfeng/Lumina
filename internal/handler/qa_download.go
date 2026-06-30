package handler

import (
	"fmt"
	"io"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xResult "github.com/bamboo-services/bamboo-base-go/major/result"
	"github.com/gin-gonic/gin"
)

// Download Q&A 文件下载（无认证，令牌本身即凭据）
//
// 下载流程：消费一次性令牌 → 打开缓存文件 → 流式写入响应体。
// 令牌无效/过期 → 404；文件不存在 → 404；成功 → 文件流。
//
// @Summary     [公开] Q&A 文件下载
// @Description 通过一次性下载令牌下载 Q&A 附件文件，令牌消费后立即失效
// @Tags        Q&A接口
// @Produce     octet-stream
// @Param       token  path  string  true  "下载令牌"
// @Success     200    {file}  binary  "文件流"
// @Router      /api/v1/qa/download/{token} [GET]
func (h *QaHandler) Download(ctx *gin.Context) {
	h.log.Info(ctx, "Download - Q&A 文件下载")

	token := ctx.Param("token")

	// 消费令牌（原子 GET+DEL，一次性使用）
	info, xErr := h.service.qaLogic.ConsumeDownloadToken(ctx.Request.Context(), token)
	if xErr != nil {
		xResult.AbortError(ctx, xError.NotExist, "下载链接无效或已过期，请重新获取", false)
		return
	}

	// 打开缓存文件
	reader, xErr := h.service.qaLogic.OpenDownloadFile(ctx.Request.Context(), info.FilePath)
	if xErr != nil {
		xResult.AbortError(ctx, xError.FileNotFound, "文件不存在或已被清理", false)
		return
	}
	defer reader.Close()

	// 设置响应头
	ctx.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, info.Filename))
	if info.MimeType != "" {
		ctx.Header("Content-Type", info.MimeType)
	}
	ctx.Status(200)

	// 流式写入响应体
	if _, err := io.Copy(ctx.Writer, reader); err != nil {
		// 客户端断开等 IO 错误：日志记录即可，响应头已发送无法回退
		return
	}
}
