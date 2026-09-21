package logic

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"github.com/xiaolfeng/Lumina/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// newPreviewEditLogic 构建基于 sqlite 的 PreviewLogic，验证行级编辑/读取/按名删除全链路
func newPreviewEditLogic(t *testing.T) *PreviewLogic {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "preview-edit.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&entity.PreviewSession{}, &entity.PreviewFile{}); err != nil {
		t.Fatalf("migrate sqlite: %v", err)
	}
	return &PreviewLogic{
		logic: logic{log: xLog.WithName(xLog.NamedLOGC, "PreviewEditTest")},
		repo: previewRepo{
			session: repository.NewPreviewSessionRepo(db),
			file:    repository.NewPreviewFileRepo(db),
		},
	}
}

// newPreviewEditSession 在测试库中创建一个 active 预览会话并返回其 ID
//
// ExpiresAt 置 nil：域上无过期时间的 active 会话合法，且规避 sqlite 驱动对 *time.Time 的扫描差异。
func newPreviewEditSession(t *testing.T, l *PreviewLogic) xSnowflake.SnowflakeID {
	t.Helper()
	id := xSnowflake.GenerateID(bConst.GenePreviewSession)
	session := &entity.PreviewSession{
		BaseEntity: xModels.BaseEntity{ID: id},
		ProjectID:  xSnowflake.GenerateID(bConst.GeneProject),
		Title:      "行级编辑测试",
		Hash:       "edit" + id.String(),
		Status:     bConst.PreviewSessionStatusActive,
	}
	if xErr := l.repo.session.Create(context.Background(), session); xErr != nil {
		t.Fatalf("create session: %s", xErr.Error())
	}
	return id
}

// TestPreviewLogicEditFileFlow 走查 上传 → insert/replace/delete → 行级读取 → 按名删除 全链路
func TestPreviewLogicEditFileFlow(t *testing.T) {
	l := newPreviewEditLogic(t)
	ctx := context.Background()
	sessionID := newPreviewEditSession(t, l)

	// 编辑不存在的文件 → NotFound（创建须走 UploadFile）
	if _, xErr := l.EditFile(ctx, sessionID, "app.js", bConst.PreviewEditOperationReplace, 1, 1, "x"); xErr == nil || xErr.GetErrorCode() != xError.NotFound {
		t.Fatalf("edit missing file: xErr = %v, want NotFound", xErr)
	}

	// 上传初始文件（3 行，带结尾换行）
	uploaded, xErr := l.UploadFile(ctx, sessionID, "app.js", "const a = 1;\nconst b = 2;\nconst c = 3;\n")
	if xErr != nil {
		t.Fatalf("upload: %s", xErr.Error())
	}

	// insert：第 2 行前插入注释
	inserted, xErr := l.EditFile(ctx, sessionID, "app.js", bConst.PreviewEditOperationInsert, 2, 0, "// header")
	if xErr != nil {
		t.Fatalf("insert: %s", xErr.Error())
	}
	if inserted.TotalLines != 4 {
		t.Fatalf("insert TotalLines = %d, want 4", inserted.TotalLines)
	}
	if inserted.ID != uploaded.ID {
		t.Fatalf("编辑应保留原文件 ID：insert ID = %v, upload ID = %v", inserted.ID, uploaded.ID)
	}
	if !strings.Contains(inserted.Region, "2| // header") {
		t.Fatalf("insert Region 应含带行号新行，got %q", inserted.Region)
	}

	// replace：第 2 行替换
	replaced, xErr := l.EditFile(ctx, sessionID, "app.js", bConst.PreviewEditOperationReplace, 2, 2, "// changed")
	if xErr != nil {
		t.Fatalf("replace: %s", xErr.Error())
	}
	if replaced.TotalLines != 4 || !strings.Contains(replaced.Region, "2| // changed") {
		t.Fatalf("replace 结果异常: total=%d region=%q", replaced.TotalLines, replaced.Region)
	}

	// delete：删除第 3-4 行
	deleted, xErr := l.EditFile(ctx, sessionID, "app.js", bConst.PreviewEditOperationDelete, 3, 4, "")
	if xErr != nil {
		t.Fatalf("delete lines: %s", xErr.Error())
	}
	if deleted.TotalLines != 2 {
		t.Fatalf("delete TotalLines = %d, want 2", deleted.TotalLines)
	}

	// 全量读取：原始内容（结尾换行保留）
	full, xErr := l.GetFileLinesBySession(ctx, sessionID, "app.js", 0, 0)
	if xErr != nil {
		t.Fatalf("get full: %s", xErr.Error())
	}
	if full.Content != "const a = 1;\n// changed\n" || full.TotalLines != 2 || full.StartLine != 1 || full.EndLine != 2 {
		t.Fatalf("full = %#v", full)
	}

	// 行区间读取：带行号
	head, xErr := l.GetFileLinesBySession(ctx, sessionID, "app.js", 1, 1)
	if xErr != nil {
		t.Fatalf("get range: %s", xErr.Error())
	}
	if head.Content != "1| const a = 1;" {
		t.Fatalf("range content = %q", head.Content)
	}

	// end_line 超出总行数钳制到末行
	tail, xErr := l.GetFileLinesBySession(ctx, sessionID, "app.js", 2, 99)
	if xErr != nil {
		t.Fatalf("get clamped range: %s", xErr.Error())
	}
	if tail.EndLine != 2 || tail.Content != "2| // changed" {
		t.Fatalf("clamped = %#v", tail)
	}

	// start_line 越界 → ParameterError
	if _, xErr = l.GetFileLinesBySession(ctx, sessionID, "app.js", 5, 9); xErr == nil || xErr.GetErrorCode() != xError.ParameterError {
		t.Fatalf("invalid start: xErr = %v, want ParameterError", xErr)
	}

	// delete 操作携带 content → ParameterError
	if _, xErr = l.EditFile(ctx, sessionID, "app.js", bConst.PreviewEditOperationDelete, 1, 1, "x"); xErr == nil || xErr.GetErrorCode() != xError.ParameterError {
		t.Fatalf("delete with content: xErr = %v, want ParameterError", xErr)
	}

	// 按名删除 → 文件消失 → 再删 NotFound
	snapshot, xErr := l.DeleteFileByName(ctx, sessionID, "app.js")
	if xErr != nil || snapshot.Filename != "app.js" {
		t.Fatalf("delete by name: snapshot=%v xErr=%v", snapshot, xErr)
	}
	if _, xErr = l.GetFileLinesBySession(ctx, sessionID, "app.js", 0, 0); xErr == nil || xErr.GetErrorCode() != xError.NotFound {
		t.Fatalf("get after delete: xErr = %v, want NotFound", xErr)
	}
	if _, xErr = l.DeleteFileByName(ctx, sessionID, "app.js"); xErr == nil || xErr.GetErrorCode() != xError.NotFound {
		t.Fatalf("delete twice: xErr = %v, want NotFound", xErr)
	}
}

