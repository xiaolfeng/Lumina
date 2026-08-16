package handler

import (
	"net/http"
	"strings"

	xResult "github.com/bamboo-services/bamboo-base-go/major/result"
	"github.com/gin-gonic/gin"
	apiCommon "github.com/xiaolfeng/Lumina/api/common"
	apiPreview "github.com/xiaolfeng/Lumina/api/preview"
)

// 确保 apiCommon 包被编译器识别（swag 注释依赖此导入）
var _ = apiCommon.BaseResponse{}

// CreateSession 创建预览会话（管理端）
//
// @Summary     [管理] 创建预览会话
// @Description 提交关联项目 ID 与会话标题创建预览会话，返回会话元数据与访问哈希
// @Tags        Preview接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string                        true  "Bearer Access Token"
// @Param       request        body      apiPreview.CreateSessionRequest  true  "创建会话请求"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiPreview.PreviewSessionResponse}  "创建成功"
// @Failure     400  {object}  apiCommon.BaseResponse  "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Router      /api/v1/preview/sessions [POST]
func (h *PreviewHandler) CreateSession(ctx *gin.Context) {
	h.log.Info(ctx, "CreateSession - 创建预览会话")

	var req apiPreview.CreateSessionRequest
	if !BindJSON(ctx, &req) {
		return
	}

	resp, xErr := h.service.previewLogic.CreateSession(ctx.Request.Context(), req.ProjectID, req.Title)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "创建成功", resp)
}

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

	// 禁止缓存预览文件（内容可能随 WS 实时变更，需保证 iframe 每次加载最新版本）
	ctx.Header("Cache-Control", "no-cache")
	// 对可执行 MIME 施加 sandbox，防止上传的恶意 HTML/SVG 在 Lumina 同源下执行脚本窃取令牌
	// 存储的 HTML MIME 为常量 PreviewMimeHTML（"text/html; charset=utf-8"），故用 HasPrefix 匹配
	if mime := strings.ToLower(file.MimeType); strings.HasPrefix(mime, "text/html") || mime == "image/svg+xml" || strings.Contains(mime, "javascript") {
		ctx.Header("Content-Security-Policy", "sandbox allow-scripts")
	}
	ctx.Data(http.StatusOK, file.MimeType, []byte(file.Content))
}

// GetFileByID 获取预览文件详情（管理端，按 file_id，含会话哈希）
//
// @Summary     [管理] 获取预览文件详情
// @Description 根据文件 ID 查询预览文件详情与关联会话哈希，供 Q&A supplement preview 类型渲染解析 serve 地址
// @Tags        Preview接口
// @Produce     json
// @Param       Authorization  header  string  true  "Bearer Access Token"
// @Param       id             path    string  true  "预览文件 ID"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiPreview.PreviewFileDetailResponse}  "查询成功"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Failure     404  {object}  apiCommon.BaseResponse  "预览文件不存在"
// @Router      /api/v1/preview/files/{id} [GET]
func (h *PreviewHandler) GetFileByID(ctx *gin.Context) {
	h.log.Info(ctx, "GetFileByID - 获取预览文件详情")

	id, xErr := ParseSnowflakeID(ctx, ctx.Param("id"))
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	resp, xErr := h.service.previewLogic.GetFileByID(ctx.Request.Context(), id)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "查询成功", resp)
}

// ListSessions 获取预览会话列表（管理端，分页 + 可选项目筛选）
//
// @Summary     [管理] 获取预览会话列表
// @Description 分页获取预览会话列表，可按项目 ID 筛选（不传则返回全部）
// @Tags        Preview接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string  true  "Bearer Access Token"
// @Param       project_id     query     string  false "项目ID筛选"
// @Param       page           query     int     false "页码"  default(1)
// @Param       size           query     int     false "每页数量"  default(20)
// @Success     200  {object}  apiCommon.BaseResponse{data=apiPreview.PreviewSessionListResponse}  "查询成功"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Router      /api/v1/preview/sessions [GET]
func (h *PreviewHandler) ListSessions(ctx *gin.Context) {
	h.log.Info(ctx, "ListSessions - 获取预览会话列表")

	var req apiPreview.PreviewSessionListRequest
	if !BindQuery(ctx, &req) {
		return
	}

	resp, xErr := h.service.previewLogic.ListSessions(ctx.Request.Context(), req.ProjectID, req.Page, req.Size)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "查询成功", resp)
}

// DeleteSession 删除预览会话（级联删除其下全部文件）
//
// @Summary     [管理] 删除预览会话
// @Description 根据会话 ID 删除预览会话，并级联删除该会话下的全部预览文件
// @Tags        Preview接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string   true  "Bearer Access Token"
// @Param       id             path      string   true  "预览会话 ID"
// @Success     200  {object}  apiCommon.BaseResponse  "删除成功"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Failure     404  {object}  apiCommon.BaseResponse  "预览会话不存在"
// @Router      /api/v1/preview/sessions/{id} [DELETE]
func (h *PreviewHandler) DeleteSession(ctx *gin.Context) {
	h.log.Info(ctx, "DeleteSession - 删除预览会话")

	id, xErr := ParseSnowflakeID(ctx, ctx.Param("id"))
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	if xErr := h.service.previewLogic.DeleteSession(ctx.Request.Context(), id); xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.Success(ctx, "删除成功")
}

// DeleteFile 删除单个预览文件
//
// @Summary     [管理] 删除预览文件
// @Description 根据文件 ID 删除单个预览文件
// @Tags        Preview接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string   true  "Bearer Access Token"
// @Param       id             path      string   true  "预览文件 ID"
// @Success     200  {object}  apiCommon.BaseResponse  "删除成功"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Failure     404  {object}  apiCommon.BaseResponse  "预览文件不存在"
// @Router      /api/v1/preview/files/{id} [DELETE]
func (h *PreviewHandler) DeleteFile(ctx *gin.Context) {
	h.log.Info(ctx, "DeleteFile - 删除预览文件")

	id, xErr := ParseSnowflakeID(ctx, ctx.Param("id"))
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	if xErr := h.service.previewLogic.DeleteFile(ctx.Request.Context(), id); xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.Success(ctx, "删除成功")
}
