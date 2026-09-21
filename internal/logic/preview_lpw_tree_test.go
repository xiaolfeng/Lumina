package logic

import (
	"fmt"
	"strings"
	"testing"
)

// ── 1. 合法层级用例 ──────────────────────────────────────────

func TestHierarchy_RootLayout_Container_Block(t *testing.T) {
	t.Parallel()
	doc := &lpwDocument{
		Version: "1.1",
		Content: []lpwNode{
			{
				ID:   "lay-1",
				Kind: "layout",
				Type: "layout",
				Props: map[string]any{
					"pattern": "split",
				},
				Children: []lpwNode{
					{
						ID:   "sec-1",
						Kind: "container",
						Type: "section",
						Props: map[string]any{
							"variant": "evidence",
							"title":   "技术证据",
						},
						Children: []lpwNode{
							{
								ID:    "diff-1",
								Kind:  "block",
								Type:  "diff",
								Props: map[string]any{"oldCode": "a", "newCode": "b"},
							},
						},
					},
					{
						ID:    "blk-2",
						Kind:  "block",
						Type:  "markdown",
						Props: map[string]any{"content": "右侧说明"},
					},
				},
			},
		},
	}
	if err := validateDocument11(doc); err != nil {
		t.Fatalf("expected legal root layout->container->block to pass, got: %v", err)
	}
}

func TestHierarchy_RootLayout_DirectBlock(t *testing.T) {
	t.Parallel()
	doc := &lpwDocument{
		Version: "1.1",
		Content: []lpwNode{
			{
				ID:   "lay-1",
				Kind: "layout",
				Type: "layout",
				Props: map[string]any{
					"pattern": "split",
				},
				Children: []lpwNode{
					{
						ID:    "img-1",
						Kind:  "block",
						Type:  "image",
						Props: map[string]any{"src": "https://example.com/a.png"},
					},
					{
						ID:    "md-1",
						Kind:  "block",
						Type:  "markdown",
						Props: map[string]any{"content": "说明文本"},
					},
				},
			},
		},
	}
	if err := validateDocument11(doc); err != nil {
		t.Fatalf("expected legal layout direct blocks to pass, got: %v", err)
	}
}

func TestHierarchy_RootContainer_Block(t *testing.T) {
	t.Parallel()
	doc := &lpwDocument{
		Version: "1.1",
		Content: []lpwNode{
			{
				ID:   "pan-1",
				Kind: "container",
				Type: "panel",
				Props: map[string]any{
					"variant": "aside",
				},
				Children: []lpwNode{
					{
						ID:    "co-1",
						Kind:  "block",
						Type:  "callout",
						Props: map[string]any{"content": "侧栏提示"},
					},
				},
			},
		},
	}
	if err := validateDocument11(doc); err != nil {
		t.Fatalf("expected root container->block to pass, got: %v", err)
	}
}

func TestHierarchy_RootBlock(t *testing.T) {
	t.Parallel()
	doc := &lpwDocument{
		Version: "1.1",
		Content: []lpwNode{
			{
				ID:    "md-1",
				Kind:  "block",
				Type:  "markdown",
				Props: map[string]any{"content": "单篇长文"},
			},
		},
	}
	if err := validateDocument11(doc); err != nil {
		t.Fatalf("expected root block to pass, got: %v", err)
	}
}

func TestHierarchy_EditorialWrap_ImagePlusMarkdown(t *testing.T) {
	t.Parallel()
	doc := &lpwDocument{
		Version: "1.1",
		Content: []lpwNode{
			{
				ID:   "lay-wrap",
				Kind: "layout",
				Type: "layout",
				Props: map[string]any{
					"pattern": "editorial-wrap",
				},
				Children: []lpwNode{
					{
						ID:    "img-1",
						Kind:  "block",
						Type:  "image",
						Props: map[string]any{"src": "https://example.com/cover.png"},
					},
					{
						ID:    "md-1",
						Kind:  "block",
						Type:  "markdown",
						Props: map[string]any{"content": "图文环绕正文"},
					},
				},
			},
		},
	}
	if err := validateDocument11(doc); err != nil {
		t.Fatalf("expected editorial-wrap image+markdown to pass, got: %v", err)
	}
}

