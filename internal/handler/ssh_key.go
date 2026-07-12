package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xResult "github.com/bamboo-services/bamboo-base-go/major/result"
	"github.com/gin-gonic/gin"
	apiCommon "github.com/xiaolfeng/Lumina/api/common"
	apiSsh "github.com/xiaolfeng/Lumina/api/ssh"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"github.com/xiaolfeng/Lumina/internal/logic"
)

// 确保 apiCommon 包被编译器识别（swag 注释依赖此导入）
var _ = apiCommon.BaseResponse{}

// ──────────────────────────────────────────────────────────────────────
// SshKeyHandler
// ──────────────────────────────────────────────────────────────────────

// CreateSshKey 创建 SSH 密钥
//
// @Summary     [管理] 创建 SSH 密钥
// @Description 创建 SSH 密钥（生成 Ed25519 密钥对或导入已有 PEM 私钥），响应中永不包含私钥
// @Tags        SSH密钥接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string                          true  "Bearer Access Token"
// @Param       request        body      apiSsh.CreateSshKeyRequest      true  "创建 SSH 密钥请求"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiSsh.CreateSshKeyResponse}  "创建成功"
// @Failure     400  {object}  apiCommon.BaseResponse  "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Router      /api/v1/ssh [POST]
func (h *SshKeyHandler) CreateSshKey(ctx *gin.Context) {
	h.log.Info(ctx, "CreateSshKey - 创建 SSH 密钥")

	var req apiSsh.CreateSshKeyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(err)
		return
	}

	// Handler 层格式校验：source 仅接受 generated / imported
	if req.Source != "generated" && req.Source != "imported" {
		_ = ctx.Error(xError.NewError(ctx, xError.ValidationError, "密钥来源仅支持 generated 或 imported", false, nil))
		return
	}
	// imported 时 private_key 必填
	if req.Source == "imported" && req.PrivateKey == "" {
		_ = ctx.Error(xError.NewError(ctx, xError.ValidationError, "导入密钥时 private_key 不能为空", false, nil))
		return
	}

	created, xErr := h.service.sshKeyLogic.CreateSshKey(ctx.Request.Context(), logic.CreateSshKeyRequest{
		Source:      req.Source,
		Name:        req.Name,
		Description: req.Description,
		PrivateKey:  req.PrivateKey,
	})
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	keyResp := sshKeyToResponse(created)
	xResult.SuccessHasData(ctx, "创建成功", apiSsh.CreateSshKeyResponse{
		SshKeyResponse:       keyResp,
		PublicKeyDownloadURL: fmt.Sprintf("/api/v1/ssh/%s/public-key", keyResp.ID),
	})
}

// ListSshKeys 获取 SSH 密钥列表
//
// @Summary     [管理] 获取 SSH 密钥列表
// @Description 分页查询 SSH 密钥列表，返回密钥信息与总数（不含私钥）
// @Tags        SSH密钥接口
// @Produce     json
// @Param       Authorization  header    string   true  "Bearer Access Token"
// @Param       page           query     int      false  "页码"
// @Param       size           query     int      false  "每页数量"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiSsh.SshKeyListResponse}  "获取成功"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Router      /api/v1/ssh [GET]
func (h *SshKeyHandler) ListSshKeys(ctx *gin.Context) {
	h.log.Info(ctx, "ListSshKeys - 获取 SSH 密钥列表")

	pageStr := ctx.DefaultQuery("page", "1")
	sizeStr := ctx.DefaultQuery("size", "20")
	page, _ := strconv.Atoi(pageStr)
	size, _ := strconv.Atoi(sizeStr)

	keys, total, xErr := h.service.sshKeyLogic.ListSshKeys(ctx.Request.Context(), page, size)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	items := make([]apiSsh.SshKeyResponse, 0, len(keys))
	for _, k := range keys {
		items = append(items, sshKeyToResponse(k))
	}

	xResult.SuccessHasData(ctx, "获取成功", apiSsh.SshKeyListResponse{
		Total: total,
		Items: items,
	})
}

// GetSshKey 获取 SSH 密钥详情
//
// @Summary     [管理] 获取 SSH 密钥详情
// @Description 根据密钥 ID 查询 SSH 密钥详情（不含私钥）
// @Tags        SSH密钥接口
// @Produce     json
// @Param       Authorization  header    string   true  "Bearer Access Token"
// @Param       id             path      string   true  "密钥ID"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiSsh.SshKeyResponse}  "获取成功"
// @Failure     400  {object}  apiCommon.BaseResponse  "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Failure     404  {object}  apiCommon.BaseResponse  "密钥不存在"
// @Router      /api/v1/ssh/{id} [GET]
func (h *SshKeyHandler) GetSshKey(ctx *gin.Context) {
	h.log.Info(ctx, "GetSshKey - 获取 SSH 密钥详情")

	id, err := xSnowflake.ParseSnowflakeID(ctx.Param("id"))
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	sshKey, xErr := h.service.sshKeyLogic.GetSshKey(ctx.Request.Context(), id)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "获取成功", sshKeyToResponse(sshKey))
}

