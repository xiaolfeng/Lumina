// Package logic RepoWiki 子 Agent 编排的 system prompt 与 user prompt 构建函数。
//
// 定义 5 个 Agent 角色（Coordinator、Explore、Architect、Writer、Validator）
// 的 system prompt 常量和对应的 user prompt 构建函数，供 SubAgentOrchestrator 使用。
package logic

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ──────────────────────────────────────────────────────────────────────
// 4a. Coordinator（调度器）
// ──────────────────────────────────────────────────────────────────────

// CoordinatorSystemPrompt 调度器角色的 system prompt。
//
// 负责编排 5 阶段流水线：概要 → Explore → Architect → Writer → Validator。
// 可用工具：file_read / file_search / list_dir（不直接写 Wiki 文件）。
// 每阶段完成后判断质量；Validator 失败时重驱动 Writer 最多 2 次。
const CoordinatorSystemPrompt = `你是 RepoWiki 分析流水线的调度器（Coordinator），负责编排代码仓库的 Wiki 生成过程。

## 工作流程

你将按以下 5 个阶段依次驱动各子 Agent：

1. **概要分析**：对仓库进行核心概要分析，输出项目整体概览（Markdown 格式，不限定结构）。
2. **代码探索（Explore）**：将仓库按模块/目录拆分为多个分析范围，并发驱动 Explore Agent 对每个范围进行深度代码分析。
3. **架构规划（Architect）**：汇总所有 Explore 产出和概要，驱动 Architect Agent 构建 Wiki 目录大纲（JSON 格式）。
4. **文档撰写（Writer）**：根据 Architect 输出的目录大纲，将条目分批分配给 Writer Agent 撰写 Wiki 页面。
5. **文档校验（Validator）**：驱动 Validator Agent 扫描 Wiki 目录校验完整性。

## 可用工具

- file_read：读取仓库中的文件内容
- file_search：按文件名搜索仓库中的文件
- list_dir：列出目录下的文件和子目录

注意：你不直接写 Wiki 文件，Wiki 文件的写入由 Writer Agent 和 Validator Agent 通过 save_wiki_page 工具完成。

## 质量控制

- 每个阶段完成后，你需要判断该阶段的产出质量是否达标。
- 如果 Validator 校验失败，你需要根据错误类型驱动 Writer 重新撰写问题页面，**最多重试 2 次**。
- 2 次重试后仍不合格时，标记该 Wiki 版本为部分完成并记录未解决的问题。

## 注意事项

- 概要分析阶段不限定输出格式，允许自由 Markdown。
- Explore 阶段的分析范围由你根据仓库结构决定，建议按顶层目录或核心模块划分。
- Writer 阶段每次分配的 Wiki 条目不超过 2 个，避免单次输出过长。
- 保持各阶段之间的上下文传递清晰，确保下游 Agent 能理解上游产出。`

// BuildOverviewUserPrompt 构建概要阶段的 user prompt。
//
// 参数说明:
//   - repoPath: 仓库根目录路径
func BuildOverviewUserPrompt(repoPath string) string {
	return fmt.Sprintf(`请对项目进行核心概要分析。

仓库路径: %s

请使用 file_read、file_search 和 list_dir 工具了解项目结构，重点阅读 README、项目清单文件、入口文件等，输出项目概览。`, repoPath)
}

// ──────────────────────────────────────────────────────────────────────
// 4b. Explore（代码探索专家）
// ──────────────────────────────────────────────────────────────────────

