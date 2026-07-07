package route

import (
	"github.com/gin-gonic/gin"

	"github.com/xiaolfeng/Lumina/internal/handler"
)

// webhookRouter registers the webhook receiver endpoint BEFORE engine.Use().
// This bypasses ResponseMiddleware so the handler can read raw body for HMAC verification.
// No Auth middleware — signature verification is in the logic layer.
func (r *route) webhookRouter(engine *gin.Engine) {
	h := handler.NewHandler[handler.WebhookHandler](r.context, "WebhookHandler")
	engine.POST("/api/v1/webhooks/repowiki/:token", h.HandleRepoWikiWebhook)
}