// UpdateSshKey 更新 SSH 密钥
//
// @Summary     [管理] 更新 SSH 密钥
// @Description 更新指定 ID 的 SSH 密钥名称和描述（密钥材料不可变）
// @Tags        SSH密钥接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string                          true  "Bearer Access Token"
// @Param       id             path      string                          true  "密钥ID"
// @Param       request        body      apiSsh.UpdateSshKeyRequest      true  "更新 SSH 密钥请求"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiSsh.SshKeyResponse}  "更新成功"
// @Failure     400  {object}  apiCommon.BaseResponse  "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Failure     404  {object}  apiCommon.BaseResponse  "密钥不存在"
// @Router      /api/v1/ssh/{id} [PUT]
func (h *SshKeyHandler) UpdateSshKey(ctx *gin.Context) {
	h.log.Info(ctx, "UpdateSshKey - 更新 SSH 密钥")

	id, err := xSnowflake.ParseSnowflakeID(ctx.Param("id"))
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	var req apiSsh.UpdateSshKeyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(err)
		return
	}

	updated, xErr := h.service.sshKeyLogic.UpdateSshKey(ctx.Request.Context(), id, logic.UpdateSshKeyRequest{
		Name:        req.Name,
		Description: req.Description,
	})
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "更新成功", sshKeyToResponse(updated))
}

// DeleteSshKey 删除 SSH 密钥
//
// @Summary     [管理] 删除 SSH 密钥
// @Description 根据密钥 ID 删除 SSH 密钥，被 RepoWiki 配置引用时禁止删除
// @Tags        SSH密钥接口
// @Produce     json
// @Param       Authorization  header    string   true  "Bearer Access Token"
// @Param       id             path      string   true  "密钥ID"
// @Success     200  {object}  apiCommon.BaseResponse  "删除成功"
// @Failure     400  {object}  apiCommon.BaseResponse  "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Failure     404  {object}  apiCommon.BaseResponse  "密钥不存在"
// @Failure     409  {object}  apiCommon.BaseResponse  "密钥被引用，禁止删除"
// @Router      /api/v1/ssh/{id} [DELETE]
func (h *SshKeyHandler) DeleteSshKey(ctx *gin.Context) {
	h.log.Info(ctx, "DeleteSshKey - 删除 SSH 密钥")

	id, err := xSnowflake.ParseSnowflakeID(ctx.Param("id"))
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	if xErr := h.service.sshKeyLogic.DeleteSshKey(ctx.Request.Context(), id); xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.Success(ctx, "删除成功")
}

// GetPublicKey 下载 SSH 公钥
//
// @Summary     [管理] 下载 SSH 公钥
// @Description 下载指定密钥的 OpenSSH 格式公钥文件（.pub），以附件形式返回
// @Tags        SSH密钥接口
// @Produce     octet-stream
// @Param       Authorization  header    string   true  "Bearer Access Token"
// @Param       id             path      string   true  "密钥ID"
// @Success     200  {file}  binary  "公钥文件"
// @Failure     400  {object}  apiCommon.BaseResponse  "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Failure     404  {object}  apiCommon.BaseResponse  "密钥不存在"
// @Router      /api/v1/ssh/{id}/public-key [GET]
func (h *SshKeyHandler) GetPublicKey(ctx *gin.Context) {
	h.log.Info(ctx, "GetPublicKey - 下载 SSH 公钥")

	id, err := xSnowflake.ParseSnowflakeID(ctx.Param("id"))
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	// 获取密钥实体（需要 Name 构造下载文件名 + PublicKey 作为文件内容）
	sshKey, xErr := h.service.sshKeyLogic.GetSshKey(ctx.Request.Context(), id)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	ctx.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.pub"`, sshKey.Name))
	ctx.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(sshKey.PublicKey))
}

// ──────────────────────────────────────────────────────────────────────
// Entity → DTO 转换辅助函数
// ──────────────────────────────────────────────────────────────────────

// sshKeyToResponse 将 SshKey 实体转为响应 DTO
//
// 安全约束：此函数绝不映射 PrivateKey 字段，确保 API 响应中永不包含私钥。
// entity.SshKey.PrivateKey 虽有 json:"-" tag，但手动映射到 SshKeyResponse 是双重保险。
func sshKeyToResponse(sshKey *entity.SshKey) apiSsh.SshKeyResponse {
	return apiSsh.SshKeyResponse{
		ID:          strconv.FormatInt(sshKey.ID.Int64(), 10),
		Name:        sshKey.Name,
		Description: sshKey.Description,
		KeyType:     sshKey.KeyType,
		PublicKey:   sshKey.PublicKey,
		Fingerprint: sshKey.Fingerprint,
		Source:      sshKey.Source,
		CreatedAt:   sshKey.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   sshKey.UpdatedAt.Format(time.RFC3339),
	}
}
