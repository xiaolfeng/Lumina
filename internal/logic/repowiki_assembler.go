// Package logic RepoWiki 文档组装器。
//
// DocumentAssembler 负责将 4 个 Agent Pass 的 JSON 输出组装为最终 Wiki 文档。
// 纯模板转换，不调用 LLM；即使某个 Pass 失败也生成占位内容，保证产物完整。
package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"
	"sort"
	"strings"
	"time"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"

	"github.com/xiaolfeng/Lumina/internal/service"
)

// ──────────────────────────────────────────────────────────────────────
// DocumentAssembler
// ──────────────────────────────────────────────────────────────────────

// DocumentAssembler 文档组装器
//
// 读取 4 个 Pass 的 JSON 输出，生成结构化 Markdown Wiki 文档 + manifest.json。
// 职责仅限于格式转换（JSON → Markdown），不调用 LLM，不修改 Pass JSON 源文件。
//
// 生成产物布局（{wikiPath} = storage.GetWikiPath(projectID)）：
//
//	{wikiPath}/
//	├── 主页.md                       # 项目概览摘要 + 导航
//	├── content/
//	│   ├── 项目概览.md               # Pass 1
//	│   ├── 模块分析.md               # Pass 2 + Mermaid 模块依赖图
//	│   ├── 架构设计.md               # Pass 3 + Mermaid 图 + 时序图
//	│   └── 阅读指南.md               # Pass 4
//	└── meta/
//	    └── repowiki-metadata.json   # 侧边栏导航
type DocumentAssembler struct {
	storage *service.WikiStorageService
	log     *xLog.LogNamedLogger
}

// NewDocumentAssembler 创建 DocumentAssembler 实例
func NewDocumentAssembler(storage *service.WikiStorageService, log *xLog.LogNamedLogger) *DocumentAssembler {
	if log == nil {
		log = xLog.WithName(xLog.NamedLOGC, "DocumentAssembler")
	}
	return &DocumentAssembler{
		storage: storage,
		log:     log,
	}
}

// ──────────────────────────────────────────────────────────────────────
// Assemble
// ──────────────────────────────────────────────────────────────────────

