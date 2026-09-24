package logic

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	apiProject "github.com/xiaolfeng/Lumina/api/project"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"github.com/xiaolfeng/Lumina/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupProjectWorkspaceTest(t *testing.T) (*ProjectLogic, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "project.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(&entity.Workspace{}, &entity.Project{}); err != nil {
		t.Fatal(err)
	}
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	l := &ProjectLogic{
		logic: newWorkspaceLogicForValidate().logic,
		repo: projectRepo{
			project:   repository.NewProjectRepo(db, rdb),
			workspace: repository.NewWorkspaceRepo(db, rdb),
		},
	}
	for i, slug := range []string{"default", "work"} {
		workspace := &entity.Workspace{Name: slug, Slug: slug, IsDefault: i == 0}
		workspace.ID = 100000000000000001
		if i == 1 {
			workspace.ID = 100000000000000002
		}
		if err := db.Create(workspace).Error; err != nil {
			t.Fatal(err)
		}
	}
	project := &entity.Project{Name: "migrate-test", WorkspaceID: 100000000000000001, AliasName: "alias", MatchPath: []string{"/test"}, Description: "original"}
	project.ID = 100000000000000003
	if xErr := l.repo.project.Create(context.Background(), project); xErr != nil {
		t.Fatal(xErr)
	}
	return l, db
}

func TestProjectUpdateWorkspace(t *testing.T) {
	for _, test := range []struct {
		name          string
		body          string
		wantWorkspace string
		wantError     bool
	}{
		{"omitted", `{"name":"updated"}`, "100000000000000001", false},
		{"null", `{"name":"updated","workspace_id":null}`, "100000000000000001", false},
		{"same", `{"name":"updated","workspace_id":"100000000000000001"}`, "100000000000000001", false},
		{"move string", `{"name":"updated","workspace_id":"100000000000000002"}`, "100000000000000002", false},
		{"move number", `{"name":"updated","workspace_id":100000000000000002}`, "100000000000000002", false},
		{"missing target", `{"name":"updated","workspace_id":"100000000000000099"}`, "100000000000000001", true},
		{"zero", `{"name":"updated","workspace_id":0}`, "100000000000000001", true},
		{"negative", `{"name":"updated","workspace_id":-1}`, "100000000000000001", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			l, db := setupProjectWorkspaceTest(t)
			var req apiProject.UpdateProjectRequest
			if err := json.Unmarshal([]byte(test.body), &req); err != nil {
				t.Fatal(err)
			}
			resp, xErr := l.Update(context.Background(), "100000000000000003", &req)
			if (xErr != nil) != test.wantError {
				t.Fatalf("Update error = %v, wantError %v", xErr, test.wantError)
			}
			if !test.wantError && resp.WorkspaceID.String() != test.wantWorkspace {
				t.Fatalf("response workspace = %v", resp.WorkspaceID)
			}
			var stored entity.Project
			if err := db.First(&stored, "id = ?", "100000000000000003").Error; err != nil {
				t.Fatal(err)
			}
			if stored.WorkspaceID.String() != test.wantWorkspace {
				t.Fatalf("stored workspace = %v, want %s", stored.WorkspaceID, test.wantWorkspace)
			}
			if test.wantError && (stored.Name != "migrate-test" || stored.Description != "original") {
				t.Fatal("failed migration modified project fields")
			}
			cached, xErr := l.GetByID(context.Background(), "100000000000000003")
			if xErr != nil || cached.WorkspaceID.String() != test.wantWorkspace {
				t.Fatalf("cached project = %+v, error = %v", cached, xErr)
			}
		})
	}
}

func TestProjectUpdateWorkspaceRejectsInvalidJSON(t *testing.T) {
	for _, value := range []string{`""`, `"invalid"`, `{}`, `true`} {
		var req apiProject.UpdateProjectRequest
		if err := json.Unmarshal([]byte(`{"name":"test","workspace_id":`+value+`}`), &req); err == nil {
			t.Errorf("accepted invalid workspace_id %s", value)
		}
	}
}

func TestProjectUpdateWorkspacePreservesProjectAndLists(t *testing.T) {
	l, _ := setupProjectWorkspaceTest(t)
	ctx := context.Background()
	target := json.Number("100000000000000002")
	resp, xErr := l.Update(ctx, "100000000000000003", &apiProject.UpdateProjectRequest{
		Name: "migrate-test", AliasName: "alias", MatchPath: []string{"/test"}, Description: "original", WorkspaceID: &target,
	})
	if xErr != nil {
		t.Fatal(xErr)
	}
	if resp.ID.String() != "100000000000000003" || resp.AliasName != "alias" || resp.Description != "original" || len(resp.MatchPath) != 1 || resp.MatchPath[0] != "/test" {
		t.Fatalf("migration changed project identity or fields: %+v", resp)
	}
	oldList, xErr := l.List(ctx, 1, 10, 100000000000000001, "")
	if xErr != nil || oldList.Total != 0 {
		t.Fatalf("old list = %+v, error = %v", oldList, xErr)
	}
	newList, xErr := l.List(ctx, 1, 10, 100000000000000002, "")
	if xErr != nil || newList.Total != 1 || newList.Items[0].ID != resp.ID {
		t.Fatalf("new list = %+v, error = %v", newList, xErr)
	}
}