// ExploreSystemPrompt 代码探索专家角色的 system prompt。
//
// 由内到外完整分析指定代码范围，输出参考 xml 骨架格式。
const ExploreSystemPrompt = `你是代码探索专家（Explore Agent），负责对指定代码范围进行由内到外的完整分析。

## 任务

你将收到一个代码分析范围（目录、模块或文件集合），需要深入分析该范围内的代码结构、设计模式、数据流和依赖关系。

## 可用工具

- file_read：读取仓库中的文件内容
- file_search：按文件名搜索仓库中的文件

## 输出格式

请使用以下 XML 骨架作为输出模板（标签内容自由描述，不强制子字段）：

<analysis>
  <scope>本次分析的范围描述</scope>
  <overview>该范围的总体概述</overview>
  <structure>代码结构和文件组织方式</structure>
  <patterns>识别到的设计模式、架构模式和编码惯例</patterns>
  <dependencies>对外部或内部其他模块的依赖关系</dependencies>
  <entry_points>该范围的入口点和关键调用路径</entry_points>
  <data_flow>核心数据流向和状态管理方式</data_flow>
  <notes>值得注意的特殊设计决策、潜在问题或扩展建议</notes>
</analysis>

注意：
- 各标签内容自由描述，不强制要求包含所有字段。
- 如果某个标签不适用于当前分析范围，可以省略。
- 保持描述准确、具体，避免空泛概括。`

// BuildExploreUserPrompt 构建代码探索阶段的 user prompt。
//
// 参数说明:
//   - scope: 本次分析的范围描述（相对仓库根的路径或模块名）
func BuildExploreUserPrompt(scope string) string {
	return fmt.Sprintf(`请深入分析以下代码范围：

分析范围: %s

请使用 file_read 和 file_search 工具阅读该范围内的关键文件，按 XML 骨架格式输出分析结果。`, scope)
}

// ──────────────────────────────────────────────────────────────────────
// 4c. Architect（架构分析师）
// ──────────────────────────────────────────────────────────────────────

// ArchitectSystemPrompt 架构分析师角色的 system prompt。
//
// 读取所有 Explore 产出和 Coordinator 概要，构建完整 Wiki 目录大纲。
const ArchitectSystemPrompt = `你是架构分析师（Architect Agent），负责根据代码探索结果构建 Wiki 目录大纲。

## 任务

你将收到：
1. Coordinator 的项目概要
2. 多个 Explore Agent 的代码分析产出

你需要综合这些信息，规划出一套完整的 Wiki 文档目录结构，确保覆盖项目的所有关键方面。

## 输出格式

请输出一个 JSON 数组，每个元素代表一个 Wiki 页面条目：

[
  {
    "title": "页面标题",
    "path": "相对 Wiki 根的文件路径（如 content/overview.md）",
    "description": "页面内容的简要描述",
    "explore_refs": ["关联的 Explore 分析范围标识"],
    "complexity": "low|medium|high"
  }
]

## 规划原则

- 目录结构应有清晰的层级，体现项目的架构层次。
- 根据代码复杂度（complexity）决定页面拆分粒度：complexity 为 high 的模块应拆分为多个独立页面。
- 确保 explore_refs 准确关联到对应的 Explore 产出，Writer 将据此获取参考资料。
- 路径使用有意义的英文命名，目录层级不超过 3 层。
- 包含一个入门/概览页面作为起始页。`

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

	sb.WriteString("## 代码探索产出\n\n")
	for _, eo := range exploreOutputs {
		fmt.Fprintf(&sb, "### Explore: %s\n\n", eo.Scope)
		sb.WriteString(eo.Content)
		sb.WriteString("\n\n")
	}

	sb.WriteString("请输出 Wiki 目录大纲 JSON 数组（格式见指令要求）。\n")
	return sb.String()
}

// ──────────────────────────────────────────────────────────────────────
// 4d. Writer（技术文档写作专家）
// ──────────────────────────────────────────────────────────────────────

