package logic

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	"github.com/redis/go-redis/v9"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"github.com/xiaolfeng/Lumina/internal/repository"
	"github.com/xiaolfeng/Lumina/internal/service"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestRepoWikiLogic 构造测试用 RepoWikiLogic 实例（in-memory SQLite + miniredis）
func setupTestRepoWikiLogic(t *testing.T) (*RepoWikiLogic, *gorm.DB, *miniredis.Miniredis, string) {
	t.Helper()

	// in-memory SQLite (raw table creation to avoid timestamptz incompatibility)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	createTestTables(t, db)

	// miniredis
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	// temp storage path
	tmpDir, err := os.MkdirTemp("", "repowiki-logic-test-*")
	if err != nil {
		t.Fatalf("MkdirTemp failed: %v", err)
	}
	t.Setenv("REPOWIKI_STORAGE_PATH", tmpDir)

	l := &RepoWikiLogic{
		logic: logic{
			log: xLog.WithName(xLog.NamedLOGC, "RepoWikiLogicTest"),
		},
		repo: repowikiRepo{
			config:       repository.NewRepoWikiConfigRepo(db, rdb),
			version:      repository.NewWikiVersionRepo(db, rdb),
			webhookEvent: repository.NewWebhookEventRepo(db, rdb),
		},
		svc: repowikiSvc{
			storage: service.NewWikiStorageService(),
		},
		semaphore: make(chan struct{}, 1),
	}

	return l, db, mr, tmpDir
}

// seedConfigAndVersions 创建测试用 config + versions，返回 configID 和 versionIDs
func seedConfigAndVersions(t *testing.T, db *gorm.DB, statuses []string) (xSnowflake.SnowflakeID, []xSnowflake.SnowflakeID) {
	t.Helper()

	configID := xSnowflake.GenerateID(39)
	config := &entity.RepoWikiConfig{
		BaseEntity:      xBaseEntityWithID(configID),
		ProjectID:       xSnowflake.GenerateID(32),
		GitURL:          "https://github.com/test/repo.git",
		DefaultBranch:   "main",
		DefaultLanguage: "zh",
		Status:          "completed",
		WebhookToken:    "test-token",
		WebhookSecret:   "test-secret",
	}
	if err := db.Create(config).Error; err != nil {
		t.Fatalf("failed to seed config: %v", err)
	}

	versionIDs := make([]xSnowflake.SnowflakeID, 0, len(statuses))
	for _, status := range statuses {
		vid := xSnowflake.GenerateID(40)
		version := &entity.WikiVersion{
			BaseEntity: xBaseEntityWithID(vid),
			ConfigID:   configID,
			Branch:     "main",
			Language:   "zh",
			Status:     status,
		}
		if err := db.Create(version).Error; err != nil {
			t.Fatalf("failed to seed version (status=%s): %v", status, err)
		}
		versionIDs = append(versionIDs, vid)
	}

	return configID, versionIDs
}

// seedWebhookEvent 创建测试用 WebhookEvent 行
func seedWebhookEvent(t *testing.T, db *gorm.DB, configID xSnowflake.SnowflakeID) {
	t.Helper()
	event := &entity.WebhookEvent{
		BaseEntity: xBaseEntityWithID(xSnowflake.GenerateID(43)),
		ConfigID:   &configID,
		Provider:   "github",
		EventType:  "push",
		Branch:     "main",
		Status:     "received",
		ReceivedAt: time.Now(),
	}
	if err := db.Create(event).Error; err != nil {
		t.Fatalf("failed to seed webhook event: %v", err)
	}
}

// xBaseEntityWithID 构造指定 ID 的 BaseEntity
func xBaseEntityWithID(id xSnowflake.SnowflakeID) xModels.BaseEntity {
	return xModels.BaseEntity{ID: id}
}

