package logic

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	apiPreview "github.com/xiaolfeng/Lumina/api/preview"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"github.com/xiaolfeng/Lumina/internal/repository"
	"github.com/xiaolfeng/Lumina/internal/service"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestIncrementPatchVersion(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"v1.0.0": "v1.0.1",
		"v1.1.9": "v1.1.10",
		"1.2.3":  "v1.2.4",
		"v2.0":   "v1.0.0",
		"":       "v1.0.0",
	}
	for input, want := range cases {
		if got := incrementPatchVersion(input); got != want {
			t.Errorf("incrementPatchVersion(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestIsHTMLFilename(t *testing.T) {
	t.Parallel()
	if !isHTMLFilename("index.html") || !isHTMLFilename("about.HTM") {
		t.Fatal("html filenames should be recognized")
	}
	if isHTMLFilename("style.css") {
		t.Fatal("css should not be treated as html entry")
	}
}

func TestFindHTMLEntry(t *testing.T) {
	t.Parallel()
	files := []*entity.PreviewFile{
		{Filename: "app.js", MimeType: bConst.PreviewMimeJS},
		{Filename: "index.html", MimeType: bConst.PreviewMimeHTML},
	}
	if got := findHTMLEntry(files); got != "index.html" {
		t.Fatalf("findHTMLEntry() = %q, want index.html", got)
	}
	if got := findHTMLEntry([]*entity.PreviewFile{{Filename: "app.js", MimeType: bConst.PreviewMimeJS}}); got != "" {
		t.Fatalf("findHTMLEntry(no html) = %q, want empty", got)
	}
}

func TestValidateSlug(t *testing.T) {
	t.Parallel()
	logic := &PagesLogic{}
	if xErr := logic.validateSlug(t.Context(), "design-system"); xErr != nil {
		t.Fatalf("valid slug rejected: %v", xErr)
	}
	for _, slug := range []string{"", "Bad_Slug", "has space", "UPPER", "a/b"} {
		if xErr := logic.validateSlug(t.Context(), slug); xErr == nil {
			t.Fatalf("invalid slug %q accepted", slug)
		}
	}
}

func TestValidateFilename(t *testing.T) {
	t.Parallel()
	if err := validateFilename("index.html"); err != nil {
		t.Fatalf("valid filename rejected: %v", err)
	}
	for _, name := range []string{"", "assets/app.js", "..", "foo\\bar", "a..b"} {
		if err := validateFilename(name); err == nil {
			t.Fatalf("invalid filename %q accepted", name)
		}
	}
}

// ── Pages 密码门限流与归档可见性 ──

const testAuthProjectID int64 = 100000000000000010

// setupPagesAuthTest 构造带 in-memory SQLite 的 PagesLogic（独立限流器，避免用例间计数串扰）
func setupPagesAuthTest(t *testing.T) (*PagesLogic, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "pages.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(&entity.Project{}, &entity.Page{}); err != nil {
		t.Fatal(err)
	}

	return &PagesLogic{
		logic: logic{log: xLog.WithName(xLog.NamedLOGC, "PagesLogicTest")},
		repo: pagesRepo{
			page:    repository.NewPageRepo(db),
			project: repository.NewProjectRepo(db, nil),
			info:    repository.NewInfoRepo(db),
		},
		authToken: service.NewPageAuthTokenService(),
		authLimit: newPagesAuthLimiter(),
	}, db
}

// seedAuthPage 创建项目与 password 模式页面
func seedAuthPage(t *testing.T, db *gorm.DB, slug, status, passwordHash string) {
	t.Helper()

	project := &entity.Project{WorkspaceID: 100000000000000001, Name: "auth-test"}
	project.ID = xSnowflake.SnowflakeID(testAuthProjectID)
	if err := db.Create(project).Error; err != nil {
		t.Fatal(err)
	}
	page := &entity.Page{
		BaseEntity:   xModels.BaseEntity{ID: xSnowflake.SnowflakeID(testAuthProjectID + 10)},
		ProjectID:    project.ID,
		Slug:         slug,
		Title:        "auth page",
		Status:       status,
		AccessMode:   bConst.PageAccessModePassword,
		PasswordHash: passwordHash,
	}
	if err := db.Create(page).Error; err != nil {
		t.Fatal(err)
	}
}

// TestPagesUnlockRateLimitLockedAfterFailures 连续失败达到阈值后锁定，锁定期内即使密码正确也 429
func TestPagesUnlockRateLimitLockedAfterFailures(t *testing.T) {
	l, db := setupPagesAuthTest(t)
	ctx := context.Background()

	hash, err := service.HashPassword("correct-password")
	if err != nil {
		t.Fatal(err)
	}
	seedAuthPage(t, db, "locked", bConst.PageStatusPublished, hash)

	// 连续失败 pagesAuthMaxFailures 次触发锁定
	for i := 0; i < pagesAuthMaxFailures; i++ {
		_, _, _, xErr := l.Unlock(ctx, "auth-test", "locked", "wrong-password")
		if xErr == nil || xErr.GetErrorCode() != xError.Unauthorized {
			t.Fatalf("第 %d 次失败应返回密码错误，got %v", i+1, xErr)
		}
	}

	// 锁定期内即使密码正确也 429
	_, _, _, xErr := l.Unlock(ctx, "auth-test", "locked", "correct-password")
	if xErr == nil || xErr.GetErrorCode() != xError.TooManyRequests {
		t.Fatalf("锁定后正确密码应返回 TooManyRequests，got %v", xErr)
	}
}

// TestPagesUnlockRateLimitResetOnSuccess 解锁成功后失败计数清零，不再累计历史失败
func TestPagesUnlockRateLimitResetOnSuccess(t *testing.T) {
	l, db := setupPagesAuthTest(t)
	ctx := context.Background()

	hash, err := service.HashPassword("correct-password")
	if err != nil {
		t.Fatal(err)
	}
	seedAuthPage(t, db, "reset", bConst.PageStatusPublished, hash)

	// 失败 maxFailures-1 次（未达锁定阈值）后成功解锁
	for i := 0; i < pagesAuthMaxFailures-1; i++ {
		if _, _, _, xErr := l.Unlock(ctx, "auth-test", "reset", "wrong"); xErr == nil {
			t.Fatal("expected password error")
		}
	}
	pageID, token, maxAge, xErr := l.Unlock(ctx, "auth-test", "reset", "correct-password")
	if xErr != nil || token == "" || maxAge <= 0 {
		t.Fatalf("expected successful unlock, got token=%q maxAge=%d xErr=%v", token, maxAge, xErr)
	}
	if pageID != testAuthProjectID+10 {
		t.Fatalf("unexpected pageID %d", pageID)
	}

	// 计数已清零：再次失败 maxFailures-1 次后仍可成功（若未清零会累计到阈值触发 429）
	for i := 0; i < pagesAuthMaxFailures-1; i++ {
		if _, _, _, xErr := l.Unlock(ctx, "auth-test", "reset", "wrong"); xErr == nil {
			t.Fatal("expected password error")
		}
	}
	if _, _, _, xErr := l.Unlock(ctx, "auth-test", "reset", "correct-password"); xErr != nil {
		t.Fatalf("counter should have been reset on success: %v", xErr)
	}
}

// TestPagesAuthLimiterConcurrent 并发调用下无数据竞争（mutex 保护），最终状态一致
func TestPagesAuthLimiterConcurrent(t *testing.T) {
	t.Parallel()

	limiter := newPagesAuthLimiter()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			limiter.allow(1)
			limiter.recordFailure(1)
			limiter.reset(2)
		}()
	}
	wg.Wait()

	// page 2 不断被 reset，永远允许
	if !limiter.allow(2) {
		t.Fatal("page 2 should stay allowed after concurrent resets")
	}
	// 串行复现锁定与解除路径
	locked := false
	for i := 0; i < pagesAuthMaxFailures; i++ {
		limiter.recordFailure(3)
		if !limiter.allow(3) {
			locked = true
		}
	}
	if !locked || limiter.allow(3) {
		t.Fatal("page 3 should be locked after max failures")
	}
	limiter.reset(3)
	if !limiter.allow(3) {
		t.Fatal("page 3 should be allowed after reset")
	}
}

