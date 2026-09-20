package logic

import (
	"strings"
	"testing"
)

func TestPreviewLpwTreeOperations(t *testing.T) {
	t.Parallel()

	// 1. insertBlock 正常与异常
	t.Run("insertBlock", func(t *testing.T) {
		doc := &lpwDocument{
			Version: "1.0",
			Blocks: []lpwBlock{
				{ID: "b1", Type: "markdown", Props: map[string]any{"content": "hello"}},
			},
		}

		// 顶层末尾插入
		err := insertBlock(doc, "", nil, lpwBlock{ID: "b2", Type: "heading", Props: map[string]any{"content": "title"}})
		if err != nil {
			t.Fatalf("unexpected insert error: %v", err)
		}
		if len(doc.Blocks) != 2 || doc.Blocks[1].ID != "b2" {
			t.Errorf("expected b2 at index 1")
		}

		// 指定位置插入
		pos := 0
		err = insertBlock(doc, "", &pos, lpwBlock{ID: "b0", Type: "divider", Props: map[string]any{}})
		if err != nil {
			t.Fatalf("unexpected insert error: %v", err)
		}
		if doc.Blocks[0].ID != "b0" {
			t.Errorf("expected b0 at index 0")
		}

		// ID 冲突错误
		err = insertBlock(doc, "", nil, lpwBlock{ID: "b1", Type: "markdown", Props: map[string]any{}})
		if err == nil || !strings.Contains(err.Error(), "已存在") {
			t.Errorf("expected id duplicate error, got: %v", err)
		}

		// 目标父级不是容器
		err = insertBlock(doc, "b1", nil, lpwBlock{ID: "c1", Type: "markdown", Props: map[string]any{}})
		if err == nil || !strings.Contains(err.Error(), "不是容器组件") {
			t.Errorf("expected not-a-container error, got: %v", err)
		}

		// 目标父级不存在
		err = insertBlock(doc, "non-exist", nil, lpwBlock{ID: "c1", Type: "markdown", Props: map[string]any{}})
		if err == nil || !strings.Contains(err.Error(), "不存在") {
			t.Errorf("expected parent not found error, got: %v", err)
		}

		// Q-08：负数 position 显式报错，不静默降级
		neg := -1
		err = insertBlock(doc, "", &neg, lpwBlock{ID: "neg-1", Type: "divider", Props: map[string]any{}})
		if err == nil || !strings.Contains(err.Error(), "不能为负数") {
			t.Errorf("expected negative position error, got: %v", err)
		}
		if _, total := collectBlockIDs(doc); total != 3 {
			t.Errorf("negative position insert must not mutate document, got %d blocks", total)
		}
	})

	// 2. removeBlocks 正常与异常
	t.Run("removeBlocks", func(t *testing.T) {
		doc := &lpwDocument{
			Version: "1.0",
			Blocks: []lpwBlock{
				{ID: "b1", Type: "markdown"},
				{ID: "b2", Type: "heading"},
				{
					ID:   "sec-1",
					Type: "section",
					Children: []lpwBlock{
						{ID: "sub-1", Type: "callout"},
					},
				},
			},
		}

		// 删除子块和顶层块
		err := removeBlocks(doc, []string{"b1", "sub-1"})
		if err != nil {
			t.Fatalf("unexpected remove error: %v", err)
		}
		ids, total := collectBlockIDs(doc)
		if total != 2 {
			t.Errorf("expected 2 blocks left, got %d", total)
		}
		if _, exists := ids["b1"]; exists {
			t.Errorf("expected b1 to be deleted")
		}
		if _, exists := ids["sub-1"]; exists {
			t.Errorf("expected sub-1 to be deleted")
		}

		// 删除不存在的块报错
		err = removeBlocks(doc, []string{"b1"})
		if err == nil || !strings.Contains(err.Error(), "不存在") {
			t.Errorf("expected error when deleting non-existing block, got: %v", err)
		}
	})

	// 3. reorderSiblings 正常与异常
	t.Run("reorderSiblings", func(t *testing.T) {
		doc := &lpwDocument{
			Version: "1.0",
			Blocks: []lpwBlock{
				{ID: "b1", Type: "markdown"},
				{ID: "b2", Type: "heading"},
				{ID: "b3", Type: "callout"},
			},
		}

		err := reorderSiblings(doc, "", []string{"b3", "b1", "b2"})
		if err != nil {
			t.Fatalf("unexpected reorder error: %v", err)
		}
		if doc.Blocks[0].ID != "b3" || doc.Blocks[1].ID != "b1" || doc.Blocks[2].ID != "b2" {
			t.Errorf("reorder failed: %v", doc.Blocks)
		}

		// 非完整排列（缺项）
		err = reorderSiblings(doc, "", []string{"b3", "b1"})
		if err == nil || !strings.Contains(err.Error(), "完整子块排列") {
			t.Errorf("expected incomplete order error, got: %v", err)
		}
	})

	// 4. patchProps
	t.Run("patchProps", func(t *testing.T) {
		doc := &lpwDocument{
			Version: "1.0",
			Blocks: []lpwBlock{
				{ID: "b1", Type: "heading", Props: map[string]any{"level": 2, "content": "old"}},
			},
		}

		err := patchProps(doc, "b1", map[string]any{"content": "new", "level": nil})
		if err != nil {
			t.Fatalf("unexpected patch error: %v", err)
		}
		if doc.Blocks[0].Props["content"] != "new" {
			t.Errorf("expected content updated")
		}
		if _, exists := doc.Blocks[0].Props["level"]; exists {
			t.Errorf("expected level to be deleted when nil")
		}

		// 不存在
		err = patchProps(doc, "non-exist", map[string]any{})
		if err == nil || !strings.Contains(err.Error(), "不存在") {
			t.Errorf("expected block not found, got: %v", err)
		}
	})

	// 5. replaceBlock
	t.Run("replaceBlock", func(t *testing.T) {
		doc := &lpwDocument{
			Version: "1.0",
			Blocks: []lpwBlock{
				{ID: "b1", Type: "markdown"},
			},
		}

		err := replaceBlock(doc, "b1", lpwBlock{ID: "b1-new", Type: "heading", Props: map[string]any{"content": "ok"}})
		if err != nil {
			t.Fatalf("unexpected replace error: %v", err)
		}
		if doc.Blocks[0].ID != "b1-new" {
			t.Errorf("expected b1-new, got: %s", doc.Blocks[0].ID)
		}
	})
}