func TestHierarchy_Newspaper_BodyMarkdownFullChart(t *testing.T) {
	t.Parallel()
	doc := &lpwDocument{
		Version: "1.1",
		Content: []lpwNode{
			{
				ID:   "lay-news",
				Kind: "layout",
				Type: "layout",
				Props: map[string]any{
					"pattern": "newspaper",
					"placements": []any{
						map[string]any{"nodeId": "md-body", "role": "body"},
						map[string]any{"nodeId": "ch-full", "role": "full"},
					},
				},
				Children: []lpwNode{
					{
						ID:    "md-body",
						Kind:  "block",
						Type:  "markdown",
						Props: map[string]any{"content": "报纸正文多栏流"},
					},
					{
						ID:    "ch-full",
						Kind:  "block",
						Type:  "chart",
						Props: map[string]any{"chartType": "line", "series": []any{}},
					},
				},
			},
		},
	}
	if err := validateDocument11(doc); err != nil {
		t.Fatalf("expected newspaper body markdown + full chart to pass, got: %v", err)
	}
}

func TestContract_TabsItemsMatchChildren(t *testing.T) {
	t.Parallel()
	doc := &lpwDocument{
		Version: "1.1",
		Content: []lpwNode{
			{
				ID:   "tb-1",
				Kind: "container",
				Type: "tabs",
				Props: map[string]any{
					"variant": "reference",
					"items": []any{
						map[string]any{"key": "k1", "label": "L1"},
						map[string]any{"key": "k2", "label": "L2"},
						map[string]any{"key": "k3", "label": "L3"},
					},
					"defaultKey": "k2",
				},
				Children: []lpwNode{
					{ID: "b1", Kind: "block", Type: "markdown", Props: map[string]any{"content": "1"}},
					{ID: "b2", Kind: "block", Type: "code", Props: map[string]any{"content": "2"}},
					{ID: "b3", Kind: "block", Type: "table", Props: map[string]any{}},
				},
			},
		},
	}
	if err := validateDocument11(doc); err != nil {
		t.Fatalf("expected tabs items == children to pass, got: %v", err)
	}
}

// ── 2. 非法层级用例（必须断言错误含路径或关键词） ─────────────

func TestHierarchy_RejectsLayoutInsideLayout(t *testing.T) {
	t.Parallel()
	doc := &lpwDocument{
		Version: "1.1",
		Content: []lpwNode{
			{
				ID:    "lay-1",
				Kind:  "layout",
				Type:  "layout",
				Props: map[string]any{"pattern": "split"},
				Children: []lpwNode{
					{
						ID:    "lay-nested",
						Kind:  "layout",
						Type:  "layout",
						Props: map[string]any{"pattern": "grid"},
						Children: []lpwNode{
							{ID: "b1", Kind: "block", Type: "markdown", Props: map[string]any{"content": "x"}},
						},
					},
					{ID: "b2", Kind: "block", Type: "markdown", Props: map[string]any{"content": "y"}},
				},
			},
		},
	}
	err := validateDocument11(doc)
	if err == nil {
		t.Fatalf("expected nested layout to be rejected")
	}
	if !strings.Contains(err.Error(), "layout") || !strings.Contains(err.Error(), "不允许包含") {
		t.Errorf("expected error message explaining layout cannot contain layout, got: %v", err)
	}
}