// createTestTables 手动创建 SQLite 表（用 datetime 替代 timestamptz，避免 time.Time 扫描错误）
func createTestTables(t *testing.T, db *gorm.DB) {
	t.Helper()

	tables := []string{"repo_wiki_configs", "wiki_versions", "webhook_events"}
	for _, tbl := range tables {
		db.Migrator().DropTable(tbl)
	}

	if err := db.Exec(`CREATE TABLE repo_wiki_configs (
		id INTEGER PRIMARY KEY,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		project_id INTEGER NOT NULL,
		git_url TEXT NOT NULL,
		default_branch TEXT NOT NULL DEFAULT 'main',
		local_path TEXT,
		default_language TEXT NOT NULL DEFAULT 'zh',
		ssh_key_id INTEGER,
		wiki_password_hash TEXT,
		status TEXT NOT NULL DEFAULT 'pending',
		selected_version_id INTEGER,
		last_accessed_at DATETIME,
		webhook_token TEXT,
		webhook_secret TEXT,
		webhook_branches TEXT,
		custom_prompt TEXT
	)`).Error; err != nil {
		t.Fatalf("failed to create repo_wiki_configs: %v", err)
	}

	if err := db.Exec(`CREATE TABLE wiki_versions (
		id INTEGER PRIMARY KEY,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		config_id INTEGER NOT NULL,
		commit_hash TEXT NOT NULL,
		branch TEXT NOT NULL DEFAULT 'main',
		language TEXT NOT NULL DEFAULT 'zh',
		llm_model TEXT,
		llm_provider TEXT,
		status TEXT NOT NULL DEFAULT 'pending',
		current_stage TEXT,
		file_count INTEGER DEFAULT 0,
		token_count INTEGER DEFAULT 0,
		duration_ms INTEGER DEFAULT 0,
		storage_path TEXT,
		wiki_path TEXT,
		architecture_path TEXT,
		explore_outputs_path TEXT,
		file_scan_path TEXT,
		dep_summary_path TEXT,
		manifest_path TEXT,
		error_msg TEXT,
		retry_count INTEGER NOT NULL DEFAULT 0,
		started_at DATETIME,
		completed_at DATETIME
	)`).Error; err != nil {
		t.Fatalf("failed to create wiki_versions: %v", err)
	}

	if err := db.Exec(`CREATE TABLE webhook_events (
		id INTEGER PRIMARY KEY,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		config_id INTEGER,
		provider TEXT NOT NULL,
		event_type TEXT NOT NULL,
		branch TEXT,
		commit_before TEXT,
		commit_after TEXT,
		changed_count INTEGER DEFAULT 0,
		status TEXT NOT NULL DEFAULT 'received',
		reason TEXT,
		version_id INTEGER,
		response_code INTEGER,
		received_at DATETIME NOT NULL,
		processed_at DATETIME
	)`).Error; err != nil {
		t.Fatalf("failed to create webhook_events: %v", err)
	}
}

// createFakeDirs 创建假的 version 目录和 git 缓存目录
func createFakeDirs(t *testing.T, storage *service.WikiStorageService, configID xSnowflake.SnowflakeID, versionIDs []xSnowflake.SnowflakeID) {
	t.Helper()
	for _, vid := range versionIDs {
		dir := storage.GetVersionPath(vid.Int64())
		if err := os.MkdirAll(filepath.Join(dir, "raw"), 0755); err != nil {
			t.Fatalf("failed to create version dir %s: %v", dir, err)
		}
	}
	repoDir := storage.GetRepoPath(configID.Int64())
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatalf("failed to create repo dir %s: %v", repoDir, err)
	}
}

// assertDirNotExists 断言目录不存在
func assertDirNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("expected dir %s to not exist, but it does (err=%v)", path, err)
	}
}

// assertDirExists 断言目录存在
func assertDirExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("expected dir %s to exist, but it doesn't", path)
	}
}

// ──────────────────────────────────────────────────────────────────────
// Test Cases
// ──────────────────────────────────────────────────────────────────────

// TestRepoWikiLogic_DeleteConfig_Cascade 级联删除全部终态版本
func TestRepoWikiLogic_DeleteConfig_Cascade(t *testing.T) {
	l, db, mr, tmpDir := setupTestRepoWikiLogic(t)
	defer mr.Close()
	defer os.RemoveAll(tmpDir)
	ctx := context.Background()

	// seed: config + 3 versions (all terminal) + webhook event
	configID, versionIDs := seedConfigAndVersions(t, db, []string{"completed", "failed", "cancelled"})
	seedWebhookEvent(t, db, configID)

	// create fake dirs
	createFakeDirs(t, l.svc.storage, configID, versionIDs)

	// verify dirs exist before delete
	for _, vid := range versionIDs {
		assertDirExists(t, l.svc.storage.GetVersionPath(vid.Int64()))
	}
	assertDirExists(t, l.svc.storage.GetRepoPath(configID.Int64()))

	// delete
	xErr := l.DeleteConfig(ctx, configID)
	if xErr != nil {
		t.Fatalf("DeleteConfig failed: %v", xErr)
	}

	// assert config row deleted
	var count int64
	db.Model(&entity.RepoWikiConfig{}).Where("id = ?", configID).Count(&count)
	if count != 0 {
		t.Errorf("expected config row deleted, got count=%d", count)
	}

	// assert version rows deleted
	db.Model(&entity.WikiVersion{}).Where("config_id = ?", configID).Count(&count)
	if count != 0 {
		t.Errorf("expected version rows deleted, got count=%d", count)
	}

	// assert webhook event rows deleted
	db.Model(&entity.WebhookEvent{}).Where("config_id = ?", configID).Count(&count)
	if count != 0 {
		t.Errorf("expected webhook event rows deleted, got count=%d", count)
	}

	// assert file dirs deleted
	for _, vid := range versionIDs {
		assertDirNotExists(t, l.svc.storage.GetVersionPath(vid.Int64()))
	}
	assertDirNotExists(t, l.svc.storage.GetRepoPath(configID.Int64()))
}