// TestPagesArchivedPageAuthNotFound 归档页在公开端点一律 404：LookupAuth / CheckAuth / Unlock 均不可探测
func TestPagesArchivedPageAuthNotFound(t *testing.T) {
	l, db := setupPagesAuthTest(t)
	ctx := context.Background()

	hash, err := service.HashPassword("correct-password")
	if err != nil {
		t.Fatal(err)
	}
	seedAuthPage(t, db, "archived", bConst.PageStatusArchived, hash)

	// LookupAuth → NotFound
	_, _, lookupErr := l.LookupAuth(ctx, "auth-test", "archived")
	if lookupErr == nil {
		t.Fatal("LookupAuth on archived page should fail")
	}
	var lookupXErr *xError.Error
	if !errors.As(lookupErr, &lookupXErr) || lookupXErr.GetErrorCode() != xError.NotFound {
		t.Fatalf("expected NotFound from LookupAuth, got %v", lookupErr)
	}

	// CheckAuth → NotFound
	if _, xErr := l.CheckAuth(ctx, "auth-test", "archived", ""); xErr == nil || xErr.GetErrorCode() != xError.NotFound {
		t.Fatalf("CheckAuth on archived page should 404, got %v", xErr)
	}

	// Unlock → NotFound（即使密码正确也不可解锁，且不消耗限流计数）
	if _, _, _, xErr := l.Unlock(ctx, "auth-test", "archived", "correct-password"); xErr == nil || xErr.GetErrorCode() != xError.NotFound {
		t.Fatalf("Unlock on archived page should 404, got %v", xErr)
	}
}

