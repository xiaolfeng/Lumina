package logic

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	apiPreview "github.com/xiaolfeng/Lumina/api/preview"
	"github.com/xiaolfeng/Lumina/internal/service"
)

// memStorage 内存模拟存储
type memStorage struct {
	mu    sync.Mutex
	files map[string]memFile
}

type memFile struct {
	content   string
	updatedAt time.Time
}

func newMemStorage() *memStorage {
	return &memStorage{files: make(map[string]memFile)}
}

func (s *memStorage) key(sessionID xSnowflake.SnowflakeID, filename string) string {
	return fmt.Sprintf("%d:%s", sessionID.Int64(), filename)
}

func (s *memStorage) GetFile(ctx context.Context, sessionID xSnowflake.SnowflakeID, filename string) (string, time.Time, *xError.Error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	f, exists := s.files[s.key(sessionID, filename)]
	if !exists {
		return "", time.Time{}, xError.NewError(ctx, xError.NotFound, "文件不存在", false, nil)
	}
	return f.content, f.updatedAt, nil
}

func (s *memStorage) SaveFile(ctx context.Context, sessionID xSnowflake.SnowflakeID, filename, content string) (*apiPreview.PreviewFileResponse, *xError.Error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	s.files[s.key(sessionID, filename)] = memFile{
		content:   content,
		updatedAt: now,
	}
	return &apiPreview.PreviewFileResponse{
		ID:        xSnowflake.SnowflakeID(100),
		SessionID: sessionID,
		Filename:  filename,
		Size:      len(content),
		UpdatedAt: now.Format(time.RFC3339),
	}, nil
}

func setupTestLpwLogic(t *testing.T) (*PreviewLpwLogic, *memStorage) {
	schemaLoader, err := service.NewLpwSchemaLoader()
	if err != nil {
		t.Fatalf("failed to init schema loader: %v", err)
	}
	mem := newMemStorage()
	l := &PreviewLpwLogic{
		logic:   logic{log: xLog.WithName(xLog.NamedLOGC, "PreviewLpwLogic")},
		schema:  schemaLoader,
		storage: mem,
	}
	return l, mem
}

func TestPreviewLpwLogicFlow(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	l, _ := setupTestLpwLogic(t)
	sid := xSnowflake.SnowflakeID(1)
	fn := "doc.lpw"

	// 1. init
	res, err := l.InitDocument(ctx, sid, fn, lpwMeta{Title: "测试文档", Description: "描述", Author: "作者", Tags: []string{"tag1"}}, nil, "")
	if err != nil {
		t.Fatalf("InitDocument error: %v", err)
	}
	if res.TotalNodes != 0 {
		t.Errorf("expected 0 nodes after init, got %d", res.TotalNodes)
	}

	// 2. add x 3 (heading / markdown / divider)
	b1 := lpwNode{ID: "h1", Kind: "block", Type: "heading", Props: map[string]any{"level": float64(1), "content": "标题"}}
	res1, err := l.AddNode(ctx, sid, fn, b1, "", nil, "")
	if err != nil {
		t.Fatalf("AddNode b1 error: %v", err)
	}
	if res1.TotalNodes != 1 {
		t.Errorf("expected 1 node, got %d", res1.TotalNodes)
	}

	b2 := lpwNode{ID: "m1", Kind: "block", Type: "markdown", Props: map[string]any{"content": "正文内容"}}
	res2, err := l.AddNode(ctx, sid, fn, b2, "", nil, "")
	if err != nil {
		t.Fatalf("AddNode b2 error: %v", err)
	}
	if res2.TotalNodes != 2 {
		t.Errorf("expected 2 nodes, got %d", res2.TotalNodes)
	}

	b3 := lpwNode{ID: "d1", Kind: "block", Type: "divider", Props: map[string]any{}}
	res3, err := l.AddNode(ctx, sid, fn, b3, "", nil, "")
	if err != nil {
		t.Fatalf("AddNode b3 error: %v", err)
	}
	if res3.TotalNodes != 3 {
		t.Errorf("expected 3 nodes, got %d", res3.TotalNodes)
	}

	// 3. outline
	outline, err := l.Outline(ctx, sid, fn)
	if err != nil {
		t.Fatalf("Outline error: %v", err)
	}
	if outline.NodeCount != 3 {
		t.Errorf("expected 3 outline items, got %d", outline.NodeCount)
	}
	if outline.Items[0].ID != "h1" || outline.Items[1].ID != "m1" || outline.Items[2].ID != "d1" {
		t.Errorf("outline order unexpected: %v", outline.Items)
	}

	// 4. edit patch
	_, err = l.EditNode(ctx, sid, fn, "h1", map[string]any{"content": "新标题"}, nil, nil, "")
	if err != nil {
		t.Fatalf("EditNode error: %v", err)
	}

	// 5. sort
	_, err = l.SortNodes(ctx, sid, fn, "", []string{"d1", "h1", "m1"}, "")
	if err != nil {
		t.Fatalf("SortNodes error: %v", err)
	}
	outlineAfterSort, _ := l.Outline(ctx, sid, fn)
	if outlineAfterSort.Items[0].ID != "d1" {
		t.Errorf("expected d1 at index 0 after sort")
	}

	// 6. remove
	_, err = l.RemoveNodes(ctx, sid, fn, []string{"d1"}, "")
	if err != nil {
		t.Fatalf("RemoveNodes error: %v", err)
	}
	outlineAfterRemove, _ := l.Outline(ctx, sid, fn)
	if outlineAfterRemove.NodeCount != 2 {
		t.Errorf("expected 2 nodes after remove, got %d", outlineAfterRemove.NodeCount)
	}
}