// TestRepoWikiLogic_DeleteConfig_RejectedByActiveTask 有非终态版本时拒绝删除
func TestRepoWikiLogic_DeleteConfig_RejectedByActiveTask(t *testing.T) {
	l, db, mr, tmpDir := setupTestRepoWikiLogic(t)
	defer mr.Close()
	defer os.RemoveAll(tmpDir)
	ctx := context.Background()

	// seed: config + 1 pending + 1 completed
	configID, versionIDs := seedConfigAndVersions(t, db, []string{"pending", "completed"})
	seedWebhookEvent(t, db, configID)
	createFakeDirs(t, l.svc.storage, configID, versionIDs)

	// delete should be rejected
	xErr := l.DeleteConfig(ctx, configID)
	if xErr == nil {
		t.Fatal("expected error for active task, got nil")
	}

	// assert BusinessError
	if xErr.GetErrorCode() != xError.BusinessError {
		t.Errorf("expected BusinessError, got %v", xErr.GetErrorCode())
	}

	// assert config row still exists
	var count int64
	db.Model(&entity.RepoWikiConfig{}).Where("id = ?", configID).Count(&count)
	if count != 1 {
		t.Errorf("expected config row to still exist, got count=%d", count)
	}

	// assert version rows still exist
	db.Model(&entity.WikiVersion{}).Where("config_id = ?", configID).Count(&count)
	if count != 2 {
		t.Errorf("expected 2 version rows to still exist, got count=%d", count)
	}

	// assert file dirs still exist
	for _, vid := range versionIDs {
		assertDirExists(t, l.svc.storage.GetVersionPath(vid.Int64()))
	}
	assertDirExists(t, l.svc.storage.GetRepoPath(configID.Int64()))
}

// TestRepoWikiLogic_DeleteConfig_NotFound 删除不存在的配置返回错误不 panic
func TestRepoWikiLogic_DeleteConfig_NotFound(t *testing.T) {
	l, _, mr, tmpDir := setupTestRepoWikiLogic(t)
	defer mr.Close()
	defer os.RemoveAll(tmpDir)
	ctx := context.Background()

	nonExistentID := xSnowflake.GenerateID(39)
	xErr := l.DeleteConfig(ctx, nonExistentID)
	if xErr == nil {
		t.Fatal("expected error for non-existent config, got nil")
	}
}

// TestRepoWikiLogic_CleanFailedVersions 清理失败版本，保留 completed/pending
func TestRepoWikiLogic_CleanFailedVersions(t *testing.T) {
	l, db, mr, tmpDir := setupTestRepoWikiLogic(t)
	defer mr.Close()
	defer os.RemoveAll(tmpDir)
	ctx := context.Background()

	// seed: config + 2 failed + 1 completed + 1 pending
	configID, versionIDs := seedConfigAndVersions(t, db, []string{"failed", "failed", "completed", "pending"})

	// create fake dirs for all versions
	createFakeDirs(t, l.svc.storage, configID, versionIDs)

	// verify all dirs exist before clean
	for _, vid := range versionIDs {
		assertDirExists(t, l.svc.storage.GetVersionPath(vid.Int64()))
	}

	// clean failed versions
	n, xErr := l.CleanFailedVersions(ctx, configID)
	if xErr != nil {
		t.Fatalf("CleanFailedVersions failed: %v", xErr)
	}
	if n != 2 {
		t.Errorf("expected cleaned=2, got %d", n)
	}

	// assert failed rows deleted
	var failedCount int64
	db.Model(&entity.WikiVersion{}).Where("config_id = ? AND status = ?", configID, "failed").Count(&failedCount)
	if failedCount != 0 {
		t.Errorf("expected 0 failed rows, got %d", failedCount)
	}

	// assert completed + pending rows preserved
	var remainingCount int64
	db.Model(&entity.WikiVersion{}).Where("config_id = ?", configID).Count(&remainingCount)
	if remainingCount != 2 {
		t.Errorf("expected 2 remaining rows (completed+pending), got %d", remainingCount)
	}

	// assert config row still exists
	var configCount int64
	db.Model(&entity.RepoWikiConfig{}).Where("id = ?", configID).Count(&configCount)
	if configCount != 1 {
		t.Errorf("expected config row to still exist, got count=%d", configCount)
	}

	// assert failed version dirs deleted (first 2 versionIDs)
	for i := 0; i < 2; i++ {
		assertDirNotExists(t, l.svc.storage.GetVersionPath(versionIDs[i].Int64()))
	}
	// assert completed + pending version dirs preserved (last 2 versionIDs)
	for i := 2; i < 4; i++ {
		assertDirExists(t, l.svc.storage.GetVersionPath(versionIDs[i].Int64()))
	}
}

