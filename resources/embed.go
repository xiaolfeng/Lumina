// Package resources 提供项目级内嵌资源（prompts、模板等）。
//
// 将静态资源文件集中在项目根的 resources/ 目录下管理，
// 通过 go:embed 暴露给各业务包引用，避免资源文件散落在业务包内部。
package resources

import "embed"

// PromptFiles 内嵌 RepoWiki 5 角色 system prompt 文件。
//
// 文件布局：resources/prompts/{role}.md
// 如 prompts/coordinator.md、prompts/writer.md 等。
//
//go:embed prompts/*.md
var PromptFiles embed.FS
