package handler

import (
	xResult "github.com/bamboo-services/bamboo-base-go/major/result"
	"github.com/gin-gonic/gin"
	apiCommon "github.com/xiaolfeng/Lumina/api/common"
	apiHealth "github.com/xiaolfeng/Lumina/api/health"
)

// 确保 apiCommon 包被编译器识别（swag 注释依赖此导入）
var _ = apiCommon.BaseResponse{}

// 确保 apiHealth 响应类型被 swag 识别（Ping 接口返回 PingResponse）
var _ = apiHealth.PingResponse{}

// Ping 健康检查
//
// @Summary     健康检查
// @Description 探测服务存活状态，返回应用版本、数据库与 Redis 就绪情况
// @Tags        健康检查接口
// @Accept      json
// @Produce     json
// @Success     200  {object}  apiCommon.BaseResponse{data=apiHealth.PingResponse}  "服务正常"
// @Failure     500  {object}  apiCommon.BaseResponse                              "依赖服务不可用"
// @Router      /api/v1/health/ping [GET]
func (h *HealthHandler) Ping(ctx *gin.Context) {
	h.log.Info(ctx, "Ping - 健康检查")

	status, xErr := h.service.healthLogic.Ping(ctx.Request.Context())
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "pong", status)
}
