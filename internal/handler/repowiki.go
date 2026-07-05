package handler

import (
	"strconv"

	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xResult "github.com/bamboo-services/bamboo-base-go/major/result"
	"github.com/gin-gonic/gin"
	apiCommon "github.com/xiaolfeng/Lumina/api/common"
	apiRepowiki "github.com/xiaolfeng/Lumina/api/repowiki"
	"github.com/xiaolfeng/Lumina/internal/entity"
)

// 确保 apiCommon 包被编译器识别（swag 注释依赖此导入）
var _ = apiCommon.BaseResponse{}

// ──────────────────────────────────────────────────────────────────────
// RepoWikiHandler
// ──────────────────────────────────────────────────────────────────────

// CreateConfig 创建 RepoWiki 配置
//
// @Summary     [管理] 创建 RepoWiki 配置
// @Description 提交仓库地址、分支、语言等参数创建 RepoWiki 分析配置，可选 SSH Key 与 Wiki 访问密码
// @Tags        RepoWiki接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string                              true  "Bearer Access Token"
// @Param       request        body      apiRepowiki.CreateConfigRequest     true  "创建配置请求"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiRepowiki.ConfigResponse}  "创建成功"
// @Failure     400  {object}  apiCommon.BaseResponse  "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Router      /api/v1/repowiki/configs [POST]
func (h *RepoWikiHandler) CreateConfig(ctx *gin.Context) {
	h.log.Info(ctx, "CreateConfig - 创建 RepoWiki 配置")

	var req apiRepowiki.CreateConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(err)
		return
	}

	config, xErr := h.service.repoWikiLogic.CreateConfig(ctx.Request.Context(), &req)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "创建成功", configToResponse(config, nil))
}

// ListConfigs 获取 RepoWiki 配置列表
//
// @Summary     [管理] 获取 RepoWiki 配置列表
// @Description 按 page/size 分页查询 RepoWiki 配置列表，返回配置信息与总数
// @Tags        RepoWiki接口
// @Produce     json
// @Param       Authorization  header    string   true  "Bearer Access Token"
// @Param       page           query     int      false  "页码"
// @Param       size           query     int      false  "每页数量"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiRepowiki.ConfigListResponse}  "获取成功"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Router      /api/v1/repowiki/configs [GET]
func (h *RepoWikiHandler) ListConfigs(ctx *gin.Context) {
	h.log.Info(ctx, "ListConfigs - 获取 RepoWiki 配置列表")

	pageStr := ctx.DefaultQuery("page", "1")
	sizeStr := ctx.DefaultQuery("size", "20")
	page, _ := strconv.Atoi(pageStr)
	size, _ := strconv.Atoi(sizeStr)

	configs, total, xErr := h.service.repoWikiLogic.ListConfigs(ctx.Request.Context(), page, size)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	items := make([]apiRepowiki.ConfigResponse, 0, len(configs))
	for _, c := range configs {
		items = append(items, *configToResponse(c, nil))
	}

	xResult.SuccessHasData(ctx, "获取成功", apiRepowiki.ConfigListResponse{
		Total: total,
		Items: items,
	})
}

// GetConfig 获取 RepoWiki 配置详情
//
// @Summary     [管理] 获取 RepoWiki 配置详情
// @Description 根据配置 ID 查询单个 RepoWiki 配置详情
// @Tags        RepoWiki接口
// @Produce     json
// @Param       Authorization  header    string   true  "Bearer Access Token"
// @Param       id             path      string   true  "配置ID"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiRepowiki.ConfigResponse}  "获取成功"
// @Failure     400  {object}  apiCommon.BaseResponse  "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Failure     404  {object}  apiCommon.BaseResponse  "配置不存在"
// @Router      /api/v1/repowiki/configs/{id} [GET]
func (h *RepoWikiHandler) GetConfig(ctx *gin.Context) {
	h.log.Info(ctx, "GetConfig - 获取 RepoWiki 配置详情")

	id, err := xSnowflake.ParseSnowflakeID(ctx.Param("id"))
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	config, xErr := h.service.repoWikiLogic.GetConfig(ctx.Request.Context(), id)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "获取成功", configToResponse(config, nil))
}

