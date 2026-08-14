package mcp

import (
	"encoding/json"
	"strings"
	"testing"

	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	apiPreview "github.com/xiaolfeng/Lumina/api/preview"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

func TestPreviewToolDefinitions(t *testing.T) {
	wantNames := []string{
		"preview_session_create",
		"preview_session_list",
		"preview_file_upload",
		"preview_file_list",
		"preview_file_get",
	}
	if len(previewToolDefs) != len(wantNames) {
		t.Fatalf("previewToolDefs count = %d, want %d", len(previewToolDefs), len(wantNames))
	}

	for i, def := range previewToolDefs {
		if def.name != wantNames[i] {
			t.Errorf("previewToolDefs[%d].name = %q, want %q", i, def.name, wantNames[i])
		}
		if def.title == "" || !strings.Contains(def.description, "何时调用") {
			t.Errorf("tool %s 缺少标题或调用时机说明", def.name)
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
		if def.annotations == nil || def.annotations.OpenWorldHint == nil || *def.annotations.OpenWorldHint {
			t.Errorf("tool %s 应声明为 Lumina 内部闭合世界操作", def.name)
		}
	}

	server := mcp.NewServer(&mcp.Implementation{Name: "preview-test", Version: "test"}, nil)
	RegisterPreviewTools(server)
}

func TestPreviewSupplementData(t *testing.T) {
	data := previewSupplementData("123", "456")
	if data["content_type"] != "preview" {
		t.Fatalf("content_type = %v, want preview", data["content_type"])
	}
	if data["content"] != `{"session_id":"123","file_id":"456"}` {
		t.Fatalf("content = %v", data["content"])
	}
	if err := validatePreviewSupplementContent(data["content"].(string)); err != nil {
		t.Fatalf("generated supplement content should be valid: %v", err)
	}
}

func TestValidatePreviewSupplementContent(t *testing.T) {
	valid := `{"session_id":"123","file_id":"456"}`
	if err := validatePreviewSupplementContent(valid); err != nil {
		t.Fatalf("valid content rejected: %v", err)
	}

	invalid := []string{
		`{"session_id":"123","hash":"abc"}`,
		`{"session_id":"123","file_id":"456","preview_url":"https://example.com"}`,
		`{"session_id":123,"file_id":"456"}`,
		`not-json`,
	}
	for _, content := range invalid {
		if err := validatePreviewSupplementContent(content); err == nil {
			t.Errorf("invalid content accepted: %s", content)
		}
	}
}

func TestFindPreviewEntry(t *testing.T) {
	// 与 logic 层 inferMimeType 存储的真实 MIME 一致（text/html; charset=utf-8）
	files := []apiPreview.PreviewFileResponse{
		{ID: xSnowflake.SnowflakeID(1), Filename: "app.js", MimeType: "application/javascript; charset=utf-8"},
		{ID: xSnowflake.SnowflakeID(2), Filename: "index.html", MimeType: bConst.PreviewMimeHTML},
	}
	entry := findPreviewEntry(files)
	if entry == nil || entry.Filename != "index.html" {
		t.Fatalf("findPreviewEntry() = %#v", entry)
	}

	// 负例：裸 "text/html"（不含 charset）不是生产代码产出的 MIME，不应命中
	bare := []apiPreview.PreviewFileResponse{
		{ID: xSnowflake.SnowflakeID(1), Filename: "index.html", MimeType: "text/html"},
	}
	if entry := findPreviewEntry(bare); entry != nil {
		t.Fatalf("findPreviewEntry(bare text/html) = %#v, want nil", entry)
	}
}

func TestPreviewStructuredResult(t *testing.T) {
	structured := map[string]any{"status": "success", "message": "ok"}
	result := previewStructuredResult(structured)
	if result.IsError || result.StructuredContent == nil || len(result.Content) != 1 {
		t.Fatalf("unexpected result: %#v", result)
	}

	text := result.Content[0]
	payload, err := text.MarshalJSON()
	if err != nil {
		t.Fatalf("marshal text content: %v", err)
	}
	var wire struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(payload, &wire); err != nil {
		t.Fatalf("unmarshal text content: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal([]byte(wire.Text), &decoded); err != nil {
		t.Fatalf("compatibility text is not JSON: %v", err)
	}
}
