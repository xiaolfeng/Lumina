package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

func newWebAuthnTestCtx(origin, referer, host, forwardedProto string) *gin.Context {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/", nil)
	c.Request.Host = host
	if origin != "" {
		c.Request.Header.Set("Origin", origin)
	}
	if referer != "" {
		c.Request.Header.Set("Referer", referer)
	}
	if forwardedProto != "" {
		c.Request.Header.Set("X-Forwarded-Proto", forwardedProto)
	}
	return c
}

func TestResolveWebAuthnOrigin(t *testing.T) {
	tests := []struct {
		name           string
		origin         string
		referer        string
		host           string
		forwardedProto string
		want           string
	}{
		{"Origin 头优先", "https://lumina.example.com", "", "", "", "https://lumina.example.com"},
		{"Origin 非法回退 Referer", "not-a-url", "https://lumina.example.com/path", "", "", "https://lumina.example.com"},
		{"Referer 兜底", "", "http://localhost:3000/login", "", "", "http://localhost:3000"},
		{"Host + X-Forwarded-Proto 推导 https", "", "", "lumina.example.com", "https", "https://lumina.example.com"},
		{"Host 无代理走 http", "", "", "localhost:8080", "", "http://localhost:8080"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newWebAuthnTestCtx(tt.origin, tt.referer, tt.host, tt.forwardedProto)
			if got := resolveWebAuthnOrigin(c); got != tt.want {
				t.Fatalf("resolveWebAuthnOrigin() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWebAuthnOriginMiddlewareInjectsContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c := newWebAuthnTestCtx("https://lumina.example.com", "", "", "")

	WebAuthnOrigin()(c)

	got, ok := c.Request.Context().Value(bConst.WebAuthnOriginContextKey).(string)
	if !ok || got != "https://lumina.example.com" {
		t.Fatalf("middleware 未正确注入 Origin: got %q, ok=%v", got, ok)
	}
}
