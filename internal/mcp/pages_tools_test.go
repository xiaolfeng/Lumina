package mcp

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestPagesToolDefinitions(t *testing.T) {
		wantNames := []string{"pages_list", "pages_promote", "pages_fork"}
	if len(pagesToolDefs) != len(wantNames) {
		t.Fatalf("pagesToolDefs count = %d, want %d", len(pagesToolDefs), len(wantNames))
	}
	for i, def := range pagesToolDefs {
		if def.name != wantNames[i] {
			t.Errorf("pagesToolDefs[%d].name = %q, want %q", i, def.name, wantNames[i])
		}
		if def.title == "" || !strings.Contains(def.description, "何时调用") {
			t.Errorf("tool %s 缺少标题或调用时机说明", def.name)
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
	server := mcp.NewServer(&mcp.Implementation{Name: "pages-test", Version: "test"}, nil)
	RegisterPagesTools(server)
}