func TestHierarchy_RejectsContainerInsideContainer(t *testing.T) {
	t.Parallel()
	doc := &lpwDocument{
		Version: "1.1",
		Content: []lpwNode{
			{
				ID:    "sec-1",
				Kind:  "container",
				Type:  "section",
				Props: map[string]any{"variant": "article", "title": "t1"},
				Children: []lpwNode{
					{
						ID:    "sec-2",
						Kind:  "container",
						Type:  "section",
						Props: map[string]any{"variant": "article", "title": "t2"},
						Children: []lpwNode{
							{ID: "b1", Kind: "block", Type: "markdown", Props: map[string]any{"content": "x"}},
						},
					},
				},
			},
		},
	}
	err := validateDocument11(doc)
	if err == nil {
		t.Fatalf("expected nested container to be rejected")
	}
	if !strings.Contains(err.Error(), "container") || !strings.Contains(err.Error(), "只能包含 block") {
		t.Errorf("expected error message explaining container cannot contain container, got: %v", err)
	}
}

func TestHierarchy_RejectsContainerInsideLayout_WrongVariant(t *testing.T) {
	t.Parallel()
	doc := &lpwDocument{
		Version: "1.1",
		Content: []lpwNode{
			{
				ID:    "lay-1",
				Kind:  "layout",
				Type:  "layout",
				Props: map[string]any{"pattern": "split"},
				Children: []lpwNode{
					{
						ID:    "pan-1",
						Kind:  "container",
						Type:  "panel",
						Props: map[string]any{"variant": "summary"},
						Children: []lpwNode{
							// panel/summary 只能放 decision / data，不允许 code
							{ID: "c1", Kind: "block", Type: "code", Props: map[string]any{"content": "x"}},
						},
					},
					{ID: "b2", Kind: "block", Type: "markdown", Props: map[string]any{"content": "y"}},
				},
			},
		},
	}
	err := validateDocument11(doc)
	if err == nil {
		t.Fatalf("expected panel/summary with code block to be rejected")
	}
	if !strings.Contains(err.Error(), "不允许子块类型 code") {
		t.Errorf("expected variant contract rejection, got: %v", err)
	}
}

func TestHierarchy_RejectsBlockWithChildren(t *testing.T) {
	t.Parallel()
	doc := &lpwDocument{
		Version: "1.1",
		Content: []lpwNode{
			{
				ID:    "b1",
				Kind:  "block",
				Type:  "markdown",
				Props: map[string]any{"content": "x"},
				Children: []lpwNode{
					{ID: "b2", Kind: "block", Type: "markdown", Props: map[string]any{"content": "y"}},
				},
			},
		},
	}
	err := validateDocument11(doc)
	if err == nil {
		t.Fatalf("expected block with children to be rejected")
	}
	if !strings.Contains(err.Error(), "block 不允许 children") {
		t.Errorf("expected 'block 不允许 children' error, got: %v", err)
	}
}

func TestHierarchy_RejectsBlockAsParent(t *testing.T) {
	t.Parallel()
	doc := &lpwDocument{
		Version: "1.1",
		Content: []lpwNode{
			{ID: "b1", Kind: "block", Type: "markdown", Props: map[string]any{"content": "parent"}},
		},
	}
	err := insertNode(doc, "b1", nil, lpwNode{ID: "b2", Kind: "block", Type: "markdown", Props: map[string]any{}})
	if err == nil {
		t.Fatalf("expected insert into block to be rejected")
	}
	if !strings.Contains(err.Error(), "不能作为 parent") {
		t.Errorf("expected cannot be parent error, got: %v", err)
	}
}

func TestHierarchy_RejectsEditorialWrapWithContainer(t *testing.T) {
	t.Parallel()
	doc := &lpwDocument{
		Version: "1.1",
		Content: []lpwNode{
			{
				ID:    "lay-wrap",
				Kind:  "layout",
				Type:  "layout",
				Props: map[string]any{"pattern": "editorial-wrap"},
				Children: []lpwNode{
					{
						ID:    "sec-1",
						Kind:  "container",
						Type:  "section",
						Props: map[string]any{"variant": "article", "title": "t"},
						Children: []lpwNode{
							{ID: "m1", Kind: "block", Type: "markdown", Props: map[string]any{"content": "c"}},
						},
					},
					{ID: "b2", Kind: "block", Type: "markdown", Props: map[string]any{"content": "y"}},
				},
			},
		},
	}
	err := validateDocument11(doc)
	if err == nil {
		t.Fatalf("expected editorial-wrap with container to be rejected")
	}
	if !strings.Contains(err.Error(), "editorial-wrap") {
		t.Errorf("expected editorial-wrap error, got: %v", err)
	}
}