// UpdateConfig 更新 RepoWiki 配置
//
// @Summary     [管理] 更新 RepoWiki 配置
// @Description 更新指定 ID 的 RepoWiki 配置信息（支持部分更新，指针 nil = 不更新）
// @Tags        RepoWiki接口
// @Produce     json
// @Param       Authorization  header    string   true  "Bearer Access Token"
// @Param       id             path      string   true  "配置ID"
// @Param       request        body      apiRepowiki.UpdateConfigRequest     true  "更新配置请求"
// @Success     200  {object}  apiCommon.BaseResponse  "更新成功"
// @Failure     400  {object}  apiCommon.BaseResponse  "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Failure     404  {object}  apiCommon.BaseResponse  "配置不存在"
// @Router      /api/v1/repowiki/configs/{id} [PUT]
func (h *RepoWikiHandler) UpdateConfig(ctx *gin.Context) {
	h.log.Info(ctx, "UpdateConfig - 更新 RepoWiki 配置")

	id, err := xSnowflake.ParseSnowflakeID(ctx.Param("id"))
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	var req apiRepowiki.UpdateConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(err)
		return
	}

	xErr := h.service.repoWikiLogic.UpdateConfig(ctx.Request.Context(), id, &req)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.Success(ctx, "更新成功")
}

// DeleteConfig 删除 RepoWiki 配置
//
// @Summary     [管理] 删除 RepoWiki 配置
// @Description 根据配置 ID 删除指定 RepoWiki 配置，删除后不可恢复
// @Tags        RepoWiki接口
// @Produce     json
// @Param       Authorization  header    string   true  "Bearer Access Token"
// @Param       id             path      string   true  "配置ID"
// @Success     200  {object}  apiCommon.BaseResponse  "删除成功"
// @Failure     400  {object}  apiCommon.BaseResponse  "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Failure     404  {object}  apiCommon.BaseResponse  "配置不存在"
// @Router      /api/v1/repowiki/configs/{id} [DELETE]
func (h *RepoWikiHandler) DeleteConfig(ctx *gin.Context) {
	h.log.Info(ctx, "DeleteConfig - 删除 RepoWiki 配置")

	id, err := xSnowflake.ParseSnowflakeID(ctx.Param("id"))
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	xErr := h.service.repoWikiLogic.DeleteConfig(ctx.Request.Context(), id)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.Success(ctx, "删除成功")
}

// Analyze 触发仓库分析
//
// @Summary     [管理] 触发仓库分析
// @Description 根据配置 ID 触发 RepoWiki 仓库分析（异步执行），返回版本 ID 用于轮询状态
// @Tags        RepoWiki接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string                        true  "Bearer Access Token"
// @Param       id             path      string                        true  "配置ID"
// @Param       request        body      apiRepowiki.AnalyzeRequest    false "分析请求（语言/分支可选）"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiRepowiki.AnalyzeResponse}  "分析任务已触发"
// @Failure     400  {object}  apiCommon.BaseResponse  "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Failure     404  {object}  apiCommon.BaseResponse  "配置不存在"
// @Failure     409  {object}  apiCommon.BaseResponse  "分析任务已达并发上限"
// @Failure     500  {object}  apiCommon.BaseResponse  "LLM Provider 未就绪"
// @Router      /api/v1/repowiki/configs/{id}/analyze [POST]
func (h *RepoWikiHandler) Analyze(ctx *gin.Context) {
	h.log.Info(ctx, "Analyze - 触发仓库分析")

	id, err := xSnowflake.ParseSnowflakeID(ctx.Param("id"))
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	var req apiRepowiki.AnalyzeRequest
	// body 可选，绑定失败时使用零值（默认分支和语言由 Logic 层填充）
	_ = ctx.ShouldBindJSON(&req)

	version, xErr := h.service.repoWikiLogic.AnalyzeRepo(ctx.Request.Context(), id, &req)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "分析任务已触发", apiRepowiki.AnalyzeResponse{
		VersionID: version.ID.Int64(),
		Status:    version.Status,
	})
}