// TestPagesPublishedPageAuthFlow published 页面密码门正常流转：LookupAuth 返回哈希，CheckAuth 标记必填，Unlock 签发 Token
func TestPagesPublishedPageAuthFlow(t *testing.T) {
	l, db := setupPagesAuthTest(t)
	ctx := context.Background()

	hash, err := service.HashPassword("correct-password")
	if err != nil {
		t.Fatal(err)
	}
	seedAuthPage(t, db, "published", bConst.PageStatusPublished, hash)

	pageID, passwordHash, lookupErr := l.LookupAuth(ctx, "auth-test", "published")
	if lookupErr != nil || pageID != testAuthProjectID+10 || passwordHash != hash {
		t.Fatalf("LookupAuth = (%d, %q, %v), want (%d, hash, nil)", pageID, passwordHash, lookupErr, testAuthProjectID+10)
	}

	check, xErr := l.CheckAuth(ctx, "auth-test", "published", "")
	if xErr != nil || !check.PasswordRequired || check.Authenticated {
		t.Fatalf("CheckAuth = %+v, xErr = %v; want password required and unauthenticated", check, xErr)
	}

	_, token, _, xErr := l.Unlock(ctx, "auth-test", "published", "correct-password")
	if xErr != nil || token == "" {
		t.Fatalf("Unlock failed: token=%q xErr=%v", token, xErr)
	}
	if !l.authToken.ValidateToken(token, testAuthProjectID+10) {
		t.Fatal("unlocked token should validate for the page")
	}
}

// ── Pages 版本解析与晋升 / Fork ──

const testPromoteProjectID int64 = 200000000000000001

// setupPagesPromoteTest 构造带 in-memory SQLite + miniredis 的 PagesLogic（含页面/版本/文件与 Preview 会话表）
func setupPagesPromoteTest(t *testing.T) (*PagesLogic, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "pages.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(&entity.Project{}, &entity.Page{}, &entity.PageVersion{}, &entity.PageFile{},
		&entity.PreviewSession{}, &entity.PreviewFile{}, &entity.Info{}); err != nil {
		t.Fatal(err)
	}
	// ProjectRepo 为 Cache-First，RDB 为 nil 时 GetByID 会 panic，按现有测试模式注入 miniredis
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	return &PagesLogic{
		logic: logic{log: xLog.WithName(xLog.NamedLOGC, "PagesLogicTest")},
		repo: pagesRepo{
			page:           repository.NewPageRepo(db),
			version:        repository.NewPageVersionRepo(db),
			file:           repository.NewPageFileRepo(db),
			previewSession: repository.NewPreviewSessionRepo(db),
			previewFile:    repository.NewPreviewFileRepo(db),
			project:        repository.NewProjectRepo(db, rdb),
			info:           repository.NewInfoRepo(db),
		},
		authToken: service.NewPageAuthTokenService(),
		authLimit: newPagesAuthLimiter(),
	}, db
}

