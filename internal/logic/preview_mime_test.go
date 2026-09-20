package logic

import (
	"fmt"
	"strconv"
	"strings"
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

	// 2. .lpw 文件合法 JSON 且结构合规
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

func TestValidateLpwContentDeepValidation(t *testing.T) {
	t.Parallel()

	// Q-03 回归：语法合法但结构非法的文档必须拒绝（上传失败不得覆盖已有文件）
	t.Run("未注册块类型被 Schema 拒绝", func(t *testing.T) {
		doc := `{"version":"1.0","blocks":[{"id":"b1","type":"not-a-type","props":{}}]}`
		if err := validateLpwContent("index.lpw", doc); err == nil {
			t.Errorf("expected unknown block type to be rejected")
		}
	})

	t.Run("叶子块携带 children 被拒绝", func(t *testing.T) {
		doc := `{"version":"1.0","blocks":[{"id":"b1","type":"markdown","props":{"content":"x"},"children":[]}]}`
		if err := validateLpwContent("index.lpw", doc); err == nil {
			t.Errorf("expected leaf with children to be rejected")
		}
	})

	// Q-03 回归：深嵌套（6 层容器）必须被结构规则走查拒绝
	t.Run("深嵌套被结构规则拒绝", func(t *testing.T) {
		inner := `{"id":"leaf","type":"markdown","props":{"content":"x"}}`
		doc := `{"version":"1.0","blocks":[` + wrapContainers(inner, 6) + `]}`
		if err := validateLpwContent("index.lpw", doc); err == nil {
			t.Errorf("expected deep nesting to be rejected by structure rules")
		}
	})

	// Q-07 回归：重复 id 必须被走查拒绝（Schema 不表达唯一性）
	t.Run("重复块 id 被拒绝", func(t *testing.T) {
		doc := `{"version":"1.0","blocks":[{"id":"dup","type":"markdown","props":{"content":"a"}},{"id":"dup","type":"markdown","props":{"content":"b"}}]}`
		if err := validateLpwContent("index.lpw", doc); err == nil {
			t.Errorf("expected duplicate ids to be rejected")
		}
	})

	// 规则 8 回归：顶层不超 500 但含子孙总数超 500 必须被拒绝
	t.Run("全文档总块数超过 500 被拒绝", func(t *testing.T) {
		var blocks strings.Builder
		blocks.WriteByte('[')
		for i := 0; i < 250; i++ {
			if i > 0 {
				blocks.WriteByte(',')
			}
			fmt.Fprintf(&blocks, `{"id":"s-%d","type":"section","props":{"title":"t"},"children":[{"id":"c-%d-1","type":"markdown","props":{"content":"a"}},{"id":"c-%d-2","type":"markdown","props":{"content":"b"}},{"id":"c-%d-3","type":"markdown","props":{"content":"c"}}]}`, i, i, i, i)
		}
		blocks.WriteByte(']')
		doc := `{"version":"1.0","blocks":` + blocks.String() + `}`
		if err := validateLpwContent("index.lpw", doc); err == nil {
			t.Errorf("expected total block count > 500 to be rejected")
		}
	})

	// 合法 3 层容器 + 叶子（块深度 4）必须通过——与前端渲染契约一致
	t.Run("合法 3 层容器嵌套通过", func(t *testing.T) {
		doc := `{"version":"1.0","blocks":[{"id":"s1","type":"section","props":{"title":"a"},"children":[{"id":"t1","type":"tabs","props":{"items":[{"key":"k","label":"K"}]},"children":[{"id":"c1","type":"columns","props":{"ratio":"1:1"},"children":[{"id":"leaf","type":"markdown","props":{"content":"x"}},{"id":"leaf2","type":"markdown","props":{"content":"y"}}]}]}]}]}`
		if err := validateLpwContent("index.lpw", doc); err != nil {
			t.Errorf("expected legal 3-container nesting to pass, got: %v", err)
		}
	})
}

// wrapContainers 把 innerJSON 包裹 n 层 section 容器
func wrapContainers(innerJSON string, n int) string {
	out := innerJSON
	for i := 0; i < n; i++ {
		out = `{"id":"w-` + strconv.Itoa(i) + `","type":"section","props":{"title":"t"},"children":[` + out + `]}`
	}
	return out
}
