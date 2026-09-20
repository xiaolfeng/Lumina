package logic

import "testing"

func TestBuildPreviewURL(t *testing.T) {
	tests := []struct {
		name     string
		domain   string
		hash     string
		filename string
		want     string
	}{
		{
			name:   "会话首页",
			domain: "https://lumina.example.com/",
			hash:   "abc123",
			want:   "https://lumina.example.com/preview/abc123",
		},
		{
			name:     "文件深链",
			domain:   "https://lumina.example.com",
			hash:     "abc123",
			filename: "demo page.html",
			want:     "https://lumina.example.com/preview/abc123/demo%20page.html",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildPreviewURL(tt.domain, tt.hash, tt.filename); got != tt.want {
				t.Fatalf("buildPreviewURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildPagesURL(t *testing.T) {
	got := buildPagesURL("https://lumina.example.com/", "lumina", "design-system", "index.html")
	want := "https://lumina.example.com/pages/lumina/design-system/index.html"
	if got != want {
		t.Fatalf("buildPagesURL() = %q, want %q", got, want)
	}
}
