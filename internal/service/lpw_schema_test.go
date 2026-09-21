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

	if !loader.Supports("1.1") {
		t.Errorf("expected support for version 1.1")
	}
	if loader.Supports("1.0") {
		t.Errorf("did not expect support for version 1.0")
	}
	if loader.Supports("2.0") {
		t.Errorf("did not expect support for version 2.0")
	}

	// 1. 合法文档
	validDoc := []byte(`{
		"version": "1.1",
		"meta": { "title": "测试文档" },
		"content": [
			{ "id": "b1", "kind": "block", "type": "markdown", "props": { "content": "hello" } }
		]
	}`)
	failPath, reason, ok := loader.Validate(validDoc, "1.1")
	if !ok {
		t.Errorf("expected valid document to pass, got path=%q reason=%q", failPath, reason)
	}

	// 2. 缺 version
	missingVersion := []byte(`{
		"content": []
	}`)
	failPath, reason, ok = loader.Validate(missingVersion, "1.1")
	if ok {
		t.Errorf("expected document missing version to fail")
	}
	if failPath != "/version" {
		t.Errorf("expected failPath /version, got %q (reason: %s)", failPath, reason)
	}

	// 3. 根字段为旧 blocks
	oldBlocksDoc := []byte(`{
		"version": "1.1",
		"blocks": []
	}`)
	failPath, reason, ok = loader.Validate(oldBlocksDoc, "1.1")
	if ok {
		t.Errorf("expected document with blocks root to fail")
	}
	if failPath != "/blocks" {
		t.Errorf("expected failPath /blocks, got %q (reason: %s)", failPath, reason)
	}

	// 4. 叶子块带 children
	leafWithChildren := []byte(`{
		"version": "1.1",
		"content": [
			{ "id": "b1", "kind": "block", "type": "markdown", "props": { "content": "hi" }, "children": [] }
		]
	}`)
	failPath, reason, ok = loader.Validate(leafWithChildren, "1.1")
	if ok {
		t.Errorf("expected leaf with children to fail")
	}
	if failPath != "/content[0]/children" {
		t.Errorf("expected failPath /content[0]/children, got %q (reason: %s)", failPath, reason)
	}

	// 5. metrics items 为空数组（minItems 1）
	emptyMetrics := []byte(`{
		"version": "1.1",
		"content": [
			{ "id": "m1", "kind": "block", "type": "metrics", "props": { "items": [] } }
		]
	}`)
	failPath, reason, ok = loader.Validate(emptyMetrics, "1.1")
	if ok {
		t.Errorf("expected empty metrics items to fail")
	}
	if !strings.HasPrefix(failPath, "/content[0]/props/items") {
		t.Errorf("expected failPath to start with /content[0]/props/items, got %q (reason: %s)", failPath, reason)
	}

	// 6. id 包含非法字符（大写或下划线）
	badID := []byte(`{
		"version": "1.1",
		"content": [
			{ "id": "BAD_ID", "kind": "block", "type": "markdown", "props": { "content": "hi" } }
		]
	}`)
	failPath, reason, ok = loader.Validate(badID, "1.1")
	if ok {
		t.Errorf("expected bad id to fail")
	}
	if failPath != "/content[0]/id" {
		t.Errorf("expected failPath /content[0]/id, got %q (reason: %s)", failPath, reason)
	}

	// 7. layoutStrategy 分支验证
	layoutDoc := []byte(`{
		"version": "1.1",
		"meta": { "title": "布局测试" },
		"content": [
			{
				"id": "lay-1",
				"kind": "layout",
				"type": "layout",
				"props": {
					"pattern": "split",
					"strategy": { "type": "ratio", "tracks": [1, 2] }
				},
				"children": [
					{ "id": "b1", "kind": "block", "type": "markdown", "props": { "content": "left" } },
					{ "id": "b2", "kind": "block", "type": "markdown", "props": { "content": "right" } }
				]
			}
		]
	}`)
	failPath, reason, ok = loader.Validate(layoutDoc, "1.1")
	if !ok {
		t.Errorf("expected valid layout to pass, got path=%q reason=%q", failPath, reason)
	}
}
