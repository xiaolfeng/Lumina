package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestIsSandboxStaticSubresource(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name     string
		filepath string
		dest     string
		accept   string
		query    string
		want     bool
	}{
		{name: "css style dest", filepath: "/style.css", dest: "style", want: true},
		{name: "js script dest", filepath: "/app.js", dest: "script", want: true},
		{name: "tsx subresource dest", filepath: "/demo.tsx", dest: "script", want: true},
		{name: "jsx subresource dest", filepath: "/demo.jsx", dest: "script", want: true},
		{name: "vue subresource dest", filepath: "/App.vue", dest: "script", want: true},
		{name: "svelte subresource dest", filepath: "/App.svelte", dest: "script", want: true},
		{name: "ts subresource dest", filepath: "/utils.ts", dest: "script", want: true},
		{name: "scss subresource dest", filepath: "/style.scss", dest: "style", want: true},
		{name: "wasm subresource dest", filepath: "/math.wasm", dest: "empty", want: true},
		{name: "gltf subresource dest", filepath: "/model.gltf", dest: "empty", want: true},
		{name: "mp4 subresource dest", filepath: "/video.mp4", dest: "video", want: true},
		{name: "yaml subresource dest", filepath: "/config.yaml", dest: "empty", want: true},
		{name: "lpw subresource dest", filepath: "/doc.lpw", dest: "empty", want: true},
		{name: "svg image dest", filepath: "/logo.svg", dest: "image", want: true},
		{name: "html iframe still gated", filepath: "/index.html", dest: "iframe", want: false},
		{name: "md iframe still gated", filepath: "/readme.md", dest: "iframe", want: false},
		{name: "top-level document html", filepath: "/index.html", dest: "document", want: false},
		{name: "top-level empty dest html accept", filepath: "/style.css", accept: "text/html", want: false},
		{name: "empty filepath", filepath: "", dest: "style", want: false},
		{name: "lumina_frame css", filepath: "/theme.css", dest: "", query: "lumina_frame=1", want: true},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			url := "/preview/abc" + tt.filepath
			if tt.query != "" {
				url += "?" + tt.query
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)
			if tt.dest != "" {
				req.Header.Set("Sec-Fetch-Dest", tt.dest)
			}
			if tt.accept != "" {
				req.Header.Set("Accept", tt.accept)
			}
			ctx.Request = req
			ctx.Params = gin.Params{{Key: "filepath", Value: tt.filepath}}
			if got := IsSandboxStaticSubresource(ctx); got != tt.want {
				t.Fatalf("IsSandboxStaticSubresource() = %v, want %v", got, tt.want)
			}
		})
	}
}
