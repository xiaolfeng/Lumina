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

func TestRenderTSXVirtualHost(t *testing.T) {
	t.Parallel()
	tsx := `import React, { useState } from 'react'; export default function App() { return <h1>Hello Lumina</h1> }`
	got := renderTSXVirtualHost("App.tsx", tsx)
	if !strings.Contains(got, "<title>App.tsx · Lumina Preview</title>") {
		t.Fatal("missing title in virtual host")
	}
	if !strings.Contains(got, "babel.min.js") {
		t.Fatal("missing babel standalone in virtual host")
	}
	if !strings.Contains(got, "ReactDOM.createRoot") {
		t.Fatal("missing ReactDOM createRoot in virtual host")
	}
	if !strings.Contains(got, "modules: 'commonjs'") {
		t.Fatal("missing commonjs module transformation in babel preset")
	}
	if !strings.Contains(got, "function require(name)") {
		t.Fatal("missing runtime mock require implementation")
	}
	if strings.Contains(got, "react@19/umd") || strings.Contains(got, "react-dom@19/umd") {
		t.Fatal("react@19 does not have UMD builds, must use react@18 UMD")
	}
	if !strings.Contains(got, "react@18") || !strings.Contains(got, "react-dom@18") {
		t.Fatal("missing react 18 UMD script in virtual host")
	}
	if !strings.Contains(got, "window.React") {
		t.Fatal("missing runtime defensive guard for window.React")
	}
}

func TestWriteServedFile_CORS(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/preview/abc/demo.tsx", nil)
	req.Header.Set("Origin", "null")
	ctx.Request = req

	writeServedFile(ctx, "demo.tsx", "text/typescript-jsx", "export default function App() {}")

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" && got != "null" {
		t.Fatalf("writeServedFile missing Access-Control-Allow-Origin, got %q", got)
	}
}

func TestRenderTSXVirtualHost_SecurityEscaping(t *testing.T) {
	t.Parallel()
	maliciousFilename := `demo</title><script>alert(1)</script>.tsx`
	maliciousContent := `export default function App() { return <div></Script><script>alert(2)</script></div> }`
	got := renderTSXVirtualHost(maliciousFilename, maliciousContent)

	if strings.Contains(got, "<title>demo</title>") {
		t.Fatal("filename XSS was not escaped")
	}
	if !strings.Contains(got, "&lt;/title&gt;&lt;script&gt;") {
		t.Fatal("filename did not have HTML entity escaping")
	}
	// Content must not have raw unescaped closing script tag
	if strings.Contains(got, "</Script>") || strings.Contains(got, "</script>") {
		// Only the script closing tag of the container is allowed, inside the container must be safe
		sourceTagStart := strings.Index(got, `id="__lumina_tsx_source__"`)
		if sourceTagStart != -1 {
			sourceTagEnd := strings.Index(got[sourceTagStart:], `</script>`)
			if sourceTagEnd != -1 {
				inner := got[sourceTagStart : sourceTagStart+sourceTagEnd]
				if strings.Contains(strings.ToLower(inner), "</script") {
					t.Fatal("raw closing script tag found inside tsx source container")
				}
			}
		}
	}
}
