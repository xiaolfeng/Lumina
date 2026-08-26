package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestBaseURLPrefersForwardedHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("host from request", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/plugins/marketplace.json", nil)
		ctx.Request.Host = "127.0.0.1:8080"
		if got := requestBaseURL(ctx); got != "http://127.0.0.1:8080" {
			t.Fatalf("requestBaseURL = %q, want http://127.0.0.1:8080", got)
		}
	})

	t.Run("forwarded proto and host", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/plugins/marketplace.json", nil)
		ctx.Request.Host = "localhost:8080"
		ctx.Request.Header.Set("X-Forwarded-Proto", "HTTPS, http")
		ctx.Request.Header.Set("X-Forwarded-Host", "lumina.example, localhost")
		if got := requestBaseURL(ctx); got != "https://lumina.example" {
			t.Fatalf("requestBaseURL = %q, want https://lumina.example", got)
		}
	})
}