// Assemble 组装 Wiki 文档
//
// 将 4 个 Pass 的 JSON 结果转换为 Markdown 文档并写入文件系统。
// 即使某个 Pass 失败（Success=false 或 JSON 为空），也会生成占位段落，不中断整体组装。
//
// 参数说明:
//   - ctx: 上下文（用于错误构造 + 取消检查）
//   - passResults: 4 个 Pass 的结果（key 为 "pass1"/"pass2"/"pass3"/"pass4"）
//   - fileScan: 文件扫描结果（用于主页统计信息，可为 nil）
//   - depSummary: 依赖摘要（用于模块分析 Mermaid 图，可为 nil）
//   - projectID: 项目 ID（确定 Wiki 输出路径）
//   - language: Wiki 语言标识（写入 metadata，不影响路径——路径由 storage 决定）
func (a *DocumentAssembler) Assemble(
	ctx context.Context,
	passResults map[string]*PassResult,
	fileScan *service.FileScanResult,
	depSummary *service.DependencySummary,
	projectID int64,
	language string,
) *xError.Error {
	wikiPath := a.storage.GetWikiPath(projectID)
	a.log.Info(ctx, "开始组装 Wiki 文档",
		slog.String("wikiPath", wikiPath),
		slog.Int64("projectID", projectID),
		slog.String("language", language))

	// 确保核心子目录存在
	contentDir := filepath.Join(wikiPath, "content")
	metaDir := filepath.Join(wikiPath, "meta")
	for _, dir := range []string{wikiPath, contentDir, metaDir} {
		if err := a.storage.EnsureDir(dir); err != nil {
			return xError.NewError(ctx, xError.ServerInternalError,
				xError.ErrMessage("创建 Wiki 目录失败 "+dir), true, err)
		}
	}

	// 解析各 Pass JSON（失败返回空 map，后续段落降级为占位）
	pass1 := asPassMap(passResults["pass1"])
	pass2 := asPassMap(passResults["pass2"])
	pass3 := asPassMap(passResults["pass3"])
	pass4 := asPassMap(passResults["pass4"])

	projectName := getString(pass1, "project_name", "未命名项目")

	// ── 生成各文档 ──
	homeMD := a.buildHomePage(projectName, pass1, pass2, fileScan, language)
	overviewMD := a.buildOverviewPage(pass1, fileScan)
	moduleMD := a.buildModulePage(pass2, depSummary)
	archMD := a.buildArchitecturePage(pass3)
	guideMD := a.buildGuidePage(pass4)

	// ── 写入 Markdown 文件 ──
	mdFiles := map[string]string{
		filepath.Join(wikiPath, "主页.md"):     homeMD,
		filepath.Join(contentDir, "项目概览.md"): overviewMD,
		filepath.Join(contentDir, "模块分析.md"): moduleMD,
		filepath.Join(contentDir, "架构设计.md"): archMD,
		filepath.Join(contentDir, "阅读指南.md"): guideMD,
	}
	for path, content := range mdFiles {
		if xErr := a.storage.WriteMarkdown(path, content); xErr != nil {
			return xError.NewError(ctx, xError.ServerInternalError,
				xError.ErrMessage("写入 Markdown 失败 "+path), true, xErr)
		}
	}

	// ── 写入 metadata ──
	manifest := buildMetadata(projectName, language)
	manifestPath := a.storage.GetManifestPath(projectID)
	if xErr := a.storage.WriteJSON(manifestPath, manifest); xErr != nil {
		return xError.NewError(ctx, xError.ServerInternalError,
			xError.ErrMessage("写入 metadata 失败 "+manifestPath), true, xErr)
	}

	a.log.Info(ctx, "Wiki 文档组装完成",
		slog.Int("files", len(mdFiles)+1),
		slog.String("wikiPath", wikiPath))
	return nil
}

// ──────────────────────────────────────────────────────────────────────
// 页面构建方法
// ──────────────────────────────────────────────────────────────────────

// buildHomePage 构建主页 Markdown
func (a *DocumentAssembler) buildHomePage(
	projectName string,
	pass1 map[string]interface{},
	_ map[string]interface{},
	fileScan *service.FileScanResult,
	language string,
) string {
	desc := getString(pass1, "description", "暂无项目简介")
	techStack := getStringSlice(pass1, "tech_stack")

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", escapeMD(projectName)))
	sb.WriteString(fmt.Sprintf("%s\n\n", desc))
	sb.WriteString("## 快速导航\n\n")
	sb.WriteString("- [项目概览](content/项目概览.md)\n")
	sb.WriteString("- [模块分析](content/模块分析.md)\n")
	sb.WriteString("- [架构设计](content/架构设计.md)\n")
	sb.WriteString("- [阅读指南](content/阅读指南.md)\n\n")

	// 技术栈
	sb.WriteString("## 技术栈\n\n")
	if len(techStack) > 0 {
		for _, t := range techStack {
			sb.WriteString(fmt.Sprintf("- %s\n", t))
		}
	} else {
		sb.WriteString("_暂无技术栈信息_\n")
	}
	sb.WriteString("\n")

	// 项目统计
	sb.WriteString("## 项目统计\n\n")
	if fileScan != nil {
		sb.WriteString(fmt.Sprintf("- 文件数：%d\n", fileScan.TotalFiles))
		sb.WriteString(fmt.Sprintf("- 总大小：%s\n", humanSize(fileScan.TotalSize)))
		sb.WriteString(fmt.Sprintf("- 主要语言：%s\n", topLanguages(fileScan.LanguageStats)))
	} else {
		sb.WriteString("- 文件数：_未提供_\n")
		sb.WriteString("- 主要语言：_未提供_\n")
	}
	sb.WriteString(fmt.Sprintf("- 分析时间：%s\n", time.Now().Format("2006-01-02")))
	sb.WriteString(fmt.Sprintf("- 语言：%s\n", language))
	sb.WriteString("\n")
	return sb.String()
}

