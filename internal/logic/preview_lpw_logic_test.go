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
	res, err := l.InitDocument(ctx, sid, fn, "测试文档", "描述", "作者", []string{"tag1"}, "")
	if err != nil {
		t.Fatalf("InitDocument error: %v", err)
	}
	if res.TotalBlocks != 0 {
		t.Errorf("expected 0 blocks after init, got %d", res.TotalBlocks)
	}

	// 2. add x 3 (heading / markdown / divider)
	b1 := lpwBlock{ID: "h1", Type: "heading", Props: map[string]any{"level": float64(1), "content": "标题"}}
	res1, err := l.AddBlock(ctx, sid, fn, "", nil, b1, "")
	if err != nil {
		t.Fatalf("AddBlock b1 error: %v", err)
	}
	if res1.TotalBlocks != 1 {
		t.Errorf("expected 1 block, got %d", res1.TotalBlocks)
	}

	b2 := lpwBlock{ID: "m1", Type: "markdown", Props: map[string]any{"content": "正文内容"}}
	res2, err := l.AddBlock(ctx, sid, fn, "", nil, b2, "")
	if err != nil {
		t.Fatalf("AddBlock b2 error: %v", err)
	}
	if res2.TotalBlocks != 2 {
		t.Errorf("expected 2 blocks, got %d", res2.TotalBlocks)
	}

	b3 := lpwBlock{ID: "d1", Type: "divider", Props: map[string]any{}}
	res3, err := l.AddBlock(ctx, sid, fn, "", nil, b3, "")
	if err != nil {
		t.Fatalf("AddBlock b3 error: %v", err)
	}
	if res3.TotalBlocks != 3 {
		t.Errorf("expected 3 blocks, got %d", res3.TotalBlocks)
	}

	// 3. outline
	outline, err := l.Outline(ctx, sid, fn)
	if err != nil {
		t.Fatalf("Outline error: %v", err)
	}
	if outline.BlockCount != 3 {
		t.Errorf("expected 3 outline items, got %d", outline.BlockCount)
	}
	if outline.Items[0].ID != "h1" || outline.Items[1].ID != "m1" || outline.Items[2].ID != "d1" {
		t.Errorf("outline order unexpected: %v", outline.Items)
	}

	// 4. edit patch
	_, err = l.EditBlock(ctx, sid, fn, "h1", map[string]any{"content": "新标题"}, nil, "")
	if err != nil {
		t.Fatalf("EditBlock error: %v", err)
	}

	// 5. sort
	_, err = l.SortBlocks(ctx, sid, fn, "", []string{"d1", "h1", "m1"}, "")
	if err != nil {
		t.Fatalf("SortBlocks error: %v", err)
	}
	outlineAfterSort, _ := l.Outline(ctx, sid, fn)
	if outlineAfterSort.Items[0].ID != "d1" {
		t.Errorf("expected d1 at index 0 after sort")
	}

	// 6. remove
	_, err = l.RemoveBlocks(ctx, sid, fn, []string{"d1"}, "")
	if err != nil {
		t.Fatalf("RemoveBlocks error: %v", err)
	}
	outlineAfterRemove, _ := l.Outline(ctx, sid, fn)
	if outlineAfterRemove.BlockCount != 2 {
		t.Errorf("expected 2 blocks after remove, got %d", outlineAfterRemove.BlockCount)
	}
}

func TestPreviewLpwLogicAtomicity(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	l, mem := setupTestLpwLogic(t)
	sid := xSnowflake.SnowflakeID(2)
	fn := "atomic.lpw"

	_, _ = l.InitDocument(ctx, sid, fn, "原子性测试", "", "", nil, "")
	initialContent, _, _ := mem.GetFile(ctx, sid, fn)

	// 构造一个 Schema 校验失败的 Add（metrics items 为空）
	invalidBlock := lpwBlock{
		ID:    "bad-metrics",
		Type:  "metrics",
		Props: map[string]any{"items": []any{}}, // 非法：minItems 1
	}

	_, err := l.AddBlock(ctx, sid, fn, "", nil, invalidBlock, "")
	if err == nil {
		t.Fatalf("expected add invalid block to fail")
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

	_, _ = l.InitDocument(ctx, sid, fn, "并发测试", "", "", nil, "")

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		_, _ = l.AddBlock(ctx, sid, fn, "", nil, lpwBlock{
			ID:    "c-1",
			Type:  "markdown",
			Props: map[string]any{"content": "并发内容 1"},
		}, "")
	}()

	go func() {
		defer wg.Done()
		_, _ = l.AddBlock(ctx, sid, fn, "", nil, lpwBlock{
			ID:    "c-2",
			Type:  "markdown",
			Props: map[string]any{"content": "并发内容 2"},
		}, "")
	}()

	wg.Wait()

	outline, err := l.Outline(ctx, sid, fn)
	if err != nil {
		t.Fatalf("outline failed: %v", err)
	}
	if outline.BlockCount != 2 {
		t.Errorf("expected 2 blocks after concurrent writes, got %d", outline.BlockCount)
	}
}

func TestPreviewLpwLogicRevisionConflict(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	l, _ := setupTestLpwLogic(t)
	sid := xSnowflake.SnowflakeID(4)
	fn := "conflict.lpw"

	_, _ = l.InitDocument(ctx, sid, fn, "版本测试", "", "", nil, "")

	// 传入过期的 revision
	staleRevision := "2020-01-01T00:00:00Z"
	_, err := l.AddBlock(ctx, sid, fn, "", nil, lpwBlock{
		ID:    "b1",
		Type:  "markdown",
		Props: map[string]any{"content": "hi"},
	}, staleRevision)

	if err == nil {
		t.Fatalf("expected revision conflict error")
	}
	if !strings.Contains(err.Error(), "已被修改") {
		t.Errorf("expected revision conflict message, got: %v", err)
	}
}