// seedPromoteProject 创建项目（供版本解析 / 晋升 / Fork 用例复用）
func seedPromoteProject(t *testing.T, db *gorm.DB) *entity.Project {
	t.Helper()

	project := &entity.Project{WorkspaceID: 100000000000000001, Name: "promote-test"}
	project.ID = xSnowflake.SnowflakeID(testPromoteProjectID)
	if err := db.Create(project).Error; err != nil {
		t.Fatal(err)
	}
	return project
}

// seedPromotePageWithVersions 创建 published 页面与两个版本（v1.0.0 为生效指针，v1.1.0 为额外标签）
func seedPromotePageWithVersions(t *testing.T, db *gorm.DB, projectID xSnowflake.SnowflakeID) (*entity.Page, *entity.PageVersion, *entity.PageVersion) {
	t.Helper()

	page := &entity.Page{
		BaseEntity:      xModels.BaseEntity{ID: xSnowflake.SnowflakeID(testPromoteProjectID + 100)},
		ProjectID:       projectID,
		Slug:            "site",
		Title:           "promote page",
		Status:          bConst.PageStatusPublished,
		AccessMode:      bConst.PageAccessModePublic,
		LatestVersionID: xSnowflake.SnowflakeID(testPromoteProjectID + 101),
	}
	latest := &entity.PageVersion{
		BaseEntity:    xModels.BaseEntity{ID: page.LatestVersionID},
		PageID:        page.ID,
		Version:       "v1.0.0",
		EntryFilename: "index.html",
	}
	extra := &entity.PageVersion{
		BaseEntity:    xModels.BaseEntity{ID: xSnowflake.SnowflakeID(testPromoteProjectID + 102)},
		PageID:        page.ID,
		Version:       "v1.1.0",
		EntryFilename: "index.html",
	}
	for _, record := range []any{page, latest, extra} {
		if err := db.Create(record).Error; err != nil {
			t.Fatal(err)
		}
	}
	return page, latest, extra
}

// seedPromoteSession 创建 active 预览会话与 index.html 文件（无 SourcePageID，走新建页面分支）。
//
// ExpiresAt 留空：Promote/Fork 路径不依赖过期时间，且 mattn sqlite 驱动不解析
// timestamptz 列的字符串值（非 NULL 会触发 *time.Time 扫描错误）。
func seedPromoteSession(t *testing.T, db *gorm.DB, projectID xSnowflake.SnowflakeID, idOffset int64) *entity.PreviewSession {
	t.Helper()

	session := &entity.PreviewSession{
		BaseEntity: xModels.BaseEntity{ID: xSnowflake.SnowflakeID(testPromoteProjectID + idOffset)},
		ProjectID:  projectID,
		Title:      "draft",
		Hash:       generateSessionHash(xSnowflake.SnowflakeID(testPromoteProjectID + idOffset)),
		Status:     bConst.PreviewSessionStatusActive,
	}
	file := &entity.PreviewFile{
		BaseEntity: xModels.BaseEntity{ID: xSnowflake.SnowflakeID(testPromoteProjectID + idOffset + 1)},
		SessionID:  session.ID,
		Filename:   "index.html",
		MimeType:   bConst.PreviewMimeHTML,
		Content:    "<html><body>draft</body></html>",
		Size:       len("<html><body>draft</body></html>"),
	}
	if err := db.Create(session).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(file).Error; err != nil {
		t.Fatal(err)
	}
	return session
}

// TestNormalizeVersionLabel 版本标签规范化：合法语义化版本归一为 vX.Y.Z，垃圾输入一律拒绝
func TestNormalizeVersionLabel(t *testing.T) {
	t.Parallel()

	valid := map[string]string{
		"v1.0.0":     "v1.0.0",
		"1.2.3":      "v1.2.3",
		" v2.10.3\t": "v2.10.3",
	}
	for input, want := range valid {
		if got, ok := normalizeVersionLabel(input); !ok || got != want {
			t.Errorf("normalizeVersionLabel(%q) = (%q, %v), want (%q, true)", input, got, ok, want)
		}
	}

	invalid := []string{
		"",
		"latest",
		"main",
		"2026-09-16T12:00:00Z", // RFC3339 缓存戳：必须拒绝而非拼成 v2026-...
		"1757980800000",         // 毫秒时间戳
		"v1.0",
		"1",
		"1.0.0.0",
		"+1.0.0",
		"-1.0.0",
		"1.0.a",
		"vv1.0.0",
	}
	for _, input := range invalid {
		if got, ok := normalizeVersionLabel(input); ok {
			t.Errorf("normalizeVersionLabel(%q) = (%q, true), want rejected", input, got)
		}
	}
}

