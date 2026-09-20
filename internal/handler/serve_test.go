package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestIsDocumentRequest(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name   string
		dest   string
		accept string
		query  string
		want   bool
	}{
		{name: "top-level document", dest: "document", want: true},
		{name: "iframe", dest: "iframe", want: false},
		{name: "style", dest: "style", want: false},
		{name: "script", dest: "script", want: false},
		{name: "empty dest html accept", accept: "text/html,application/xhtml+xml", want: true},
		{name: "empty dest css accept", accept: "text/css,*/*", want: false},
		{name: "empty dest and accept", want: false},
		{name: "frame param html accept no dest", accept: "text/html", query: "lumina_frame=1", want: false},
		{name: "frame param html accept with dest", dest: "document", accept: "text/html", query: "lumina_frame=1", want: false},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			req := httptest.NewRequest(http.MethodGet, "/preview/abc/style.css", nil)
			if tt.query != "" {
				req.URL.RawQuery = tt.query
			}
			if tt.dest != "" {
				req.Header.Set("Sec-Fetch-Dest", tt.dest)
			}
			if tt.accept != "" {
				req.Header.Set("Accept", tt.accept)
			}
			ctx.Request = req
			if got := IsDocumentRequest(ctx); got != tt.want {
				t.Fatalf("IsDocumentRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInjectNavigateShim(t *testing.T) {
	t.Parallel()
	html := "<html><body><a href=\"about.html\">go</a></body></html>"
	got := injectNavigateShim(html)
	if !strings.Contains(got, `data-lumina-nav="1"`) {
		t.Fatal("shim not injected before </body>")
	}
	if !strings.Contains(got, "lumina:navigate") {
		t.Fatal("shim missing navigate postMessage")
	}
}
