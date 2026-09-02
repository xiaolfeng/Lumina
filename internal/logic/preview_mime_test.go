package logic

import (
	"testing"

	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

func TestInferMimeType(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"index.html":     bConst.PreviewMimeHTML,
		"theme.css":      bConst.PreviewMimeCSS,
		"app.js":         bConst.PreviewMimeJS,
		"app.mjs":        bConst.PreviewMimeJS,
		"data.json":      bConst.PreviewMimeJSON,
		"README.md":      bConst.PreviewMimeMarkdown,
		"notes.markdown": bConst.PreviewMimeMarkdown,
		"main.ts":        bConst.PreviewMimePlain,
		"icon.svg":       bConst.PreviewMimeSVG,
		"unknown.bin":    bConst.PreviewMimePlain,
	}

	for filename, want := range cases {
		if got := inferMimeType(filename); got != want {
			t.Errorf("inferMimeType(%q) = %q, want %q", filename, got, want)
		}
	}
}
