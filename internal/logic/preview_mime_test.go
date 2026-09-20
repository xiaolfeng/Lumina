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
		"index.lpw":      bConst.PreviewMimeLPW,
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

func TestValidateLpwContent(t *testing.T) {
	t.Parallel()

	// 1. 非 .lpw 文件，即使坏 json 也忽略
	if err := validateLpwContent("index.html", "{ bad json"); err != nil {
		t.Errorf("expected non-lpw file to bypass json validation, got: %v", err)
	}

	// 2. .lpw 文件合法 JSON
	if err := validateLpwContent("index.lpw", `{"version":"1.0","blocks":[]}`); err != nil {
		t.Errorf("expected valid lpw to pass, got: %v", err)
	}

	// 3. .lpw 文件非法 JSON
	if err := validateLpwContent("index.lpw", `{`); err == nil {
		t.Errorf("expected invalid lpw to fail validation, got nil")
	}

	// 4. 大小写扩展名 .LPW
	if err := validateLpwContent("test.LPW", `not json`); err == nil {
		t.Errorf("expected uppercase .LPW with bad content to fail validation, got nil")
	}
}
