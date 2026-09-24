package handler

import (
	"net/http"
	"strings"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xResult "github.com/bamboo-services/bamboo-base-go/major/result"
	"github.com/gin-gonic/gin"
	apiCommon "github.com/xiaolfeng/Lumina/api/common"
	apiPages "github.com/xiaolfeng/Lumina/api/pages"
	pageService "github.com/xiaolfeng/Lumina/internal/service"
)

var _ = apiCommon.BaseResponse{}

// ListPages 分页获取页面列表
//
// @Summary     [管理] 获取页面列表
// @Description 分页获取 Pages 列表，可按项目或空间筛选
// @Tags        页面接口
// @Produce     json
// @Param       Authorization  header  string  true  "Bearer Access Token"
// @Param       project_id     query   string  false "项目ID"
// @Param       workspace_id   query   string  false "空间ID"
// @Param       status         query   string  false "状态筛选（published/archived）"
// @Param       page           query   int     false "页码"
// @Param       size           query   int     false "每页数量"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiPages.PageListResponse}  "查询成功"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Router      /api/v1/pages [GET]
func (h *PagesHandler) ListPages(ctx *gin.Context) {
	var req apiPages.PageListRequest
	if !BindQuery(ctx, &req) {
		return
	}
	resp, xErr := h.service.pagesLogic.List(ctx.Request.Context(), req.ProjectID, req.WorkspaceID, req.Status, req.Page, req.Size)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}
	xResult.SuccessHasData(ctx, "查询成功", resp)
}

// GetPage 获取页面详情
//
// @Summary     [管理] 获取页面详情
// @Description 根据页面 ID 获取详情（不含密码哈希）
// @Tags        页面接口
// @Produce     json
// @Param       Authorization  header  string  true  "Bearer Access Token"
// @Param       id             path    string  true  "页面 ID"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiPages.PageResponse}  "查询成功"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Failure     404  {object}  apiCommon.BaseResponse  "页面不存在"
// @Router      /api/v1/pages/{id} [GET]
func (h *PagesHandler) GetPage(ctx *gin.Context) {
	id, xErr := ParseSnowflakeID(ctx, ctx.Param("id"))
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}
	resp, xErr := h.service.pagesLogic.GetByID(ctx.Request.Context(), id)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}
	xResult.SuccessHasData(ctx, "查询成功", resp)
}

// ListPageVersions 获取页面版本列表
//
// @Summary     [管理] 获取页面版本列表
// @Description 列出指定页面的全部不可变快照版本
// @Tags        页面接口
// @Produce     json
// @Param       Authorization  header  string  true  "Bearer Access Token"
// @Param       id             path    string  true  "页面 ID"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiPages.PageVersionListResponse}  "查询成功"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Router      /api/v1/pages/{id}/versions [GET]
func (h *PagesHandler) ListPageVersions(ctx *gin.Context) {
	id, xErr := ParseSnowflakeID(ctx, ctx.Param("id"))
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}
	resp, xErr := h.service.pagesLogic.ListVersions(ctx.Request.Context(), id)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}
	xResult.SuccessHasData(ctx, "查询成功", resp)
}

// SwitchActiveVersion 切换生效版本指针
//
// @Summary     [管理] 切换生效版本
// @Description 原子更新 latest_version_id，零文件搬迁
// @Tags        页面接口
// @Produce     json
// @Param       Authorization  header  string  true  "Bearer Access Token"
// @Param       id             path    string  true  "页面 ID"
// @Param       version_id     path    string  true  "版本 ID"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiPages.SwitchActiveResponse}  "切换成功"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Router      /api/v1/pages/{id}/versions/{version_id}/switch-active [POST]
func (h *PagesHandler) SwitchActiveVersion(ctx *gin.Context) {
	id, xErr := ParseSnowflakeID(ctx, ctx.Param("id"))
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}
	versionID, xErr := ParseSnowflakeID(ctx, ctx.Param("version_id"))
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}
	resp, xErr := h.service.pagesLogic.SwitchActive(ctx.Request.Context(), id, versionID)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}
	xResult.SuccessHasData(ctx, "切换成功", resp)
}