// TestRepoWikiLogic_CleanFailedVersions_NoneFound 无 failed 版本时返回 cleaned=0 不报错
func TestRepoWikiLogic_CleanFailedVersions_NoneFound(t *testing.T) {
	l, db, mr, tmpDir := setupTestRepoWikiLogic(t)
	defer mr.Close()
	defer os.RemoveAll(tmpDir)
	ctx := context.Background()

	// seed: config + 1 completed + 1 pending (no failed)
	configID, _ := seedConfigAndVersions(t, db, []string{"completed", "pending"})

	n, xErr := l.CleanFailedVersions(ctx, configID)
	if xErr != nil {
		t.Fatalf("expected nil error for no failed versions, got: %v", xErr)
	}
	if n != 0 {
		t.Errorf("expected cleaned=0, got %d", n)
	}

	// assert all version rows preserved
	var count int64
	db.Model(&entity.WikiVersion{}).Where("config_id = ?", configID).Count(&count)
	if count != 2 {
		t.Errorf("expected 2 version rows preserved, got %d", count)
	}
}

// ──────────────────────────────────────────────────────────────────────
// KeepLatestVersions tests
// ──────────────────────────────────────────────────────────────────────

// seedConfigWithSelected 创建 config 并设置 SelectedVersionID
func seedConfigWithSelected(t *testing.T, db *gorm.DB, selectedVersionID *xSnowflake.SnowflakeID) xSnowflake.SnowflakeID {
	t.Helper()
	configID := xSnowflake.GenerateID(39)
	config := &entity.RepoWikiConfig{
		BaseEntity:        xBaseEntityWithID(configID),
		ProjectID:         xSnowflake.GenerateID(32),
		GitURL:            "https://github.com/test/repo.git",
		DefaultBranch:     "main",
		DefaultLanguage:   "zh",
		Status:            "completed",
		WebhookToken:      "test-token",
		WebhookSecret:     "test-secret",
		SelectedVersionID: selectedVersionID,
	}
	if err := db.Create(config).Error; err != nil {
		t.Fatalf("failed to seed config: %v", err)
	}
	return configID
}

// seedVersionWithTimestamp 创建带显式 created_at 的版本（控制排序）
func seedVersionWithTimestamp(t *testing.T, db *gorm.DB, configID xSnowflake.SnowflakeID, status string, createdAt time.Time) xSnowflake.SnowflakeID {
	t.Helper()
	vid := xSnowflake.GenerateID(40)
	version := &entity.WikiVersion{
		BaseEntity: xBaseEntityWithID(vid),
		ConfigID:   configID,
		Branch:     "main",
		Language:   "zh",
		Status:     status,
	}
	version.CreatedAt = createdAt
	version.UpdatedAt = createdAt
	if err := db.Create(version).Error; err != nil {
		t.Fatalf("failed to seed version (status=%s): %v", status, err)
	}
	return vid
}