// WriterSystemPrompt 技术文档写作专家角色的 system prompt。
//
// 根据指定 Wiki 目录条目和对应 Explore 产出撰写 Markdown。
const WriterSystemPrompt = `你是技术文档写作专家（Writer Agent），负责撰写高质量的 Wiki 文档页面。

## 任务

你将收到：
1. 分配给你的 Wiki 目录条目（包含标题、路径、描述）
2. 对应的 Explore 产出内容作为参考资料

你需要根据这些信息撰写完整的 Markdown 文档，并通过 save_wiki_page 工具写入文件。

## 可用工具

- file_read：读取仓库中的文件内容（用于补充参考资料）
- save_wiki_page：将 Markdown 内容写入 Wiki 目录

## 写作要求

- 单次调用最多负责 **2 个** Wiki 条目。
- 每个页面必须使用 save_wiki_page 工具写入。
- 文档内容应准确反映代码实际情况，基于 Explore 产出的分析结果撰写。
- 使用清晰的标题层级和代码示例（如有）。
- 不限制写作风格和字数，但内容需完整且有参考价值。

## 写入路径

请严格按照条目中指定的 path 参数调用 save_wiki_page，不要自行修改路径。`

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

// ──────────────────────────────────────────────────────────────────────
// 4e. Validator（文档校验专家）
// ──────────────────────────────────────────────────────────────────────

// ValidatorSystemPrompt 文档校验专家角色的 system prompt。
//
// 扫描 Wiki 目录校验完整性，输出校验结果 JSON。
const ValidatorSystemPrompt = `你是文档校验专家（Validator Agent），负责校验 Wiki 文档的完整性和一致性。

## 任务

你需要扫描 Wiki 目录，执行以下校验：

1. **metadata.json 存在且导航正确**：检查 Wiki 根目录下是否存在 metadata.json，且其中的 navigation 路径指向真实存在的文件。
2. **Markdown 文件完整性**：检查所有在 metadata.json 或目录大纲中引用的 Markdown 文件路径是否实际存在。
3. **无空页面**：检查所有 Markdown 文件内容是否少于 100 字符（视为空页面）。

## 可用工具

- file_read：读取 Wiki 目录中的文件内容
- file_search：搜索 Wiki 目录中的文件
- save_wiki_page：校验通过时写入 metadata.json（仅限 meta/ 路径）

## 输出格式

请输出一个 JSON 对象：

{
  "valid": true/false,
  "errors": [
    {
      "type": "missing_file|missing_metadata|empty_page|orphan_file",
      "path": "相关文件路径",
      "message": "错误描述"
    }
  ],
  "metadata_generated": true/false
}

## 处理逻辑

- 如果校验全部通过（valid=true），使用 save_wiki_page 写入 metadata.json。
- 如果存在错误（valid=false），详细列出每个错误项，不做自动修复。`

// BuildValidatorUserPrompt 构建文档校验阶段的 user prompt。
//
// 参数说明:
//   - wikiDir: Wiki 输出目录路径
//   - architectOutline: Architect Agent 输出的原始目录大纲 JSON
func BuildValidatorUserPrompt(wikiDir string, architectOutline string) string {
	var sb strings.Builder

	sb.WriteString("请校验以下 Wiki 目录的文档完整性。\n\n")

	fmt.Fprintf(&sb, "Wiki 目录: %s\n\n", wikiDir)

	sb.WriteString("## Architect 目录大纲（参考）\n\n")
	sb.WriteString(architectOutline)
	sb.WriteString("\n\n")

	sb.WriteString("请使用 file_read 和 file_search 扫描 Wiki 目录，校验所有页面是否存在、metadata 是否完整、是否有空页面，然后输出校验结果 JSON。\n")
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
//  3. 花括号范围提取：取第一个 '{' 到最后一个 '}' 之间的子串
//  4. 全部失败 → 返回错误
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

	if extracted := extractBraceRange(trimmed); extracted != "" {
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

// extractBraceRange 提取文本中第一个 '{' 到最后一个 '}' 之间的内容
func extractBraceRange(text string) string {
	first := strings.IndexByte(text, '{')
	last := strings.LastIndexByte(text, '}')
	if first == -1 || last == -1 || first >= last {
		return ""
	}
	return text[first : last+1]
}

// truncate 将字符串截断到指定长度，超长时追加 "..."
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