// TestResolveVersion 版本解析：合法标签精确查询（含无 v 前缀兼容），垃圾输入回退生效指针而非盲目拼 v 前缀
func TestResolveVersion(t *testing.T) {
	l, db := setupPagesPromoteTest(t)
	ctx := context.Background()

	project := seedPromoteProject(t, db)
	page, latest, extra := seedPromotePageWithVersions(t, db, project.ID)

	cases := []struct {
		name    string
		label   string
		wantID  xSnowflake.SnowflakeID
		wantErr bool
	}{
		{name: "空标签走生效指针", label: "", wantID: latest.ID},
		{name: "v前缀标签原样查询", label: "v1.1.0", wantID: extra.ID},
		{name: "无v前缀兼容映射", label: "1.1.0", wantID: extra.ID},
		{name: "RFC3339时间戳回退生效指针", label: "2026-09-16T12:00:00Z", wantID: latest.ID},
		{name: "随机串回退生效指针", label: "latest", wantID: latest.ID},
		{name: "合法但不存在的版本仍404", label: "v9.9.9", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			version, xErr := l.resolveVersion(ctx, page, tc.label)
			if tc.wantErr {
				if xErr == nil || xErr.GetErrorCode() != xError.NotFound {
					t.Fatalf("resolveVersion(%q) 期望 NotFound，got %v", tc.label, xErr)
				}
				return
			}
			if xErr != nil {
				t.Fatalf("resolveVersion(%q) failed: %v", tc.label, xErr)
			}
			if version.ID != tc.wantID {
				t.Fatalf("resolveVersion(%q) 命中版本 %d，want %d", tc.label, version.ID.Int64(), tc.wantID.Int64())
			}
		})
	}
}

// TestResolveVersionUnpublishedPage 未发布页面在回退生效指针时保持「页面尚未发布版本」的 404 语义
func TestResolveVersionUnpublishedPage(t *testing.T) {
	l, db := setupPagesPromoteTest(t)
	ctx := context.Background()

	project := seedPromoteProject(t, db)
	page, _, _ := seedPromotePageWithVersions(t, db, project.ID)
	page.LatestVersionID = 0

	_, xErr := l.resolveVersion(ctx, page, "2026-09-16T12:00:00Z")
	if xErr == nil || xErr.GetErrorCode() != xError.NotFound {
		t.Fatalf("垃圾输入回退到零值生效指针应返回 NotFound，got %v", xErr)
	}
}

