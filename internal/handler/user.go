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