func TestHierarchy_RejectsNewspaperBodyChart(t *testing.T) {
	t.Parallel()
	doc := &lpwDocument{
		Version: "1.1",
		Content: []lpwNode{
			{
				ID:   "lay-news",
				Kind: "layout",
				Type: "layout",
				Props: map[string]any{
					"pattern": "newspaper",
					"placements": []any{
						map[string]any{"nodeId": "ch-1", "role": "body"},
						map[string]any{"nodeId": "md-2", "role": "full"},
					},
				},
				Children: []lpwNode{
					{ID: "ch-1", Kind: "block", Type: "chart", Props: map[string]any{}},
					{ID: "md-2", Kind: "block", Type: "markdown", Props: map[string]any{"content": "m"}},
				},
			},
		},
	}
	err := validateDocument11(doc)
	if err == nil {
		t.Fatalf("expected newspaper body chart to be rejected")
	}
	if !strings.Contains(err.Error(), "role=body 的节点必须是 markdown 块") {
		t.Errorf("expected newspaper body markdown error, got: %v", err)
	}
}

func TestContract_NewspaperRequiresRoleBody(t *testing.T) {
	t.Parallel()

	// 1. 省略 placements 时，必须拦截
	docWithoutPlacements := &lpwDocument{
		Version: "1.1",
		Content: []lpwNode{
			{
				ID:   "lay-news",
				Kind: "layout",
				Type: "layout",
				Props: map[string]any{
					"pattern": "newspaper",
				},
				Children: []lpwNode{
					{ID: "md-1", Kind: "block", Type: "markdown", Props: map[string]any{"content": "正文 1"}},
					{ID: "md-2", Kind: "block", Type: "markdown", Props: map[string]any{"content": "正文 2"}},
				},
			},
		},
	}
	err1 := validateDocument11(docWithoutPlacements)
	if err1 == nil {
		t.Fatalf("expected newspaper without placements to fail role=body check")
	}
	if !strings.Contains(err1.Error(), "必须有且仅有一个 role=body 的 markdown 块") {
		t.Errorf("expected role=body error, got: %v", err1)
	}

	// 2. 提供了 placements 但缺少 role=body 时，必须拦截
	docWithoutBody := &lpwDocument{
		Version: "1.1",
		Content: []lpwNode{
			{
				ID:   "lay-news",
				Kind: "layout",
				Type: "layout",
				Props: map[string]any{
					"pattern": "newspaper",
					"placements": []any{
						map[string]any{"nodeId": "md-1", "role": "full"},
						map[string]any{"nodeId": "md-2", "role": "full"},
					},
				},
				Children: []lpwNode{
					{ID: "md-1", Kind: "block", Type: "markdown", Props: map[string]any{"content": "正文 1"}},
					{ID: "md-2", Kind: "block", Type: "markdown", Props: map[string]any{"content": "正文 2"}},
				},
			},
		},
	}
	err2 := validateDocument11(docWithoutBody)
	if err2 == nil {
		t.Fatalf("expected newspaper without role=body to fail")
	}
	if !strings.Contains(err2.Error(), "必须有且仅有一个 role=body 的 markdown 块") {
		t.Errorf("expected role=body error, got: %v", err2)
	}
}

