package handler

import (
	"net/http"

	xResult "github.com/bamboo-services/bamboo-base-go/major/result"
	"github.com/gin-gonic/gin"
	apiCommon "github.com/xiaolfeng/Lumina/api/common"
	apiPreview "github.com/xiaolfeng/Lumina/api/preview"
)

// 确保 apiCommon 包被编译器识别（swag 注释依赖此导入）
var _ = apiCommon.BaseResponse{}

// GetSession 获取预览会话详情与文件列表（公开，hash 鉴权）
//
// @Summary     [公开] 获取预览会话详情
// @Description 根据访问哈希获取预览会话元数据与文件列表，用于预览页加载
// @Tags        Preview接口
// @Produce     json
// @Param       hash  path  string  true  "预览会话访问哈希"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiPreview.PreviewSessionDetailResponse}  "获取成功"
// @Failure     404  {object}  apiCommon.BaseResponse  "预览会话不存在"
// @Router      /api/v1/preview/sessions/{hash} [GET]
func (h *PreviewHandler) GetSession(ctx *gin.Context) {
	h.log.Info(ctx, "GetSession - 获取预览会话详情")

	hash := ctx.Param("hash")
	session, xErr := h.service.previewLogic.GetSessionByHash(ctx.Request.Context(), hash)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	files, xErr := h.service.previewLogic.ListFiles(ctx.Request.Context(), session.ID)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "查询成功", &apiPreview.PreviewSessionDetailResponse{
		Session: *session,
		Files:   files,
	})
}

// GetFile 获取预览文件内容（公开，serve 原始内容，供 iframe src 相对引用）
//
// @Summary     [公开] 获取预览文件内容
// @Description 根据访问哈希与文件名返回文件原始内容（带正确 Content-Type），供预览 iframe 加载
// @Tags        Preview接口
// @Produce     octet-stream
// @Param       hash      path  string  true  "预览会话访问哈希"
// @Param       filename  path  string  true  "文件名（如 index.html）"
// @Success     200  "文件内容"
// @Failure     404  {object}  apiCommon.BaseResponse  "预览文件不存在"
// @Router      /api/v1/preview/sessions/{hash}/files/{filename} [GET]
func (h *PreviewHandler) GetFile(ctx *gin.Context) {
	h.log.Info(ctx, "GetFile - 获取预览文件内容")

	hash := ctx.Param("hash")
	filename := ctx.Param("filename")

	file, xErr := h.service.previewLogic.GetFileContent(ctx.Request.Context(), hash, filename)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	ctx.Data(http.StatusOK, file.MimeType, []byte(file.Content))
}
