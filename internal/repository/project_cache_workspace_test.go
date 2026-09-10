package repository

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	"github.com/redis/go-redis/v9"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupProjectRepoWithCache(t *testing.T) *ProjectRepo {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&entity.Project{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	t.Cleanup(mr.Close)

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	return NewProjectRepo(db, rdb)
}

func TestGetByIDReloadsZeroWorkspaceCache(t *testing.T) {
	repo := setupProjectRepoWithCache(t)
	ctx := context.Background()

	workspaceID := xSnowflake.GenerateID(bConst.GeneWorkspace)
	project := &entity.Project{
		BaseEntity:  xModels.BaseEntity{ID: xSnowflake.GenerateID(bConst.GeneProject)},
		WorkspaceID: workspaceID,
		Name:        "cache-zero-ws",
		AliasName:   "cache-zero",
		MatchPath:   []string{"/tmp/cache-zero"},
	}
	if err := repo.db.Create(project).Error; err != nil {
		t.Fatalf("create project: %v", err)
	}

	stale := *project
	stale.WorkspaceID = 0
	if xErr := repo.cache.SetProject(ctx, &stale); xErr != nil {
		t.Fatalf("seed stale cache: %v", xErr)
	}

	got, xErr := repo.GetByID(ctx, project.ID)
	if xErr != nil {
		t.Fatalf("GetByID: %v", xErr)
	}
	if got.WorkspaceID != workspaceID {
		t.Fatalf("WorkspaceID = %d, want %d", got.WorkspaceID.Int64(), workspaceID.Int64())
	}
}

func TestReplaceWorkspaceCacheWritesWorkspaceID(t *testing.T) {
	repo := setupProjectRepoWithCache(t)
	ctx := context.Background()

	defaultID := xSnowflake.GenerateID(bConst.GeneWorkspace)
	project := &entity.Project{
		BaseEntity:  xModels.BaseEntity{ID: xSnowflake.GenerateID(bConst.GeneProject)},
		WorkspaceID: 0,
		Name:        "cache-replace-ws",
		AliasName:   "cache-replace",
		MatchPath:   []string{"/tmp/cache-replace"},
	}
	if xErr := repo.cache.SetProject(ctx, project); xErr != nil {
		t.Fatalf("seed cache: %v", xErr)
	}

	if xErr := repo.ReplaceWorkspaceCache(ctx, []*entity.Project{project}, defaultID); xErr != nil {
		t.Fatalf("ReplaceWorkspaceCache: %v", xErr)
	}

	got, ok, xErr := repo.cache.GetByID(ctx, project.ID.Int64())
	if xErr != nil {
		t.Fatalf("GetByID cache: %v", xErr)
	}
	if !ok {
		t.Fatal("expected cache hit after replace")
	}
	if got.WorkspaceID != defaultID {
		t.Fatalf("WorkspaceID = %d, want %d", got.WorkspaceID.Int64(), defaultID.Int64())
	}
}

func TestUpdateClearsOldWorkspaceCache(t *testing.T) {
	repo := setupProjectRepoWithCache(t)
	ctx := context.Background()
	project := &entity.Project{
		WorkspaceID: 100000000000000001,
		Name:        "cache-move-original", AliasName: "cache-move-alias",
		MatchPath: []string{"/cache-move"},
	}
	project.ID = 100000000000000003
	if xErr := repo.Create(ctx, project); xErr != nil {
		t.Fatal(xErr)
	}
	project.WorkspaceID = 100000000000000002
	project.Name = "cache-move-renamed"
	project.AliasName = "cache-move-new-alias"
	if xErr := repo.Update(ctx, project); xErr != nil {
		t.Fatal(xErr)
	}
	if _, ok, _ := repo.cache.GetIDByMatchPath(ctx, 100000000000000001, "/cache-move"); ok {
		t.Fatal("old workspace path cache remains")
	}
	if _, ok, _ := repo.cache.GetIDByName(ctx, "cache-move-original"); ok {
		t.Fatal("old name cache remains")
	}
	if _, ok, _ := repo.cache.GetIDByAlias(ctx, "cache-move-alias"); ok {
		t.Fatal("old alias cache remains")
	}
	if id, ok, _ := repo.cache.GetIDByMatchPath(ctx, 100000000000000002, "/cache-move"); !ok || id != project.ID.String() {
		t.Fatal("new workspace path cache missing")
	}
	cached, ok, xErr := repo.cache.GetByID(ctx, project.ID.Int64())
	if xErr != nil || !ok || cached.WorkspaceID != project.WorkspaceID {
		t.Fatalf("cached project = %+v, error = %v", cached, xErr)
	}
}
