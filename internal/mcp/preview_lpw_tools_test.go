package mcp

import (
	"encoding/json"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestPreviewLpwToolDefinitions(t *testing.T) {
	expectedTools := map[string]bool{
		"preview_lpw_init":         true,
		"preview_lpw_block_add":    true,
		"preview_lpw_block_edit":   true,
		"preview_lpw_block_remove": true,
		"preview_lpw_block_sort":   true,
		"preview_lpw_meta_set":     true,
		"preview_lpw_outline":      true,
	}

	if len(previewLpwToolDefs) != 7 {
		t.Fatalf("expected 7 preview lpw tools, got %d", len(previewLpwToolDefs))
	}

	for _, def := range previewLpwToolDefs {
		if !expectedTools[def.name] {
			t.Errorf("unexpected tool: %s", def.name)
		}
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

		// 检查写操作必填输出字段
		if def.name != "preview_lpw_outline" {
			reqFields, _ := def.outputSchema["required"].([]string)
			reqMap := make(map[string]bool)
			for _, f := range reqFields {
				reqMap[f] = true
			}
			for _, required := range []string{"status", "workflow", "revision", "total_blocks", "file_size"} {
				if !reqMap[required] {
					t.Errorf("tool %s outputSchema 缺少必填字段 %s", def.name, required)
				}
			}
		} else {
			reqFields, _ := def.outputSchema["required"].([]string)
			reqMap := make(map[string]bool)
			for _, f := range reqFields {
				reqMap[f] = true
			}
			if !reqMap["status"] || !reqMap["outline"] {
				t.Errorf("tool %s outputSchema 缺少必填字段 status 或 outline", def.name)
			}
		}

		// Q-05 回归：init 的 inputSchema 必须包含 design 0003 契约声明的 blocks 属性
		if def.name == "preview_lpw_init" {
			props, _ := def.inputSchema["properties"].(map[string]any)
			if props == nil || props["blocks"] == nil {
				t.Errorf("preview_lpw_init inputSchema 缺少 design 0003 契约的 blocks 属性")
			}
		}
	}

	server := mcp.NewServer(&mcp.Implementation{Name: "preview-lpw-test", Version: "test"}, nil)
	RegisterPreviewLpwTools(server)
}
