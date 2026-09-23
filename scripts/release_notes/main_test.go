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
		Commits: []string{
			"4432306 feat(预览): 支持 MDX 独立文档渲染与交互预览 (xiaolfeng)",
			"122a495 ci: 优化 GoReleaser 构建平台并限制并发 (xiaolfeng)",
			"ce843da fix(权限): 修复普通用户可查看私有仓库的越权缺陷 (alice)",
			"a1b2c3d perf(缓存): 优化多级缓存查询响应时延 (bob)",
			"e5f6a7b 文档与规范调整 (charlie)",
		},
	}

	result := formatReleaseNotes(renderFallbackTemplate(ctx), ctx)
	for _, want := range []string{
		"> [!IMPORTANT]\n> 本次更新重点：支持 MDX 独立文档渲染与交互预览；修复普通用户可查看私有仓库的越权缺陷。",
		"### 新增特性\n- 支持 MDX 独立文档渲染与交互预览",
		"### 修复与改进\n- 修复普通用户可查看私有仓库的越权缺陷\n- 优化多级缓存查询响应时延",
		"### 其他变更\n- 优化 GoReleaser 构建平台并限制并发",
		"### 变更记录\n- 4432306 feat(预览): 支持 MDX 独立文档渲染与交互预览 (xiaolfeng)",
		"完整变更记录：[v1.0.0...v1.0.1-beta.1](https://github.com/xiaolfeng/Lumina/compare/v1.0.0...v1.0.1-beta.1)",
	} {
		if !strings.Contains(result, want) {
			t.Errorf("生成结果缺少 %q\n完整输出:\n%s", want, result)
		}
	}
	if strings.Contains(result, "@alice") || strings.Contains(result, "(Features)") {
		t.Errorf("生成结果出现不符合中文模板的内容:\n%s", result)
	}
}

func TestFormatReleaseNotes(t *testing.T) {
	ctx := &gitContext{
		Repo:        "xiaolfeng/Lumina",
		CurrentTag:  "v1.0.2",
		PreviousTag: "v1.0.1",
		Commits:     []string{"abcdef1 fix: 修复登录失败 (Alice Smith)"},
	}
	body := "> [!IMPORTANT]\n> 登录现已恢复。\n\n### 修复与改进\n- 修复登录失败。"
	result := formatReleaseNotes(body, ctx)
	if !strings.HasPrefix(result, body) || strings.Count(result, "### 变更记录") != 1 ||
		!strings.Contains(result, "- abcdef1 fix: 修复登录失败 (Alice Smith)") ||
		!strings.HasSuffix(result, "[v1.0.1...v1.0.2](https://github.com/xiaolfeng/Lumina/compare/v1.0.1...v1.0.2)") {
		t.Errorf("正文或变更记录格式错误:\n%s", result)
	}

	firstRelease := formatReleaseNotes(renderFallbackTemplate(&gitContext{CurrentTag: "v1.0.0"}), &gitContext{CurrentTag: "v1.0.0"})
	if strings.Contains(firstRelease, "### 变更记录") || strings.Contains(firstRelease, "/compare/") {
		t.Errorf("首发空提交产生了多余栏目:\n%s", firstRelease)
	}
}

func TestBuildPrompts(t *testing.T) {
	ctx := &gitContext{
		Repo:        "xiaolfeng/Lumina",
		CurrentTag:  "v1.0.2",
		PreviousTag: "v1.0.1",
		Range:       "v1.0.1..v1.0.2",
		Commits:     []string{"abcdef1 feat(mcp): 新增 Pin 工具支持 (xiaolfeng)"},
		DiffStat:    " 1 file changed, 10 insertions(+)",
		DiffSnippet: "+// new feature",
	}

	sysPrompt := buildSystemPrompt()
	for _, want := range []string{"简体中文", "[!IMPORTANT]", "### 新增特性", "### 修复与改进", "程序会在正文之后附加提交记录"} {
		if !strings.Contains(sysPrompt, want) {
			t.Errorf("System prompt 缺少 %q", want)
		}
	}

	userPrompt := buildUserPrompt(ctx)
	for _, want := range []string{"v1.0.2", "v1.0.1..v1.0.2", "Diff Stat", "Diff Snippet"} {
		if !strings.Contains(userPrompt, want) {
			t.Errorf("User prompt 缺少 %q:\n%s", want, userPrompt)
		}
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