// TestRepoWikiLogic_KeepLatestVersions_SelectedExists SelectedVersionID 存在时保留它
func TestRepoWikiLogic_KeepLatestVersions_SelectedExists(t *testing.T) {
	l, db, mr, tmpDir := setupTestRepoWikiLogic(t)
	defer mr.Close()
	defer os.RemoveAll(tmpDir)
	ctx := context.Background()

	base := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	configID := seedConfigWithSelected(t, db, nil)
	v1 := seedVersionWithTimestamp(t, db, configID, "completed", base.Add(1*time.Second))
	v2 := seedVersionWithTimestamp(t, db, configID, "completed", base.Add(2*time.Second))
	v3 := seedVersionWithTimestamp(t, db, configID, "failed", base.Add(3*time.Second))
	v4 := seedVersionWithTimestamp(t, db, configID, "pending", base.Add(4*time.Second))

	// set SelectedVersionID = v2
	db.Model(&entity.RepoWikiConfig{}).Where("id = ?", configID).Update("selected_version_id", v2)

	createFakeDirs(t, l.svc.storage, configID, []xSnowflake.SnowflakeID{v1, v2, v3, v4})

	cleaned, skipped, keptID, xErr := l.KeepLatestVersions(ctx, configID)
	if xErr != nil {
		t.Fatalf("KeepLatestVersions failed: %v", xErr)
	}
	if *keptID != v2 {
		t.Errorf("expected keptID=v2, got %d", keptID.Int64())
	}
	if cleaned != 2 {
		t.Errorf("expected cleaned=2 (V1+V3), got %d", cleaned)
	}
	if skipped != 1 {
		t.Errorf("expected skipped=1 (V4 pending), got %d", skipped)
	}

	// V2 and V4 rows preserved
	var count int64
	db.Model(&entity.WikiVersion{}).Where("id = ?", v2).Count(&count)
	if count != 1 {
		t.Errorf("expected V2 row preserved, got count=%d", count)
	}
	db.Model(&entity.WikiVersion{}).Where("id = ?", v4).Count(&count)
	if count != 1 {
		t.Errorf("expected V4 row preserved, got count=%d", count)
	}
	db.Model(&entity.WikiVersion{}).Where("id = ?", v1).Count(&count)
	if count != 0 {
		t.Errorf("expected V1 row deleted, got count=%d", count)
	}
	db.Model(&entity.WikiVersion{}).Where("id = ?", v3).Count(&count)
	if count != 0 {
		t.Errorf("expected V3 row deleted, got count=%d", count)
	}

	// SelectedVersionID unchanged (still v2)
	var cfg entity.RepoWikiConfig
	db.Where("id = ?", configID).First(&cfg)
	if cfg.SelectedVersionID == nil || *cfg.SelectedVersionID != v2 {
		t.Errorf("expected SelectedVersionID unchanged=v2, got %v", cfg.SelectedVersionID)
	}
}

// TestRepoWikiLogic_KeepLatestVersions_NoSelected_KeepLatestCompleted 无 SelectedVersionID 时保留最近 completed
func TestRepoWikiLogic_KeepLatestVersions_NoSelected_KeepLatestCompleted(t *testing.T) {
	l, db, mr, tmpDir := setupTestRepoWikiLogic(t)
	defer mr.Close()
	defer os.RemoveAll(tmpDir)
	ctx := context.Background()

	configID := seedConfigWithSelected(t, db, nil)
	base := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	v1 := seedVersionWithTimestamp(t, db, configID, "completed", base.Add(1*time.Second))
	v2 := seedVersionWithTimestamp(t, db, configID, "completed", base.Add(2*time.Second))
	v3 := seedVersionWithTimestamp(t, db, configID, "failed", base.Add(3*time.Second))

	createFakeDirs(t, l.svc.storage, configID, []xSnowflake.SnowflakeID{v1, v2, v3})

	cleaned, skipped, keptID, xErr := l.KeepLatestVersions(ctx, configID)
	if xErr != nil {
		t.Fatalf("KeepLatestVersions failed: %v", xErr)
	}
	// latest completed by created_at DESC = v2 (created at base+2s, v3 is failed)
	if *keptID != v2 {
		t.Errorf("expected keptID=v2 (latest completed), got %d", keptID.Int64())
	}
	if cleaned != 2 {
		t.Errorf("expected cleaned=2 (V1+V3), got %d", cleaned)
	}
	if skipped != 0 {
		t.Errorf("expected skipped=0, got %d", skipped)
	}

	// SelectedVersionID updated to v2
	var cfg entity.RepoWikiConfig
	db.Where("id = ?", configID).First(&cfg)
	if cfg.SelectedVersionID == nil || *cfg.SelectedVersionID != v2 {
		t.Errorf("expected SelectedVersionID updated to v2, got %v", cfg.SelectedVersionID)
	}
}