func TestPreviewLpwLogicAtomicity(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	l, mem := setupTestLpwLogic(t)
	sid := xSnowflake.SnowflakeID(2)
	fn := "atomic.lpw"

	_, _ = l.InitDocument(ctx, sid, fn, lpwMeta{Title: "原子性测试"}, nil, "")
	initialContent, _, _ := mem.GetFile(ctx, sid, fn)

	// 构造一个 Schema 校验失败的 Add（metrics items 为空）
	invalidBlock := lpwNode{
		ID:    "bad-metrics",
		Kind:  "block",
		Type:  "metrics",
		Props: map[string]any{"items": []any{}}, // 非法：minItems 1
	}

	_, err := l.AddNode(ctx, sid, fn, invalidBlock, "", nil, "")
	if err == nil {
		t.Fatalf("expected add invalid node to fail")
	}

	// 断言：失败后文件字节完全未变！
	afterContent, _, _ := mem.GetFile(ctx, sid, fn)
	if initialContent != afterContent {
		t.Errorf("expected file content to remain untouched on failure\nbefore: %s\nafter: %s", initialContent, afterContent)
	}
}

func TestPreviewLpwLogicConcurrency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	l, _ := setupTestLpwLogic(t)
	sid := xSnowflake.SnowflakeID(3)
	fn := "concurrent.lpw"

	_, _ = l.InitDocument(ctx, sid, fn, lpwMeta{Title: "并发测试"}, nil, "")

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		_, _ = l.AddNode(ctx, sid, fn, lpwNode{
			ID:    "c-1",
			Kind:  "block",
			Type:  "markdown",
			Props: map[string]any{"content": "并发内容 1"},
		}, "", nil, "")
	}()

	go func() {
		defer wg.Done()
		_, _ = l.AddNode(ctx, sid, fn, lpwNode{
			ID:    "c-2",
			Kind:  "block",
			Type:  "markdown",
			Props: map[string]any{"content": "并发内容 2"},
		}, "", nil, "")
	}()

	wg.Wait()

	outline, err := l.Outline(ctx, sid, fn)
	if err != nil {
		t.Fatalf("outline failed: %v", err)
	}
	if outline.NodeCount != 2 {
		t.Errorf("expected 2 nodes after concurrent writes, got %d", outline.NodeCount)
	}
}

