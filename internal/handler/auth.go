package handler

import (
	apiAuth "github.com/xiaolfeng/Lumina/api/auth"
	apiCommon "github.com/xiaolfeng/Lumina/api/common"
	xResult "github.com/bamboo-services/bamboo-base-go/major/result"
	"github.com/gin-gonic/gin"
)

// 确保 apiCommon 包被编译器识别（swag 注释依赖此导入）
var _ = apiCommon.BaseResponse{}

// Initialize 系统初始化
//
// @Summary		系统初始化
// @Tags		auth
// @Accept		json
// @Produce		json
// @Param		request	body		apiAuth.InitializeRequest	true	"初始化请求"
// @Success		200		{object}	apiCommon.BaseResponse				"初始化成功"
// @Failure		400		{object}	apiCommon.BaseResponse				"参数错误"
// @Failure		409		{object}	apiCommon.BaseResponse				"系统已初始化"
// @Router		/api/v1/auth/initialize [post]
func (h *AuthHandler) Initialize(ctx *gin.Context) {
	h.log.Info(ctx, "Initialize - 系统初始化")

	var req apiAuth.InitializeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(err)
		return
	}

	xErr := h.service.authLogic.Initialize(ctx, &req)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.Success(ctx, "系统初始化成功")
}

// Login 用户登录
//
// @Summary		用户登录
// @Tags		auth
// @Accept		json
// @Produce		json
// @Param		request	body		apiAuth.LoginRequest	true	"登录请求"
// @Success		200		{object}	apiCommon.BaseResponse			"登录成功，返回Token"
// @Failure		400		{object}	apiCommon.BaseResponse			"参数错误"
// @Failure		401		{object}	apiCommon.BaseResponse			"账号或密码错误"
// @Router		/api/v1/auth/login [post]
func (h *AuthHandler) Login(ctx *gin.Context) {
	h.log.Info(ctx, "Login - 用户登录")

	var req apiAuth.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(err)
		return
	}

	resp, xErr := h.service.authLogic.Login(ctx, &req)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "登录成功", resp)
}

// Refresh 刷新令牌
//
// @Summary		刷新令牌
// @Tags		auth
// @Accept		json
// @Produce		json
// @Param		request	body		apiAuth.RefreshRequest	true	"刷新请求"
// @Success		200		{object}	apiCommon.BaseResponse			"刷新成功，返回新Token"
// @Failure		400		{object}	apiCommon.BaseResponse			"参数错误"
// @Failure		401		{object}	apiCommon.BaseResponse			"刷新令牌无效"
// @Router		/api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(ctx *gin.Context) {
	h.log.Info(ctx, "Refresh - 刷新令牌")

	var req apiAuth.RefreshRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(err)
		return
	}

	resp, xErr := h.service.authLogic.Refresh(ctx, &req)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "令牌刷新成功", resp)
}

// Logout 用户登出
//
// @Summary		用户登出
// @Tags		auth
// @Accept		json
// @Produce		json
// @Param		request	body		apiAuth.RefreshRequest	true	"登出请求（传入refresh_token）"
// @Success		200		{object}	apiCommon.BaseResponse			"登出成功"
// @Failure		401		{object}	apiCommon.BaseResponse			"未授权"
// @Security	ApiKeyAuth
// @Router		/api/v1/auth/logout [post]
func (h *AuthHandler) Logout(ctx *gin.Context) {
	h.log.Info(ctx, "Logout - 用户登出")

	var req apiAuth.RefreshRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(err)
		return
	}

	xErr := h.service.authLogic.Logout(ctx, req.RefreshToken)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.Success(ctx, "登出成功")
}

// Status 系统初始化状态
//
// @Summary		系统初始化状态
// @Tags		auth
// @Accept		json
// @Produce		json
// @Success		200		{object}	apiCommon.BaseResponse	"返回系统初始化状态（true=未初始化）"
// @Router		/api/v1/auth/status [get]
func (h *AuthHandler) Status(ctx *gin.Context) {
	h.log.Info(ctx, "Status - 查询系统状态")

	isInitial, xErr := h.service.authLogic.GetInitialStatus(ctx)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "查询成功", &apiAuth.StatusResponse{IsInitial: isInitial})
}
