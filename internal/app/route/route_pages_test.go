package route

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestPagesRouteStructure 确保 pages 路由组注册时不会因通配符冲突导致 panic，
// 并验证各端点能够正常分流匹配。
func TestPagesRouteStructure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	apiRouter := engine.Group("/api/v1")

	// 模拟 route_pages.go 中的路由定义结构
	public := apiRouter.Group("/pages/by-project")
	public.GET("/:project_name/:slug/auth-check", func(c *gin.Context) {
		c.String(http.StatusOK, "auth-check:"+c.Param("project_name")+":"+c.Param("slug"))
	})
	public.POST("/:project_name/:slug/unlock", func(c *gin.Context) {
		c.String(http.StatusOK, "unlock:"+c.Param("project_name")+":"+c.Param("slug"))
	})
	public.GET("/:project_name/:slug/meta", func(c *gin.Context) {
		c.String(http.StatusOK, "meta:"+c.Param("project_name")+":"+c.Param("slug"))
	})

	admin := apiRouter.Group("/pages")
	admin.GET("", func(c *gin.Context) {
		c.String(http.StatusOK, "list")
	})
	admin.GET("/:id", func(c *gin.Context) {
		c.String(http.StatusOK, "detail:"+c.Param("id"))
	})
	admin.GET("/:id/versions", func(c *gin.Context) {
		c.String(http.StatusOK, "versions:"+c.Param("id"))
	})
	admin.POST("/:id/versions/:version_id/switch-active", func(c *gin.Context) {
		c.String(http.StatusOK, "switch:"+c.Param("id")+":"+c.Param("version_id"))
	})
	admin.POST("/:id/fork", func(c *gin.Context) {
		c.String(http.StatusOK, "fork:"+c.Param("id"))
	})
	admin.POST("/:id/archive", func(c *gin.Context) {
		c.String(http.StatusOK, "archive:"+c.Param("id"))
	})
	admin.PUT("/:id/access-policy", func(c *gin.Context) {
		c.String(http.StatusOK, "policy:"+c.Param("id"))
	})

	tests := []struct {
		method string
		path   string
		expect string
	}{
		{"GET", "/api/v1/pages", "list"},
		{"GET", "/api/v1/pages/415994701439575040", "detail:415994701439575040"},
		{"GET", "/api/v1/pages/415994701439575040/versions", "versions:415994701439575040"},
		{"POST", "/api/v1/pages/415994701439575040/versions/v101/switch-active", "switch:415994701439575040:v101"},
		{"POST", "/api/v1/pages/415994701439575040/fork", "fork:415994701439575040"},
		{"POST", "/api/v1/pages/415994701439575040/archive", "archive:415994701439575040"},
		{"PUT", "/api/v1/pages/415994701439575040/access-policy", "policy:415994701439575040"},
		{"GET", "/api/v1/pages/by-project/lumina-web/overview/auth-check", "auth-check:lumina-web:overview"},
		{"POST", "/api/v1/pages/by-project/lumina-web/overview/unlock", "unlock:lumina-web:overview"},
		{"GET", "/api/v1/pages/by-project/lumina-web/overview/meta", "meta:lumina-web:overview"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("%s %s got status %d, want %d", tt.method, tt.path, w.Code, http.StatusOK)
		}
		if w.Body.String() != tt.expect {
			t.Errorf("%s %s got %q, want %q", tt.method, tt.path, w.Body.String(), tt.expect)
		}
	}
}
