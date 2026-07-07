package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	apiWebhook "github.com/xiaolfeng/Lumina/api/webhook"
)

// HandleRepoWikiWebhook receives Git Provider webhook push events.
// No Auth middleware — signature verification is in logic layer.
// Route registered BEFORE engine.Use() to bypass ResponseMiddleware.
//
// @Summary     [Webhook] Git Provider Webhook 接收端点
// @Description 接收 GitHub/Gitee/GitLab/Gitea 的 push 事件，验证签名后触发 Wiki 增量分析
// @Tags        Webhook接口
// @Accept      json
// @Produce     json
// @Param       token  path    string  true  "Webhook Token"
// @Success     200  {object}  apiWebhook.WebhookResponse  "处理结果"
// @Router      /api/v1/webhooks/repowiki/{token} [POST]
func (h *WebhookHandler) HandleRepoWikiWebhook(ctx *gin.Context) {
	// 1. Read raw body with 10MB limit
	body, err := io.ReadAll(io.LimitReader(ctx.Request.Body, 10*1024*1024))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"code": 400, "message": "读取请求体失败"})
		return
	}
	if len(body) == 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"code": 400, "message": "empty body"})
		return
	}

	// 2. Extract token from URL path
	token := ctx.Param("token")
	if token == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"code": 400, "message": "missing token"})
		return
	}

	// 3. Call logic — provider detection happens inside logic for full audit
	event, xErr := h.service.repoWikiLogic.HandleWebhookPush(ctx.Request.Context(), token, ctx.Request.Header, body)

	// 4. Build response — use ctx.JSON directly (no xResult, ResponseMiddleware bypassed)
	if xErr != nil {
		// Error case — event still logged in logic
		statusCode := http.StatusInternalServerError
		if event != nil {
			statusCode = event.ResponseCode
		}
		ctx.JSON(statusCode, apiWebhook.WebhookResponse{
			Status:  "failed",
			Message: xErr.Error(),
			Reason:  "processing_error",
		})
		return
	}

	// Success or skip — event has the final status
	if event != nil {
		ctx.JSON(event.ResponseCode, apiWebhook.WebhookResponse{
			Status:    event.Status,
			VersionID: event.VersionID.Int64(),
			Message:   getWebhookMessage(event.Status),
			Reason:    event.Reason,
		})
		return
	}

	// Fallback (should not reach here)
	ctx.JSON(http.StatusOK, apiWebhook.WebhookResponse{
		Status:  "ignored",
		Message: "事件已接收",
	})
}

func getWebhookMessage(status string) string {
	switch status {
	case "accepted":
		return "增量分析已触发"
	case "ignored":
		return "事件已接收但未触发分析"
	case "failed":
		return "处理失败"
	default:
		return "事件已接收"
	}
}