// ForkPage 从页面版本派生预览会话
//
// @Summary     [管理] Fork 页面到预览
// @Description 深拷贝指定版本文件到新 Preview 会话并写入溯源字段
// @Tags        页面接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header  string  true  "Bearer Access Token"
// @Param       id             path    string  true  "页面 ID"
// @Param       request        body    apiPages.ForkPageRequest  false  "可选版本"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiPages.ForkPageResponse}  "创建成功"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Router      /api/v1/pages/{id}/fork [POST]
func (h *PagesHandler) ForkPage(ctx *gin.Context) {
	id, xErr := ParseSnowflakeID(ctx, ctx.Param("id"))
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}
	var req apiPages.ForkPageRequest
	_ = ctx.ShouldBindJSON(&req)
	resp, xErr := h.service.pagesLogic.Fork(ctx.Request.Context(), id, req.VersionID)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}
	xResult.SuccessHasData(ctx, "Fork 成功", resp)
}

// ArchivePage 归档页面
//
// @Summary     [管理] 归档页面
// @Description 将页面状态改为 archived
// @Tags        页面接口
// @Produce     json
// @Param       Authorization  header  string  true  "Bearer Access Token"
// @Param       id             path    string  true  "页面 ID"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiPages.PageResponse}  "归档成功"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Router      /api/v1/pages/{id}/archive [POST]
func (h *PagesHandler) ArchivePage(ctx *gin.Context) {
	id, xErr := ParseSnowflakeID(ctx, ctx.Param("id"))
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}
	resp, xErr := h.service.pagesLogic.Archive(ctx.Request.Context(), id)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}
	xResult.SuccessHasData(ctx, "归档成功", resp)
}

// UpdateAccessPolicy 更新访问策略
//
// @Summary     [管理] 更新页面访问策略
// @Description 仅控制台可设置公开或密码保护
// @Tags        页面接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header  string  true  "Bearer Access Token"
// @Param       id             path    string  true  "页面 ID"
// @Param       request        body    apiPages.UpdateAccessPolicyRequest  true  "访问策略"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiPages.PageResponse}  "更新成功"
// @Failure     400  {object}  apiCommon.BaseResponse  "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Router      /api/v1/pages/{id}/access-policy [PUT]
func (h *PagesHandler) UpdateAccessPolicy(ctx *gin.Context) {
	id, xErr := ParseSnowflakeID(ctx, ctx.Param("id"))
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}
	var req apiPages.UpdateAccessPolicyRequest
	if !BindJSON(ctx, &req) {
		return
	}
	resp, xErr := h.service.pagesLogic.UpdateAccessPolicy(ctx.Request.Context(), id, req.AccessMode, req.Password)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}
	xResult.SuccessHasData(ctx, "更新成功", resp)
}

// CheckPageAuth 检查密码门状态
//
// @Summary     [公开] 检查页面认证状态
// @Description 返回是否需要密码以及当前 Cookie 是否有效
// @Tags        页面接口
// @Produce     json
// @Param       project_name  path  string  true  "项目名称"
// @Param       slug          path  string  true  "页面标识"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiPages.PageAuthCheckResponse}  "认证状态"
// @Router      /api/v1/pages/by-project/{project_name}/{slug}/auth-check [GET]
func (h *PagesHandler) CheckPageAuth(ctx *gin.Context) {
	projectName := ctx.Param("project_name")
	slug := ctx.Param("slug")
	page, _, xErr := h.service.pagesLogic.GetByProjectNameAndSlug(ctx.Request.Context(), projectName, slug)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}
	cookieValue, _ := ctx.Cookie(pageService.CookieName(page.ID.Int64()))
	resp, xErr := h.service.pagesLogic.CheckAuth(ctx.Request.Context(), projectName, slug, cookieValue)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}
	xResult.SuccessHasData(ctx, "OK", resp)
}

// UnlockPage 解锁密码门
//
// @Summary     [公开] 解锁页面密码门
// @Description 校验密码后签发 HttpOnly Cookie
// @Tags        页面接口
// @Accept      json
// @Produce     json
// @Param       project_name  path  string  true  "项目名称"
// @Param       slug          path  string  true  "页面标识"
// @Param       request       body  apiPages.UnlockPageRequest  true  "密码"
// @Success     200  {object}  apiCommon.BaseResponse  "解锁成功"
// @Failure     401  {object}  apiCommon.BaseResponse  "密码错误"
// @Router      /api/v1/pages/by-project/{project_name}/{slug}/unlock [POST]
func (h *PagesHandler) UnlockPage(ctx *gin.Context) {
	var req apiPages.UnlockPageRequest
	if !BindJSON(ctx, &req) {
		return
	}
	pageID, token, maxAge, xErr := h.service.pagesLogic.Unlock(ctx.Request.Context(), ctx.Param("project_name"), ctx.Param("slug"), req.Password)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}
	if token != "" {
		ctx.SetSameSite(http.SameSiteLaxMode)
		ctx.SetCookie(pageService.CookieName(pageID), token, maxAge, "/", "", ctx.Request.TLS != nil, true)
	}
	xResult.Success(ctx, "解锁成功")
}