func TestPreviewLpwLogicInitWithBlocks(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	l, _ := setupTestLpwLogic(t)
	sid := xSnowflake.SnowflakeID(5)
	fn := "init-nodes.lpw"

	// 1. 合法初始节点
	nodes := []lpwNode{
		{ID: "h-1", Kind: "block", Type: "heading", Props: map[string]any{"level": float64(2), "content": "标题"}},
		{ID: "m-1", Kind: "block", Type: "markdown", Props: map[string]any{"content": "正文"}},
	}
	res, err := l.InitDocument(ctx, sid, fn, lpwMeta{Title: "初始节点文档"}, nodes, "")
	if err != nil {
		t.Fatalf("InitDocument with nodes error: %v", err)
	}
	if res.TotalNodes != 2 {
		t.Errorf("expected 2 nodes after init-with-nodes, got %d", res.TotalNodes)
	}

	outline, err := l.Outline(ctx, sid, fn)
	if err != nil {
		t.Fatalf("Outline error: %v", err)
	}
	if outline.NodeCount != 2 || outline.Items[0].ID != "h-1" || outline.Items[1].ID != "m-1" {
		t.Errorf("unexpected outline after init-with-nodes: %+v", outline.Items)
	}

	// 2. 非法初始节点（未注册类型）：init 必须整体失败且不落库
	badNodes := []lpwNode{
		{ID: "bad-1", Kind: "block", Type: "not-a-type", Props: map[string]any{}},
	}
	_, err = l.InitDocument(ctx, sid, fn, lpwMeta{Title: "非法初始节点"}, badNodes, "")
	if err == nil {
		t.Fatalf("expected init with invalid node to fail")
	}

	// 文件内容保持上一次合法版本
	outlineAfter, err := l.Outline(ctx, sid, fn)
	if err != nil {
		t.Fatalf("Outline after failed init error: %v", err)
	}
	if outlineAfter.NodeCount != 2 {
		t.Errorf("expected file untouched after failed init (2 nodes), got %d", outlineAfter.NodeCount)
	}
}

func TestPreviewLpwLogicRevisionConflict(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	l, _ := setupTestLpwLogic(t)
	sid := xSnowflake.SnowflakeID(4)
	fn := "conflict.lpw"

	_, _ = l.InitDocument(ctx, sid, fn, lpwMeta{Title: "版本测试"}, nil, "")

	// 传入过期的 revision
	staleRevision := "2020-01-01T00:00:00Z"
	_, err := l.AddNode(ctx, sid, fn, lpwNode{
		ID:    "b1",
		Kind:  "block",
		Type:  "markdown",
		Props: map[string]any{"content": "hi"},
	}, "", nil, staleRevision)

	if err == nil {
		t.Fatalf("expected revision conflict error")
	}
}