// TestPromoteConsumesDraftSession 晋升即消耗草稿：同一事务内源会话标记 deleted、预览文件保留、
// 快照深拷贝独立落库，并广播 delete_session 让工作台 WS 及时下线
func TestPromoteConsumesDraftSession(t *testing.T) {
	l, db := setupPagesPromoteTest(t)
	ctx := context.Background()

	project := seedPromoteProject(t, db)
	session := seedPromoteSession(t, db, project.ID, 200)

	// 捕获 WS 广播（OnPreviewChanged 为包级钩子，测试内替换并在结束后恢复）
	var mu sync.Mutex
	events := make([]string, 0, 1)
	orig := OnPreviewChanged
	OnPreviewChanged = func(sessionID, eventType string) {
		mu.Lock()
		defer mu.Unlock()
		events = append(events, sessionID+":"+eventType)
	}
	t.Cleanup(func() { OnPreviewChanged = orig })

	resp, xErr := l.Promote(ctx, session.ID, &apiPreview.PromoteSessionRequest{Title: "站点", Slug: "site"})
	if xErr != nil {
		t.Fatalf("Promote failed: %v", xErr)
	}
	if resp.Conflict != nil {
		t.Fatal("unexpected conflict response")
	}

	// 源会话已软删除，不再出现在 active 列表
	var consumed entity.PreviewSession
	if err := db.Where("id = ?", session.ID).First(&consumed).Error; err != nil {
		t.Fatal(err)
	}
	if consumed.Status != bConst.PreviewSessionStatusDeleted {
		t.Fatalf("晋升后会话状态 = %q, want deleted", consumed.Status)
	}

	// 软删除保留预览文件；Pages 快照为深拷贝独立落库，与草稿互不影响
	var previewCount, snapshotCount int64
	db.Model(&entity.PreviewFile{}).Where("session_id = ?", session.ID).Count(&previewCount)
	db.Model(&entity.PageFile{}).Where("version_id = ?", resp.Version.ID).Count(&snapshotCount)
	if previewCount != 1 || snapshotCount != 1 {
		t.Fatalf("文件计数异常: preview=%d snapshot=%d, want 1/1", previewCount, snapshotCount)
	}

	// 版本溯源指向被消耗的源会话（审计保留）
	var version entity.PageVersion
	if err := db.Where("id = ?", resp.Version.ID).First(&version).Error; err != nil {
		t.Fatal(err)
	}
	if version.SourceSessionID == nil || *version.SourceSessionID != session.ID {
		t.Fatalf("版本溯源 = %v, want 源会话 %d", version.SourceSessionID, session.ID.Int64())
	}

	// 工作台 WS 收到会话结束广播
	mu.Lock()
	defer mu.Unlock()
	wantEvent := session.ID.String() + ":delete_session"
	if len(events) != 1 || events[0] != wantEvent {
		t.Fatalf("广播事件 = %v, want [%s]", events, wantEvent)
	}

	// 草稿已消耗：再次晋升同一会话按不存在处理
	if _, xErr := l.Promote(ctx, session.ID, &apiPreview.PromoteSessionRequest{Title: "再次", Slug: "site"}); xErr == nil || xErr.GetErrorCode() != xError.NotFound {
		t.Fatalf("已消耗草稿再次晋升应 NotFound，got %v", xErr)
	}
}

// TestForkCreatesNewActiveSession Fork 新建 active 草稿并携带溯源，不改动源 Pages 生效指针与快照
func TestForkCreatesNewActiveSession(t *testing.T) {
	l, db := setupPagesPromoteTest(t)
	ctx := context.Background()

	project := seedPromoteProject(t, db)
	page, latest, _ := seedPromotePageWithVersions(t, db, project.ID)
	content := "<html><body>page</body></html>"
	if err := db.Create(&entity.PageFile{
		BaseEntity: xModels.BaseEntity{ID: xSnowflake.SnowflakeID(testPromoteProjectID + 110)},
		VersionID:  latest.ID,
		Filename:   "index.html",
		MimeType:   bConst.PreviewMimeHTML,
		Content:    content,
		Size:       len(content),
	}).Error; err != nil {
		t.Fatal(err)
	}

	resp, xErr := l.Fork(ctx, page.ID, 0)
	if xErr != nil {
		t.Fatalf("Fork failed: %v", xErr)
	}

	// 新会话 active 且带溯源指针
	forked := resp.Session
	if forked.Status != bConst.PreviewSessionStatusActive {
		t.Fatalf("Fork 会话状态 = %q, want active", forked.Status)
	}
	if forked.SourcePageID == nil || *forked.SourcePageID != page.ID ||
		forked.SourceVersionID == nil || *forked.SourceVersionID != latest.ID {
		t.Fatal("Fork 会话缺少 source_page_id / source_version_id 溯源")
	}
	if forked.FileCount != 1 {
		t.Fatalf("Fork 文件数 = %d, want 1", forked.FileCount)
	}

	// 源 Pages 不受影响：生效指针与快照文件原样保留
	var sourcePage entity.Page
	if err := db.Where("id = ?", page.ID).First(&sourcePage).Error; err != nil {
		t.Fatal(err)
	}
	if sourcePage.LatestVersionID != page.LatestVersionID {
		t.Fatalf("源页面生效指针被改动: %d, want %d", sourcePage.LatestVersionID.Int64(), page.LatestVersionID.Int64())
	}
	var pageFiles int64
	db.Model(&entity.PageFile{}).Where("version_id = ?", latest.ID).Count(&pageFiles)
	if pageFiles != 1 {
		t.Fatalf("源快照文件数 = %d, want 1", pageFiles)
	}
}
