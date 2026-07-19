// Package logic RepoWiki 子 Agent user prompt 构建测试。
package logic

import (
	"strings"
	"testing"
)

// TestBuildUserPrompt 验证 user prompt 包含 frontmatter / .mdx / {outline,metas} 引导文本。
func TestBuildUserPrompt(t *testing.T) {
	t.Run("writer_contains_frontmatter_guidance", func(t *testing.T) {
		entries := []WikiEntry{
			{Title: "概览", Path: "overview", Description: "项目概览", Complexity: "low", ExploreRefs: []string{"scope-a"}},
		}
		explores := map[string]string{"scope-a": "explore content"}

		result := BuildWriterUserPrompt(entries, explores)

		if !strings.Contains(result, "frontmatter") {
			t.Errorf("expected writer user prompt to contain frontmatter guidance, got:\n%s", result)
		}
		if !strings.Contains(result, ".mdx") {
			t.Errorf("expected writer user prompt to contain .mdx path reference, got:\n%s", result)
		}
		if !strings.Contains(result, "save_wiki_page") {
			t.Errorf("expected writer user prompt to mention save_wiki_page tool, got:\n%s", result)
		}
	})

	t.Run("architect_requests_outline_metas_object", func(t *testing.T) {
		overview := "项目概要文本"
		explores := []ExploreOutput{{Scope: "scope-a", Content: "explore content"}}

		result := BuildArchitectUserPrompt(overview, explores)

		if !strings.Contains(result, "outline") {
			t.Errorf("expected architect user prompt to mention outline key, got:\n%s", result)
		}
		if !strings.Contains(result, "metas") {
			t.Errorf("expected architect user prompt to mention metas key, got:\n%s", result)
		}
		if !strings.Contains(result, "JSON 对象") {
			t.Errorf("expected architect user prompt to request JSON object (not array), got:\n%s", result)
		}
		if strings.Contains(result, "JSON 数组（格式见指令要求）") {
			t.Errorf("architect user prompt should not request bare JSON array, got:\n%s", result)
		}
		if !strings.Contains(result, "无扩展名") {
			t.Errorf("expected architect user prompt to mention path has no extension, got:\n%s", result)
		}
	})

	t.Run("validator_contains_mdx_and_frontmatter_guidance", func(t *testing.T) {
		result := BuildValidatorUserPrompt("/wiki", `{"outline":[],"metas":[]}`)

		if !strings.Contains(result, ".mdx") {
			t.Errorf("expected validator user prompt to mention .mdx files, got:\n%s", result)
		}
		if !strings.Contains(result, "frontmatter") {
			t.Errorf("expected validator user prompt to mention frontmatter, got:\n%s", result)
		}
		if !strings.Contains(result, "{outline, metas}") {
			t.Errorf("expected validator user prompt to mention {outline, metas} object shape, got:\n%s", result)
		}
	})
}

// TestBuildArchitectRetryHint 验证 retry hint 要求 { } 对象而非 [ ] 数组。
func TestBuildArchitectRetryHint(t *testing.T) {
	hint := buildArchitectRetryHint(1)

	if !strings.Contains(hint, "{") {
		t.Errorf("expected retry hint to contain '{' (object opener), got:\n%s", hint)
	}
	if !strings.Contains(hint, "}") {
		t.Errorf("expected retry hint to contain '}' (object closer), got:\n%s", hint)
	}
	if strings.Contains(hint, "以 '[' 开头") {
		t.Errorf("retry hint must not request '[' opener (old array format), got:\n%s", hint)
	}
	if strings.Contains(hint, "']' 结尾") {
		t.Errorf("retry hint must not request ']' closer (old array format), got:\n%s", hint)
	}
	if !strings.Contains(hint, "outline") || !strings.Contains(hint, "metas") {
		t.Errorf("expected retry hint to mention outline and metas keys, got:\n%s", hint)
	}
}