func TestPreviewLpwLogicProgressiveLayoutAndContainer(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	l, _ := setupTestLpwLogic(t)
	sid := xSnowflake.SnowflakeID(10)
	fn := "progressive.lpw"

	_, err := l.InitDocument(ctx, sid, fn, lpwMeta{Title: "分步构建测试"}, nil, "")
	if err != nil {
		t.Fatalf("InitDocument error: %v", err)
	}

	// 1. 插入空白 split layout
	topLayout := lpwNode{
		ID:    "top",
		Kind:  "layout",
		Type:  "layout",
		Props: map[string]any{"pattern": "split"},
	}
	res1, err := l.AddNode(ctx, sid, fn, topLayout, "", nil, "")
	if err != nil {
		t.Fatalf("AddNode empty layout failed: %v", err)
	}
	if res1.TotalNodes != 1 {
		t.Errorf("expected 1 node, got %d", res1.TotalNodes)
	}

	// 2. 插入空白 section container (variant=feature, 要求 MinItems: 2) 到 split layout 下
	secContainer := lpwNode{
		ID:    "sec-1",
		Kind:  "container",
		Type:  "section",
		Props: map[string]any{"variant": "feature", "title": "特性区域"},
	}
	res2, err := l.AddNode(ctx, sid, fn, secContainer, "top", nil, "")
	if err != nil {
		t.Fatalf("AddNode empty container failed: %v", err)
	}
	if res2.TotalNodes != 2 {
		t.Errorf("expected 2 nodes, got %d", res2.TotalNodes)
	}

	// 3. 插入首个 child image (满足 FirstOf: image/gallery/heading，但此时只有 1 个子节点 < MinItems: 2)
	imgBlock := lpwNode{
		ID:    "img-1",
		Kind:  "block",
		Type:  "image",
		Props: map[string]any{"src": "https://example.com/img.png", "alt": "示例图"},
	}
	res3, err := l.AddNode(ctx, sid, fn, imgBlock, "sec-1", nil, "")
	if err != nil {
		t.Fatalf("AddNode first block failed: %v", err)
	}
	if res3.TotalNodes != 3 {
		t.Errorf("expected 3 nodes, got %d", res3.TotalNodes)
	}

	// 4. 插入第二个 child markdown (此时达到 MinItems: 2)
	mdBlock := lpwNode{
		ID:    "md-1",
		Kind:  "block",
		Type:  "markdown",
		Props: map[string]any{"content": "正文说明"},
	}
	res4, err := l.AddNode(ctx, sid, fn, mdBlock, "sec-1", nil, "")
	if err != nil {
		t.Fatalf("AddNode second block failed: %v", err)
	}
	if res4.TotalNodes != 4 {
		t.Errorf("expected 4 nodes, got %d", res4.TotalNodes)
	}
}

func TestPreviewLpwLogicInitRejectsIncompleteNodes(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	l, _ := setupTestLpwLogic(t)
	sid := xSnowflake.SnowflakeID(11)
	fn := "init-incomplete.lpw"

	// init 一次性提交语义：split layout 只挂 1 个 child（低于 MinChildren=2）必须整体失败
	incomplete := []lpwNode{
		{
			ID:      "lay-1",
			Kind:    "layout",
			Type:    "layout",
			Props:   map[string]any{"pattern": "split"},
			Children: []lpwNode{
				{ID: "md-1", Kind: "block", Type: "markdown", Props: map[string]any{"content": "只有一半"}},
			},
		},
	}
	_, err := l.InitDocument(ctx, sid, fn, lpwMeta{Title: "残缺初始文档"}, incomplete, "")
	if err == nil {
		t.Fatalf("expected init with incomplete split layout to fail")
	}
	if !strings.Contains(err.Error(), "超出限制") {
		t.Errorf("expected min-children violation error, got: %v", err)
	}

	// 失败后文件不得落库
	if _, oErr := l.Outline(ctx, sid, fn); oErr == nil {
		t.Errorf("expected outline to fail after rejected init (file must not persist)")
	}
}

