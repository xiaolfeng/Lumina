package handler

import (
	xResult "github.com/bamboo-services/bamboo-base-go/major/result"
	"github.com/gin-gonic/gin"
)

func (h *HealthHandler) Ping(ctx *gin.Context) {
	h.log.Info(ctx, "Ping - 健康检查")

	status, xErr := h.service.healthLogic.Ping(ctx)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "pong", status)
}