func TestHierarchy_RejectsAnnotationOnLayout(t *testing.T) {
	t.Parallel()
	doc := &lpwDocument{
		Version: "1.1",
		Content: []lpwNode{
			{
				ID:         "lay-1",
				Kind:       "layout",
				Type:       "layout",
				Props:      map[string]any{"pattern": "split"},
				Annotation: &lpwAnnotation{Kind: "note", Message: "非法批注"},
				Children: []lpwNode{
					{ID: "b1", Kind: "block", Type: "markdown", Props: map[string]any{"content": "1"}},
					{ID: "b2", Kind: "block", Type: "markdown", Props: map[string]any{"content": "2"}},
				},
			},
		},
	}
	err := validateDocument11(doc)
	if err == nil {
		t.Fatalf("expected annotation on layout to be rejected")
	}
	if !strings.Contains(err.Error(), "不允许设置 annotation") {
		t.Errorf("expected error rejecting annotation on layout, got: %v", err)
	}
}

func TestHierarchy_RejectsAnnotationOnContainer(t *testing.T) {
	t.Parallel()
	doc := &lpwDocument{
		Version: "1.1",
		Content: []lpwNode{
			{
				ID:         "sec-1",
				Kind:       "container",
				Type:       "section",
				Props:      map[string]any{"variant": "article", "title": "t"},
				Annotation: &lpwAnnotation{Kind: "note", Message: "非法批注"},
				Children: []lpwNode{
					{ID: "b1", Kind: "block", Type: "markdown", Props: map[string]any{"content": "1"}},
				},
			},
		},
	}
	err := validateDocument11(doc)
	if err == nil {
		t.Fatalf("expected annotation on container to be rejected")
	}
	if !strings.Contains(err.Error(), "不允许设置 annotation") {
		t.Errorf("expected error rejecting annotation on container, got: %v", err)
	}
}

func TestHierarchy_RejectsUnknownAnnotationField(t *testing.T) {
	t.Parallel()
	doc := &lpwDocument{
		Version: "1.1",
		Content: []lpwNode{
			{
				ID:    "md-1",
				Kind:  "block",
				Type:  "markdown",
				Props: map[string]any{"content": "abc"},
				Annotation: &lpwAnnotation{
					Kind:    "note",
					Message: "字段不对",
					Targets: []lpwAnnotationTarget{
						{Field: "src", Pattern: "abc"},
					},
				},
			},
		},
	}
	err := validateDocument11(doc)
	if err == nil {
		t.Fatalf("expected annotation with target field=src on markdown to be rejected")
	}
	if !strings.Contains(err.Error(), "不在 markdown 的可批注字段") {
		t.Errorf("expected unknown field error, got: %v", err)
	}
}

func TestHierarchy_RejectsBadAnnotationRegex(t *testing.T) {
	t.Parallel()
	doc := &lpwDocument{
		Version: "1.1",
		Content: []lpwNode{
			{
				ID:    "md-1",
				Kind:  "block",
				Type:  "markdown",
				Props: map[string]any{"content": "abc"},
				Annotation: &lpwAnnotation{
					Kind:    "note",
					Message: "坏正则",
					Targets: []lpwAnnotationTarget{
						{Field: "content", Pattern: "(("},
					},
				},
			},
		},
	}
	err := validateDocument11(doc)
	if err == nil {
		t.Fatalf("expected bad regex to be rejected")
	}
	if !strings.Contains(err.Error(), "批注正则无效") {
		t.Errorf("expected invalid regex error, got: %v", err)
	}
}

func TestHierarchy_RejectsDuplicateIDs(t *testing.T) {
	t.Parallel()
	doc := &lpwDocument{
		Version: "1.1",
		Content: []lpwNode{
			{ID: "dup-id", Kind: "block", Type: "markdown", Props: map[string]any{"content": "1"}},
			{ID: "dup-id", Kind: "block", Type: "markdown", Props: map[string]any{"content": "2"}},
		},
	}
	err := validateDocument11(doc)
	if err == nil {
		t.Fatalf("expected duplicate IDs to be rejected")
	}
	if !strings.Contains(err.Error(), "全文档必须唯一") {
		t.Errorf("expected duplicate ID error, got: %v", err)
	}
}

