package handler

import (
	xResult "github.com/bamboo-services/bamboo-base-go/major/result"
	"github.com/gin-gonic/gin"
	apiBiometric "github.com/xiaolfeng/Lumina/api/biometric"
	apiCommon "github.com/xiaolfeng/Lumina/api/common"
)

// 确保 apiCommon 和 apiBiometric 包被编译器识别（swag 注释依赖此导入）
var _ = apiCommon.BaseResponse{}
var _ = apiBiometric.AvailabilityResponse{}

// Availability 生物特征登录可用性
//
// @Summary     生物特征登录可用性
// @Description 查询是否已注册生物特征凭证，登录页据此决定是否显示生物特征登录按钮
// @Tags        生物特征接口
// @Produce     json
// @Success     200  {object}  apiCommon.BaseResponse{data=apiBiometric.AvailabilityResponse}  "查询成功"
// @Router      /api/v1/auth/biometric/availability [GET]
func (h *BiometricHandler) Availability(ctx *gin.Context) {
	h.log.Info(ctx, "Availability - 查询生物特征可用性")

	resp, xErr := h.service.biometricLogic.GetAvailability(ctx)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "查询成功", resp)
}

// RegisterStart 生物特征注册开始
//
// @Summary     [管理] 生物特征注册开始
// @Description 发起 WebAuthn 注册流程，返回创建选项供浏览器调用 navigator.credentials.create
// @Tags        生物特征接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string                            true  "Bearer Access Token"
// @Param       request        body      apiBiometric.RegisterStartRequest true  "注册开始请求"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiBiometric.RegisterStartResponse}  "返回创建选项"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Router      /api/v1/auth/biometric/register/start [POST]
func (h *BiometricHandler) RegisterStart(ctx *gin.Context) {
	h.log.Info(ctx, "RegisterStart - 生物特征注册开始")

	var req apiBiometric.RegisterStartRequest
	if !BindJSON(ctx, &req) {
		return
	}

	resp, xErr := h.service.biometricLogic.RegisterStart(ctx, &req)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "注册开始", resp)
}

// RegisterFinish 生物特征注册完成
//
// @Summary     [管理] 生物特征注册完成
// @Description 提交浏览器创建的 PublicKeyCredential 完成注册
// @Tags        生物特征接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string                             true  "Bearer Access Token"
// @Param       request        body      apiBiometric.RegisterFinishRequest true  "注册完成请求"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiBiometric.RegisterFinishResponse}  "注册成功"
// @Failure     400  {object}  apiCommon.BaseResponse  "凭证数据无效"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Router      /api/v1/auth/biometric/register/finish [POST]
func (h *BiometricHandler) RegisterFinish(ctx *gin.Context) {
	h.log.Info(ctx, "RegisterFinish - 生物特征注册完成")

	var req apiBiometric.RegisterFinishRequest
	if !BindJSON(ctx, &req) {
		return
	}

	resp, xErr := h.service.biometricLogic.RegisterFinish(ctx, &req)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "注册成功", resp)
}

// LoginStart 生物特征登录开始
//
// @Summary     生物特征登录开始
// @Description 发起 WebAuthn 登录流程，返回获取选项供浏览器调用 navigator.credentials.get
// @Tags        生物特征接口
// @Produce     json
// @Success     200  {object}  apiCommon.BaseResponse{data=apiBiometric.LoginStartResponse}  "返回获取选项"
// @Router      /api/v1/auth/biometric/login/start [POST]
func (h *BiometricHandler) LoginStart(ctx *gin.Context) {
	h.log.Info(ctx, "LoginStart - 生物特征登录开始")

	resp, xErr := h.service.biometricLogic.LoginStart(ctx)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "登录开始", resp)
}

// LoginFinish 生物特征登录完成
//
// @Summary     生物特征登录完成
// @Description 提交浏览器验证的 PublicKeyCredential 完成登录，返回访问令牌
// @Tags        生物特征接口
// @Accept      json
// @Produce     json
// @Param       request  body      apiBiometric.LoginFinishRequest true  "登录完成请求"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiBiometric.LoginFinishResponse}  "登录成功，返回令牌"
// @Failure     401  {object}  apiCommon.BaseResponse  "生物特征验证失败"
// @Router      /api/v1/auth/biometric/login/finish [POST]
func (h *BiometricHandler) LoginFinish(ctx *gin.Context) {
	h.log.Info(ctx, "LoginFinish - 生物特征登录完成")

	var req apiBiometric.LoginFinishRequest
	if !BindJSON(ctx, &req) {
		return
	}

	resp, xErr := h.service.biometricLogic.LoginFinish(ctx, &req)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "登录成功", resp)
}