func TestPreviewLpwLogicOutlineCompleteness(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	l, _ := setupTestLpwLogic(t)
	sid := xSnowflake.SnowflakeID(12)
	fn := "completeness.lpw"

	if _, err := l.InitDocument(ctx, sid, fn, lpwMeta{Title: "完备性测试"}, nil, ""); err != nil {
		t.Fatalf("InitDocument error: %v", err)
	}

	// 渐进构建中间态：split layout 只挂 1 个 child
	_, err := l.AddNode(ctx, sid, fn, lpwNode{
		ID: "lay-1", Kind: "layout", Type: "layout",
		Props: map[string]any{"pattern": "split"},
	}, "", nil, "")
	if err != nil {
		t.Fatalf("AddNode layout failed: %v", err)
	}
	_, err = l.AddNode(ctx, sid, fn, lpwNode{
		ID: "md-1", Kind: "block", Type: "markdown",
		Props: map[string]any{"content": "第一栏"},
	}, "lay-1", nil, "")
	if err != nil {
		t.Fatalf("AddNode block failed: %v", err)
	}

	// 中间态 outline 必须给出完备性告警
	midOutline, err := l.Outline(ctx, sid, fn)
	if err != nil {
		t.Fatalf("Outline error: %v", err)
	}
	if len(midOutline.CompletenessWarnings) == 0 {
		t.Fatalf("expected completeness warnings for incomplete layout")
	}
	if !strings.Contains(strings.Join(midOutline.CompletenessWarnings, "; "), "split") {
		t.Errorf("expected split min-children warning, got: %v", midOutline.CompletenessWarnings)
	}

	// 补齐第二个 child 后告警清零
	_, err = l.AddNode(ctx, sid, fn, lpwNode{
		ID: "md-2", Kind: "block", Type: "markdown",
		Props: map[string]any{"content": "第二栏"},
	}, "lay-1", nil, "")
	if err != nil {
		t.Fatalf("AddNode second block failed: %v", err)
	}
	finalOutline, err := l.Outline(ctx, sid, fn)
	if err != nil {
		t.Fatalf("Outline error: %v", err)
	}
	if len(finalOutline.CompletenessWarnings) != 0 {
		t.Errorf("expected no completeness warnings on complete doc, got: %v", finalOutline.CompletenessWarnings)
	}
}

func TestPreviewLpwLogicOutlineVersionErrorNotWarning(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	l, storage := setupTestLpwLogic(t)
	sid := xSnowflake.SnowflakeID(13)
	fn := "bad_ver.lpw"

	// 存入一个 version 不是 1.1 的文档
	badDoc := `{"version":"1.0","content":[]}`
	if _, err := storage.SaveFile(ctx, sid, fn, badDoc); err != nil {
		t.Fatalf("SaveFile failed: %v", err)
	}

	// Outline 必须直接返回 ParameterError，严禁混入 CompletenessWarnings
	outline, xErr := l.Outline(ctx, sid, fn)
	if xErr == nil {
		t.Fatalf("expected Outline to return error for unsupported version, got outline: %#v", outline)
	}
	if xErr.GetErrorCode() != xError.ParameterError {
		t.Fatalf("expected ParameterError, got: %v", xErr.GetErrorCode())
	}

	// 存入一个包含旧版 blocks 的文档
	fnBlocks := "legacy_blocks.lpw"
	legacyDoc := `{"version":"1.1","blocks":[]}`
	if _, err := storage.SaveFile(ctx, sid, fnBlocks, legacyDoc); err != nil {
		t.Fatalf("SaveFile legacy failed: %v", err)
	}
	_, xErr = l.Outline(ctx, sid, fnBlocks)
	if xErr == nil {
		t.Fatalf("expected Outline to return error for legacy blocks, got nil")
	}
	if xErr.GetErrorCode() != xError.ParameterError {
		t.Fatalf("expected ParameterError, got: %v", xErr.GetErrorCode())
	}
}

func TestInitDocument_ExceedsMaxNodes(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	l, _ := setupTestLpwLogic(t)
	sid := xSnowflake.SnowflakeID(14)
	fn := "init_501.lpw"

	nodes := make([]LpwNodeExport, 501)
	for i := 0; i < 501; i++ {
		nodes[i] = LpwNodeExport{
			ID:    fmt.Sprintf("n-%d", i),
			Kind:  "block",
			Type:  "markdown",
			Props: map[string]any{"content": "t"},
		}
	}

	_, xErr := l.InitDocument(ctx, sid, fn, lpwMeta{Title: "超额初始化"}, nodes, "")
	if xErr == nil {
		t.Fatalf("expected InitDocument exceeding 500 nodes to fail, got nil")
	}
	if !strings.Contains(xErr.Error(), "超过上限 500") {
		t.Errorf("expected error message containing '超过上限 500', got: %v", xErr)
	}
}