// Update 触发增量更新
//
// @Summary     [管理] 触发增量更新
// @Description 根据配置 ID 触发 RepoWiki 增量更新（异步执行），自动对比 commit hash 决定全量或增量分析
// @Tags        RepoWiki接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string                        true  "Bearer Access Token"
// @Param       id             path      string                        true  "配置ID"
// @Param       request        body      apiRepowiki.AnalyzeRequest    false "更新请求（语言/分支可选）"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiRepowiki.AnalyzeResponse}  "更新任务已触发"
// @Failure     400  {object}  apiCommon.BaseResponse  "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Failure     404  {object}  apiCommon.BaseResponse  "配置不存在"
// @Failure     409  {object}  apiCommon.BaseResponse  "分析任务已达并发上限"
// @Failure     500  {object}  apiCommon.BaseResponse  "LLM Provider 未就绪"
// @Router      /api/v1/repowiki/configs/{id}/update [PUT]
func (h *RepoWikiHandler) Update(ctx *gin.Context) {
	h.log.Info(ctx, "Update - 触发增量更新")

	id, err := xSnowflake.ParseSnowflakeID(ctx.Param("id"))
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	var req apiRepowiki.AnalyzeRequest
	_ = ctx.ShouldBindJSON(&req)

	version, xErr := h.service.repoWikiLogic.AnalyzeRepo(ctx.Request.Context(), id, &req)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "更新任务已触发", apiRepowiki.AnalyzeResponse{
		VersionID: version.ID.Int64(),
		Status:    version.Status,
	})
}

// GetVersionStatus 获取版本分析状态
//
// @Summary     [管理] 获取版本分析状态
// @Description 根据版本 ID 查询分析状态（pending/analyzing/completed/failed），用于轮询分析进度
// @Tags        RepoWiki接口
// @Produce     json
// @Param       Authorization  header    string   true  "Bearer Access Token"
// @Param       id             path      string   true  "版本ID"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiRepowiki.VersionStatusResponse}  "获取成功"
// @Failure     400  {object}  apiCommon.BaseResponse  "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Failure     404  {object}  apiCommon.BaseResponse  "版本不存在"
// @Router      /api/v1/repowiki/versions/{id}/status [GET]
func (h *RepoWikiHandler) GetVersionStatus(ctx *gin.Context) {
	h.log.Info(ctx, "GetVersionStatus - 获取版本分析状态")

	id, err := xSnowflake.ParseSnowflakeID(ctx.Param("id"))
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	version, xErr := h.service.repoWikiLogic.GetVersionStatus(ctx.Request.Context(), id)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "获取成功", versionToStatusResponse(version))
}

// GetVersionDetail 获取版本详情
//
// @Summary     [管理] 获取版本详情
// @Description 根据版本 ID 查询完整的版本信息（含 commit hash、分析耗时、token 消耗等）
// @Tags        RepoWiki接口
// @Produce     json
// @Param       Authorization  header    string   true  "Bearer Access Token"
// @Param       id             path      string   true  "版本ID"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiRepowiki.VersionStatusResponse}  "获取成功"
// @Failure     400  {object}  apiCommon.BaseResponse  "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Failure     404  {object}  apiCommon.BaseResponse  "版本不存在"
// @Router      /api/v1/repowiki/versions/{id} [GET]
func (h *RepoWikiHandler) GetVersionDetail(ctx *gin.Context) {
	h.log.Info(ctx, "GetVersionDetail - 获取版本详情")

	id, err := xSnowflake.ParseSnowflakeID(ctx.Param("id"))
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	version, xErr := h.service.repoWikiLogic.GetVersionStatus(ctx.Request.Context(), id)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "获取成功", versionToStatusResponse(version))
}

