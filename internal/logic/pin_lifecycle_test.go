package logic

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	"github.com/redis/go-redis/v9"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"github.com/xiaolfeng/Lumina/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupPinTestEnv 构建基于 SQLite 与 MiniRedis 的 Pin 测试环境
func setupPinTestEnv(t *testing.T) (*PinLogic, *gorm.DB, *entity.Project) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "pin-lifecycle.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&entity.Project{}); err != nil {
		t.Fatalf("auto migrate sqlite project: %v", err)
	}
	// 在 SQLite 中，entity.Pin 的 ConsumedAt 标签为 type:timestamptz（用于 PostgreSQL），
	// mattn/go-sqlite3 仅对 DATETIME/TIMESTAMP 识别并将其转换为 time.Time，
	// 若声明为 timestamptz 则会作为 string 返回导致 Scan 错误。
	// 因此在 SQLite 测试中创建 pins 表时显式使用 DATETIME。
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS pins (
		id INTEGER PRIMARY KEY,
		created_at DATETIME,
		updated_at DATETIME,
		from_project_id INTEGER NOT NULL,
		to_project_id INTEGER NOT NULL,
		title VARCHAR(255) NOT NULL,
		content TEXT NOT NULL,
		category VARCHAR(16) DEFAULT 'notice',
		status VARCHAR(16) DEFAULT 'pending',
		priority VARCHAR(16) DEFAULT 'medium',
		consumed_at DATETIME
	)`).Error; err != nil {
		t.Fatalf("create pins table in sqlite: %v", err)
	}

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	t.Cleanup(mr.Close)

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	projectRepo := repository.NewProjectRepo(db, rdb)
	pRepo := repository.NewPinRepo(db)

	workspaceID := xSnowflake.GenerateID(bConst.GeneWorkspace)
	projectID := xSnowflake.GenerateID(bConst.GeneProject)
	project := &entity.Project{
		BaseEntity:  xModels.BaseEntity{ID: projectID},
		WorkspaceID: workspaceID,
		Name:        "Pin Target Project",
		AliasName:   "target-proj",
	}
	if err := db.Create(project).Error; err != nil {
		t.Fatalf("create test project: %v", err)
	}

	l := &PinLogic{
		logic: logic{
			log: xLog.WithName(xLog.NamedLOGC, "PinLifecycleTest"),
		},
		repo: pinRepo{
			pin:     pRepo,
			project: projectRepo,
		},
	}

	return l, db, project
}

// TestPinLifecycle_PeekAndConsume 严格验证 Pin 生命周期闭环：
// 1. 插入 pending 约束；
// 2. 调用 PeekOldestPending，断言返回正确标题与正文；
// 3. 关键断言：断言该约束在数据库/存储中的状态依然为 pending，consumed_at 为空；
// 4. 再次调用 Peek，断言仍然能够重复只读读取，状态未变；
// 5. 随后调用 Consume，断言状态被原子更新为 consumed，consumed_at 被正确设置；
// 6. 再次调用 PeekOldestPending，断言此时无待消费约束（NotFound）。
func TestPinLifecycle_PeekAndConsume(t *testing.T) {
	l, db, project := setupPinTestEnv(t)
	ctx := context.Background()

	// Step 0: 尚未插入任何约束时，PeekOldestPending 应返回 NotFound
	emptyResp, xErr := l.PeekOldestPending(ctx, project.ID)
	if emptyResp != nil {
		t.Fatalf("expected nil resp on empty queue, got %v", emptyResp)
	}
	if xErr == nil || xErr.GetErrorCode() != xError.NotFound {
		t.Fatalf("expected NotFound on empty queue, got %v", xErr)
	}

	// 验证 repo.GetOldestPending 在无记录时返回 nil, nil
	repoEmptyPin, repoErr := l.repo.pin.GetOldestPending(ctx, project.ID)
	if repoErr != nil {
		t.Fatalf("repo.GetOldestPending error on empty queue: %v", repoErr)
	}
	if repoEmptyPin != nil {
		t.Fatalf("expected nil repoEmptyPin, got %v", repoEmptyPin)
	}

	// Step 1: 插入一条 pending 约束
	pinID := xSnowflake.GenerateID(bConst.GenePin)
	expectedTitle := "接口 breaking change 注意事项"
	expectedContent := "用户鉴权端点重构为 OAuth 2.1，需要同步移除所有旧 token 依赖。"
	now := time.Now().Truncate(time.Second)

	pin := &entity.Pin{
		BaseEntity: xModels.BaseEntity{
			ID:        pinID,
			CreatedAt: now,
			UpdatedAt: now,
		},
		FromProjectID: project.ID,
		ToProjectID:   project.ID,
		Title:         expectedTitle,
		Content:       expectedContent,
		Category:      "breaking",
		Status:        bConst.PinStatusPending,
		Priority:      bConst.PinPriorityHigh,
		ConsumedAt:    nil,
	}
	if err := db.Create(pin).Error; err != nil {
		t.Fatalf("create pin entity: %v", err)
	}

	// Step 2: 调用 PeekOldestPending，断言返回正确的标题与正文
	peekResp, xErr := l.PeekOldestPending(ctx, project.ID)
	if xErr != nil {
		t.Fatalf("PeekOldestPending failed: %s", xErr.Error())
	}
	if peekResp == nil {
		t.Fatal("PeekOldestPending returned nil response")
	}
	if peekResp.ID != pinID {
		t.Errorf("peekResp.ID = %s, want %s", peekResp.ID, pinID)
	}
	if peekResp.Title != expectedTitle {
		t.Errorf("peekResp.Title = %q, want %q", peekResp.Title, expectedTitle)
	}
	if peekResp.Content != expectedContent {
		t.Errorf("peekResp.Content = %q, want %q", peekResp.Content, expectedContent)
	}
	if peekResp.Status != bConst.PinStatusPending {
		t.Errorf("peekResp.Status = %s, want %s", peekResp.Status, bConst.PinStatusPending)
	}
	if peekResp.ConsumedAt != "" {
		t.Errorf("peekResp.ConsumedAt = %q, want empty string", peekResp.ConsumedAt)
	}

	// 同时验证 repo.GetOldestPending 的只读返回
	repoPin, repoXErr := l.repo.pin.GetOldestPending(ctx, project.ID)
	if repoXErr != nil {
		t.Fatalf("repo.GetOldestPending failed: %s", repoXErr.Error())
	}
	if repoPin == nil || repoPin.ID != pinID {
		t.Fatalf("repo.GetOldestPending id mismatch: got %v, want %s", repoPin, pinID)
	}

	// Step 3 (关键断言): 断言该约束在数据库/存储中的状态依然为 pending，consumed_at 为空
	var dbPinAfterPeek entity.Pin
	if err := db.Where("id = ?", pinID).First(&dbPinAfterPeek).Error; err != nil {
		t.Fatalf("failed to query pin from db: %v", err)
	}
	if dbPinAfterPeek.Status != bConst.PinStatusPending {
		t.Fatalf("CRITICAL: DB status was modified by peek! got %s, want %s", dbPinAfterPeek.Status, bConst.PinStatusPending)
	}
	if dbPinAfterPeek.ConsumedAt != nil {
		t.Fatalf("CRITICAL: DB consumed_at was set by peek! got %v, want nil", dbPinAfterPeek.ConsumedAt)
	}

	// Step 4: 再次调用 PeekOldestPending 与 Peek(id)，断言仍然能够重复只读读取，状态未变
	for i := 0; i < 3; i++ {
		repeatPeekResp, repeatErr := l.PeekOldestPending(ctx, project.ID)
		if repeatErr != nil {
			t.Fatalf("repeat %d PeekOldestPending failed: %v", i, repeatErr)
		}
		if repeatPeekResp.Status != bConst.PinStatusPending || repeatPeekResp.ConsumedAt != "" {
			t.Fatalf("repeat %d peek returned non-pending status: %s", i, repeatPeekResp.Status)
		}

		repeatByIDResp, repeatByIDErr := l.Peek(ctx, pinID)
		if repeatByIDErr != nil {
			t.Fatalf("repeat %d Peek by ID failed: %v", i, repeatByIDErr)
		}
		if repeatByIDResp.Status != bConst.PinStatusPending || repeatByIDResp.ConsumedAt != "" {
			t.Fatalf("repeat %d Peek by ID returned non-pending status: %s", i, repeatByIDResp.Status)
		}
	}

	// 再次检查 DB 状态确认绝对只读幂等
	if err := db.Where("id = ?", pinID).First(&dbPinAfterPeek).Error; err != nil {
		t.Fatalf("query db after repeat peek: %v", err)
	}
	if dbPinAfterPeek.Status != bConst.PinStatusPending || dbPinAfterPeek.ConsumedAt != nil {
		t.Fatalf("DB state polluted after repeated peeks: status=%s, consumed_at=%v", dbPinAfterPeek.Status, dbPinAfterPeek.ConsumedAt)
	}

	// Step 5: 随后调用 Consume，断言状态被原子更新为 consumed，consumed_at 被正确设置
	consumeResp, consumeErr := l.Consume(ctx, project.ID, nil)
	if consumeErr != nil {
		t.Fatalf("Consume failed: %s", consumeErr.Error())
	}
	if consumeResp == nil {
		t.Fatal("Consume returned nil response")
	}
	if consumeResp.ID != pinID {
		t.Errorf("consumeResp.ID = %s, want %s", consumeResp.ID, pinID)
	}
	if consumeResp.Status != bConst.PinStatusConsumed {
		t.Errorf("consumeResp.Status = %s, want %s", consumeResp.Status, bConst.PinStatusConsumed)
	}
	if consumeResp.ConsumedAt == "" {
		t.Error("consumeResp.ConsumedAt should not be empty")
	}

	// 关键断言：直接查询数据库验证底层存储状态已被正确更新为 consumed，consumed_at 非空
	var dbPinAfterConsume entity.Pin
	if err := db.Where("id = ?", pinID).First(&dbPinAfterConsume).Error; err != nil {
		t.Fatalf("failed to query pin from db after consume: %v", err)
	}
	if dbPinAfterConsume.Status != bConst.PinStatusConsumed {
		t.Fatalf("DB status not consumed: got %s", dbPinAfterConsume.Status)
	}
	if dbPinAfterConsume.ConsumedAt == nil {
		t.Fatal("DB consumed_at is nil after consume")
	}

	// Step 6: 再次调用 PeekOldestPending，断言此时无待消费约束（NotFound）
	postConsumePeek, postConsumeErr := l.PeekOldestPending(ctx, project.ID)
	if postConsumePeek != nil {
		t.Fatalf("expected nil response after consume, got %v", postConsumePeek)
	}
	if postConsumeErr == nil {
		t.Fatal("expected error from PeekOldestPending after all pins consumed")
	}
	if postConsumeErr.GetErrorCode() != xError.NotFound {
		t.Fatalf("expected NotFound error code, got %v", postConsumeErr.GetErrorCode())
	}

	// 同时验证 repo.GetOldestPending 也返回 nil, nil
	postConsumeRepoPin, postConsumeRepoErr := l.repo.pin.GetOldestPending(ctx, project.ID)
	if postConsumeRepoErr != nil {
		t.Fatalf("repo.GetOldestPending failed after consume: %v", postConsumeRepoErr)
	}
	if postConsumeRepoPin != nil {
		t.Fatalf("repo.GetOldestPending should be nil after consume, got %v", postConsumeRepoPin)
	}

	// 补充校验：通过 ID 依然可以 Peek 到已消费的 Pin 历史记录
	consumedHistory, peekHistoryErr := l.Peek(ctx, pinID)
	if peekHistoryErr != nil {
		t.Fatalf("Peek by ID for consumed pin failed: %v", peekHistoryErr)
	}
	if consumedHistory.Status != bConst.PinStatusConsumed {
		t.Errorf("consumedHistory.Status = %s, want %s", consumedHistory.Status, bConst.PinStatusConsumed)
	}
	if consumedHistory.ConsumedAt == "" {
		t.Error("consumedHistory.ConsumedAt should show timestamp")
	}
}

// TestPinLifecycle_FIFOOrdering 验证多条待处理约束时的 FIFO 队列顺序：
// 1. 插入多条 pending 约束（不同创建时间）；
// 2. Peek 始终返回最早创建的约束；
// 3. 消费队首后，Peek 顺延至下一条约束；
// 4. 全部消费后，Peek 返回 NotFound。
func TestPinLifecycle_FIFOOrdering(t *testing.T) {
	l, db, project := setupPinTestEnv(t)
	ctx := context.Background()

	t1 := time.Now().Add(-10 * time.Minute)
	t2 := time.Now().Add(-5 * time.Minute)

	pin1ID := xSnowflake.GenerateID(bConst.GenePin)
	pin1 := &entity.Pin{
		BaseEntity:    xModels.BaseEntity{ID: pin1ID, CreatedAt: t1, UpdatedAt: t1},
		FromProjectID: project.ID,
		ToProjectID:   project.ID,
		Title:         "第一条约束（旧）",
		Content:       "这是较早创建的约束内容",
		Status:        bConst.PinStatusPending,
		Priority:      bConst.PinPriorityLow,
	}
	pin2ID := xSnowflake.GenerateID(bConst.GenePin)
	pin2 := &entity.Pin{
		BaseEntity:    xModels.BaseEntity{ID: pin2ID, CreatedAt: t2, UpdatedAt: t2},
		FromProjectID: project.ID,
		ToProjectID:   project.ID,
		Title:         "第二条约束（新）",
		Content:       "这是较晚创建的约束内容",
		Status:        bConst.PinStatusPending,
		Priority:      bConst.PinPriorityHigh,
	}

	if err := db.Create(pin1).Error; err != nil {
		t.Fatalf("create pin1: %v", err)
	}
	if err := db.Create(pin2).Error; err != nil {
		t.Fatalf("create pin2: %v", err)
	}

	// 1. Peek 最早的一条，应为 pin1
	peek1, xErr := l.PeekOldestPending(ctx, project.ID)
	if xErr != nil {
		t.Fatalf("peek oldest failed: %v", xErr)
	}
	if peek1.ID != pin1ID {
		t.Fatalf("expected pin1 (%s) as oldest, got %s", pin1ID, peek1.ID)
	}

	// 2. 消费队首（pin1）
	c1, xErr := l.Consume(ctx, project.ID, nil)
	if xErr != nil {
		t.Fatalf("consume pin1 failed: %v", xErr)
	}
	if c1.ID != pin1ID {
		t.Fatalf("consumed id = %s, want %s", c1.ID, pin1ID)
	}

	// 3. 再次 Peek，应该顺延到 pin2
	peek2, xErr := l.PeekOldestPending(ctx, project.ID)
	if xErr != nil {
		t.Fatalf("peek next oldest failed: %v", xErr)
	}
	if peek2.ID != pin2ID {
		t.Fatalf("expected pin2 (%s), got %s", pin2ID, peek2.ID)
	}

	// 4. 精确消费 pin2
	c2, xErr := l.Consume(ctx, project.ID, &pin2ID)
	if xErr != nil {
		t.Fatalf("consume pin2 by id failed: %v", xErr)
	}
	if c2.ID != pin2ID {
		t.Fatalf("consumed id = %s, want %s", c2.ID, pin2ID)
	}

	// 5. 再次 Peek，队列为空
	peekEmpty, xErr := l.PeekOldestPending(ctx, project.ID)
	if peekEmpty != nil || xErr == nil || xErr.GetErrorCode() != xError.NotFound {
		t.Fatalf("expected NotFound on exhausted queue, got resp=%v, err=%v", peekEmpty, xErr)
	}
}
