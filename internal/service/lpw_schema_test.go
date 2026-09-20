package service

import (
	"strings"
	"testing"
)

func TestLpwSchemaLoader(t *testing.T) {
	t.Parallel()

	loader, err := NewLpwSchemaLoader()
	if err != nil {
		t.Fatalf("NewLpwSchemaLoader() failed: %v", err)
	}

	if !loader.Supports("1.0") {
		t.Errorf("expected support for version 1.0")
	}
	if loader.Supports("2.0") {
		t.Errorf("did not expect support for version 2.0")
	}

	// 1. 合法文档
	validDoc := []byte(`{
		"version": "1.0",
		"meta": { "title": "测试文档" },
		"blocks": [
			{ "id": "b1", "type": "markdown", "props": { "content": "hello" } }
		]
	}`)
	failPath, reason, ok := loader.Validate(validDoc, "1.0")
	if !ok {
		t.Errorf("expected valid document to pass, got path=%q reason=%q", failPath, reason)
	}

	// 2. 缺 version
	missingVersion := []byte(`{
		"blocks": []
	}`)
	failPath, reason, ok = loader.Validate(missingVersion, "1.0")
	if ok {
		t.Errorf("expected document missing version to fail")
	}
	if failPath != "/version" {
		t.Errorf("expected failPath /version, got %q (reason: %s)", failPath, reason)
	}

	// 3. 叶子块带 children
	leafWithChildren := []byte(`{
		"version": "1.0",
		"blocks": [
			{ "id": "b1", "type": "markdown", "props": { "content": "hi" }, "children": [] }
		]
	}`)
	failPath, reason, ok = loader.Validate(leafWithChildren, "1.0")
	if ok {
		t.Errorf("expected leaf with children to fail")
	}
	if failPath != "/blocks/0/children" {
		t.Errorf("expected failPath /blocks/0/children, got %q (reason: %s)", failPath, reason)
	}

	// 4. metrics items 为空数组（minItems 1）
	emptyMetrics := []byte(`{
		"version": "1.0",
		"blocks": [
			{ "id": "m1", "type": "metrics", "props": { "items": [] } }
		]
	}`)
	failPath, reason, ok = loader.Validate(emptyMetrics, "1.0")
	if ok {
		t.Errorf("expected empty metrics items to fail")
	}
	if !strings.HasPrefix(failPath, "/blocks/0/props/items") {
		t.Errorf("expected failPath to start with /blocks/0/props/items, got %q (reason: %s)", failPath, reason)
	}

	// 5. id 包含非法字符（大写或下划线）
	badID := []byte(`{
		"version": "1.0",
		"blocks": [
			{ "id": "BAD_ID", "type": "markdown", "props": { "content": "hi" } }
		]
	}`)
	failPath, reason, ok = loader.Validate(badID, "1.0")
	if ok {
		t.Errorf("expected bad id to fail")
	}
	if failPath != "/blocks/0/id" {
		t.Errorf("expected failPath /blocks/0/id, got %q (reason: %s)", failPath, reason)
	}
}