// buildOverviewPage 构建项目概览页 Markdown
func (a *DocumentAssembler) buildOverviewPage(pass1 map[string]interface{}, fileScan *service.FileScanResult) string {
	var sb strings.Builder
	sb.WriteString("# 项目概览\n\n")

	desc := getString(pass1, "description", "")
	if desc == "" {
		sb.WriteString("> ⚠️ 项目概览分析未产出（Pass 1 失败或未执行），以下为占位内容。\n\n")
		sb.WriteString("_暂无项目描述_\n\n")
	} else {
		sb.WriteString(desc)
		sb.WriteString("\n\n")
	}

	// 技术栈
	sb.WriteString("## 技术栈\n\n")
	techStack := getStringSlice(pass1, "tech_stack")
	if len(techStack) > 0 {
		for _, t := range techStack {
			sb.WriteString(fmt.Sprintf("- %s\n", t))
		}
	} else {
		sb.WriteString("_暂无_\n")
	}
	sb.WriteString("\n")

	// 目录结构
	sb.WriteString("## 目录结构\n\n")
	dirStruct := getString(pass1, "directory_structure", "")
	if dirStruct != "" {
		sb.WriteString("```\n")
		sb.WriteString(dirStruct)
		sb.WriteString("\n```\n\n")
	} else {
		sb.WriteString("_暂无_\n\n")
	}

	// 入口文件
	sb.WriteString("## 入口文件\n\n")
	entryPoints := getStringSlice(pass1, "entry_points")
	if len(entryPoints) > 0 {
		for _, e := range entryPoints {
			sb.WriteString(fmt.Sprintf("- `%s`\n", e))
		}
	} else if fileScan != nil && len(fileScan.EntryPoints) > 0 {
		for _, e := range fileScan.EntryPoints {
			sb.WriteString(fmt.Sprintf("- `%s`\n", e))
		}
	} else {
		sb.WriteString("_暂无_\n")
	}
	sb.WriteString("\n")

	// 语言分布
	sb.WriteString("## 语言分布\n\n")
	langStats := pass1["language_stats"]
	if table := jsonToMarkdownTable(langStats); table != "" {
		sb.WriteString(table)
		sb.WriteString("\n")
	} else if fileScan != nil && len(fileScan.LanguageStats) > 0 {
		sb.WriteString("| 语言 | 文件数 |\n| --- | --- |\n")
		for lang, count := range fileScan.LanguageStats {
			name := lang
			if name == "" {
				name = "（未识别）"
			}
			sb.WriteString(fmt.Sprintf("| %s | %d |\n", name, count))
		}
		sb.WriteString("\n")
	} else {
		sb.WriteString("_暂无_\n\n")
	}
	return sb.String()
}

// buildModulePage 构建模块分析页 Markdown
func (a *DocumentAssembler) buildModulePage(pass2 map[string]interface{}, depSummary *service.DependencySummary) string {
	var sb strings.Builder
	sb.WriteString("# 模块分析\n\n")

	pass2Failed := len(pass2) == 0

	// 模块依赖关系图（Mermaid）
	sb.WriteString("## 模块依赖关系图\n\n")
	if depSummary != nil && len(depSummary.Modules) > 0 {
		mermaid := mermaidModuleGraph(depSummary)
		sb.WriteString("```mermaid\n")
		sb.WriteString(mermaid)
		sb.WriteString("\n```\n\n")
	} else {
		sb.WriteString("_暂无依赖摘要数据_\n\n")
	}

	// 模块详情
	sb.WriteString("## 模块详情\n\n")
	modules := getSlice(pass2, "modules")
	if len(modules) > 0 {
		for _, m := range modules {
			if mm, ok := m.(map[string]interface{}); ok {
				sb.WriteString(generateModuleSection(mm))
			}
		}
	} else if depSummary != nil && len(depSummary.Modules) > 0 {
		if pass2Failed {
			sb.WriteString("> ⚠️ 模块分析（Pass 2）未产出，以下信息来自依赖摘要（降级展示）。\n\n")
		}
		for _, m := range depSummary.Modules {
			sb.WriteString(generateModuleSectionFromInfo(m))
		}
	} else {
		sb.WriteString("_暂无模块信息_\n\n")
	}
	return sb.String()
}

