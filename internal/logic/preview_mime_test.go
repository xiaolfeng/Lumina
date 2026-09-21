package logic

import (
	"fmt"
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

	// 2. .lpw 文件合法 JSON 且结构合规 (1.1)
	if err := validateLpwContent("index.lpw", `{"version":"1.1","content":[]}`); err != nil {
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

func TestValidateLpwContent_RejectsV10(t *testing.T) {
	t.Parallel()
	doc := `{"version":"1.0","content":[]}`
	err := validateLpwContent("index.lpw", doc)
	if err == nil {
		t.Fatalf("expected v1.0 document to be rejected")
	}
	if !strings.Contains(err.Error(), "不支持的 LPW 版本") {
		t.Errorf("expected version error message, got: %v", err)
	}
}

func TestValidateLpwContent_RejectsBlocksRootField(t *testing.T) {
	t.Parallel()
	doc := `{"version":"1.1","blocks":[]}`
	err := validateLpwContent("index.lpw", doc)
	if err == nil {
		t.Fatalf("expected document with blocks root field to be rejected")
	}
	if !strings.Contains(err.Error(), "根字段 blocks") {
		t.Errorf("expected blocks error message, got: %v", err)
	}
}

func TestValidateLpwContentDeepValidation(t *testing.T) {
	t.Parallel()

	t.Run("未注册块类型被 Schema 拒绝", func(t *testing.T) {
		doc := `{"version":"1.1","content":[{"id":"b1","kind":"block","type":"not-a-type","props":{}}]}`
		if err := validateLpwContent("index.lpw", doc); err == nil {
			t.Errorf("expected unknown block type to be rejected")
		}
	})

	t.Run("叶子块携带 children 被拒绝", func(t *testing.T) {
		doc := `{"version":"1.1","content":[{"id":"b1","kind":"block","type":"markdown","props":{"content":"x"},"children":[]}]}`
		if err := validateLpwContent("index.lpw", doc); err == nil {
			t.Errorf("expected leaf with children to be rejected")
		}
	})

	t.Run("重复块 id 被拒绝", func(t *testing.T) {
		doc := `{"version":"1.1","content":[{"id":"dup","kind":"block","type":"markdown","props":{"content":"a"}},{"id":"dup","kind":"block","type":"markdown","props":{"content":"b"}}]}`
		if err := validateLpwContent("index.lpw", doc); err == nil {
			t.Errorf("expected duplicate ids to be rejected")
		}
	})

	t.Run("全文档总节点数超过 500 被拒绝", func(t *testing.T) {
		var content strings.Builder
		content.WriteByte('[')
		for i := 0; i < 251; i++ {
			if i > 0 {
				content.WriteByte(',')
			}
			fmt.Fprintf(&content, `{"id":"s-%d","kind":"container","type":"section","props":{"variant":"article","title":"t"},"children":[{"id":"c-%d-1","kind":"block","type":"markdown","props":{"content":"a"}},{"id":"c-%d-2","kind":"block","type":"markdown","props":{"content":"b"}}]}`, i, i, i)
		}
		content.WriteByte(']')
		doc := `{"version":"1.1","content":` + content.String() + `}`
		if err := validateLpwContent("index.lpw", doc); err == nil {
			t.Errorf("expected total node count > 500 to be rejected")
		}
	})

	t.Run("合法 layout 与 container 层级通过", func(t *testing.T) {
		doc := `{
			"version":"1.1",
			"content":[
				{
					"id":"lay-1",
					"kind":"layout",
					"type":"layout",
					"props":{"pattern":"split"},
					"children":[
						{
							"id":"sec-1",
							"kind":"container",
							"type":"section",
							"props":{"variant":"article","title":"架构"},
							"children":[
								{"id":"b1","kind":"block","type":"markdown","props":{"content":"正文"}}
							]
						},
						{
							"id":"b2",
							"kind":"block",
							"type":"markdown",
							"props":{"content":"右侧正文"}
						}
					]
				}
			]
		}`
		if err := validateLpwContent("index.lpw", doc); err != nil {
			t.Errorf("expected legal layout+container to pass, got: %v", err)
		}
	})
}

// TestValidateLpwContent_ProgressiveVsStrict 验证 Q-13：validateLpwContent 默认渐进式允许中间态合规结构，严格模式拒绝未达完成态下限
func TestValidateLpwContent_ProgressiveVsStrict(t *testing.T) {
	t.Parallel()
	// 渐进构建中间态：split 布局仅挂载了 1 个子节点（未达到完备态的最小 2 个节点要求）
	intermediateDoc := `{
		"version": "1.1",
		"content": [
			{
				"id": "lay-1",
				"kind": "layout",
				"type": "layout",
				"props": {"pattern": "split"},
				"children": [
					{"id": "b1", "kind": "block", "type": "markdown", "props": {"content": "左侧"}}
				]
			}
		]
	}`

	// 渐进模式（默认）允许合规的构建中间态，使得 JSON 损坏后可被修复
	if err := validateLpwContent("index.lpw", intermediateDoc); err != nil {
		t.Fatalf("expected intermediate doc to pass progressive validation, got: %v", err)
	}

	// 严格模式必须拒绝未达到完备态的文档
	if err := validateLpwContentWithOptions("index.lpw", intermediateDoc, false); err == nil {
		t.Fatalf("expected intermediate doc to fail strict validation, got nil")
	}
}
