// Package logic RepoWiki 子 Agent 编排的 user prompt 构建与输出解析函数。
//
// System prompt 已从硬编码常量迁移到 service/prompts/*.md 资源文件，
// 通过 service.LoadSystemPrompt(role) 在运行时加载。
// 本文件仅保留 user prompt 构建函数和 Agent 输出解析辅助函数。
package logic

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ──────────────────────────────────────────────────────────────────────
// User Prompt 构建函数
// ──────────────────────────────────────────────────────────────────────

// BuildOverviewUserPrompt 构建概要阶段的 user prompt。
//
// 参数说明:
//   - repoPath: 仓库根目录路径
func BuildOverviewUserPrompt(repoPath string) string {
	return fmt.Sprintf(`请对项目进行核心概要分析。

仓库路径: %s

请使用 file_read、file_search 和 list_dir 工具了解项目结构，重点阅读 README、项目清单文件、入口文件等，输出项目概览。`, repoPath)
}

// BuildExploreUserPrompt 构建代码探索阶段的 user prompt。
//
// 参数说明:
//   - scope: 本次分析的范围描述（相对仓库根的路径或模块名）
func BuildExploreUserPrompt(scope string) string {
	return fmt.Sprintf(`请深入分析以下代码范围：

分析范围: %s

请使用 file_read 和 file_search 工具阅读该范围内的关键文件，按 XML 骨架格式输出分析结果。`, scope)
}

// BuildArchitectUserPrompt 构建架构规划阶段的 user prompt。
//
// 参数说明:
//   - overviewSummary: Coordinator 的项目概要
//   - exploreOutputs: 所有 Explore Agent 的产出列表
func BuildArchitectUserPrompt(overviewSummary string, exploreOutputs []ExploreOutput) string {
	var sb strings.Builder

	sb.WriteString("请根据以下信息构建 Wiki 目录大纲。\n\n")

	sb.WriteString("## 项目概要\n\n")
	sb.WriteString(overviewSummary)
	sb.WriteString("\n\n")

	sb.WriteString("## 可用的 Explore scope 列表（explore_refs 必须逐字从此列表选择）\n\n")
	sb.WriteString("```\n")
	for i, eo := range exploreOutputs {
		fmt.Fprintf(&sb, "%d. %s\n", i+1, eo.Scope)
	}
	sb.WriteString("```\n\n")

	sb.WriteString("## 代码探索产出（按 scope 标识组织）\n\n")
	for _, eo := range exploreOutputs {
		fmt.Fprintf(&sb, "### Explore: %s\n\n", eo.Scope)
		sb.WriteString(eo.Content)
		sb.WriteString("\n\n")
	}

	sb.WriteString("请输出 Wiki 目录大纲 JSON 数组（格式见指令要求）。\n")
	return sb.String()
}

// BuildWriterUserPrompt 构建文档撰写阶段的 user prompt。
//
// 参数说明:
//   - entries: 分配给本次调用的 Wiki 目录条目
//   - exploreOutputs: 对应 Explore 产出内容，key 为分析范围 scope
func BuildWriterUserPrompt(entries []WikiEntry, exploreOutputs map[string]string) string {
	var sb strings.Builder

	sb.WriteString("请撰写以下 Wiki 页面，并通过 save_wiki_page 工具写入。\n\n")

	for _, entry := range entries {
		fmt.Fprintf(&sb, "## 页面: %s\n", entry.Title)
		fmt.Fprintf(&sb, "- 路径: %s\n", entry.Path)
		fmt.Fprintf(&sb, "- 描述: %s\n", entry.Description)
		fmt.Fprintf(&sb, "- 复杂度: %s\n", entry.Complexity)
		fmt.Fprintf(&sb, "- 关联 Explore: %v\n\n", entry.ExploreRefs)
	}

	sb.WriteString("## 参考的 Explore 产出\n\n")
	for scope, content := range exploreOutputs {
		fmt.Fprintf(&sb, "### %s\n\n", scope)
		sb.WriteString(content)
		sb.WriteString("\n\n")
	}

	sb.WriteString("请依次撰写每个页面，使用 save_wiki_page 写入指定的路径。\n")
	return sb.String()
}