// buildArchitecturePage 构架构设计页 Markdown
func (a *DocumentAssembler) buildArchitecturePage(pass3 map[string]interface{}) string {
	var sb strings.Builder
	sb.WriteString("# 架构设计\n\n")

	if len(pass3) == 0 {
		sb.WriteString("> ⚠️ 架构分析（Pass 3）未产出，以下为占位内容。\n\n")
	}

	// 架构模式
	sb.WriteString("## 架构模式\n\n")
	pattern := getString(pass3, "architecture_pattern", "")
	if pattern != "" {
		sb.WriteString(pattern)
		sb.WriteString("\n\n")
	} else {
		sb.WriteString("_暂无_\n\n")
	}

	// 设计决策
	sb.WriteString("## 设计决策\n\n")
	decisions := getStringSlice(pass3, "design_decisions")
	if len(decisions) > 0 {
		for _, d := range decisions {
			sb.WriteString(fmt.Sprintf("- %s\n", d))
		}
	} else {
		sb.WriteString("_暂无_\n")
	}
	sb.WriteString("\n")

	// 数据流
	sb.WriteString("## 数据流\n\n")
	dataFlow := getString(pass3, "data_flow", "")
	if dataFlow != "" {
		sb.WriteString(dataFlow)
		sb.WriteString("\n\n")
	} else {
		sb.WriteString("_暂无_\n\n")
	}

	// 架构图
	sb.WriteString("## 架构图\n\n")
	mermaidGraph := getString(pass3, "mermaid_graph", "")
	if mermaidGraph != "" {
		sb.WriteString("```mermaid\n")
		sb.WriteString(mermaidGraph)
		sb.WriteString("\n```\n\n")
	} else {
		sb.WriteString("_暂无_\n\n")
	}

	// 时序图
	sb.WriteString("## 时序图\n\n")
	mermaidSeq := getString(pass3, "mermaid_sequence", "")
	if mermaidSeq != "" {
		sb.WriteString("```mermaid\n")
		sb.WriteString(mermaidSeq)
		sb.WriteString("\n```\n\n")
	} else {
		sb.WriteString("_暂无_\n\n")
	}
	return sb.String()
}

// buildGuidePage 构建阅读指南页 Markdown
func (a *DocumentAssembler) buildGuidePage(pass4 map[string]interface{}) string {
	var sb strings.Builder
	sb.WriteString("# 阅读指南\n\n")

	if len(pass4) == 0 {
		sb.WriteString("> ⚠️ 阅读指南（Pass 4）未产出，以下为占位内容。\n\n")
	}

	// 推荐阅读顺序
	sb.WriteString("## 推荐阅读顺序\n\n")
	readingOrder := getStringSlice(pass4, "reading_order")
	if len(readingOrder) > 0 {
		for i, step := range readingOrder {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, step))
		}
	} else {
		sb.WriteString("_暂无_\n")
	}
	sb.WriteString("\n")

	// 新人上手路径
	sb.WriteString("## 新人上手路径\n\n")
	onboarding := getStringSlice(pass4, "onboarding_path")
	if len(onboarding) > 0 {
		for i, step := range onboarding {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, step))
		}
	} else {
		sb.WriteString("_暂无_\n")
	}
	sb.WriteString("\n")

	// 关键代码段落
	sb.WriteString("## 关键代码段落\n\n")
	keySections := getSlice(pass4, "key_code_sections")
	if len(keySections) > 0 {
		for _, ks := range keySections {
			if m, ok := ks.(map[string]interface{}); ok {
				file := getString(m, "file", "未知文件")
				description := getString(m, "description", "")
				sb.WriteString(fmt.Sprintf("- **`%s`**：%s\n", file, description))
			}
		}
	} else {
		sb.WriteString("_暂无_\n")
	}
	sb.WriteString("\n")

	// 常见问题
	sb.WriteString("## 常见问题\n\n")
	faqs := getSlice(pass4, "faq")
	if len(faqs) > 0 {
		for _, f := range faqs {
			if m, ok := f.(map[string]interface{}); ok {
				q := getString(m, "q", "未命名问题")
				ans := getString(m, "a", "")
				sb.WriteString(fmt.Sprintf("### %s\n\n%s\n\n", q, ans))
			}
		}
	} else {
		sb.WriteString("_暂无_\n\n")
	}
	return sb.String()
}

