package main

import (
	"strings"
	"testing"
)

func TestRenderFallbackTemplate(t *testing.T) {
	ctx := &gitContext{
		Repo:        "xiaolfeng/Lumina",
		CurrentTag:  "v1.0.1-beta.1",
		PreviousTag: "v1.0.0",
		Range:       "v1.0.0..v1.0.1-beta.1",
		Commits: []string{
			"4432306 feat(预览): 支持 MDX 独立文档渲染与交互预览 (xiaolfeng)",
			"122a495 ci: 优化 GoReleaser 构建平台并限制并发 (xiaolfeng)",
			"ce843da fix(权限): 修复普通用户可查看私有仓库的越权缺陷 (alice)",
			"a1b2c3d perf(缓存): 优化多级缓存查询响应时延 (bob)",
			"e5f6g7h 文档与规范调整 (charlie)",
		},
		DiffStat:    " 5 files changed, 100 insertions(+), 20 deletions(-)",
		DiffSnippet: "diff --git a/main.go b/main.go\n...",
	}

	result := renderFallbackTemplate(ctx)

	// 验证各分组标题
	expectedSubstrings := []string{
		"## Lumina · 微明 v1.0.1-beta.1",
		"### 🚀 新增特性 (Features)",
		"- **预览**: 支持 MDX 独立文档渲染与交互预览 (4432306) by @xiaolfeng",
		"### ⚡ 优化与改进 (Improvements & Performance)",
		"- **缓存**: 优化多级缓存查询响应时延 (a1b2c3d) by @bob",
		"### 🐛 缺陷修复 (Bug Fixes)",
		"- **权限**: 修复普通用户可查看私有仓库的越权缺陷 (ce843da) by @alice",
		"### 🛠️ 基础设施与工程构建 (Infrastructure & CI/CD)",
		"- 优化 GoReleaser 构建平台并限制并发 (122a495) by @xiaolfeng",
		"### 📝 其他变更 (Other Changes)",
		"- e5f6g7h 文档与规范调整 (charlie)",
		"**完整变更对比**: [v1.0.0...v1.0.1-beta.1](https://github.com/xiaolfeng/Lumina/compare/v1.0.0...v1.0.1-beta.1)",
	}

	for _, sub := range expectedSubstrings {
		if !strings.Contains(result, sub) {
			t.Errorf("生成结果缺少预期内容 %q\n完整输出:\n%s", sub, result)
		}
	}
}

func TestBuildPrompts(t *testing.T) {
	ctx := &gitContext{
		Repo:        "xiaolfeng/Lumina",
		CurrentTag:  "v1.0.2",
		PreviousTag: "v1.0.1",
		Range:       "v1.0.1..v1.0.2",
		Commits: []string{
			"abcdef1 feat(mcp): 新增 Pin 工具支持 (xiaolfeng)",
		},
		DiffStat:    " 1 file changed, 10 insertions(+)",
		DiffSnippet: "+// new feature",
	}

	sysPrompt := buildSystemPrompt()
	if !strings.Contains(sysPrompt, "GitHub Release 发版说明") {
		t.Errorf("System prompt 缺少预期核心描述")
	}

	userPrompt := buildUserPrompt(ctx)
	if !strings.Contains(userPrompt, "v1.0.2") || !strings.Contains(userPrompt, "v1.0.1..v1.0.2") {
		t.Errorf("User prompt 缺少版本与区间信息:\n%s", userPrompt)
	}
	if !strings.Contains(userPrompt, "Diff Stat") || !strings.Contains(userPrompt, "Diff Snippet") {
		t.Errorf("User prompt 缺少统计或代码变动摘要:\n%s", userPrompt)
	}
}

func TestGetEnvOrDefault(t *testing.T) {
	t.Setenv("TEST_KEY_EXIST", "custom_val")
	if got := getEnvOrDefault("TEST_KEY_EXIST", "fallback"); got != "custom_val" {
		t.Errorf("期望 custom_val，实际得到 %s", got)
	}
	if got := getEnvOrDefault("TEST_KEY_NOT_EXIST", "fallback"); got != "fallback" {
		t.Errorf("期望 fallback，实际得到 %s", got)
	}
}