// TestPreviewLogicEditFileEmptyFileInsert 空文件经 insert 建立内容，删除全部后回到空文件
func TestPreviewLogicEditFileEmptyFileInsert(t *testing.T) {
	l := newPreviewEditLogic(t)
	ctx := context.Background()
	sessionID := newPreviewEditSession(t, l)

	if _, xErr := l.UploadFile(ctx, sessionID, "empty.css", ""); xErr != nil {
		t.Fatalf("upload empty: %s", xErr.Error())
	}

	// 空文件省略 start_line（传 0）追加内容
	resp, xErr := l.EditFile(ctx, sessionID, "empty.css", bConst.PreviewEditOperationInsert, 0, 0, "body {\n  margin: 0;\n}")
	if xErr != nil {
		t.Fatalf("insert into empty: %s", xErr.Error())
	}
	if resp.TotalLines != 3 || resp.RegionStart != 1 || resp.RegionEnd != 3 {
		t.Fatalf("insert into empty: %#v", resp)
	}

	// 删除全部行 → 空文件（0 行、region 归零、内容为空）
	resp, xErr = l.EditFile(ctx, sessionID, "empty.css", bConst.PreviewEditOperationDelete, 1, 3, "")
	if xErr != nil {
		t.Fatalf("delete all: %s", xErr.Error())
	}
	if resp.TotalLines != 0 || resp.RegionStart != 0 || resp.RegionEnd != 0 || resp.Region != "" {
		t.Fatalf("delete all: %#v", resp)
	}

	// 空文件再传行区间 → 明确报错
	if _, xErr := l.GetFileLinesBySession(ctx, sessionID, "empty.css", 1, 1); xErr == nil || xErr.GetErrorCode() != xError.ParameterError {
		t.Fatalf("range on empty: xErr = %v, want ParameterError", xErr)
	}
}