// ──────────────────────────────────────────────────────────────────────
// Markdown 格式化 Helper
// ──────────────────────────────────────────────────────────────────────

// navEntry metadata 导航条目
type navEntry struct {
	Title string `json:"title"`
	Path  string `json:"path"`
}

// repowikiMetadata Wiki 元数据清单
type repowikiMetadata struct {
	Navigation  []navEntry `json:"navigation"`
	Home        string     `json:"home"`
	Language    string     `json:"language"`
	ProjectName string     `json:"project_name"`
}

// buildMetadata 构建侧边栏导航元数据
func buildMetadata(projectName, language string) repowikiMetadata {
	if language == "" {
		language = "zh"
	}
	return repowikiMetadata{
		Navigation: []navEntry{
			{Title: "主页", Path: "主页.md"},
			{Title: "项目概览", Path: "content/项目概览.md"},
			{Title: "模块分析", Path: "content/模块分析.md"},
			{Title: "架构设计", Path: "content/架构设计.md"},
			{Title: "阅读指南", Path: "content/阅读指南.md"},
		},
		Home:        "主页.md",
		Language:    language,
		ProjectName: projectName,
	}
}

// jsonToMarkdownTable 将 JSON 对象/数组转为 Markdown 表格
//
// 支持类型：
//   - map[string]interface{} → 双列表格（字段 | 值）
//   - []map[string]interface{} → 多列表格（列名取自首个元素）
//   - []interface{} 元素为 map → 同上
//
// 无法格式化时返回空字符串。
func jsonToMarkdownTable(data interface{}) string {
	if data == nil {
		return ""
	}

	switch v := data.(type) {
	case map[string]interface{}:
		if len(v) == 0 {
			return ""
		}
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var sb strings.Builder
		sb.WriteString("| 字段 | 值 |\n| --- | --- |\n")
		for _, k := range keys {
			sb.WriteString(fmt.Sprintf("| %s | %v |\n", k, v[k]))
		}
		return sb.String()

	case []interface{}:
		if len(v) == 0 {
			return ""
		}
		// 取首个元素作为 map 提取列名
		first, ok := v[0].(map[string]interface{})
		if !ok {
			return ""
		}
		keys := make([]string, 0, len(first))
		for k := range first {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		var sb strings.Builder
		// 表头
		header := make([]string, 0, len(keys))
		for _, k := range keys {
			header = append(header, k)
		}
		sb.WriteString("| " + strings.Join(header, " | ") + " |\n")
		sb.WriteString("| " + strings.Repeat("--- | ", len(keys)) + "\n")
		// 行
		for _, item := range v {
			m, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			row := make([]string, 0, len(keys))
			for _, k := range keys {
				row = append(row, fmt.Sprintf("%v", m[k]))
			}
			sb.WriteString("| " + strings.Join(row, " | ") + " |\n")
		}
		return sb.String()

	default:
		return ""
	}
}

// mermaidModuleGraph 从 DependencySummary 生成 Mermaid 模块依赖图
//
// 输出示例：
//
//	graph TD
//	  handler --> logic
//	  main --> handler
//	  logic --> fmt
//
// 为避免图表过于密集，模块名中的 "/" 会被替换为 "_" 作为节点 ID。
// 当模块数量超过 40 时仅保留 CoreModules 及其直接依赖，防止图表爆炸。
func mermaidModuleGraph(depSummary *service.DependencySummary) string {
	if depSummary == nil || len(depSummary.Modules) == 0 {
		return "graph TD\n  %% 空依赖图"
	}

	// 节点超过阈值时裁剪
	moduleList := depSummary.Modules
	if len(moduleList) > 40 && len(depSummary.CoreModules) > 0 {
		coreSet := make(map[string]bool, len(depSummary.CoreModules))
		for _, c := range depSummary.CoreModules {
			coreSet[c] = true
		}
		filtered := make([]service.ModuleInfo, 0, len(depSummary.CoreModules))
		for _, m := range moduleList {
			if coreSet[m.Name] {
				filtered = append(filtered, m)
			}
		}
		moduleList = filtered
	}

	var sb strings.Builder
	sb.WriteString("graph TD\n")
	for _, m := range moduleList {
		fromNode := sanitizeMermaidNode(m.Name)
		for _, dep := range m.Dependencies {
			toNode := sanitizeMermaidNode(dep)
			if toNode == "" {
				continue
			}
			sb.WriteString(fmt.Sprintf("  %s --> %s\n", fromNode, toNode))
		}
	}
	return strings.TrimSuffix(sb.String(), "\n")
}

// sanitizeMermaidNode 将模块名转换为合法 Mermaid 节点 ID
//
// Mermaid 节点 ID 不允许含 "/"、"-"（除非用引号包裹）。
// 将 "/" 替换为 "_"，"." 替换为 "root"。
func sanitizeMermaidNode(name string) string {
	if name == "" {
		return ""
	}
	if name == "." {
		return "root"
	}
	r := strings.NewReplacer("/", "_", "-", "_", ".", "_")
	return r.Replace(name)
}

// generateModuleSection 生成单个模块的 Markdown 段落（来自 Pass 2 JSON map）
//
// Pass 2 模块结构：{name, responsibility, key_files[], dependencies[], interfaces[]}
func generateModuleSection(module map[string]interface{}) string {
	name := getString(module, "name", "未命名模块")
	responsibility := getString(module, "responsibility", "")
	keyFiles := getStringSlice(module, "key_files")
	dependencies := getStringSlice(module, "dependencies")
	interfaces := getStringSlice(module, "interfaces")

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("### %s\n\n", escapeMD(name)))
	if responsibility != "" {
		sb.WriteString(fmt.Sprintf("- **职责**：%s\n", responsibility))
	} else {
		sb.WriteString("- **职责**：_未提供_\n")
	}
	if len(keyFiles) > 0 {
		sb.WriteString(fmt.Sprintf("- **关键文件**：`%s`\n", strings.Join(keyFiles, "`, `")))
	} else {
		sb.WriteString("- **关键文件**：_无_\n")
	}
	if len(dependencies) > 0 {
		sb.WriteString(fmt.Sprintf("- **依赖**：%s\n", strings.Join(dependencies, ", ")))
	} else {
		sb.WriteString("- **依赖**：无\n")
	}
	if len(interfaces) > 0 {
		sb.WriteString(fmt.Sprintf("- **接口**：%s\n", strings.Join(interfaces, ", ")))
	} else {
		sb.WriteString("- **接口**：_暂无_\n")
	}
	sb.WriteString("\n")
	return sb.String()
}