// TestRepoWikiLogic_KeepLatestVersions_NoCompleted 无 completed 时保留最新版本不写 SelectedVersionID
func TestRepoWikiLogic_KeepLatestVersions_NoCompleted(t *testing.T) {
	l, db, mr, tmpDir := setupTestRepoWikiLogic(t)
	defer mr.Close()
	defer os.RemoveAll(tmpDir)
	ctx := context.Background()

	configID := seedConfigWithSelected(t, db, nil)
	base := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	v1 := seedVersionWithTimestamp(t, db, configID, "failed", base.Add(1*time.Second))
	v2 := seedVersionWithTimestamp(t, db, configID, "cancelled", base.Add(2*time.Second))

	createFakeDirs(t, l.svc.storage, configID, []xSnowflake.SnowflakeID{v1, v2})

	cleaned, skipped, keptID, xErr := l.KeepLatestVersions(ctx, configID)
	if xErr != nil {
		t.Fatalf("KeepLatestVersions failed: %v", xErr)
	}
	// no completed → keep versions[0] by created_at DESC = v2 (latest)
	if *keptID != v2 {
		t.Errorf("expected keptID=v2 (latest by created_at), got %d", keptID.Int64())
	}
	if cleaned != 1 {
		t.Errorf("expected cleaned=1 (V1), got %d", cleaned)
	}
	if skipped != 0 {
		t.Errorf("expected skipped=0, got %d", skipped)
	}

	// SelectedVersionID not written (still nil)
	var cfg entity.RepoWikiConfig
	db.Where("id = ?", configID).First(&cfg)
	if cfg.SelectedVersionID != nil {
		t.Errorf("expected SelectedVersionID nil (no completed to select), got %d", cfg.SelectedVersionID.Int64())
	}
}

// TestRepoWikiLogic_KeepLatestVersions_SelectedNotFound SelectedVersionID 指向不存在的版本时回退到最近 completed
func TestRepoWikiLogic_KeepLatestVersions_SelectedNotFound(t *testing.T) {
	l, db, mr, tmpDir := setupTestRepoWikiLogic(t)
	defer mr.Close()
	defer os.RemoveAll(tmpDir)
	ctx := context.Background()

	nonExistentVID := xSnowflake.GenerateID(40)
	configID := seedConfigWithSelected(t, db, &nonExistentVID)
	base := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	v1 := seedVersionWithTimestamp(t, db, configID, "completed", base.Add(1*time.Second))
	v2 := seedVersionWithTimestamp(t, db, configID, "completed", base.Add(2*time.Second))

	createFakeDirs(t, l.svc.storage, configID, []xSnowflake.SnowflakeID{v1, v2})

	cleaned, skipped, keptID, xErr := l.KeepLatestVersions(ctx, configID)
	if xErr != nil {
		t.Fatalf("KeepLatestVersions failed: %v", xErr)
	}
	// SelectedVersionID=V99 not found → fallback to latest completed = v2
	if *keptID != v2 {
		t.Errorf("expected keptID=v2 (fallback to latest completed), got %d", keptID.Int64())
	}
	if cleaned != 1 {
		t.Errorf("expected cleaned=1 (V1), got %d", cleaned)
	}
	if skipped != 0 {
		t.Errorf("expected skipped=0, got %d", skipped)
	}

	// SelectedVersionID updated to v2
	var cfg entity.RepoWikiConfig
	db.Where("id = ?", configID).First(&cfg)
	if cfg.SelectedVersionID == nil || *cfg.SelectedVersionID != v2 {
		t.Errorf("expected SelectedVersionID updated to v2, got %v", cfg.SelectedVersionID)
	}
}

// TestRepoWikiLogic_KeepLatestVersions_ConfigNotFound 配置不存在时返回错误不 panic
func TestRepoWikiLogic_KeepLatestVersions_ConfigNotFound(t *testing.T) {
	l, _, mr, tmpDir := setupTestRepoWikiLogic(t)
	defer mr.Close()
	defer os.RemoveAll(tmpDir)
	ctx := context.Background()

	nonExistentID := xSnowflake.GenerateID(39)
	cleaned, skipped, keptID, xErr := l.KeepLatestVersions(ctx, nonExistentID)
	if xErr == nil {
		t.Fatal("expected error for non-existent config, got nil")
	}
	if cleaned != 0 || skipped != 0 {
		t.Errorf("expected 0,0 for non-existent config, got cleaned=%d skipped=%d", cleaned, skipped)
	}
	if keptID != nil {
		t.Errorf("expected nil keptVersionID, got %d", keptID.Int64())
	}
}
