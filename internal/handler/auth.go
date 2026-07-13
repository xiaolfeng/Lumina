package handler

import (
	xResult "github.com/bamboo-services/bamboo-base-go/major/result"
	"github.com/gin-gonic/gin"
	apiAuth "github.com/xiaolfeng/Lumina/api/auth"
	apiCommon "github.com/xiaolfeng/Lumina/api/common"
)

// 确保 apiCommon 包被编译器识别（swag 注释依赖此导入）
var _ = apiCommon.BaseResponse{}

// Initialize 系统初始化
//
// @Summary     [超管] 系统初始化
// @Description 首次部署时提交用户名、邮箱、密码，初始化系统管理员账户；已初始化时拒绝
// @Tags        认证接口
// @Accept      json
// @Produce     json
// @Param       request  body      apiAuth.InitializeRequest  true  "初始化请求"
// @Success     200      {object}  apiCommon.BaseResponse      "初始化成功"
// @Failure     400      {object}  apiCommon.BaseResponse      "请求参数错误"
// @Failure     409      {object}  apiCommon.BaseResponse      "系统已初始化"
// @Router      /api/v1/auth/initialize [POST]
func (h *AuthHandler) Initialize(ctx *gin.Context) {
	h.log.Info(ctx, "Initialize - 系统初始化")

	var req apiAuth.InitializeRequest
	if !BindJSON(ctx, &req) {
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
// @Summary     [管理] 用户登录
// @Description 校验账号（用户名或邮箱）与密码，签发访问令牌与刷新令牌
// @Tags        认证接口
// @Accept      json
// @Produce     json
// @Param       request  body      apiAuth.LoginRequest                    true  "登录请求"
// @Success     200      {object}  apiCommon.BaseResponse{data=apiAuth.TokenResponse}  "登录成功，返回令牌"
// @Failure     400      {object}  apiCommon.BaseResponse                  "请求参数错误"
// @Failure     401      {object}  apiCommon.BaseResponse                  "账号或密码错误"
// @Router      /api/v1/auth/login [POST]
func (h *AuthHandler) Login(ctx *gin.Context) {
	h.log.Info(ctx, "Login - 用户登录")

	var req apiAuth.LoginRequest
	if !BindJSON(ctx, &req) {
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
// @Summary     [管理] 刷新令牌
// @Description 使用刷新令牌换取新的访问令牌与刷新令牌
// @Tags        认证接口
// @Accept      json
// @Produce     json
// @Param       request  body      apiAuth.RefreshRequest                   true  "刷新请求"
// @Success     200      {object}  apiCommon.BaseResponse{data=apiAuth.TokenResponse}  "刷新成功，返回新令牌"
// @Failure     400      {object}  apiCommon.BaseResponse                  "请求参数错误"
// @Failure     401      {object}  apiCommon.BaseResponse                  "刷新令牌无效"
// @Router      /api/v1/auth/refresh [POST]
func (h *AuthHandler) Refresh(ctx *gin.Context) {
	h.log.Info(ctx, "Refresh - 刷新令牌")

	var req apiAuth.RefreshRequest
	if !BindJSON(ctx, &req) {
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
// @Summary     [管理] 用户登出
// @Description 根据刷新令牌撤销登录会话，需携带有效访问令牌
// @Tags        认证接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string                       true  "Bearer Access Token"
// @Param       request        body      apiAuth.RefreshRequest       true  "登出请求（传入 refresh_token）"
// @Success     200            {object}  apiCommon.BaseResponse       "登出成功"
// @Failure     401            {object}  apiCommon.BaseResponse       "未授权"
// @Router      /api/v1/auth/logout [POST]
func (h *AuthHandler) Logout(ctx *gin.Context) {
	h.log.Info(ctx, "Logout - 用户登出")

	var req apiAuth.RefreshRequest
	if !BindJSON(ctx, &req) {
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
// @Summary     系统初始化状态
// @Description 查询系统是否已完成初始化，前端据此决定展示初始化页或登录页
// @Tags        认证接口
// @Accept      json
// @Produce     json
// @Success     200  {object}  apiCommon.BaseResponse{data=apiAuth.StatusResponse}  "查询成功，返回初始化状态"
// @Router      /api/v1/auth/status [GET]
func (h *AuthHandler) Status(ctx *gin.Context) {
	h.log.Info(ctx, "Status - 查询系统状态")

	isInitial, xErr := h.service.authLogic.GetInitialStatus(ctx)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "查询成功", &apiAuth.StatusResponse{IsInitial: isInitial})
}
