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
// @Summary		获取当前用户信息
// @Tags		user
// @Accept		json
// @Produce		json
// @Success		200		{object}	apiCommon.BaseResponse	"获取成功，返回用户信息"
// @Failure		401		{object}	apiCommon.BaseResponse	"未授权"
// @Security	ApiKeyAuth
// @Router		/api/v1/user/current [get]
func (h *UserHandler) Current(ctx *gin.Context) {
	h.log.Info(ctx, "Current - 获取当前用户信息")

	resp, xErr := h.service.authLogic.GetCurrentUser(ctx)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "获取成功", resp)
}
