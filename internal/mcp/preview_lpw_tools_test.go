package mcp

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/xiaolfeng/Lumina/internal/logic"
)

func TestPreviewLpwToolDefinitions(t *testing.T) {
	expectedTools := map[string]bool{
		"preview_lpw_init":        true,
		"preview_lpw_node_add":    true,
		"preview_lpw_node_edit":   true,
		"preview_lpw_node_remove": true,
		"preview_lpw_node_sort":   true,
		"preview_lpw_meta_set":    true,
		"preview_lpw_outline":     true,
		"preview_lpw_schema":      true,
	}

	if len(previewLpwToolDefs) != 8 {
		t.Fatalf("expected 8 preview lpw tools, got %d", len(previewLpwToolDefs))
	}

	for _, def := range previewLpwToolDefs {
		if !expectedTools[def.name] {
			t.Errorf("unexpected tool: %s", def.name)
		}
		delete(expectedTools, def.name)

		if def.title == "" || def.description == "" {
			t.Errorf("tool %s 缺少 title 或 description", def.name)
		}
		if def.inputSchema["type"] != "object" || def.outputSchema["type"] != "object" {
			t.Errorf("tool %s 的输入/输出 Schema 根类型必须为 object", def.name)
		}

		for schemaName, schemaValue := range map[string]map[string]any{
			"input":  def.inputSchema,
			"output": def.outputSchema,
		} {
			payload, err := json.Marshal(schemaValue)
			if err != nil {
				t.Fatalf("marshal %s schema for %s: %v", schemaName, def.name, err)
			}
			var schema jsonschema.Schema
			if err := json.Unmarshal(payload, &schema); err != nil {
				t.Fatalf("unmarshal %s schema for %s: %v", schemaName, def.name, err)
			}
			if _, err := schema.Resolve(nil); err != nil {
				t.Errorf("resolve %s schema for %s: %v", schemaName, def.name, err)
			}
		}
	}
	if len(expectedTools) > 0 {
		t.Errorf("missing tools: %v", expectedTools)
	}

	server := mcp.NewServer(&mcp.Implementation{Name: "preview-lpw-test", Version: "test"}, nil)
	RegisterPreviewLpwTools(server)
}

func TestPreviewLpwLegacyToolsAbsent(t *testing.T) {
	legacy := []string{
		"preview_lpw_block_add",
		"preview_lpw_block_edit",
		"preview_lpw_block_remove",
		"preview_lpw_block_sort",
	}

	for name := range previewLpwToolHandlers {
		for _, l := range legacy {
			if name == l {
				t.Errorf("legacy handler still registered: %s", l)
			}
		}
	}
	for _, def := range previewLpwToolDefs {
		for _, l := range legacy {
			if def.name == l {
				t.Errorf("legacy tool def still present: %s", l)
			}
		}
	}
}

func TestPreviewLpwInitSchemaContract(t *testing.T) {
	var initDef *previewLpwToolDef
	for i := range previewLpwToolDefs {
		if previewLpwToolDefs[i].name == "preview_lpw_init" {
			initDef = &previewLpwToolDefs[i]
			break
		}
	}
	if initDef == nil {
		t.Fatalf("preview_lpw_init not found")
	}

	props, _ := initDef.inputSchema["properties"].(map[string]any)
	if props == nil {
		t.Fatalf("preview_lpw_init inputSchema missing properties")
	}

	if props["content"] == nil {
		t.Errorf("expected content in properties")
	}
	if props["blocks"] != nil {
		t.Errorf("blocks must not exist in 1.1 init properties")
	}
	if props["version"] != nil {
		t.Errorf("version must not be exposed in init properties (hardcoded to 1.1)")
	}

	reqList, _ := initDef.inputSchema["required"].([]string)
	hasSessionID := false
	hasMeta := false
	for _, r := range reqList {
		if r == "session_id" {
			hasSessionID = true
		}
		if r == "meta" {
			hasMeta = true
		}
	}
	if !hasSessionID || !hasMeta {
		t.Errorf("init inputSchema required must include session_id and meta, got: %v", reqList)
	}
}

