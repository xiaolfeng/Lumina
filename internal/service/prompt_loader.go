package service

import (
	"strings"

	"github.com/xiaolfeng/Lumina/resources"
)

// LoadSystemPrompt 按 role 从内嵌的 Markdown 文件加载 system prompt。
//
// 文件映射：role → prompts/{role}.md
// 如 "repowiki:coordinator" → prompts/coordinator.md（自动去除 "repowiki:" 前缀）
// 返回文件全文（已去除首尾空白）。
func LoadSystemPrompt(role string) string {
	name := strings.TrimPrefix(role, "repowiki:")
	name = "prompts/" + name + ".md"
	data, err := resources.PromptFiles.ReadFile(name)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
