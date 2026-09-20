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

func TestPreviewSessionSchemaDeclaresSourceFields(t *testing.T) {
	schema := previewSessionSchema()
	props, _ := schema["properties"].(map[string]any)
	for _, key := range []string{"expires_at", "source_page_id", "source_page_slug", "source_version_id", "preview_url"} {
		if _, ok := props[key]; !ok {
			t.Fatalf("previewSessionSchema missing %s", key)
		}
	}
}

func TestPreviewSessionDataIncludesSourceSlug(t *testing.T) {
	session := &apiPreview.PreviewSessionResponse{
		Title:          "forked",
		Hash:           "abc",
		SourcePageSlug: "design-system",
	}
	data := previewSessionData(session, "https://example.com/preview/abc/index.html")
	if data["source_page_slug"] != "design-system" {
		t.Fatalf("source_page_slug = %v", data["source_page_slug"])
	}
	if data["source_page_id"] != nil || data["source_version_id"] != nil {
		t.Fatalf("empty source ids should be null, got %v %v", data["source_page_id"], data["source_version_id"])
	}
}

func TestPreviewToolDefinitions(t *testing.T) {
	wantNames := []string{
		"preview_session_create",
		"preview_session_list",
		"preview_file_upload",
		"preview_file_edit",
		"preview_file_delete",
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
	// 1. 优先 index.lpw
	lpwFiles := []apiPreview.PreviewFileResponse{
		{ID: xSnowflake.SnowflakeID(1), Filename: "index.html", MimeType: bConst.PreviewMimeHTML},
		{ID: xSnowflake.SnowflakeID(2), Filename: "index.lpw", MimeType: bConst.PreviewMimeLPW},
	}
	if entry := findPreviewEntry(lpwFiles); entry == nil || entry.Filename != "index.lpw" {
		t.Fatalf("findPreviewEntry(lpwFiles) = %#v, want index.lpw", entry)
	}

	// 2. 首个 lpw 优先于 html
	firstLpwFiles := []apiPreview.PreviewFileResponse{
		{ID: xSnowflake.SnowflakeID(1), Filename: "index.html", MimeType: bConst.PreviewMimeHTML},
		{ID: xSnowflake.SnowflakeID(2), Filename: "doc.lpw", MimeType: bConst.PreviewMimeLPW},
	}
	if entry := findPreviewEntry(firstLpwFiles); entry == nil || entry.Filename != "doc.lpw" {
		t.Fatalf("findPreviewEntry(firstLpwFiles) = %#v, want doc.lpw", entry)
	}

	// 3. 仅 HTML：与 logic 层 inferMimeType 存储的真实 MIME 一致（text/html; charset=utf-8）
	files := []apiPreview.PreviewFileResponse{
		{ID: xSnowflake.SnowflakeID(1), Filename: "app.js", MimeType: "application/javascript; charset=utf-8"},
		{ID: xSnowflake.SnowflakeID(2), Filename: "index.html", MimeType: bConst.PreviewMimeHTML},
	}
	entry := findPreviewEntry(files)
	if entry == nil || entry.Filename != "index.html" {
		t.Fatalf("findPreviewEntry() = %#v", entry)
	}

	// 4. 负例：裸 "text/html"（不含 charset）不是生产代码产出的 MIME，不应命中
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

func TestParsePreviewLineArg(t *testing.T) {
	// 未传入：present 为 false 且无错误
	value, present, errMsg := parsePreviewLineArg(map[string]any{}, "start_line")
	if value != 0 || present || errMsg != "" {
		t.Fatalf("absent arg: value=%d present=%v err=%q", value, present, errMsg)
	}
	// 显式 null：与未传入等价
	value, present, errMsg = parsePreviewLineArg(map[string]any{"start_line": nil}, "start_line")
	if value != 0 || present || errMsg != "" {
		t.Fatalf("null arg: value=%d present=%v err=%q", value, present, errMsg)
	}
	// 合法正整数
	value, present, errMsg = parsePreviewLineArg(map[string]any{"start_line": 42.0}, "start_line")
	if value != 42 || !present || errMsg != "" {
		t.Fatalf("valid arg: value=%d present=%v err=%q", value, present, errMsg)
	}
	// 类型错误
	if _, present, errMsg = parsePreviewLineArg(map[string]any{"start_line": "42"}, "start_line"); !present || errMsg == "" {
		t.Fatalf("string arg should fail: present=%v err=%q", present, errMsg)
	}
	// 非正数与小数
	if _, _, errMsg = parsePreviewLineArg(map[string]any{"start_line": 0.0}, "start_line"); errMsg == "" {
		t.Fatal("zero start_line should fail")
	}
	if _, _, errMsg = parsePreviewLineArg(map[string]any{"start_line": 2.5}, "start_line"); errMsg == "" {
		t.Fatal("fractional start_line should fail")
	}
}

func TestPreviewFileDataWithLines(t *testing.T) {
	file := &apiPreview.PreviewFileResponse{Filename: "app.js"}
	data := previewFileDataWithLines(file, 128)
	if data["total_lines"] != 128 || data["filename"] != "app.js" {
		t.Fatalf("previewFileDataWithLines = %#v", data)
	}
}

func TestPreviewFileEditOutputSchemaDeclaresRegion(t *testing.T) {
	schema := previewFileEditOutputSchema()
	props, _ := schema["properties"].(map[string]any)
	fileSchema, _ := props["file"].(map[string]any)
	fileProps, _ := fileSchema["properties"].(map[string]any)
	if _, ok := fileProps["total_lines"]; !ok {
		t.Fatal("edit 输出 Schema 的 file 应包含 total_lines")
	}
	regionSchema, _ := props["edited_region"].(map[string]any)
	regionProps, _ := regionSchema["properties"].(map[string]any)
	for _, key := range []string{"start_line", "end_line", "content"} {
		if _, ok := regionProps[key]; !ok {
			t.Fatalf("edited_region 缺少 %s", key)
		}
	}
}