func TestHierarchy_RejectsOver500Nodes(t *testing.T) {
	t.Parallel()
	nodes := make([]lpwNode, 501)
	for i := 0; i < 501; i++ {
		nodes[i] = lpwNode{
			ID:    strings.ToLower(strings.ReplaceAll(strings.Repeat("a", 1)+string(rune('a'+i%26))+string(rune('a'+i/26%26))+string(rune('0'+i%10)), " ", "")),
			Kind:  "block",
			Type:  "markdown",
			Props: map[string]any{"content": "x"},
		}
	}
	doc := &lpwDocument{
		Version: "1.1",
		Content: nodes,
	}
	err := validateDocument11(doc)
	if err == nil {
		t.Fatalf("expected >500 nodes to be rejected")
	}
	if !strings.Contains(err.Error(), "超过上限 500") {
		t.Errorf("expected over 500 nodes error, got: %v", err)
	}
}

func TestHierarchy_RejectsMissingPattern(t *testing.T) {
	t.Parallel()
	doc := &lpwDocument{
		Version: "1.1",
		Content: []lpwNode{
			{
				ID:    "lay-1",
				Kind:  "layout",
				Type:  "layout",
				Props: map[string]any{},
				Children: []lpwNode{
					{ID: "b1", Kind: "block", Type: "markdown", Props: map[string]any{"content": "1"}},
					{ID: "b2", Kind: "block", Type: "markdown", Props: map[string]any{"content": "2"}},
				},
			},
		},
	}
	err := validateDocument11(doc)
	if err == nil {
		t.Fatalf("expected layout missing pattern to be rejected")
	}
	if !strings.Contains(err.Error(), "缺少必填属性 pattern") {
		t.Errorf("expected missing pattern error, got: %v", err)
	}
}

func TestHierarchy_RejectsPlacementsForeignNodeId(t *testing.T) {
	t.Parallel()
	doc := &lpwDocument{
		Version: "1.1",
		Content: []lpwNode{
			{
				ID:   "lay-1",
				Kind: "layout",
				Type: "layout",
				Props: map[string]any{
					"pattern": "split",
					"placements": []any{
						map[string]any{"nodeId": "foreign-id", "role": "primary"},
					},
				},
				Children: []lpwNode{
					{ID: "b1", Kind: "block", Type: "markdown", Props: map[string]any{"content": "1"}},
					{ID: "b2", Kind: "block", Type: "markdown", Props: map[string]any{"content": "2"}},
				},
			},
		},
	}
	err := validateDocument11(doc)
	if err == nil {
		t.Fatalf("expected foreign nodeId in placements to be rejected")
	}
	if !strings.Contains(err.Error(), "不是该 layout 的直接子节点") {
		t.Errorf("expected foreign nodeId error, got: %v", err)
	}
}

func TestHierarchy_RejectsFirstOfViolation(t *testing.T) {
	t.Parallel()
	doc := &lpwDocument{
		Version: "1.1",
		Content: []lpwNode{
			{
				ID:   "sec-feat",
				Kind: "container",
				Type: "section",
				Props: map[string]any{
					"variant": "feature",
					"title":   "亮点特性",
				},
				Children: []lpwNode{
					// section/feature 首个必须是 image / gallery / heading，放 callout 应该违背 FirstOf
					{ID: "co-1", Kind: "block", Type: "callout", Props: map[string]any{"content": "提示"}},
					{ID: "md-1", Kind: "block", Type: "markdown", Props: map[string]any{"content": "正文"}},
				},
			},
		},
	}
	err := validateDocument11(doc)
	if err == nil {
		t.Fatalf("expected FirstOf violation to be rejected")
	}
	if !strings.Contains(err.Error(), "第一个子块类型必须在") {
		t.Errorf("expected FirstOf error, got: %v", err)
	}
}

