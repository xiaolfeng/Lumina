package handler

import (
	xResult "github.com/bamboo-services/bamboo-base-go/major/result"
	"github.com/gin-gonic/gin"
	apiCommon "github.com/xiaolfeng/Lumina/api/common"
	apiUser "github.com/xiaolfeng/Lumina/api/user"
)

// 确保 apiCommon 包被编译器识别（swag 注释依赖此导入）
var _ = apiCommon.BaseResponse{}

// 确保 apiUser 包被编译器识别（swag 注释依赖此导入）
var _ = apiUser.UserInfoResponse{}

// Current 获取当前用户信息
//
// @Summary     [管理] 获取当前用户信息
// @Description 根据访问令牌获取当前登录用户信息（用户名、邮箱）
// @Tags        用户接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string   true  "Bearer Access Token"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiUser.UserInfoResponse}  "获取成功，返回用户信息"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Router      /api/v1/user/current [GET]
func (h *UserHandler) Current(ctx *gin.Context) {
	h.log.Info(ctx, "Current - 获取当前用户信息")

	resp, xErr := h.service.authLogic.GetCurrentUser(ctx)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "获取成功", resp)
}

// UpdateProfile 更新个人资料
//
// @Summary     [管理] 更新个人资料
// @Description 修改用户名和邮箱
// @Tags        用户接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string                       true  "Bearer Access Token"
// @Param       request        body      apiUser.UpdateProfileRequest true  "更新资料请求"
// @Success     200  {object}  apiCommon.BaseResponse  "更新成功"
// @Failure     400  {object}  apiCommon.BaseResponse  "请求参数错误"
// @Router      /api/v1/user/profile [PUT]
func (h *UserHandler) UpdateProfile(ctx *gin.Context) {
	h.log.Info(ctx, "UpdateProfile - 更新个人资料")

	var req apiUser.UpdateProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(err)
		return
	}

	xErr := h.service.authLogic.UpdateProfile(ctx, &req)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.Success(ctx, "更新成功")
}

// UpdatePassword 修改密码
//
// @Summary     [管理] 修改密码
// @Description 验证旧密码并设置新密码，修改成功后撤销所有现有令牌
// @Tags        用户接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string                        true  "Bearer Access Token"
// @Param       request        body      apiUser.UpdatePasswordRequest true  "修改密码请求"
// @Success     200  {object}  apiCommon.BaseResponse  "修改成功"
// @Failure     400  {object}  apiCommon.BaseResponse  "旧密码错误"
// @Router      /api/v1/user/password [PUT]
func (h *UserHandler) UpdatePassword(ctx *gin.Context) {
	h.log.Info(ctx, "UpdatePassword - 修改密码")

	var req apiUser.UpdatePasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(err)
		return
	}

	xErr := h.service.authLogic.UpdatePassword(ctx, &req)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.Success(ctx, "密码修改成功")
}

// BiometricCredentials 获取生物特征凭证列表
//
// @Summary     [管理] 获取生物特征凭证列表
// @Description 返回当前用户已注册的所有生物特征凭证
// @Tags        用户接口
// @Produce     json
// @Param       Authorization  header    string  true  "Bearer Access Token"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiUser.BiometricCredentialListResponse}  "获取成功"
// @Router      /api/v1/user/biometric/credentials [GET]
func (h *UserHandler) BiometricCredentials(ctx *gin.Context) {
	h.log.Info(ctx, "BiometricCredentials - 获取凭证列表")

	resp, xErr := h.service.biometricLogic.ListCredentials(ctx)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "获取成功", resp)
}

// DeleteBiometricCredential 删除生物特征凭证
//
// @Summary     [管理] 删除生物特征凭证
// @Description 删除指定的生物特征凭证
// @Tags        用户接口
// @Produce     json
// @Param       Authorization  header    string  true  "Bearer Access Token"
// @Param       id             path      string  true  "凭证 ID"
// @Success     200  {object}  apiCommon.BaseResponse  "删除成功"
// @Failure     404  {object}  apiCommon.BaseResponse  "凭证不存在"
// @Router      /api/v1/user/biometric/credentials/{id} [DELETE]
func (h *UserHandler) DeleteBiometricCredential(ctx *gin.Context) {
	h.log.Info(ctx, "DeleteBiometricCredential - 删除凭证")

	id := ctx.Param("id")
	xErr := h.service.biometricLogic.DeleteCredential(ctx, id)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.Success(ctx, "删除成功")
}
