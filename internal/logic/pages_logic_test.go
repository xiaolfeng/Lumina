package logic

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"testing"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
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