// generateModuleSectionFromInfo 从 DependencySummary.ModuleInfo 生成模块段落（降级模式）
//
// ModuleInfo 没有 responsibility 字段，以 Path 替代展示。
func generateModuleSectionFromInfo(m service.ModuleInfo) string {
	var sb strings.Builder
	displayName := m.Name
	if displayName == "" || displayName == "." {
		displayName = "根模块"
	}
	sb.WriteString(fmt.Sprintf("### %s\n\n", escapeMD(displayName)))
	sb.WriteString(fmt.Sprintf("- **路径**：`%s`\n", m.Path))
	if len(m.KeyFiles) > 0 {
		sb.WriteString(fmt.Sprintf("- **关键文件**：`%s`\n", strings.Join(m.KeyFiles, "`, `")))
	} else {
		sb.WriteString("- **关键文件**：_无_\n")
	}
	if len(m.Dependencies) > 0 {
		sb.WriteString(fmt.Sprintf("- **依赖**：%s\n", strings.Join(m.Dependencies, ", ")))
	} else {
		sb.WriteString("- **依赖**：无\n")
	}
	if len(m.Interfaces) > 0 {
		sb.WriteString(fmt.Sprintf("- **接口**：%s\n", strings.Join(m.Interfaces, ", ")))
	} else {
		sb.WriteString("- **接口**：_暂无_\n")
	}
	sb.WriteString("\n")
	return sb.String()
}

