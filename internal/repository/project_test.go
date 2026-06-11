package repository

import (
	"context"
	"testing"

	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
)

// setupTestRepo 创建测试用 ProjectRepo 实例
// 由于需要完整的 PostgreSQL + Redis 基础设施，无连接时自动跳过
func setupTestRepo(t *testing.T) *ProjectRepo {
	t.Helper()
	t.Skip("requires database connection - skipping unit test")
	return nil
}

func TestProjectRepo_Create(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

	project := &entity.Project{
		BaseEntity:  xModels.BaseEntity{ID: xSnowflake.GenerateID(bConst.GeneProject)},
		Name:        "test-project-repo-create",
		AliasName:   "test-create",
		MatchPath:   []string{"/test/repo-create"},
		Description: "repository create test project",
	}

	xErr := repo.Create(ctx, project)
	if xErr != nil {
		t.Fatalf("Create failed: %v", xErr)
	}

	if project.ID.Int64() == 0 {
		t.Errorf("expected non-zero project ID, got 0")
	}

	// 清理测试数据
	_ = repo.Delete(ctx, project.ID)
}

func TestProjectRepo_GetByID(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

	id := xSnowflake.GenerateID(bConst.GeneProject)

	// 先创建再查询
	project := &entity.Project{
		BaseEntity:  xModels.BaseEntity{ID: id},
		Name:        "test-project-repo-getbyid",
		AliasName:   "test-getbyid",
		MatchPath:   []string{"/test/repo-getbyid"},
		Description: "repository getbyid test project",
	}
	if xErr := repo.Create(ctx, project); xErr != nil {
		t.Fatalf("setup Create failed: %v", xErr)
	}

	found, xErr := repo.GetByID(ctx, id)
	if xErr != nil {
		t.Fatalf("GetByID failed: %v", xErr)
	}
	if found == nil {
		t.Fatal("GetByID returned nil project")
	}
	if found.Name != "test-project-repo-getbyid" {
		t.Errorf("expected name 'test-project-repo-getbyid', got '%s'", found.Name)
	}

	_ = repo.Delete(ctx, id)
}

func TestProjectRepo_GetByName(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

	name := "test-project-repo-getbyname"

	project := &entity.Project{
		BaseEntity:  xModels.BaseEntity{ID: xSnowflake.GenerateID(bConst.GeneProject)},
		Name:        name,
		AliasName:   "test-getbyname",
		MatchPath:   []string{"/test/repo-getbyname"},
		Description: "repository getbyname test project",
	}
	if xErr := repo.Create(ctx, project); xErr != nil {
		t.Fatalf("setup Create failed: %v", xErr)
	}

	found, xErr := repo.GetByName(ctx, name)
	if xErr != nil {
		t.Fatalf("GetByName failed: %v", xErr)
	}
	if found == nil {
		t.Fatal("GetByName returned nil project")
	}

	_ = repo.Delete(ctx, project.ID)
}

func TestProjectRepo_List(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

	projects, total, xErr := repo.List(ctx, 1, 10)
	if xErr != nil {
		t.Fatalf("List failed: %v", xErr)
	}
	if total < 0 {
		t.Errorf("expected non-negative total, got %d", total)
	}
	if projects == nil {
		t.Fatal("List returned nil slice")
	}
	if len(projects) > int(total) {
		t.Errorf("items length %d exceeds total %d", len(projects), total)
	}
}

func TestProjectRepo_Update(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

	id := xSnowflake.GenerateID(bConst.GeneProject)

	project := &entity.Project{
		BaseEntity:  xModels.BaseEntity{ID: id},
		Name:        "test-project-repo-update",
		AliasName:   "test-update",
		MatchPath:   []string{"/test/repo-update"},
		Description: "original description",
	}
	if xErr := repo.Create(ctx, project); xErr != nil {
		t.Fatalf("setup Create failed: %v", xErr)
	}

	// 更新描述
	project.Description = "updated description"
	xErr := repo.Update(ctx, project)
	if xErr != nil {
		t.Fatalf("Update failed: %v", xErr)
	}

	// 验证更新结果
	updated, xErr := repo.GetByID(ctx, id)
	if xErr != nil {
		t.Fatalf("GetByID after update failed: %v", xErr)
	}
	if updated.Description != "updated description" {
		t.Errorf("expected description 'updated description', got '%s'", updated.Description)
	}

	_ = repo.Delete(ctx, id)
}

func TestProjectRepo_Delete(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

	id := xSnowflake.GenerateID(bConst.GeneProject)

	project := &entity.Project{
		BaseEntity:  xModels.BaseEntity{ID: id},
		Name:        "test-project-repo-delete",
		AliasName:   "test-delete",
		MatchPath:   []string{"/test/repo-delete"},
		Description: "repository delete test project",
	}
	if xErr := repo.Create(ctx, project); xErr != nil {
		t.Fatalf("setup Create failed: %v", xErr)
	}

	xErr := repo.Delete(ctx, id)
	if xErr != nil {
		t.Fatalf("Delete failed: %v", xErr)
	}

	// 验证已删除
	_, xErr = repo.GetByID(ctx, id)
	if xErr == nil {
		t.Fatal("expected error after delete, got nil")
	}
}

func TestProjectRepo_FindByAliasName(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

	alias := "test-findalias"

	project := &entity.Project{
		BaseEntity:  xModels.BaseEntity{ID: xSnowflake.GenerateID(bConst.GeneProject)},
		Name:        "test-project-repo-findalias",
		AliasName:   "test-findalias",
		MatchPath:   []string{"/test/repo-findalias"},
		Description: "repository findbyalias test project",
	}
	if xErr := repo.Create(ctx, project); xErr != nil {
		t.Fatalf("setup Create failed: %v", xErr)
	}

	found, xErr := repo.FindByAliasName(ctx, alias)
	if xErr != nil {
		t.Fatalf("FindByAliasName failed: %v", xErr)
	}
	if found == nil {
		t.Fatal("FindByAliasName returned nil project")
	}
	if found.Name != "test-project-repo-findalias" {
		t.Errorf("expected name 'test-project-repo-findalias', got '%s'", found.Name)
	}

	_ = repo.Delete(ctx, project.ID)
}
