package repository

import (
	"context"
	"path/filepath"
	"testing"

	apiPin "github.com/xiaolfeng/Lumina/api/pin"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func TestWorkspaceFiltersFollowProjectWithTablePrefix(t *testing.T) {
	for _, naming := range []schema.NamingStrategy{{}, {TablePrefix: "lum_", SingularTable: true}} {
		t.Run(naming.TablePrefix, func(t *testing.T) {
			db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "filters.db")), &gorm.Config{NamingStrategy: naming})
			if err != nil {
				t.Fatal(err)
			}
			sqlDB, err := db.DB()
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = sqlDB.Close() })
			if err := db.AutoMigrate(&entity.Project{}, &entity.QaSession{}, &entity.PreviewSession{}, &entity.Pin{}); err != nil {
				t.Fatal(err)
			}
			project := &entity.Project{Name: "moving", WorkspaceID: 1001}
			project.ID = 1003
			qa := &entity.QaSession{ProjectID: project.ID, Title: "qa", Hash: "test-qa"}
			preview := &entity.PreviewSession{ProjectID: project.ID, Title: "preview", Hash: "test-preview"}
			pin := &entity.Pin{FromProjectID: 1004, ToProjectID: project.ID, Title: "pin"}
			for _, record := range []any{project, qa, preview, pin} {
				if err := db.Create(record).Error; err != nil {
					t.Fatal(err)
				}
			}
			qaRepo, previewRepo, pinRepo := NewQaSessionRepo(db, nil), NewPreviewSessionRepo(db), NewPinRepo(db)
			ctx := context.Background()
			check := func(wantOld, wantNew int64) {
				t.Helper()
				for i, want := range []int64{1, wantOld, wantNew} {
					workspace := entity.Workspace{}
					if i == 1 {
						workspace.ID = 1001
					}
					if i == 2 {
						workspace.ID = 1002
					}
					qas, total, xErr := qaRepo.List(ctx, 1, 10, "", "", "", workspace.ID)
					if xErr != nil || total != want || int64(len(qas)) != want {
						t.Fatalf("QA workspace %v: total=%d, error=%v", workspace.ID, total, xErr)
					}
					previews, total, xErr := previewRepo.List(ctx, 0, workspace.ID, 1, 10)
					if xErr != nil || total != want || int64(len(previews)) != want {
						t.Fatalf("Preview workspace %v: total=%d, error=%v", workspace.ID, total, xErr)
					}
					pins, total, xErr := pinRepo.List(ctx, &apiPin.PinListRequest{Page: 1, Size: 10, WorkspaceID: workspace.ID})
					if xErr != nil || total != want || int64(len(pins)) != want {
						t.Fatalf("Pin workspace %v: total=%d, error=%v", workspace.ID, total, xErr)
					}
				}
			}
			check(1, 0)
			if err := db.Model(project).Update("workspace_id", 1002).Error; err != nil {
				t.Fatal(err)
			}
			check(0, 1)
		})
	}
}