func TestHierarchy_RejectsRepeatTakeaway(t *testing.T) {
	t.Parallel()
	doc := &lpwDocument{
		Version: "1.1",
		Content: []lpwNode{
			{
				ID:   "pan-sum",
				Kind: "container",
				Type: "panel",
				Props: map[string]any{
					"variant": "summary",
				},
				Children: []lpwNode{
					{ID: "tk-1", Kind: "block", Type: "takeaway", Props: map[string]any{"content": "结论1"}},
					{ID: "tk-2", Kind: "block", Type: "takeaway", Props: map[string]any{"content": "结论2"}},
				},
			},
		},
	}
	err := validateDocument11(doc)
	if err == nil {
		t.Fatalf("expected repeat takeaway in summary panel to be rejected")
	}
	if !strings.Contains(err.Error(), "不允许重复出现") {
		t.Errorf("expected no-repeat error, got: %v", err)
	}
}

// ── 3. 写入原子性用例 ──────────────────────────────────────────

func TestInsertNode_NegativePositionRejected(t *testing.T) {
	t.Parallel()
	doc := &lpwDocument{
		Version: "1.1",
		Content: []lpwNode{
			{ID: "b1", Kind: "block", Type: "markdown", Props: map[string]any{"content": "1"}},
		},
	}
	neg := -1
	err := insertNode(doc, "", &neg, lpwNode{ID: "b2", Kind: "block", Type: "markdown", Props: map[string]any{}})
	if err == nil {
		t.Fatalf("expected negative position to fail")
	}
	if !strings.Contains(err.Error(), "不能为负数") {
		t.Errorf("expected negative position error, got: %v", err)
	}
	if len(doc.Content) != 1 {
		t.Errorf("document should remain untouched on error, len=%d", len(doc.Content))
	}
}

func TestRemoveNodes_AllOrNothing(t *testing.T) {
	t.Parallel()
	doc := &lpwDocument{
		Version: "1.1",
		Content: []lpwNode{
			{ID: "b1", Kind: "block", Type: "markdown", Props: map[string]any{"content": "1"}},
			{ID: "b2", Kind: "block", Type: "markdown", Props: map[string]any{"content": "2"}},
		},
	}
	// b1 存在，b3 不存在，整批必须失败且不删除 b1
	err := removeNodes(doc, []string{"b1", "b3"})
	if err == nil {
		t.Fatalf("expected removeNodes to fail when some ids do not exist")
	}
	if !strings.Contains(err.Error(), "不存在") {
		t.Errorf("expected missing id error, got: %v", err)
	}
	if len(doc.Content) != 2 {
		t.Errorf("expected document to remain unchanged (len=2), got %d", len(doc.Content))
	}
}

func TestReplaceNode_ExceedsMaxNodes(t *testing.T) {
	t.Parallel()
	// 构建一个拥有 499 个节点的文档
	doc := &lpwDocument{
		Version: "1.1",
		Content: make([]lpwNode, 499),
	}
	for i := 0; i < 499; i++ {
		doc.Content[i] = lpwNode{
			ID:    fmt.Sprintf("n-%d", i),
			Kind:  "block",
			Type:  "markdown",
			Props: map[string]any{"content": "t"},
		}
	}

	// 构造一个包含 3 个节点的替换子树（替换 1 个节点后总数变为 499 - 1 + 3 = 501 > 500）
	replacement := lpwNode{
		ID:   "repl-container",
		Kind: "container",
		Type: "section",
		Props: map[string]any{"variant": "article", "title": "t"},
		Children: []lpwNode{
			{ID: "repl-1", Kind: "block", Type: "markdown", Props: map[string]any{"content": "1"}},
			{ID: "repl-2", Kind: "block", Type: "markdown", Props: map[string]any{"content": "2"}},
		},
	}

	err := replaceNode(doc, "n-0", replacement)
	if err == nil {
		t.Fatalf("expected replaceNode exceeding 500 nodes to fail, got nil")
	}
	if !strings.Contains(err.Error(), "超过上限 500") {
		t.Errorf("expected error message containing '超过上限 500', got: %v", err)
	}
}

