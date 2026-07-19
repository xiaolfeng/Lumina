package logic

import (
	"encoding/json"
	"testing"
)

// TestArchitectOutput 验证 Architect Agent 新输出格式 `{ outline, metas }` 能正确反序列化到 ArchitectOutput。
//
// 覆盖：
// - 顶层包装对象（非裸数组）
// - outline 树嵌套 children
// - 叶子节点 icon 字段
// - 目录节点 path 无扩展名
// - metas 数组与 default_open / pages 字段
func TestArchitectOutput(t *testing.T) {
	raw := `{
  "outline": [
    {
      "title": "概览",
      "path": "overview",
      "description": "项目整体介绍",
      "icon": "FileText",
      "explore_refs": ["project_overview"],
      "complexity": "low"
    },
    {
      "title": "模块",
      "path": "modules",
      "description": "业务模块文档",
      "icon": "Folder",
      "children": [
        {
          "title": "认证模块",
          "path": "modules/auth",
          "description": "认证与授权实现",
          "icon": "ShieldCheck",
          "explore_refs": ["internal_auth"],
          "complexity": "medium"
        },
        {
          "title": "API 接口",
          "path": "modules/api",
          "description": "REST API 相关文档",
          "icon": "Folder",
          "children": [
            {
              "title": "端点列表",
              "path": "modules/api/endpoints",
              "description": "所有 REST 端点说明",
              "icon": "List",
              "explore_refs": ["internal_handler"],
              "complexity": "high"
            }
          ]
        }
      ]
    }
  ],
  "metas": [
    {
      "path": "modules",
      "title": "模块",
      "icon": "Folder",
      "default_open": true,
      "pages": ["auth", "---业务模块---", "api"]
    },
    {
      "path": "modules/api",
      "title": "API 接口",
      "icon": "Folder",
      "default_open": false,
      "pages": ["endpoints"]
    }
  ]
}`

	var out ArchitectOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatalf("Unmarshal ArchitectOutput failed: %v", err)
	}

	// outline 顶层条目数
	if len(out.Outline) != 2 {
		t.Fatalf("expected 2 top-level outline entries, got %d", len(out.Outline))
	}

	// 概览页（叶子）
	overview := out.Outline[0]
	if overview.Title != "概览" {
		t.Errorf("overview title: want 概览, got %s", overview.Title)
	}
	if overview.Path != "overview" {
		t.Errorf("overview path: want overview (no extension), got %s", overview.Path)
	}
	if overview.Icon != "FileText" {
		t.Errorf("overview icon: want FileText, got %s", overview.Icon)
	}
	if len(overview.Children) != 0 {
		t.Errorf("overview should be a leaf, got %d children", len(overview.Children))
	}

	// 模块目录
	modules := out.Outline[1]
	if modules.Path != "modules" {
		t.Errorf("modules path: want modules (no extension), got %s", modules.Path)
	}
	if modules.Icon != "Folder" {
		t.Errorf("modules icon: want Folder, got %s", modules.Icon)
	}
	if len(modules.Children) != 2 {
		t.Fatalf("expected 2 children under modules, got %d", len(modules.Children))
	}

	// 嵌套目录 modules/api
	api := modules.Children[1]
	if api.Path != "modules/api" {
		t.Errorf("api path: want modules/api (no extension), got %s", api.Path)
	}
	if api.Icon != "Folder" {
		t.Errorf("api icon: want Folder, got %s", api.Icon)
	}
	if len(api.Children) != 1 {
		t.Fatalf("expected 1 child under api, got %d", len(api.Children))
	}

	// 深层叶子 modules/api/endpoints
	endpoints := api.Children[0]
	if endpoints.Path != "modules/api/endpoints" {
		t.Errorf("endpoints path: want modules/api/endpoints (no extension), got %s", endpoints.Path)
	}
	if endpoints.Icon != "List" {
		t.Errorf("endpoints icon: want List, got %s", endpoints.Icon)
	}
	if len(endpoints.Children) != 0 {
		t.Errorf("endpoints should be a leaf, got %d children", len(endpoints.Children))
	}

	// metas 数组
	if len(out.Metas) != 2 {
		t.Fatalf("expected 2 metas, got %d", len(out.Metas))
	}

	// modules meta
	modulesMeta := out.Metas[0]
	if modulesMeta.Path != "modules" {
		t.Errorf("modules meta path: want modules, got %s", modulesMeta.Path)
	}
	if modulesMeta.Title != "模块" {
		t.Errorf("modules meta title: want 模块, got %s", modulesMeta.Title)
	}
	if modulesMeta.Icon != "Folder" {
		t.Errorf("modules meta icon: want Folder, got %s", modulesMeta.Icon)
	}
	if !modulesMeta.DefaultOpen {
		t.Errorf("modules meta default_open: want true, got false")
	}
	if len(modulesMeta.Pages) != 3 {
		t.Fatalf("expected 3 pages in modules meta, got %d", len(modulesMeta.Pages))
	}
	if modulesMeta.Pages[0] != "auth" {
		t.Errorf("modules meta pages[0]: want auth, got %s", modulesMeta.Pages[0])
	}
	if modulesMeta.Pages[1] != "---业务模块---" {
		t.Errorf("modules meta pages[1]: want separator ---业务模块---, got %s", modulesMeta.Pages[1])
	}
	if modulesMeta.Pages[2] != "api" {
		t.Errorf("modules meta pages[2]: want api, got %s", modulesMeta.Pages[2])
	}

	// modules/api meta
	apiMeta := out.Metas[1]
	if apiMeta.Path != "modules/api" {
		t.Errorf("api meta path: want modules/api, got %s", apiMeta.Path)
	}
	if apiMeta.DefaultOpen {
		t.Errorf("api meta default_open: want false, got true")
	}
	if len(apiMeta.Pages) != 1 || apiMeta.Pages[0] != "endpoints" {
		t.Errorf("api meta pages: want [endpoints], got %v", apiMeta.Pages)
	}
}

// TestArchitectOutput_BareArrayRejected 验证旧的裸数组格式不再被 ArchitectOutput 接受。
//
// 这是 BREAKING 变更的回归保护：旧格式 `[{...}]` 反序列化到 ArchitectOutput 应失败。
func TestArchitectOutput_BareArrayRejected(t *testing.T) {
	oldFormat := `[{"title":"概览","path":"overview","icon":"FileText"}]`

	var out ArchitectOutput
	if err := json.Unmarshal([]byte(oldFormat), &out); err == nil {
		t.Errorf("expected error when unmarshaling bare array into ArchitectOutput, got nil; out=%+v", out)
	}
}
