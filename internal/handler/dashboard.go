package handler

import (
	xResult "github.com/bamboo-services/bamboo-base-go/major/result"
	"github.com/gin-gonic/gin"
	apiCommon "github.com/xiaolfeng/Lumina/api/common"
	apiDashboard "github.com/xiaolfeng/Lumina/api/dashboard"
)

// 确保 apiCommon / apiDashboard 包被编译器识别（swag 注释依赖此导入）
var (
	_ = apiCommon.BaseResponse{}
	_ = apiDashboard.OverviewResponse{}
)

// GetDashboardOverview 获取看板概览统计
//
// @Summary     [管理] 获取看板概览统计
// @Description 聚合返回令牌、项目、问答、预览与 RepoWiki 的计数统计及最近预览列表
// @Tags        看板接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string  true  "Bearer Access Token"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiDashboard.OverviewResponse}  "获取成功"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Router      /api/v1/dashboard/overview [GET]
func (h *DashboardHandler) GetDashboardOverview(ctx *gin.Context) {
	h.log.Info(ctx, "GetDashboardOverview - 获取看板概览统计")

	resp, xErr := h.service.dashboardLogic.Overview(ctx)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "获取成功", resp)
}