// GetPageMeta 展示态元信息与文件清单
//
// @Summary     [公开] 获取页面元信息
// @Description 返回当前生效（或指定）版本的元信息与文件清单
// @Tags        页面接口
// @Produce     json
// @Param       project_name  path   string  true  "项目名称"
// @Param       slug          path   string  true  "页面标识"
// @Param       v             query  string  false "历史版本号"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiPages.PagePublicMetaResponse}  "查询成功"
// @Failure     401  {object}  apiCommon.BaseResponse  "密码门未解锁"
// @Failure     404  {object}  apiCommon.BaseResponse  "页面不存在"
// @Router      /api/v1/pages/by-project/{project_name}/{slug}/meta [GET]
func (h *PagesHandler) GetPageMeta(ctx *gin.Context) {
	projectName := ctx.Param("project_name")
	slug := ctx.Param("slug")
	page, _, xErr := h.service.pagesLogic.GetByProjectNameAndSlug(ctx.Request.Context(), projectName, slug)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}
	cookieValue, _ := ctx.Cookie(pageService.CookieName(page.ID.Int64()))
	auth, xErr := h.service.pagesLogic.CheckAuth(ctx.Request.Context(), projectName, slug, cookieValue)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}
	if auth.PasswordRequired && !auth.Authenticated {
		xResult.AbortError(ctx, xError.Unauthorized, "page authentication required", false)
		return
	}
	resp, xErr := h.service.pagesLogic.PublicMeta(ctx.Request.Context(), projectName, slug, ctx.Query("v"))
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}
	xResult.SuccessHasData(ctx, "查询成功", resp)
}

// ListPageVersionsPublic 展示态版本列表
//
// @Summary     [公开] 获取页面版本列表
// @Description 列出指定页面的全部不可变快照版本；密码页需先解锁 Cookie
// @Tags        页面接口
// @Produce     json
// @Param       project_name  path   string  true  "项目名称"
// @Param       slug          path   string  true  "页面标识"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiPages.PageVersionListResponse}  "查询成功"
// @Failure     401  {object}  apiCommon.BaseResponse  "密码门未解锁"
// @Failure     404  {object}  apiCommon.BaseResponse  "页面不存在"
// @Router      /api/v1/pages/by-project/{project_name}/{slug}/versions [GET]
func (h *PagesHandler) ListPageVersionsPublic(ctx *gin.Context) {
	projectName := ctx.Param("project_name")
	slug := ctx.Param("slug")
	page, _, xErr := h.service.pagesLogic.GetByProjectNameAndSlug(ctx.Request.Context(), projectName, slug)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}
	cookieValue, _ := ctx.Cookie(pageService.CookieName(page.ID.Int64()))
	resp, xErr := h.service.pagesLogic.ListVersionsPublic(ctx.Request.Context(), projectName, slug, cookieValue)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}
	xResult.SuccessHasData(ctx, "查询成功", resp)
}

// ServePagesPath 引擎级路径直出
func (h *PagesHandler) ServePagesPath(ctx *gin.Context) {
	projectName := ctx.Param("project_name")
	slug := ctx.Param("slug")
	filepathParam := strings.TrimPrefix(ctx.Param("filepath"), "/")
	versionLabel := ctx.Query("v")
	if IsDocumentRequest(ctx) {
		return
	}
	if filepathParam == "" {
		entry, xErr := h.service.pagesLogic.EntryFilename(ctx.Request.Context(), projectName, slug, versionLabel)
		if xErr != nil {
			_ = ctx.Error(xErr)
			return
		}
		target := "/pages/" + projectName + "/" + slug + "/" + entry
		if versionLabel != "" {
			target += "?v=" + versionLabel
		}
		ctx.Redirect(http.StatusFound, target)
		return
	}
	file, xErr := h.service.pagesLogic.ServeFile(ctx.Request.Context(), projectName, slug, filepathParam, versionLabel)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}
	writeServedFile(ctx, file.Filename, file.MimeType, file.Content)
}