// ──────────────────────────────────────────────────────────────────────
// JSON 访问辅助函数（处理 LLM 输出类型不确定问题）
// ──────────────────────────────────────────────────────────────────────

// asPassMap 将 PassResult.JSON 解析为 map（失败或 nil 返回空 map）
func asPassMap(result *PassResult) map[string]interface{} {
	m := make(map[string]interface{})
	if result == nil || !result.Success || len(result.JSON) == 0 {
		return m
	}
	// 忽略解析错误，返回已解析部分
	_ = json.Unmarshal(result.JSON, &m)
	return m
}

// getString 安全提取 map 中的字符串值，缺失或类型不符时返回默认值
func getString(m map[string]interface{}, key, defaultVal string) string {
	if m == nil {
		return defaultVal
	}
	v, ok := m[key]
	if !ok || v == nil {
		return defaultVal
	}
	switch s := v.(type) {
	case string:
		if s == "" {
			return defaultVal
		}
		return s
	default:
		// 非 string 尝试 fmt 格式化
		return fmt.Sprintf("%v", v)
	}
}

// getStringSlice 安全提取 map 中的字符串切片
//
// 兼容 LLM 可能输出 ["a","b"] 或 "a,b" 两种形式。
func getStringSlice(m map[string]interface{}, key string) []string {
	if m == nil {
		return nil
	}
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	switch s := v.(type) {
	case []interface{}:
		result := make([]string, 0, len(s))
		for _, item := range s {
			if str, ok := item.(string); ok && str != "" {
				result = append(result, str)
			}
		}
		return result
	case []string:
		return s
	case string:
		// 兼容以逗号分隔的字符串
		parts := strings.Split(s, ",")
		result := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				result = append(result, p)
			}
		}
		return result
	default:
		return nil
	}
}

// getSlice 安全提取 map 中的切片
func getSlice(m map[string]interface{}, key string) []interface{} {
	if m == nil {
		return nil
	}
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	if s, ok := v.([]interface{}); ok {
		return s
	}
	return nil
}

// ──────────────────────────────────────────────────────────────────────
// 通用辅助函数
// ──────────────────────────────────────────────────────────────────────

// escapeMD 转义 Markdown 特殊字符（仅处理标题场景中常见的管道符）
func escapeMD(s string) string {
	// 仅转义管道符避免破坏表格，其他字符保留以保持可读性
	return strings.ReplaceAll(s, "|", "\\|")
}

// humanSize 将字节数转为人类可读的尺寸字符串
func humanSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// topLanguages 从 LanguageStats 提取 top-3 语言（按文件数降序）
func topLanguages(stats map[string]int) string {
	if len(stats) == 0 {
		return "_未提供_"
	}
	type kv struct {
		lang  string
		count int
	}
	items := make([]kv, 0, len(stats))
	for lang, count := range stats {
		name := lang
		if name == "" {
			name = "其他"
		}
		items = append(items, kv{name, count})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].count > items[j].count
	})
	limit := 3
	if len(items) < limit {
		limit = len(items)
	}
	parts := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		parts = append(parts, fmt.Sprintf("%s（%d 文件）", items[i].lang, items[i].count))
	}
	return strings.Join(parts, "、")
}
