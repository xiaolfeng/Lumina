package mcp

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xCtx "github.com/bamboo-services/bamboo-base-go/defined/context"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/redis/go-redis/v9"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"github.com/xiaolfeng/Lumina/internal/logic"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPinToolDescriptionsMatchResolveRules(t *testing.T) {
	t.Parallel()

	for _, def := range pinToolDefs {
		if strings.Contains(def.description, "项目名称/别名") || strings.Contains(def.description, "支持名称/别名") {
			t.Errorf("%s description still allows Project.Name: %s", def.name, def.description)
		}
	}

	push := pinToolDefs[0]
	if push.name != "pin_push" {
		t.Fatalf("first tool = %s, want pin_push", push.name)
	}
	if strings.Contains(push.description, "为可选") || strings.Contains(push.description, "不传时") {
		t.Errorf("pin_push still describes from_project_id as optional")
	}
	if !strings.Contains(push.description, "from_project_id 必填") {
		t.Errorf("pin_push should say from_project_id is required")
	}

	// 验证 pin_peek 工具允许按 project_name 预览队首，不再强制 id 单一必填
	var peekFound bool
	for _, def := range pinToolDefs {
		if def.name == "pin_peek" {
			peekFound = true
			props, ok := def.inputSchema["properties"].(map[string]any)
			if !ok {
				t.Fatalf("pin_peek properties is not map")
			}
			if _, hasProjectName := props["project_name"]; !hasProjectName {
				t.Errorf("pin_peek should support project_name for read-only oldest pending peek")
			}
			if _, hasID := props["id"]; !hasID {
				t.Errorf("pin_peek should support id")
			}
			break
		}
	}
	if !peekFound {
		t.Fatalf("pin_peek not found in pinToolDefs")
	}
}

// setupMcpPinTestEnv 初始化 MCP Pin 测试所需的环境与上下文
func setupMcpPinTestEnv(t *testing.T) (*gorm.DB, *entity.Project) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "mcp-pin.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&entity.Project{}); err != nil {
		t.Fatalf("auto migrate project: %v", err)
	}

	// SQLite 下兼容时间扫描，显式创建 pins 表
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

	ctx := context.WithValue(context.Background(), xCtx.DatabaseKey, db)
	ctx = context.WithValue(ctx, xCtx.RedisClientKey, rdb)

	pLogic := logic.NewPinLogic(ctx)
	SetPinLogic(pLogic)
	t.Cleanup(func() { SetPinLogic(nil) })

	workspaceID := xSnowflake.GenerateID(bConst.GeneWorkspace)
	projectID := xSnowflake.GenerateID(bConst.GeneProject)
	project := &entity.Project{
		BaseEntity:  xModels.BaseEntity{ID: projectID},
		WorkspaceID: workspaceID,
		Name:        "MCP Target Project",
		AliasName:   "mcp-target-proj",
	}
	if err := db.Create(project).Error; err != nil {
		t.Fatalf("create test project: %v", err)
	}

	return db, project
}