// BuildWriterRetryUserPrompt 构建重试阶段的 Writer user prompt，包含校验错误信息。
//
// 与 BuildWriterUserPrompt 的区别：在 prompt 中追加 Validator 返回的错误列表，
// 让 Writer 明确知道哪些页面缺失/为空/有问题，针对性地修复。
//
// 参数说明:
//   - entries: 分配给本次调用的 Wiki 目录条目
//   - exploreOutputs: 对应 Explore 产出内容，key 为分析范围 scope
//   - errors: Validator 返回的校验错误列表
func BuildWriterRetryUserPrompt(entries []WikiEntry, exploreOutputs map[string]string, errors []ValidationError) string {
	var sb strings.Builder

	sb.WriteString("⚠️ 上一轮校验发现以下问题，请针对性修复：\n\n")

	sb.WriteString("## 校验错误列表\n\n")
	for _, e := range errors {
		fmt.Fprintf(&sb, "- **[%s]** %s — %s\n", e.Type, e.Path, e.Message)
	}
	sb.WriteString("\n")

	sb.WriteString("## 需要修复的页面\n\n")
	for _, entry := range entries {
		fmt.Fprintf(&sb, "### %s\n", entry.Title)
		fmt.Fprintf(&sb, "- 路径: %s\n", entry.Path)
		fmt.Fprintf(&sb, "- 描述: %s\n", entry.Description)
		fmt.Fprintf(&sb, "- 复杂度: %s\n", entry.Complexity)
		fmt.Fprintf(&sb, "- 关联 Explore: %v\n\n", entry.ExploreRefs)
	}

	sb.WriteString("## 参考的 Explore 产出\n\n")
	for scope, content := range exploreOutputs {
		fmt.Fprintf(&sb, "### %s\n\n", scope)
		sb.WriteString(content)
		sb.WriteString("\n\n")
	}

	sb.WriteString("请检查上述错误列表中涉及的文件，使用 save_wiki_page 工具重新写入或补全每个页面。\n")
	sb.WriteString("确保每个页面内容完整（不少于 100 字符），且严格按照条目中的 path 参数写入。\n")
	return sb.String()
}

// BuildValidatorUserPrompt 构建文档校验阶段的 user prompt。
//
// 参数说明:
//   - wikiDir: Wiki 输出目录路径
//   - architectOutline: Architect Agent 输出的原始目录大纲 JSON
func BuildValidatorUserPrompt(wikiDir string, architectOutline string) string {
	var sb strings.Builder

	sb.WriteString("请校验以下 Wiki 目录的文档内容质量和一致性。\n\n")

	fmt.Fprintf(&sb, "Wiki 目录: %s\n\n", wikiDir)

	sb.WriteString("## Architect 目录大纲（参考）\n\n")
	sb.WriteString(architectOutline)
	sb.WriteString("\n\n")

	sb.WriteString("请使用 file_read、list_dir 和 file_search 扫描 Wiki 目录，校验页面内容质量和一致性（manifest 已由系统自动生成，无需检查），然后输出校验结果 JSON。\n")
	return sb.String()
}

// ──────────────────────────────────────────────────────────────────────
// Agent 输出解析辅助函数
// ──────────────────────────────────────────────────────────────────────