func TestValidateContainerRules(t *testing.T) {
	t.Parallel()

	// 1. tabs children != items
	t.Run("tabs children != items", func(t *testing.T) {
		doc := &lpwDocument{
			Version: "1.0",
			Blocks: []lpwBlock{
				{
					ID:   "t1",
					Type: "tabs",
					Props: map[string]any{
						"items": []any{
							map[string]any{"key": "a", "label": "A"},
							map[string]any{"key": "b", "label": "B"},
						},
					},
					Children: []lpwBlock{
						{ID: "c1", Type: "markdown"},
					},
				},
			},
		}
		err := validateContainerRules(doc)
		if err == nil || !strings.Contains(err.Error(), "必须等于 items 数量") {
			t.Errorf("expected tabs length mismatch error, got: %v", err)
		}
	})

	// 2. tabs defaultKey 未命中
	t.Run("tabs defaultKey not matched", func(t *testing.T) {
		doc := &lpwDocument{
			Version: "1.0",
			Blocks: []lpwBlock{
				{
					ID:   "t1",
					Type: "tabs",
					Props: map[string]any{
						"items": []any{
							map[string]any{"key": "a", "label": "A"},
						},
						"defaultKey": "bad-key",
					},
					Children: []lpwBlock{
						{ID: "c1", Type: "markdown"},
					},
				},
			},
		}
		err := validateContainerRules(doc)
		if err == nil || !strings.Contains(err.Error(), "未命中任何 items.key") {
			t.Errorf("expected defaultKey error, got: %v", err)
		}
	})

	// 3. columns children 与 ratio 不对齐
	t.Run("columns ratio mismatch", func(t *testing.T) {
		doc := &lpwDocument{
			Version: "1.0",
			Blocks: []lpwBlock{
				{
					ID:       "col1",
					Type:     "columns",
					Props:    map[string]any{"ratio": "1:1:1"},
					Children: []lpwBlock{{ID: "c1"}, {ID: "c2"}}, // 只有 2 列
				},
			},
		}
		err := validateContainerRules(doc)
		if err == nil || !strings.Contains(err.Error(), "期望 3 列子块") {
			t.Errorf("expected columns ratio mismatch error, got: %v", err)
		}
	})

	// 4. scorecard weight 之和 != 100
	t.Run("scorecard weight sum", func(t *testing.T) {
		doc := &lpwDocument{
			Version: "1.0",
			Blocks: []lpwBlock{
				{
					ID:   "sc1",
					Type: "scorecard",
					Props: map[string]any{
						"criteria": []any{
							map[string]any{"name": "C1", "weight": float64(40)},
							map[string]any{"name": "C2", "weight": float64(50)},
						},
						"plans": []any{
							map[string]any{"name": "P1", "scores": []any{float64(4), float64(5)}},
						},
					},
				},
			},
		}
		err := validateContainerRules(doc)
		if err == nil || !strings.Contains(err.Error(), "权重之和必须等于 100") {
			t.Errorf("expected scorecard weight error, got: %v", err)
		}
	})

	// 5. scorecard plan scores 长度 != criteria
	t.Run("scorecard scores length mismatch", func(t *testing.T) {
		doc := &lpwDocument{
			Version: "1.0",
			Blocks: []lpwBlock{
				{
					ID:   "sc1",
					Type: "scorecard",
					Props: map[string]any{
						"criteria": []any{
							map[string]any{"name": "C1", "weight": float64(50)},
							map[string]any{"name": "C2", "weight": float64(50)},
						},
						"plans": []any{
							map[string]any{"name": "P1", "scores": []any{float64(4)}}, // 只有 1 个得分
						},
					},
				},
			},
		}
		err := validateContainerRules(doc)
		if err == nil || !strings.Contains(err.Error(), "必须等于准则数量") {
			t.Errorf("expected scorecard scores length error, got: %v", err)
		}
	})

	// 6. comparison rows values 长度 != plans
	t.Run("comparison rows values length mismatch", func(t *testing.T) {
		doc := &lpwDocument{
			Version: "1.0",
			Blocks: []lpwBlock{
				{
					ID:   "cp1",
					Type: "comparison",
					Props: map[string]any{
						"plans": []any{map[string]any{"name": "A"}, map[string]any{"name": "B"}},
						"rows": []any{
							map[string]any{"dimension": "D1", "values": []any{map[string]any{"text": "T1"}}}, // 只有 1 个
						},
					},
				},
			},
		}
		err := validateContainerRules(doc)
		if err == nil || !strings.Contains(err.Error(), "不等于 plans 数量") {
			t.Errorf("expected comparison length mismatch error, got: %v", err)
		}
	})

	// 7. chart scatter 包含 categories
	t.Run("chart scatter has categories", func(t *testing.T) {
		doc := &lpwDocument{
			Version: "1.0",
			Blocks: []lpwBlock{
				{
					ID:   "ch1",
					Type: "chart",
					Props: map[string]any{
						"chartType":  "scatter",
						"categories": []any{"A", "B"},
						"series":     []any{},
					},
				},
			},
		}
		err := validateContainerRules(doc)
		if err == nil || !strings.Contains(err.Error(), "scatter 图表禁止指定 categories") {
			t.Errorf("expected scatter categories error, got: %v", err)
		}
	})

	// 8. cards 包含 javascript: 协议
	t.Run("cards javascript protocol", func(t *testing.T) {
		doc := &lpwDocument{
			Version: "1.0",
			Blocks: []lpwBlock{
				{
					ID:   "card1",
					Type: "cards",
					Props: map[string]any{
						"items": []any{
							map[string]any{"title": "恶意链接", "href": "javascript:alert(1)"},
						},
					},
				},
			},
		}
		err := validateContainerRules(doc)
		if err == nil || !strings.Contains(err.Error(), "不允许的协议") {
			t.Errorf("expected disallowed protocol error, got: %v", err)
		}
	})

	// 9. S-01 回归：scheme 内夹杂控制字符的伪协议串必须整体拒绝
	t.Run("cards scheme with control chars", func(t *testing.T) {
		doc := &lpwDocument{
			Version: "1.0",
			Blocks: []lpwBlock{
				{
					ID:   "card-ctrl",
					Type: "cards",
					Props: map[string]any{
						"items": []any{
							// "jav\tascript:notice"（惰性串，无攻击体）
							map[string]any{"title": "夹 TAB", "href": "jav\tascript:notice"},
							map[string]any{"title": "夹换行", "href": "jav\nascript:notice"},
						},
					},
				},
			},
		}
		err := validateContainerRules(doc)
		if err == nil || !strings.Contains(err.Error(), "控制字符") {
			t.Errorf("expected control character rejection, got: %v", err)
		}
	})

	// 10. S-01 回归：白名单放行合法形态，拒绝白名单外 scheme
	t.Run("cards url whitelist", func(t *testing.T) {
		makeCards := func(href string) *lpwDocument {
			return &lpwDocument{
				Version: "1.0",
				Blocks: []lpwBlock{
					{
						ID:   "card-wl",
						Type: "cards",
						Props: map[string]any{
							"items": []any{
								map[string]any{"title": "链接", "href": href},
							},
						},
					},
				},
			}
		}
		for _, okHref := range []string{
			"https://example.com/a",
			"http://example.com/b",
			"mailto:someone@example.com",
			"/pages/demo/landing",
			"./page.html",
			"../page.html",
			"#section",
			"detail.html",
		} {
			if err := validateContainerRules(makeCards(okHref)); err != nil {
				t.Errorf("whitelisted href %q should pass, got: %v", okHref, err)
			}
		}
		for _, badHref := range []string{
			"ftp://files.example.com/x",
			"data:text/html,notice",
			"vbscript:notice",
			"javascript:void",
		} {
			if err := validateContainerRules(makeCards(badHref)); err == nil {
				t.Errorf("non-whitelisted href %q should be rejected", badHref)
			}
		}
	})

	// 11. S-01 回归：image 与 gallery 的 src 同样走白名单
	t.Run("image and gallery src whitelist", func(t *testing.T) {
		badImage := &lpwDocument{
			Version: "1.0",
			Blocks: []lpwBlock{
				{ID: "img1", Type: "image", Props: map[string]any{"src": "jav\tascript:notice", "alt": "x"}},
			},
		}
		if err := validateContainerRules(badImage); err == nil || !strings.Contains(err.Error(), "控制字符") {
			t.Errorf("expected image control-char rejection, got: %v", err)
		}

		badGallery := &lpwDocument{
			Version: "1.0",
			Blocks: []lpwBlock{
				{
					ID:   "gal1",
					Type: "gallery",
					Props: map[string]any{
						"images": []any{
							map[string]any{"src": "ftp://x/y.png", "alt": "x"},
						},
					},
				},
			},
		}
		if err := validateContainerRules(badGallery); err == nil {
			t.Errorf("expected gallery non-whitelisted scheme rejection")
		}
	})
}