// ListVersions 获取版本列表
//
// @Summary     [管理] 获取版本列表
// @Description 按配置 ID 分页查询 Wiki 版本列表，返回版本状态与总数
// @Tags        RepoWiki接口
// @Produce     json
// @Param       Authorization  header    string   true  "Bearer Access Token"
// @Param       id             path      string   true  "配置ID"
// @Param       page           query     int      false  "页码"
// @Param       size           query     int      false  "每页数量"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiRepowiki.VersionListResponse}  "获取成功"
// @Failure     400  {object}  apiCommon.BaseResponse  "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Router      /api/v1/repowiki/configs/{id}/versions [GET]
func (h *RepoWikiHandler) ListVersions(ctx *gin.Context) {
	h.log.Info(ctx, "ListVersions - 获取版本列表")

	configID, err := xSnowflake.ParseSnowflakeID(ctx.Param("id"))
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	pageStr := ctx.DefaultQuery("page", "1")
	sizeStr := ctx.DefaultQuery("size", "20")
	page, _ := strconv.Atoi(pageStr)
	size, _ := strconv.Atoi(sizeStr)

	versions, total, xErr := h.service.repoWikiLogic.ListVersions(ctx.Request.Context(), configID, page, size)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	items := make([]apiRepowiki.VersionStatusResponse, 0, len(versions))
	for _, v := range versions {
		items = append(items, *versionToStatusResponse(v))
	}

	xResult.SuccessHasData(ctx, "获取成功", apiRepowiki.VersionListResponse{
		Total: total,
		Items: items,
	})
}

// ──────────────────────────────────────────────────────────────────────
// Entity → DTO 转换辅助函数
// ──────────────────────────────────────────────────────────────────────

// configToResponse 将 RepoWikiConfig 实体转为响应 DTO
//
// latestVersion 为 nil 时 ConfigResponse.LatestVersion 留空（omitempty）。
func configToResponse(config *entity.RepoWikiConfig, latestVersion *entity.WikiVersion) *apiRepowiki.ConfigResponse {
	resp := &apiRepowiki.ConfigResponse{
		ID:              config.ID.Int64(),
		ProjectID:       config.ProjectID.Int64(),
		Name:            config.Name,
		RepoURL:         config.GitURL,
		DefaultBranch:   config.DefaultBranch,
		DefaultLanguage: config.DefaultLanguage,
		Status:          config.Status,
		HasSSHKey:       config.SSHKeyEncrypted != "",
		HasPassword:     config.WikiPasswordHash != "",
		LastAccessedAt:  config.LastAccessedAt,
		CreatedAt:       config.CreatedAt,
		UpdatedAt:       config.UpdatedAt,
	}

	if latestVersion != nil {
		resp.LatestVersion = versionToStatusResponse(latestVersion)
	}

	return resp
}

// versionToStatusResponse 将 WikiVersion 实体转为版本状态响应 DTO
func versionToStatusResponse(version *entity.WikiVersion) *apiRepowiki.VersionStatusResponse {
	return &apiRepowiki.VersionStatusResponse{
		ID:              version.ID.Int64(),
		ConfigID:        version.ConfigID.Int64(),
		CommitHash:      version.CommitHash,
		Branch:          version.Branch,
		Language:        version.Language,
		Status:          version.Status,
		CurrentStage:    version.CurrentStage,
		ProgressPercent: 0, // 实体未存储进度百分比，Pipeline 阶段性更新通过 CurrentStage 体现
		ErrorMsg:        version.ErrorMsg,
		FileCount:       version.FileCount,
		TokenCount:      version.TokenCount,
		DurationMs:      version.DurationMs,
		StartedAt:       version.StartedAt,
		CompletedAt:     version.CompletedAt,
		CreatedAt:       version.CreatedAt,
	}
}