// parseAgentJSON 从 Agent 响应文本中提取有效 JSON
//
// 解析策略（按优先级依次尝试）:
//  1. 直接解析：TrimSpace 后直接 json.Valid
//  2. Markdown 代码块提取：提取 ```json ... ``` 或 ``` ... ``` 中的内容
//  3. 平衡括号扫描：从每个 `{` 或 `[` 起，按括号深度找到匹配的闭合位置，提取子串尝试解析
//  4. 全部失败 → 返回错误
//
// 兼容场景：LLM 在 JSON 前后输出思考文字（如 "Let me check..."）、
// JSON 中嵌入代码片段含 `{` `}`、LLM 输出多个 JSON 片段取首个有效。
func parseAgentJSON(rawOutput string) (json.RawMessage, error) {
	trimmed := strings.TrimSpace(rawOutput)
	if trimmed == "" {
		return nil, fmt.Errorf("Agent 输出为空")
	}

	if json.Valid([]byte(trimmed)) {
		return json.RawMessage(trimmed), nil
	}

	if extracted := extractMarkdownCodeBlock(trimmed); extracted != "" {
		if json.Valid([]byte(extracted)) {
			return json.RawMessage(extracted), nil
		}
	}

	if extracted := extractBalancedJSON(trimmed); extracted != "" {
		if json.Valid([]byte(extracted)) {
			return json.RawMessage(extracted), nil
		}
	}

	return nil, fmt.Errorf("无法从 Agent 输出中解析有效 JSON（前 200 字符: %s）", truncate(trimmed, 200))
}

// extractMarkdownCodeBlock 从文本中提取 ```json...``` 或 ```...``` 代码块内容
func extractMarkdownCodeBlock(text string) string {
	if content := extractFencedBlock(text, "json"); content != "" {
		return content
	}
	if content := extractFencedBlock(text, ""); content != "" {
		return content
	}
	return ""
}

// extractFencedBlock 提取指定语言标记的 markdown 围栏代码块内容
func extractFencedBlock(text, lang string) string {
	openFence := "```" + lang
	closeFence := "```"

	startIdx := strings.Index(text, openFence)
	if startIdx == -1 {
		return ""
	}
	contentStart := startIdx + len(openFence)
	for contentStart < len(text) && (text[contentStart] == '\n' || text[contentStart] == '\r') {
		contentStart++
	}

	remaining := text[contentStart:]
	before, _, found := strings.Cut(remaining, closeFence)
	if !found {
		return ""
	}

	return strings.TrimSpace(before)
}

// extractBalancedJSON 用括号深度扫描从文本中提取首个完整 JSON 对象或数组
//
// 扫描策略：遍历文本中每个 `{` 或 `[` 作为候选起点，从该位置起跟踪括号深度
// （`{`/`[` +1，`}`/`]` -1，跳过字符串字面量内的括号），深度归零即得到一段
// 完整的候选 JSON 子串。返回首个扫描到的候选（不保证 valid，由调用方校验）。
//
// 兼容 LLM 在 JSON 前后输出思考文字、JSON 中嵌入含括号的代码片段等场景。
// 相比 "first `{` to last `}`" 的粗暴做法，能正确处理嵌套和代码干扰。
func extractBalancedJSON(text string) string {
	for i := 0; i < len(text); i++ {
		c := text[i]
		if c != '{' && c != '[' {
			continue
		}
		if candidate := scanBalancedFrom(text, i); candidate != "" {
			return candidate
		}
	}
	return ""
}

// scanBalancedFrom 从 start 位置开始做括号深度扫描，返回首个闭合的候选子串
//
// 跳过字符串字面量（"..."，含转义）和 markdown 代码块（```）内的括号。
// 深度归零时返回 text[start:end+1]；扫描到文本末尾仍未闭合则返回空串。
func scanBalancedFrom(text string, start int) string {
	if start >= len(text) {
		return ""
	}
	open := text[start]
	var close byte
	switch open {
	case '{':
		close = '}'
	case '[':
		close = ']'
	default:
		return ""
	}

	depth := 0
	inString := false
	escape := false

	for i := start; i < len(text); i++ {
		c := text[i]

		if escape {
			escape = false
			continue
		}
		if inString {
			switch c {
			case '\\':
				escape = true
			case '"':
				inString = false
			}
			continue
		}

		switch c {
		case '"':
			inString = true
		case open:
			depth++
		case close:
			depth--
			if depth == 0 {
				return text[start : i+1]
			}
		}
	}
	return ""
}

// truncate 将字符串截断到指定长度，超长时追加 "..."
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
