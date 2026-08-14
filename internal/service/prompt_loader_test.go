package service

import (
	"testing"

	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

// TestLoadSystemPrompt 验证 RepoWiki 5 角色的 system prompt 均能从内嵌资源加载。
//
// 回归覆盖：bConst.AgentRoleRepoWiki* 常量使用点号前缀（repowiki.coordinator），
// LoadSystemPrompt 必须能正确剥离前缀并命中 prompts/{role}.md，否则返回空字符串。
func TestLoadSystemPrompt(t *testing.T) {
	roles := []struct {
		name string
		role string
	}{
		{name: "coordinator", role: bConst.AgentRoleRepoWikiCoordinator},
		{name: "explore", role: bConst.AgentRoleRepoWikiExplore},
		{name: "architect", role: bConst.AgentRoleRepoWikiArchitect},
		{name: "write", role: bConst.AgentRoleRepoWikiWrite},
		{name: "validator", role: bConst.AgentRoleRepoWikiValidator},
	}

	for _, tc := range roles {
		t.Run(tc.name, func(t *testing.T) {
			prompt := LoadSystemPrompt(tc.role)
			if prompt == "" {
				t.Fatalf("角色 %s 的 system prompt 加载失败", tc.role)
			}
			if len(prompt) < 50 {
				t.Fatalf("角色 %s 的 system prompt 内容过短（%d 字符），疑似加载到不完整内容", tc.role, len(prompt))
			}
		})
	}
}

// TestLoadSystemPromptUnknown 验证未知角色返回空字符串（不 panic）。
func TestLoadSystemPromptUnknown(t *testing.T) {
	if prompt := LoadSystemPrompt("repowiki.not-exist"); prompt != "" {
		t.Fatalf("未知角色应返回空字符串，实际返回 %q", prompt)
	}
}
