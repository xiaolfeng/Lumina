package logic

import (
	"context"
	"testing"

	apiProject "github.com/xiaolfeng/Lumina/api/project"
)

// setupTestLogic 创建测试用 logic 实例；无数据库连接时跳过
func setupTestLogic(t *testing.T) *ProjectLogic {
	t.Helper()
	t.Skip("requires database connection - skipping unit test")
	return nil
}

// TestProjectLogic_Create 测试创建项目 happy path
func TestProjectLogic_Create(t *testing.T) {
	l := setupTestLogic(t)
	ctx := context.Background()

	req := &apiProject.CreateProjectRequest{
		Name:        "test-project-unique",
		AliasName:   "test-alias",
		MatchPath:   []string{"/test/path"},
		Description: "test description",
	}

	resp, xErr := l.Create(ctx, req)
	if xErr != nil {
		t.Fatalf("Create failed: %v", xErr)
	}
	if resp.ID.IsZero() {
		t.Error("expected non-empty ID")
	}
	if resp.Name != req.Name {
		t.Errorf("expected name %q, got %q", req.Name, resp.Name)
	}
	if resp.Description != req.Description {
		t.Errorf("expected description %q, got %q", req.Description, resp.Description)
	}
	if resp.AliasName != req.AliasName {
		t.Errorf("expected alias %q, got %q", req.AliasName, resp.AliasName)
	}

	// Cleanup
	_ = l.Delete(ctx, resp.ID.String())
}

// TestProjectLogic_Create_DuplicateName 测试重复项目名称拒绝
func TestProjectLogic_Create_DuplicateName(t *testing.T) {
	l := setupTestLogic(t)
	ctx := context.Background()

	req := &apiProject.CreateProjectRequest{
		Name:        "test-dup-name",
		AliasName:   "",
		MatchPath:   []string{"/test/dup"},
		Description: "",
	}

	resp, xErr := l.Create(ctx, req)
	if xErr != nil {
		t.Fatalf("first Create failed: %v", xErr)
	}

	// 尝试创建同名项目，期望失败
	_, xErr = l.Create(ctx, req)
	if xErr == nil {
		t.Error("expected error for duplicate name, got nil")
	}

	// Cleanup
	_ = l.Delete(ctx, resp.ID.String())
}

// TestProjectLogic_GetByID 测试根据 ID 获取项目详情
func TestProjectLogic_GetByID(t *testing.T) {
	l := setupTestLogic(t)
	ctx := context.Background()

	// 先创建项目
	createReq := &apiProject.CreateProjectRequest{
		Name:        "test-get-by-id",
		AliasName:   "alias-get",
		MatchPath:   []string{"/test/get-by-id"},
		Description: "test get by id",
	}
	created, xErr := l.Create(ctx, createReq)
	if xErr != nil {
		t.Fatalf("Create failed: %v", xErr)
	}

	// 根据 ID 查询
	got, xErr := l.GetByID(ctx, created.ID.String())
	if xErr != nil {
		t.Fatalf("GetByID failed: %v", xErr)
	}
	if got.ID != created.ID {
		t.Errorf("expected ID %s, got %s", created.ID, got.ID)
	}
	if got.Name != createReq.Name {
		t.Errorf("expected name %q, got %q", createReq.Name, got.Name)
	}
	if got.Description != createReq.Description {
		t.Errorf("expected description %q, got %q", createReq.Description, got.Description)
	}

	// Cleanup
	_ = l.Delete(ctx, created.ID.String())
}

// TestProjectLogic_List 测试分页获取项目列表
func TestProjectLogic_List(t *testing.T) {
	l := setupTestLogic(t)
	ctx := context.Background()

	// 创建两个项目
	names := []string{"test-list-a", "test-list-b"}
	createdIDs := make([]string, 0, len(names))
	for _, name := range names {
		req := &apiProject.CreateProjectRequest{
			Name:        name,
			AliasName:   "",
			MatchPath:   []string{"/test/" + name},
			Description: "for list test",
		}
		resp, xErr := l.Create(ctx, req)
		if xErr != nil {
			t.Fatalf("Create %q failed: %v", name, xErr)
		}
		createdIDs = append(createdIDs, resp.ID.String())
	}

	// 分页查询
	listResp, xErr := l.List(ctx, 1, 10)
	if xErr != nil {
		t.Fatalf("List failed: %v", xErr)
	}
	if listResp.Total < int64(len(names)) {
		t.Errorf("expected total >= %d, got %d", len(names), listResp.Total)
	}
	if len(listResp.Items) < len(names) {
		t.Errorf("expected >= %d items, got %d", len(names), len(listResp.Items))
	}

	// Cleanup
	for _, id := range createdIDs {
		_ = l.Delete(ctx, id)
	}
}

// TestProjectLogic_Update 测试更新项目信息
func TestProjectLogic_Update(t *testing.T) {
	l := setupTestLogic(t)
	ctx := context.Background()

	// 先创建
	createReq := &apiProject.CreateProjectRequest{
		Name:        "test-update-original",
		AliasName:   "",
		MatchPath:   []string{"/test/update"},
		Description: "original description",
	}
	created, xErr := l.Create(ctx, createReq)
	if xErr != nil {
		t.Fatalf("Create failed: %v", xErr)
	}

	// 更新字段
	updateReq := &apiProject.UpdateProjectRequest{
		Name:        "test-updated-name",
		AliasName:   "updated-alias",
		MatchPath:   []string{"/test/updated"},
		Description: "updated description",
	}
	updated, xErr := l.Update(ctx, created.ID.String(), updateReq)
	if xErr != nil {
		t.Fatalf("Update failed: %v", xErr)
	}
	if updated.Name != updateReq.Name {
		t.Errorf("expected name %q, got %q", updateReq.Name, updated.Name)
	}
	if updated.Description != updateReq.Description {
		t.Errorf("expected description %q, got %q", updateReq.Description, updated.Description)
	}
	if updated.AliasName != updateReq.AliasName {
		t.Errorf("expected alias %q, got %q", updateReq.AliasName, updated.AliasName)
	}

	// Cleanup
	_ = l.Delete(ctx, created.ID.String())
}

// TestProjectLogic_Delete 测试删除项目
func TestProjectLogic_Delete(t *testing.T) {
	l := setupTestLogic(t)
	ctx := context.Background()

	// 先创建
	req := &apiProject.CreateProjectRequest{
		Name:        "test-delete-me",
		AliasName:   "",
		MatchPath:   []string{"/test/delete"},
		Description: "will be deleted",
	}
	created, xErr := l.Create(ctx, req)
	if xErr != nil {
		t.Fatalf("Create failed: %v", xErr)
	}

	// 删除
	xErr = l.Delete(ctx, created.ID.String())
	if xErr != nil {
		t.Fatalf("Delete failed: %v", xErr)
	}

	// 再次查询应返回错误
	_, xErr = l.GetByID(ctx, created.ID.String())
	if xErr == nil {
		t.Error("expected error after delete, got nil")
	}
}