func TestPreviewLpwWriteOutputContract(t *testing.T) {
	schema := previewLpwWriteOutputSchema()
	props, _ := schema["properties"].(map[string]any)
	if props == nil {
		t.Fatalf("previewLpwWriteOutputSchema missing properties")
	}

	if props["total_nodes"] == nil {
		t.Errorf("expected total_nodes in output properties")
	}
	if props["node_kind"] == nil {
		t.Errorf("expected node_kind in output properties")
	}
	if props["node_id"] == nil {
		t.Errorf("expected node_id in output properties")
	}

	if props["total_blocks"] != nil {
		t.Errorf("total_blocks must be removed from output properties")
	}
	if props["block_id"] != nil {
		t.Errorf("block_id must be removed from output properties")
	}
}

func TestPreviewLpwNodeEdit_UnmarshalErrors(t *testing.T) {
	SetPreviewLpwLogic(&logic.PreviewLpwLogic{})
	defer SetPreviewLpwLogic(nil)

	handler := previewLpwToolHandlers["preview_lpw_node_edit"]

	// 1. annotation 不是对象也不是 null
	args1, _ := json.Marshal(map[string]any{
		"session_id": "1",
		"filename":   "doc.lpw",
		"node_id":    "n1",
		"annotation": "not-an-object",
	})
	res1, err := handler(t.Context(), &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Arguments: args1}})
	if err != nil {
		t.Fatalf("unexpected handler err: %v", err)
	}
	if !res1.IsError {
		t.Errorf("expected error for invalid annotation type")
	}

	// 2. node 不是对象
	args2, _ := json.Marshal(map[string]any{
		"session_id": "1",
		"filename":   "doc.lpw",
		"node_id":    "n1",
		"node":       "not-an-object",
	})
	res2, err := handler(t.Context(), &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Arguments: args2}})
	if err != nil {
		t.Fatalf("unexpected handler err: %v", err)
	}
	if !res2.IsError {
		t.Errorf("expected error for invalid node type")
	}
}

func TestPreviewLpwSchemaTextOutput(t *testing.T) {
	t.Parallel()

	// 1. 全局大纲必须包含三层体系，且结构清晰
	fullSpec := GetLpwTextSpec("", "")
	if !strings.Contains(fullSpec, "Layout 布局体系") || !strings.Contains(fullSpec, "Container 受控容器体系") || !strings.Contains(fullSpec, "Block 原子内容组件") {
		t.Fatalf("fullSpec missing three tier headers")
	}
	if !strings.Contains(fullSpec, "chart:") || !strings.Contains(fullSpec, "diff:") || !strings.Contains(fullSpec, "annotation:") {
		t.Fatalf("fullSpec missing essential components")
	}

	// 2. 指定单个类型输出精准参数
	chartSpec := GetLpwTextSpec("", "chart")
	if !strings.Contains(chartSpec, "[Block / type: chart]") || !strings.Contains(chartSpec, "chartType") {
		t.Fatalf("chartSpec missing details, got: %s", chartSpec)
	}

	bentoSpec := GetLpwTextSpec("", "bento")
	if !strings.Contains(bentoSpec, "[Layout / pattern: bento]") || !strings.Contains(bentoSpec, "placements") {
		t.Fatalf("bentoSpec missing details, got: %s", bentoSpec)
	}

	// 3. MCP Handler 调用验证
	handler := previewLpwToolHandlers["preview_lpw_schema"]
	reqBytes, _ := json.Marshal(map[string]any{"type": "section"})
	res, err := handler(t.Context(), &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Arguments: reqBytes}})
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected error result")
	}
	if len(res.Content) == 0 {
		t.Fatalf("expected text content in response")
	}
}