// TestPreviewLogicEditFileLpwValidation 验证 Q-08：EditFile 对 .lpw 必须校验 LPW 语法与契约
func TestPreviewLogicEditFileLpwValidation(t *testing.T) {
	l := newPreviewEditLogic(t)
	ctx := context.Background()
	sessionID := newPreviewEditSession(t, l)

	validLpw := `{"version":"1.1","meta":{"title":"合法文档"},"content":[]}`
	_, xErr := l.UploadFile(ctx, sessionID, "test.lpw", validLpw)
	if xErr != nil {
		t.Fatalf("upload valid lpw failed: %s", xErr.Error())
	}

	// 1. 行级编辑为非法 JSON 必须被拦截并报 ParameterError
	_, xErr = l.EditFile(ctx, sessionID, "test.lpw", bConst.PreviewEditOperationReplace, 1, 1, "{ invalid json")
	if xErr == nil {
		t.Fatalf("expected error when editing .lpw with invalid json, got nil")
	}
	if xErr.GetErrorCode() != xError.ParameterError {
		t.Fatalf("expected ParameterError, got: %v", xErr.GetErrorCode())
	}

	// 2. 行级编辑为旧版 blocks 必须被拦截并报 ParameterError
	_, xErr = l.EditFile(ctx, sessionID, "test.lpw", bConst.PreviewEditOperationReplace, 1, 1, `{"version":"1.0","blocks":[]}`)
	if xErr == nil {
		t.Fatalf("expected error when editing .lpw with legacy blocks, got nil")
	}
	if xErr.GetErrorCode() != xError.ParameterError {
		t.Fatalf("expected ParameterError, got: %v", xErr.GetErrorCode())
	}

	// 3. 行级编辑为不支持的版本必须被拦截并报 ParameterError
	_, xErr = l.EditFile(ctx, sessionID, "test.lpw", bConst.PreviewEditOperationReplace, 1, 1, `{"version":"2.0","content":[]}`)
	if xErr == nil {
		t.Fatalf("expected error when editing .lpw with unsupported version, got nil")
	}
	if xErr.GetErrorCode() != xError.ParameterError {
		t.Fatalf("expected ParameterError, got: %v", xErr.GetErrorCode())
	}

	// 4. 行级编辑为合法 1.1 LPW 放行
	updatedLpw := `{"version":"1.1","meta":{"title":"更新后的文档"},"content":[]}`
	editResp, xErr := l.EditFile(ctx, sessionID, "test.lpw", bConst.PreviewEditOperationReplace, 1, 1, updatedLpw)
	if xErr != nil {
		t.Fatalf("valid lpw edit failed: %s", xErr.Error())
	}
	if editResp.Size != len(updatedLpw) {
		t.Fatalf("expected size %d, got %d", len(updatedLpw), editResp.Size)
	}

	// 5. 大写扩展名 .LPW 同样被拦截
	_, xErr = l.UploadFile(ctx, sessionID, "upper.LPW", validLpw)
	if xErr != nil {
		t.Fatalf("upload upper.LPW failed: %s", xErr.Error())
	}
	_, xErr = l.EditFile(ctx, sessionID, "upper.LPW", bConst.PreviewEditOperationReplace, 1, 1, "{ invalid json")
	if xErr == nil {
		t.Fatalf("expected error when editing upper.LPW with invalid json, got nil")
	}
	if xErr.GetErrorCode() != xError.ParameterError {
		t.Fatalf("expected ParameterError, got: %v", xErr.GetErrorCode())
	}

	// 6. 验证渐进中间态（split 仅 1 个 child）可通过文件通道正常上传与行级修复
	intermediateDoc := `{"version":"1.1","content":[{"id":"lay-1","kind":"layout","type":"layout","props":{"pattern":"split"},"children":[{"id":"b1","kind":"block","type":"markdown","props":{"content":"1"}}]}]}`
	_, xErr = l.UploadFile(ctx, sessionID, "mid.lpw", intermediateDoc)
	if xErr != nil {
		t.Fatalf("upload intermediate lpw failed: %s", xErr.Error())
	}
	fixedMidDoc := `{"version":"1.1","content":[{"id":"lay-1","kind":"layout","type":"layout","props":{"pattern":"split"},"children":[{"id":"b1","kind":"block","type":"markdown","props":{"content":"fixed"}}]}]}`
	_, xErr = l.EditFile(ctx, sessionID, "mid.lpw", bConst.PreviewEditOperationReplace, 1, 1, fixedMidDoc)
	if xErr != nil {
		t.Fatalf("edit intermediate lpw failed: %s", xErr.Error())
	}
}