// TestPinTools_MCP_LifecyclePeekAndConsume 全链路测试 MCP 层 Pin 工具生命周期：
// 1. 模拟 pending 约束；
// 2. handlePinPeek(project_name) 只读预览队首，断言正文、标题以及未消费状态；
// 3. 关键断言：数据库底层依然为 pending，consumed_at 为空；
// 4. 重复只读调用 handlePinPeek(project_name) 验证幂等；
// 5. handlePinList(project_name) 验证返回列表包含正文内容；
// 6. handlePinConsume(project_name) 显式确认消费；
// 7. 再次 handlePinPeek(project_name) 断言暂无待处理约束；
// 8. handlePinPeek(id) 验证已消费约束的历史回查。
func TestPinTools_MCP_LifecyclePeekAndConsume(t *testing.T) {
	db, project := setupMcpPinTestEnv(t)

	// Step 0: 缺少参数调用 handlePinPeek
	emptyArgs, _ := json.Marshal(map[string]any{})
	emptyRes, err := handlePinPeek(context.Background(), &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{Arguments: emptyArgs},
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !emptyRes.IsError {
		t.Fatal("expected error for empty args")
	}
	text := emptyRes.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "缺少参数") {
		t.Fatalf("unexpected error text: %s", text)
	}

	// Step 1: 插入一条 pending 约束
	pinID := xSnowflake.GenerateID(bConst.GenePin)
	title := "跨服务链路 trace 追踪头透传"
	content := "请在所有 HTTP 外部请求 Header 中追加 X-Trace-ID，以支持分布式全链路追踪。"
	now := time.Now().Truncate(time.Second)

	pin := &entity.Pin{
		BaseEntity:    xModels.BaseEntity{ID: pinID, CreatedAt: now, UpdatedAt: now},
		FromProjectID: project.ID,
		ToProjectID:   project.ID,
		Title:         title,
		Content:       content,
		Category:      "trace",
		Status:        bConst.PinStatusPending,
		Priority:      bConst.PinPriorityHigh,
	}
	if err := db.Create(pin).Error; err != nil {
		t.Fatalf("create pin entity: %v", err)
	}

	// Step 2: 调用 handlePinPeek(project_name)，只读预览队首待处理约束
	peekArgs, _ := json.Marshal(map[string]any{
		"project_name": project.AliasName,
	})
	peekRes, err := handlePinPeek(context.Background(), &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{Arguments: peekArgs},
	})
	if err != nil {
		t.Fatalf("handlePinPeek returned go error: %v", err)
	}
	if peekRes.IsError {
		t.Fatalf("handlePinPeek returned MCP error: %s", peekRes.Content[0].(*mcp.TextContent).Text)
	}
	peekText := peekRes.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(peekText, "标题: "+title) {
		t.Errorf("peekText missing title: %s", peekText)
	}
	if !strings.Contains(peekText, "内容: "+content) {
		t.Errorf("peekText missing content: %s", peekText)
	}
	if !strings.Contains(peekText, "状态: pending") {
		t.Errorf("peekText status is not pending: %s", peekText)
	}
	if !strings.Contains(peekText, "消费时间: （未消费）") {
		t.Errorf("peekText consumedAt should be unconsumed: %s", peekText)
	}

	// Step 3 (关键断言): 断言该约束在数据库/存储中的状态依然为 pending，consumed_at 为空
	var dbPin entity.Pin
	if err := db.Where("id = ?", pinID).First(&dbPin).Error; err != nil {
		t.Fatalf("query db pin failed: %v", err)
	}
	if dbPin.Status != bConst.PinStatusPending {
		t.Fatalf("CRITICAL: DB status was altered by handlePinPeek! got %s, want %s", dbPin.Status, bConst.PinStatusPending)
	}
	if dbPin.ConsumedAt != nil {
		t.Fatalf("CRITICAL: DB consumed_at was altered by handlePinPeek! got %v, want nil", dbPin.ConsumedAt)
	}

	// Step 4: 再次调用 handlePinPeek(project_name)，断言仍然能够重复只读读取，状态未变
	for i := 0; i < 2; i++ {
		repeatRes, err := handlePinPeek(context.Background(), &mcp.CallToolRequest{
			Params: &mcp.CallToolParamsRaw{Arguments: peekArgs},
		})
		if err != nil || repeatRes.IsError {
			t.Fatalf("repeat peek %d failed: err=%v, res=%v", i, err, repeatRes)
		}
		repeatText := repeatRes.Content[0].(*mcp.TextContent).Text
		if !strings.Contains(repeatText, "状态: pending") || !strings.Contains(repeatText, "消费时间: （未消费）") {
			t.Fatalf("repeat peek %d state modified: %s", i, repeatText)
		}
	}

	// Step 5: 调用 handlePinList(project_name)，断言返回结果包含每条约束的正文内容（Content）
	listArgs, _ := json.Marshal(map[string]any{
		"project_name": project.AliasName,
	})
	listRes, err := handlePinList(context.Background(), &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{Arguments: listArgs},
	})
	if err != nil {
		t.Fatalf("handlePinList returned go error: %v", err)
	}
	if listRes.IsError {
		t.Fatalf("handlePinList returned MCP error: %s", listRes.Content[0].(*mcp.TextContent).Text)
	}
	listText := listRes.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(listText, title) {
		t.Errorf("listText missing title: %s", listText)
	}
	if !strings.Contains(listText, "分类: trace | 优先级: high | 状态: pending") {
		t.Errorf("listText missing meta info: %s", listText)
	}
	if !strings.Contains(listText, "内容:\n"+content) {
		t.Errorf("listText missing pin content: %s", listText)
	}

	// Step 6: 随后调用 handlePinConsume(project_name)，断言状态被原子更新为 consumed，consumed_at 被正确设置
	consumeArgs, _ := json.Marshal(map[string]any{
		"project_name": project.AliasName,
	})
	consumeRes, err := handlePinConsume(context.Background(), &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{Arguments: consumeArgs},
	})
	if err != nil {
		t.Fatalf("handlePinConsume returned go error: %v", err)
	}
	if consumeRes.IsError {
		t.Fatalf("handlePinConsume returned MCP error: %s", consumeRes.Content[0].(*mcp.TextContent).Text)
	}
	consumeText := consumeRes.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(consumeText, "消费成功！") {
		t.Errorf("consumeText missing success message: %s", consumeText)
	}

	// 关键断言：直接查询底层 DB 验证状态变更
	if err := db.Where("id = ?", pinID).First(&dbPin).Error; err != nil {
		t.Fatalf("query db pin after consume failed: %v", err)
	}
	if dbPin.Status != bConst.PinStatusConsumed {
		t.Fatalf("DB status not consumed: got %s", dbPin.Status)
	}
	if dbPin.ConsumedAt == nil {
		t.Fatal("DB consumed_at is nil after consume")
	}

	// Step 7: 再次调用 handlePinPeek(project_name)，断言此时无待消费约束（NotFound）
	postConsumeRes, err := handlePinPeek(context.Background(), &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{Arguments: peekArgs},
	})
	if err != nil {
		t.Fatalf("handlePinPeek returned go error: %v", err)
	}
	if !postConsumeRes.IsError {
		t.Fatal("expected error from handlePinPeek after all pins consumed")
	}
	postConsumeText := postConsumeRes.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(postConsumeText, "查看队首待处理约束失败: 暂无待处理约束") {
		t.Errorf("unexpected error text: %s", postConsumeText)
	}

	// Step 8: 调用 handlePinPeek(id)，验证已消费的历史约束依然可以通过 ID 精确查看
	peekByIDArgs, _ := json.Marshal(map[string]any{
		"id": pinID.String(),
	})
	peekByIDRes, err := handlePinPeek(context.Background(), &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{Arguments: peekByIDArgs},
	})
	if err != nil {
		t.Fatalf("handlePinPeek by ID returned go error: %v", err)
	}
	if peekByIDRes.IsError {
		t.Fatalf("handlePinPeek by ID returned MCP error: %s", peekByIDRes.Content[0].(*mcp.TextContent).Text)
	}
	peekByIDText := peekByIDRes.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(peekByIDText, "状态: consumed") {
		t.Errorf("expected consumed status in history peek: %s", peekByIDText)
	}
	if strings.Contains(peekByIDText, "消费时间: （未消费）") {
		t.Errorf("consumedAt should show actual timestamp in history peek: %s", peekByIDText)
	}
}
