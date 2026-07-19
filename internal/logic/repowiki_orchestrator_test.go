package logic

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	xLog "github.com/bamboo-services/bamboo-base-go/common/log"

	"github.com/xiaolfeng/Lumina/internal/service"
)

// newTestOrchestrator 构造用于测试的 SubAgentOrchestrator（无 LLM client）。
//
// 通过 REPOWIKI_STORAGE_PATH 环境变量指向临时目录，返回 orchestrator 与临时目录路径。
// 调用方负责在 defer 中清理临时目录与还原环境变量。
func newTestOrchestrator(t *testing.T, versionID int64, projectName, language string) (*SubAgentOrchestrator, string) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "repowiki-orch-test-*")
	if err != nil {
		t.Fatalf("MkdirTemp failed: %v", err)
	}
	t.Setenv("REPOWIKI_STORAGE_PATH", tmpDir)
	storage := service.NewWikiStorageService()
	o := &SubAgentOrchestrator{
		roleClients: nil,
		roleModels:  nil,
		storage:     storage,
		log:         xLog.WithName(xLog.NamedLOGC, "SubAgentOrchestratorTest"),
		versionID:   versionID,
		repoPath:    tmpDir,
		projectName: projectName,
		language:    language,
	}
	return o, tmpDir
}

// ──────────────────────────────────────────────────────────────────────
// TestFindMissingEntries
// ──────────────────────────────────────────────────────────────────────

func TestFindMissingEntries(t *testing.T) {
	// outline entry.Path 为无扩展名路径（.mdx 重构后 manifest path 不带扩展名）
	outline := []WikiEntry{
		{Title: "Overview", Path: "overview"},
		{Title: "Architecture", Path: "architecture"},
		{Title: "Modules", Path: "modules"},
		{Title: "Empty Path", Path: ""},
	}

	t.Run("matches missing_file/empty_page to outline with .mdx normalization", func(t *testing.T) {
		errors := []ValidationError{
			{Type: "missing_file", Path: "overview.mdx", Message: "missing"},
			{Type: "empty_page", Path: "architecture.mdx", Message: "empty"},
		}
		got := findMissingEntries(errors, outline)
		if len(got) != 2 {
			t.Fatalf("expected 2 missing entries, got %d", len(got))
		}
		gotPaths := make(map[string]bool, len(got))
		for _, e := range got {
			gotPaths[e.Path] = true
		}
		for _, want := range []string{"overview", "architecture"} {
			if !gotPaths[want] {
				t.Errorf("expected missing entry with path %q", want)
			}
		}
	})

	t.Run("missing_frontmatter does NOT trigger rewrite", func(t *testing.T) {
		errors := []ValidationError{
			{Type: "missing_frontmatter", Path: "overview.mdx", Message: "no frontmatter"},
			{Type: "missing_file", Path: "architecture.mdx", Message: "real missing"},
		}
		got := findMissingEntries(errors, outline)
		if len(got) != 1 {
			t.Fatalf("expected 1 missing entry (missing_frontmatter skipped), got %d", len(got))
		}
		if got[0].Path != "architecture" {
			t.Errorf("expected architecture, got %q", got[0].Path)
		}
	})

	t.Run("wrong_extension and orphan_file are skipped", func(t *testing.T) {
		errors := []ValidationError{
			{Type: "wrong_extension", Path: "overview.mdx", Message: "wrong ext"},
			{Type: "orphan_file", Path: "modules.mdx", Message: "orphan"},
			{Type: "empty_page", Path: "architecture.mdx", Message: "empty"},
		}
		got := findMissingEntries(errors, outline)
		if len(got) != 1 {
			t.Fatalf("expected 1 missing entry (wrong_extension/orphan_file skipped), got %d", len(got))
		}
		if got[0].Path != "architecture" {
			t.Errorf("expected architecture, got %q", got[0].Path)
		}
	})

	t.Run("non-matching paths return nil", func(t *testing.T) {
		errors := []ValidationError{
			{Type: "missing_file", Path: "nonexistent.mdx", Message: "no match"},
			{Type: "empty_page", Path: "totally/different.mdx", Message: "no match"},
		}
		got := findMissingEntries(errors, outline)
		if got != nil {
			t.Errorf("expected nil when no error path matches outline, got %v", got)
		}
	})

	t.Run("missing_metadata errors are skipped", func(t *testing.T) {
		errors := []ValidationError{
			{Type: "missing_metadata", Path: "overview.mdx", Message: "metadata issue"},
			{Type: "missing_file", Path: "architecture.mdx", Message: "real missing"},
		}
		got := findMissingEntries(errors, outline)
		if len(got) != 1 {
			t.Fatalf("expected 1 missing entry (missing_metadata skipped), got %d", len(got))
		}
		if got[0].Path != "architecture" {
			t.Errorf("expected architecture, got %q", got[0].Path)
		}
	})

	t.Run("empty errors return nil", func(t *testing.T) {
		got := findMissingEntries(nil, outline)
		if got != nil {
			t.Errorf("expected nil for empty errors, got %v", got)
		}
		got = findMissingEntries([]ValidationError{}, outline)
		if got != nil {
			t.Errorf("expected nil for empty errors slice, got %v", got)
		}
	})

	t.Run("only skip-type errors return nil", func(t *testing.T) {
		errors := []ValidationError{
			{Type: "missing_metadata", Path: "overview.mdx", Message: "metadata"},
			{Type: "missing_frontmatter", Path: "architecture.mdx", Message: "no frontmatter"},
			{Type: "wrong_extension", Path: "modules.mdx", Message: "wrong ext"},
			{Type: "orphan_file", Path: "overview.mdx", Message: "orphan"},
		}
		got := findMissingEntries(errors, outline)
		if got != nil {
			t.Errorf("expected nil when only skip-type errors, got %v", got)
		}
	})

	t.Run("path without .mdx suffix also matches", func(t *testing.T) {
		// Validator 可能返回无扩展名路径，此时 TrimSuffix 是 no-op，仍应匹配
		errors := []ValidationError{
			{Type: "missing_file", Path: "overview", Message: "missing"},
		}
		got := findMissingEntries(errors, outline)
		if len(got) != 1 {
			t.Fatalf("expected 1 missing entry, got %d", len(got))
		}
		if got[0].Path != "overview" {
			t.Errorf("expected overview, got %q", got[0].Path)
		}
	})
}

// ──────────────────────────────────────────────────────────────────────
// TestVerifyWriterOutputs
// ──────────────────────────────────────────────────────────────────────

func TestVerifyWriterOutputs(t *testing.T) {
	o, tmpDir := newTestOrchestrator(t, 1001, "test-project", "zh")
	defer os.RemoveAll(tmpDir)

	wikiDir := o.storage.GetWikiPath(o.versionID)
	if err := os.MkdirAll(wikiDir, 0755); err != nil {
		t.Fatalf("MkdirAll wikiDir failed: %v", err)
	}

	// big.mdx: 超过 100 字节（合格）
	bigContent := make([]byte, writerFileMinSize+50)
	for i := range bigContent {
		bigContent[i] = 'x'
	}
	if err := os.WriteFile(filepath.Join(wikiDir, "big.mdx"), bigContent, 0644); err != nil {
		t.Fatalf("write big.mdx failed: %v", err)
	}

	// small.mdx: 小于 100 字节（不合格，视为缺失）
	if err := os.WriteFile(filepath.Join(wikiDir, "small.mdx"), []byte("tiny"), 0644); err != nil {
		t.Fatalf("write small.mdx failed: %v", err)
	}

	// missing: 不创建（缺失）

	// entry.Path 为无扩展名路径，磁盘文件为 {Path}.mdx
	outline := []WikiEntry{
		{Title: "Big", Path: "big"},
		{Title: "Small", Path: "small"},
		{Title: "Missing", Path: "missing"},
		{Title: "Empty Path", Path: ""},
	}

	missing := o.verifyWriterOutputs(outline, wikiDir)
	if len(missing) != 2 {
		t.Fatalf("expected 2 missing/empty entries, got %d", len(missing))
	}

	missingPaths := make(map[string]bool, len(missing))
	for _, e := range missing {
		missingPaths[e.Path] = true
	}
	if !missingPaths["small"] {
		t.Errorf("expected small in missing (size < %d)", writerFileMinSize)
	}
	if !missingPaths["missing"] {
		t.Errorf("expected missing in missing (file not found)")
	}
	if missingPaths["big"] {
		t.Errorf("big should not be in missing (size >= %d)", writerFileMinSize)
	}
}

// ──────────────────────────────────────────────────────────────────────
// TestGenerateManifest
// ──────────────────────────────────────────────────────────────────────

func TestGenerateManifest(t *testing.T) {
	o, tmpDir := newTestOrchestrator(t, 2002, "test-project", "zh")
	defer os.RemoveAll(tmpDir)

	outline := []WikiEntry{
		{Title: "概览", Path: "overview", Description: "项目概览"},
		{Title: "架构", Children: []WikiEntry{
			{Title: "架构设计", Path: "architecture", Description: "架构设计"},
			{Title: "模块说明", Path: "modules", Description: "模块说明"},
		}},
		{Title: "参考", Path: "reference", Description: "参考资料"},
	}
	archOut := ArchitectOutput{Outline: outline}

	if xErr := o.generateManifest(archOut); xErr != nil {
		t.Fatalf("generateManifest failed: %v", xErr)
	}

	manifestPath := o.storage.GetManifestPath(o.versionID)
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest failed: %v", err)
	}

	var m manifestData
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal manifest failed: %v", err)
	}

	if m.ProjectName != "test-project" {
		t.Errorf("expected project_name %q, got %q", "test-project", m.ProjectName)
	}
	if m.Language != "zh" {
		t.Errorf("expected language %q, got %q", "zh", m.Language)
	}
	if m.Home != "overview" {
		t.Errorf("expected home %q (first top-level leaf), got %q", "overview", m.Home)
	}
	if len(m.Navigation) != 3 {
		t.Fatalf("expected 3 top-level nav items, got %d", len(m.Navigation))
	}
	if m.Navigation[0].Title != "概览" || m.Navigation[0].Path != "overview" {
		t.Errorf("nav[0] = (%q, %q), want (\"概览\", \"overview\")", m.Navigation[0].Title, m.Navigation[0].Path)
	}
	if m.Navigation[0].Icon != "FileText" {
		t.Errorf("nav[0].Icon = %q, want \"FileText\" (default, no .mdx file)", m.Navigation[0].Icon)
	}
	if m.Navigation[1].Title != "架构" || m.Navigation[1].Path != "" || len(m.Navigation[1].Children) != 2 {
		t.Errorf("nav[1] should be directory node with 2 children, got title=%q path=%q children=%d",
			m.Navigation[1].Title, m.Navigation[1].Path, len(m.Navigation[1].Children))
	}
	if m.Navigation[1].Icon != "Folder" {
		t.Errorf("nav[1].Icon = %q, want \"Folder\" (default directory icon)", m.Navigation[1].Icon)
	}
	if m.Navigation[1].Children[0].Path != "architecture" || m.Navigation[1].Children[1].Path != "modules" {
		t.Errorf("nav[1].children paths mismatch, got %v", m.Navigation[1].Children)
	}
	if m.Navigation[2].Title != "参考" || m.Navigation[2].Path != "reference" {
		t.Errorf("nav[2] = (%q, %q), want (\"参考\", \"reference\")", m.Navigation[2].Title, m.Navigation[2].Path)
	}
	if m.Meta.Title != "test-project" {
		t.Errorf("expected meta.title %q (fallback to projectName), got %q", "test-project", m.Meta.Title)
	}
}

// TestGenerateManifest_EmptyOutline 验证空 outline 时 home 为空字符串
func TestGenerateManifest_EmptyOutline(t *testing.T) {
	o, tmpDir := newTestOrchestrator(t, 2003, "empty-project", "en")
	defer os.RemoveAll(tmpDir)

	if xErr := o.generateManifest(ArchitectOutput{}); xErr != nil {
		t.Fatalf("generateManifest failed: %v", xErr)
	}

	manifestPath := o.storage.GetManifestPath(o.versionID)
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest failed: %v", err)
	}

	var m manifestData
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal manifest failed: %v", err)
	}
	if m.Home != "" {
		t.Errorf("expected home fallback %q, got %q", "", m.Home)
	}
	if len(m.Navigation) != 0 {
		t.Errorf("expected empty navigation, got %d items", len(m.Navigation))
	}
}

// TestGenerateManifest_DeepTree 验证 3 层嵌套树结构保留 nesting，且 home 取最浅首个叶子
func TestGenerateManifest_DeepTree(t *testing.T) {
	o, tmpDir := newTestOrchestrator(t, 2004, "deep-project", "en")
	defer os.RemoveAll(tmpDir)

	outline := []WikiEntry{
		{Title: "Guide", Path: "guide", Description: "Top-level guide"},
		{Title: "Section A", Children: []WikiEntry{
			{Title: "Subsection A1", Children: []WikiEntry{
				{Title: "Deep Leaf", Path: "deep/leaf", Description: "Deep leaf"},
			}},
			{Title: "Leaf A2", Path: "a2", Description: "A2 leaf"},
		}},
		{Title: "Section B", Path: "b", Description: "Section B"},
	}
	archOut := ArchitectOutput{Outline: outline}

	if xErr := o.generateManifest(archOut); xErr != nil {
		t.Fatalf("generateManifest failed: %v", xErr)
	}

	manifestPath := o.storage.GetManifestPath(o.versionID)
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest failed: %v", err)
	}

	var m manifestData
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal manifest failed: %v", err)
	}

	if m.Home != "guide" {
		t.Errorf("expected home %q (shallowest first leaf), got %q", "guide", m.Home)
	}
	if len(m.Navigation) != 3 {
		t.Fatalf("expected 3 top-level nav items, got %d", len(m.Navigation))
	}
	if len(m.Navigation[1].Children) != 2 {
		t.Errorf("expected section A to have 2 children, got %d", len(m.Navigation[1].Children))
	}
	if len(m.Navigation[1].Children[0].Children) != 1 {
		t.Errorf("expected subsection A1 to have 1 child, got %d", len(m.Navigation[1].Children[0].Children))
	}
	deep := m.Navigation[1].Children[0].Children[0]
	if deep.Path != "deep/leaf" {
		t.Errorf("expected deep leaf path %q, got %q", "deep/leaf", deep.Path)
	}
}

// TestGenerateManifest_NoTopLevelLeaf 验证顶层无叶子时 home 取 DFS 首个叶子
func TestGenerateManifest_NoTopLevelLeaf(t *testing.T) {
	o, tmpDir := newTestOrchestrator(t, 2005, "no-top-leaf-project", "en")
	defer os.RemoveAll(tmpDir)

	outline := []WikiEntry{
		{Title: "Section A", Children: []WikiEntry{
			{Title: "Leaf A1", Path: "a1", Description: "A1"},
			{Title: "Leaf A2", Path: "a2", Description: "A2"},
		}},
		{Title: "Section B", Children: []WikiEntry{
			{Title: "Leaf B1", Path: "b1", Description: "B1"},
		}},
	}
	archOut := ArchitectOutput{Outline: outline}

	if xErr := o.generateManifest(archOut); xErr != nil {
		t.Fatalf("generateManifest failed: %v", xErr)
	}

	manifestPath := o.storage.GetManifestPath(o.versionID)
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest failed: %v", err)
	}

	var m manifestData
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal manifest failed: %v", err)
	}

	if m.Home != "a1" {
		t.Errorf("expected home %q (DFS first leaf in children), got %q", "a1", m.Home)
	}
	if len(m.Navigation) != 2 {
		t.Fatalf("expected 2 top-level nav items, got %d", len(m.Navigation))
	}
	if len(m.Navigation[0].Children) != 2 || len(m.Navigation[1].Children) != 1 {
		t.Errorf("expected nested children preserved, got nav[0].children=%d nav[1].children=%d",
			len(m.Navigation[0].Children), len(m.Navigation[1].Children))
	}
}

// ──────────────────────────────────────────────────────────────────────
// TestFlattenOutlineLeaves
// ──────────────────────────────────────────────────────────────────────

func TestFlattenOutlineLeaves(t *testing.T) {
	outline := []WikiEntry{
		{Title: "Root", Children: []WikiEntry{
			{Title: "Branch A", Children: []WikiEntry{
				{Title: "Leaf A1", Path: "a1", ExploreRefs: []string{"ref1"}, Complexity: "high"},
			}},
			{Title: "Leaf A2", Path: "a2", ExploreRefs: []string{"ref2"}, Complexity: "medium"},
		}},
		{Title: "Leaf B", Path: "b", ExploreRefs: []string{"ref3"}, Complexity: "low"},
		{Title: "Empty Directory", Children: []WikiEntry{}},
	}

	leaves := flattenOutlineLeaves(outline)
	if len(leaves) != 3 {
		t.Fatalf("expected 3 leaves, got %d", len(leaves))
	}

	expectedPaths := []string{"a1", "a2", "b"}
	for i, leaf := range leaves {
		if leaf.Path != expectedPaths[i] {
			t.Errorf("leaves[%d].Path = %q, want %q", i, leaf.Path, expectedPaths[i])
		}
	}
	if leaves[0].Complexity != "high" || leaves[0].ExploreRefs[0] != "ref1" {
		t.Errorf("leaf fields not preserved, got complexity=%q refs=%v", leaves[0].Complexity, leaves[0].ExploreRefs)
	}
	for _, leaf := range leaves {
		if len(leaf.Children) > 0 {
			t.Errorf("leaf %q should not have children", leaf.Path)
		}
	}
}

// ──────────────────────────────────────────────────────────────────────
// TestFindFirstLeafPath
// ──────────────────────────────────────────────────────────────────────

func TestFindFirstLeafPath(t *testing.T) {
	t.Run("top-level leaf exists", func(t *testing.T) {
		outline := []WikiEntry{
			{Title: "A", Path: "a"},
			{Title: "B", Path: "b"},
		}
		if got := findFirstLeafPath(outline); got != "a" {
			t.Errorf("findFirstLeafPath = %q, want %q", got, "a")
		}
	})

	t.Run("top-level all directories", func(t *testing.T) {
		outline := []WikiEntry{
			{Title: "A", Children: []WikiEntry{
				{Title: "B", Path: "b"},
			}},
		}
		if got := findFirstLeafPath(outline); got != "b" {
			t.Errorf("findFirstLeafPath = %q, want %q", got, "b")
		}
	})

	t.Run("empty tree", func(t *testing.T) {
		if got := findFirstLeafPath(nil); got != "" {
			t.Errorf("findFirstLeafPath = %q, want %q", got, "")
		}
		if got := findFirstLeafPath([]WikiEntry{}); got != "" {
			t.Errorf("findFirstLeafPath = %q, want %q", got, "")
		}
	})

	t.Run("skips empty directory nodes", func(t *testing.T) {
		outline := []WikiEntry{
			{Title: "Empty Dir", Children: []WikiEntry{}},
			{Title: "First Leaf", Path: "first"},
		}
		if got := findFirstLeafPath(outline); got != "first" {
			t.Errorf("findFirstLeafPath = %q, want %q", got, "first")
		}
	})
}

// ──────────────────────────────────────────────────────────────────────
// TestExecuteFlowOrder
// ──────────────────────────────────────────────────────────────────────

// TestExecuteFlowOrder 验证 Execute 流程中 generateManifest 在 runValidator 之前执行。
//
// Execute 完整流程需要 LLM client，无法在单元测试中直接驱动。
// 此测试间接验证流程顺序的关键不变量：generateManifest 调用后 manifest 文件存在于
// GetManifestPath 返回的路径，证明 manifest 在 Validator 之前已落盘
// （Execute 中 generateManifest 调用先于 runValidator，见 orchestrator.go:937 vs :944）。
//
// 完整 Execute 流程的端到端验证依赖编译通过 + 手动 QA（需数据库 + LLM 配置）。
func TestExecuteFlowOrder(t *testing.T) {
	o, tmpDir := newTestOrchestrator(t, 3003, "flow-project", "zh")
	defer os.RemoveAll(tmpDir)

	outline := []WikiEntry{
		{Title: "首页", Path: "index"},
		{Title: "指南", Path: "guide"},
	}
	archOut := ArchitectOutput{Outline: outline}

	// 模拟 Execute 中 generateManifest 调用（先于 runValidator）
	if xErr := o.generateManifest(archOut); xErr != nil {
		t.Fatalf("generateManifest failed: %v", xErr)
	}

	// 验证 manifest 文件已存在于预期路径
	manifestPath := o.storage.GetManifestPath(o.versionID)
	if _, err := os.Stat(manifestPath); err != nil {
		t.Fatalf("manifest file should exist at %s before validator runs: %v", manifestPath, err)
	}

	// 验证 manifest 内容可被正确读取（WikiStorageService.ReadJSON）
	var m manifestData
	if xErr := o.storage.ReadJSON(manifestPath, &m); xErr != nil {
		t.Fatalf("ReadJSON manifest failed: %v", xErr)
	}
	if m.Home != "index" {
		t.Errorf("expected home %q, got %q", "index", m.Home)
	}
	if len(m.Navigation) != 2 {
		t.Errorf("expected 2 nav items, got %d", len(m.Navigation))
	}

	// 验证 verifyWriterOutputs 在 manifest 生成前可正常工作
	// （Execute 流程：writers → verifyWriterOutputs → generateManifest → runValidator）
	wikiDir := o.storage.GetWikiPath(o.versionID)
	missing := o.verifyWriterOutputs(outline, wikiDir)
	if len(missing) != 2 {
		t.Errorf("expected 2 missing entries (no .mdx files written), got %d", len(missing))
	}
}

// ──────────────────────────────────────────────────────────────────────
// TestBuildOverviewUserPrompt
// ──────────────────────────────────────────────────────────────────────

func TestBuildOverviewUserPrompt(t *testing.T) {
	t.Run("both_non_empty", func(t *testing.T) {
		result := BuildOverviewUserPrompt("/repo", "关注鉴权", "重点看 JWT")
		if !strings.Contains(result, "## 项目级自定义提示词\n\n关注鉴权") {
			t.Errorf("expected result to contain custom prompt section, got:\n%s", result)
		}
		if !strings.Contains(result, "## 本次分析额外提示词\n\n重点看 JWT") {
			t.Errorf("expected result to contain extra prompt section, got:\n%s", result)
		}
		if !strings.Contains(result, "/repo") {
			t.Errorf("expected result to contain repo path /repo, got:\n%s", result)
		}
	})

	t.Run("custom_prompt_empty", func(t *testing.T) {
		result := BuildOverviewUserPrompt("/repo", "", "重点看 JWT")
		if strings.Contains(result, "## 项目级自定义提示词") {
			t.Errorf("expected result NOT to contain custom prompt section, got:\n%s", result)
		}
		if !strings.Contains(result, "## 本次分析额外提示词\n\n重点看 JWT") {
			t.Errorf("expected result to contain extra prompt section, got:\n%s", result)
		}
	})

	t.Run("extra_prompt_empty", func(t *testing.T) {
		result := BuildOverviewUserPrompt("/repo", "关注鉴权", "")
		if !strings.Contains(result, "## 项目级自定义提示词\n\n关注鉴权") {
			t.Errorf("expected result to contain custom prompt section, got:\n%s", result)
		}
		if strings.Contains(result, "## 本次分析额外提示词") {
			t.Errorf("expected result NOT to contain extra prompt section, got:\n%s", result)
		}
	})

	t.Run("both_empty", func(t *testing.T) {
		result := BuildOverviewUserPrompt("/repo", "", "")
		if strings.Contains(result, "## 项目级自定义提示词") {
			t.Errorf("expected result NOT to contain custom prompt section, got:\n%s", result)
		}
		if strings.Contains(result, "## 本次分析额外提示词") {
			t.Errorf("expected result NOT to contain extra prompt section, got:\n%s", result)
		}
		if !strings.Contains(result, "/repo") {
			t.Errorf("expected result to contain repo path /repo, got:\n%s", result)
		}
		if !strings.Contains(result, "请对项目进行核心概要分析") {
			t.Errorf("expected result to contain base overview instruction, got:\n%s", result)
		}
	})
}

// ──────────────────────────────────────────────────────────────────────
// TestGenerateManifest_WithMetas (task 9)
// ──────────────────────────────────────────────────────────────────────

// writeTestMDX 写入带 frontmatter 的 .mdx 测试文件到 wikiDir 下指定 relPath（无扩展名）
func writeTestMDX(t *testing.T, wikiDir, relPath, title, description, icon string) {
	t.Helper()
	fullPath := filepath.Join(wikiDir, relPath+".mdx")
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("MkdirAll %s failed: %v", dir, err)
	}
	var fm string
	if icon != "" {
		fm = "---\ntitle: " + title + "\ndescription: " + description + "\nicon: " + icon + "\n---\n\nbody"
	} else {
		fm = "---\ntitle: " + title + "\ndescription: " + description + "\n---\n\nbody"
	}
	if err := os.WriteFile(fullPath, []byte(fm), 0644); err != nil {
		t.Fatalf("write %s failed: %v", fullPath, err)
	}
}

// TestGenerateManifest_WithMetas 验证多目录 + meta.json + separator + icon 的完整 manifest 生成
func TestGenerateManifest_WithMetas(t *testing.T) {
	o, tmpDir := newTestOrchestrator(t, 4001, "meta-project", "zh")
	defer os.RemoveAll(tmpDir)

	wikiDir := o.storage.GetWikiPath(o.versionID)

	// 创建 .mdx 文件（带 frontmatter）
	writeTestMDX(t, wikiDir, "overview", "项目概览", "概览描述", "BookOpen")
	writeTestMDX(t, wikiDir, "modules/auth", "认证模块", "认证描述", "ShieldCheck")
	writeTestMDX(t, wikiDir, "modules/api", "API 模块", "API 描述", "Code")

	outline := []WikiEntry{
		{Title: "概览", Path: "overview"},
		{Title: "模块", Path: "modules", Children: []WikiEntry{
			{Title: "认证", Path: "modules/auth"},
			{Title: "API", Path: "modules/api"},
		}},
	}
	metas := []WikiMeta{
		{Path: "", Title: "Meta Project", Icon: "Layers", DefaultOpen: true, Pages: []string{"overview", "modules"}},
		{Path: "modules", Title: "模块目录", Icon: "Boxes", DefaultOpen: true, Pages: []string{"auth", "---分隔---", "api"}},
	}
	archOut := ArchitectOutput{Outline: outline, Metas: metas}

	if xErr := o.generateManifest(archOut); xErr != nil {
		t.Fatalf("generateManifest failed: %v", xErr)
	}

	manifestPath := o.storage.GetManifestPath(o.versionID)
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest failed: %v", err)
	}

	var m manifestData
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal manifest failed: %v", err)
	}

	// 顶层 meta
	if m.Meta.Title != "Meta Project" || m.Meta.Icon != "Layers" {
		t.Errorf("meta = (%q, %q), want (\"Meta Project\", \"Layers\")", m.Meta.Title, m.Meta.Icon)
	}

	// 顶层导航按 pages 顺序：overview, modules
	if len(m.Navigation) != 2 {
		t.Fatalf("expected 2 top-level nav items, got %d", len(m.Navigation))
	}
	if m.Navigation[0].Path != "overview" || m.Navigation[0].Title != "项目概览" {
		t.Errorf("nav[0] = (%q, %q), want title from frontmatter \"项目概览\", path \"overview\"",
			m.Navigation[0].Title, m.Navigation[0].Path)
	}
	if m.Navigation[0].Icon != "BookOpen" {
		t.Errorf("nav[0].Icon = %q, want \"BookOpen\" from frontmatter", m.Navigation[0].Icon)
	}
	if m.Navigation[0].Description != "概览描述" {
		t.Errorf("nav[0].Description = %q, want \"概览描述\" from frontmatter", m.Navigation[0].Description)
	}

	// modules 目录节点
	modNav := m.Navigation[1]
	if modNav.Path != "modules" || modNav.Title != "模块目录" || modNav.Icon != "Boxes" || !modNav.DefaultOpen {
		t.Errorf("modules nav = (title=%q, path=%q, icon=%q, defaultOpen=%v), want (\"模块目录\", \"modules\", \"Boxes\", true)",
			modNav.Title, modNav.Path, modNav.Icon, modNav.DefaultOpen)
	}

	// modules 子导航按 pages 顺序：auth, separator, api
	if len(modNav.Children) != 3 {
		t.Fatalf("expected 3 children (auth, separator, api), got %d", len(modNav.Children))
	}
	if modNav.Children[0].Path != "modules/auth" || modNav.Children[0].Title != "认证模块" || modNav.Children[0].Icon != "ShieldCheck" {
		t.Errorf("children[0] = (title=%q, path=%q, icon=%q), want (\"认证模块\", \"modules/auth\", \"ShieldCheck\")",
			modNav.Children[0].Title, modNav.Children[0].Path, modNav.Children[0].Icon)
	}
	// separator node
	if modNav.Children[1].Separator != "分隔" || modNav.Children[1].Title != "" || modNav.Children[1].Path != "" {
		t.Errorf("children[1] = (separator=%q, title=%q, path=%q), want (\"分隔\", \"\", \"\")",
			modNav.Children[1].Separator, modNav.Children[1].Title, modNav.Children[1].Path)
	}
	if modNav.Children[2].Path != "modules/api" || modNav.Children[2].Title != "API 模块" || modNav.Children[2].Icon != "Code" {
		t.Errorf("children[2] = (title=%q, path=%q, icon=%q), want (\"API 模块\", \"modules/api\", \"Code\")",
			modNav.Children[2].Title, modNav.Children[2].Path, modNav.Children[2].Icon)
	}
}

// TestGenerateManifest_NoMetaFallbackOrder 验证无 meta 时回退为 outline 原序（非字母序）
func TestGenerateManifest_NoMetaFallbackOrder(t *testing.T) {
	o, tmpDir := newTestOrchestrator(t, 4002, "order-project", "en")
	defer os.RemoveAll(tmpDir)

	// outline 顺序故意非字母序
	outline := []WikiEntry{
		{Title: "Zebra", Path: "zebra"},
		{Title: "Apple", Path: "apple"},
		{Title: "Mango", Path: "mango"},
	}
	archOut := ArchitectOutput{Outline: outline}

	if xErr := o.generateManifest(archOut); xErr != nil {
		t.Fatalf("generateManifest failed: %v", xErr)
	}

	manifestPath := o.storage.GetManifestPath(o.versionID)
	data, _ := os.ReadFile(manifestPath)
	var m manifestData
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if len(m.Navigation) != 3 {
		t.Fatalf("expected 3 nav items, got %d", len(m.Navigation))
	}
	expected := []string{"zebra", "apple", "mango"}
	for i, want := range expected {
		if m.Navigation[i].Path != want {
			t.Errorf("nav[%d].Path = %q, want %q (outline original order, NOT alphabetical)", i, m.Navigation[i].Path, want)
		}
	}
}

// TestGenerateManifest_FrontmatterMissingIcon 验证 frontmatter 缺失 icon 时默认 FileText
func TestGenerateManifest_FrontmatterMissingIcon(t *testing.T) {
	o, tmpDir := newTestOrchestrator(t, 4003, "icon-project", "en")
	defer os.RemoveAll(tmpDir)

	wikiDir := o.storage.GetWikiPath(o.versionID)
	// .mdx 有 frontmatter 但无 icon 字段
	writeTestMDX(t, wikiDir, "page", "Page Title", "desc", "")

	outline := []WikiEntry{{Title: "Page", Path: "page"}}
	archOut := ArchitectOutput{Outline: outline}

	if xErr := o.generateManifest(archOut); xErr != nil {
		t.Fatalf("generateManifest failed: %v", xErr)
	}

	manifestPath := o.storage.GetManifestPath(o.versionID)
	data, _ := os.ReadFile(manifestPath)
	var m manifestData
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if len(m.Navigation) != 1 {
		t.Fatalf("expected 1 nav item, got %d", len(m.Navigation))
	}
	if m.Navigation[0].Icon != "FileText" {
		t.Errorf("expected default icon \"FileText\" when frontmatter has no icon, got %q", m.Navigation[0].Icon)
	}
	if m.Navigation[0].Title != "Page Title" {
		t.Errorf("expected title from frontmatter \"Page Title\", got %q", m.Navigation[0].Title)
	}
}

// TestBuildNavFromMeta_PagesMatching 验证 pages 元素匹配逻辑
func TestBuildNavFromMeta_PagesMatching(t *testing.T) {
	o, tmpDir := newTestOrchestrator(t, 4004, "match-project", "en")
	defer os.RemoveAll(tmpDir)

	wikiDir := o.storage.GetWikiPath(o.versionID)
	writeTestMDX(t, wikiDir, "overview", "Overview", "desc", "BookOpen")
	writeTestMDX(t, wikiDir, "modules/auth", "Auth", "desc", "Shield")

	t.Run("folderPath empty + pages matches top-level leaf", func(t *testing.T) {
		outline := []WikiEntry{
			{Title: "Overview", Path: "overview"},
		}
		metas := []WikiMeta{
			{Path: "", Pages: []string{"overview"}},
		}
		nav := o.buildNavFromMeta("", metas, outline)
		if len(nav) != 1 || nav[0].Path != "overview" {
			t.Errorf("expected nav[0].Path=\"overview\", got %v", nav)
		}
	})

	t.Run("folderPath modules + pages auth matches modules/auth", func(t *testing.T) {
		outline := []WikiEntry{
			{Title: "Auth", Path: "modules/auth"},
		}
		metas := []WikiMeta{
			{Path: "modules", Pages: []string{"auth"}},
		}
		nav := o.buildNavFromMeta("modules", metas, outline)
		if len(nav) != 1 || nav[0].Path != "modules/auth" {
			t.Errorf("expected nav[0].Path=\"modules/auth\", got %v", nav)
		}
	})

	t.Run("folderPath empty + pages modules matches dir node and recurses", func(t *testing.T) {
		outline := []WikiEntry{
			{Title: "Modules", Path: "modules", Children: []WikiEntry{
				{Title: "Auth", Path: "modules/auth"},
			}},
		}
		metas := []WikiMeta{
			{Path: "", Pages: []string{"modules"}},
			{Path: "modules", Pages: []string{"auth"}},
		}
		nav := o.buildNavFromMeta("", metas, outline)
		if len(nav) != 1 || nav[0].Path != "modules" || len(nav[0].Children) != 1 {
			t.Errorf("expected dir node with 1 child, got %v", nav)
		}
		if nav[0].Children[0].Path != "modules/auth" {
			t.Errorf("expected child path \"modules/auth\", got %q", nav[0].Children[0].Path)
		}
	})

	t.Run("unlisted subdir appended at end", func(t *testing.T) {
		outline := []WikiEntry{
			{Title: "Overview", Path: "overview"},
			{Title: "Modules", Path: "modules", Children: []WikiEntry{
				{Title: "Auth", Path: "modules/auth"},
			}},
		}
		metas := []WikiMeta{
			{Path: "", Pages: []string{"overview"}},
		}
		nav := o.buildNavFromMeta("", metas, outline)
		if len(nav) != 2 {
			t.Fatalf("expected 2 nav items (overview + appended modules), got %d", len(nav))
		}
		if nav[0].Path != "overview" {
			t.Errorf("nav[0].Path = %q, want \"overview\" (pages order first)", nav[0].Path)
		}
		if nav[1].Path != "modules" {
			t.Errorf("nav[1].Path = %q, want \"modules\" (unlisted appended at end)", nav[1].Path)
		}
	})

	t.Run("direct child filtering: modules/auth is direct, modules/auth/login is NOT", func(t *testing.T) {
		// directChildren 只含直接子节点；modules/auth/login 是 modules/auth 的子节点，不应出现在 modules 的 directChildren 中
		outline := []WikiEntry{
			{Title: "Auth", Path: "modules/auth", Children: []WikiEntry{
				{Title: "Login", Path: "modules/auth/login"},
			}},
		}
		metas := []WikiMeta{
			{Path: "modules", Pages: []string{"auth"}},
		}
		nav := o.buildNavFromMeta("modules", metas, outline)
		if len(nav) != 1 {
			t.Fatalf("expected 1 nav item (modules/auth), got %d", len(nav))
		}
		if nav[0].Path != "modules/auth" {
			t.Errorf("nav[0].Path = %q, want \"modules/auth\"", nav[0].Path)
		}
		// modules/auth/login 不在 directChildren 中，不应出现
		for _, item := range nav {
			if item.Path == "modules/auth/login" {
				t.Errorf("modules/auth/login should NOT be in modules' direct children nav")
			}
		}
	})
}

// TestWriteMetaFiles 验证 writeMetaFiles 落盘 meta.json 到正确路径
func TestWriteMetaFiles(t *testing.T) {
	o, tmpDir := newTestOrchestrator(t, 4005, "metafiles-project", "en")
	defer os.RemoveAll(tmpDir)

	wikiDir := o.storage.GetWikiPath(o.versionID)

	metas := []WikiMeta{
		{Path: "", Title: "Root Meta", Icon: "Layers", DefaultOpen: true, Pages: []string{"overview"}},
		{Path: "modules", Title: "Modules Meta", Icon: "Boxes", Pages: []string{"auth"}},
		{Path: "api/endpoints", Title: "API Meta", Icon: "Code", Pages: []string{"users"}},
	}

	o.writeMetaFiles(metas)

	// 根 meta.json: {wikiDir}/meta.json
	rootMetaPath := filepath.Join(wikiDir, "meta.json")
	if _, err := os.Stat(rootMetaPath); err != nil {
		t.Fatalf("root meta.json should exist at %s: %v", rootMetaPath, err)
	}
	rootData, err := os.ReadFile(rootMetaPath)
	if err != nil {
		t.Fatalf("read root meta.json failed: %v", err)
	}
	var rootMeta WikiMeta
	if err := json.Unmarshal(rootData, &rootMeta); err != nil {
		t.Fatalf("unmarshal root meta.json failed: %v", err)
	}
	if rootMeta.Title != "Root Meta" || rootMeta.Path != "" {
		t.Errorf("root meta = (title=%q, path=%q), want (\"Root Meta\", \"\")", rootMeta.Title, rootMeta.Path)
	}

	// modules meta.json: {wikiDir}/modules/meta.json
	modMetaPath := filepath.Join(wikiDir, "modules", "meta.json")
	if _, err := os.Stat(modMetaPath); err != nil {
		t.Fatalf("modules meta.json should exist at %s: %v", modMetaPath, err)
	}

	// nested dir meta.json: {wikiDir}/api/endpoints/meta.json
	nestedMetaPath := filepath.Join(wikiDir, "api", "endpoints", "meta.json")
	if _, err := os.Stat(nestedMetaPath); err != nil {
		t.Fatalf("nested meta.json should exist at %s: %v", nestedMetaPath, err)
	}
	nestedData, err := os.ReadFile(nestedMetaPath)
	if err != nil {
		t.Fatalf("read nested meta.json failed: %v", err)
	}
	var nestedMeta WikiMeta
	if err := json.Unmarshal(nestedData, &nestedMeta); err != nil {
		t.Fatalf("unmarshal nested meta.json failed: %v", err)
	}
	if nestedMeta.Title != "API Meta" || nestedMeta.Path != "api/endpoints" {
		t.Errorf("nested meta = (title=%q, path=%q), want (\"API Meta\", \"api/endpoints\")", nestedMeta.Title, nestedMeta.Path)
	}
}

// ──────────────────────────────────────────────────────────────────────
// TestEndToEndManifestWithMeta (task 11)
// ──────────────────────────────────────────────────────────────────────

// TestEndToEndManifestWithMeta 端到端冒烟：构造完整 3 层 outline + 多目录 metas + .mdx 文件（带 frontmatter），
// 调用 generateManifest，断言 manifest 结构包含 icon/separator/defaultOpen，且 frontmatter 正确覆盖 entry 默认值。
//
// 覆盖关键不变量：
//   - 根 meta.Title/Icon 透传到 manifest.Meta
//   - home 取 outline 首个叶子（DFS 顺序）
//   - 目录节点从 metas 取 Title/Icon/DefaultOpen
//   - 叶子节点从 .mdx frontmatter 取 Title/Description/Icon（覆盖 entry 默认值）
//   - separator 节点（---文本---）在 pages 序列中正确生成
//   - pages 未列出的 directChildren 追加到末尾保持 outline 原序
//   - 嵌套 3 层结构（section → subsection → leaf）完整保留
func TestEndToEndManifestWithMeta(t *testing.T) {
	o, tmpDir := newTestOrchestrator(t, 5001, "e2e-project", "zh")
	defer os.RemoveAll(tmpDir)

	wikiDir := o.storage.GetWikiPath(o.versionID)

	// ── 1. 写入 .mdx 文件（全部带 frontmatter） ──
	writeTestMDX(t, wikiDir, "index", "项目首页", "首页描述", "Home")
	writeTestMDX(t, wikiDir, "guide/intro", "快速入门", "5 分钟上手", "Rocket")
	writeTestMDX(t, wikiDir, "guide/advanced", "进阶用法", "进阶描述", "GraduationCap")
	writeTestMDX(t, wikiDir, "modules/auth/login", "登录模块", "登录流程", "LogIn")
	writeTestMDX(t, wikiDir, "modules/auth/oauth", "OAuth 模块", "三方授权", "KeyRound")
	writeTestMDX(t, wikiDir, "modules/api/users", "用户 API", "用户接口", "Users")
	writeTestMDX(t, wikiDir, "modules/api/orders", "订单 API", "订单接口", "ShoppingCart")
	writeTestMDX(t, wikiDir, "reference/faq", "常见问题", "FAQ 列表", "HelpCircle")

	// ── 2. 构造完整 outline（3 层嵌套） ──
	outline := []WikiEntry{
		{Title: "首页", Path: "index", Description: "默认描述"},
		{Title: "指南", Path: "guide", Children: []WikiEntry{
			{Title: "入门", Path: "guide/intro"},
			{Title: "进阶", Path: "guide/advanced"},
		}},
		{Title: "模块", Path: "modules", Children: []WikiEntry{
			{Title: "认证", Path: "modules/auth", Children: []WikiEntry{
				{Title: "登录", Path: "modules/auth/login"},
				{Title: "OAuth", Path: "modules/auth/oauth"},
			}},
			{Title: "API", Path: "modules/api", Children: []WikiEntry{
				{Title: "用户", Path: "modules/api/users"},
				{Title: "订单", Path: "modules/api/orders"},
			}},
		}},
		{Title: "参考", Path: "reference", Children: []WikiEntry{
			{Title: "FAQ", Path: "reference/faq"},
		}},
	}

	// ── 3. 构造 metas（根 + 多目录，含 separator + defaultOpen） ──
	metas := []WikiMeta{
		{Path: "", Title: "E2E Wiki", Icon: "Layers", DefaultOpen: true, Pages: []string{"index", "guide", "modules", "reference"}},
		{Path: "guide", Title: "使用指南", Icon: "BookOpen", DefaultOpen: true, Pages: []string{"intro", "advanced"}},
		{Path: "modules", Title: "核心模块", Icon: "Boxes", DefaultOpen: true, Pages: []string{"auth", "---API 分隔---", "api"}},
		{Path: "modules/auth", Title: "认证模块", Icon: "ShieldCheck", DefaultOpen: false, Pages: []string{"login", "oauth"}},
		{Path: "modules/api", Title: "API 模块", Icon: "Code", DefaultOpen: false, Pages: []string{"users", "orders"}},
		{Path: "reference", Title: "参考资料", Icon: "Library", DefaultOpen: false, Pages: []string{"faq"}},
	}
	archOut := ArchitectOutput{Outline: outline, Metas: metas}

	// ── 4. 调用 generateManifest ──
	if xErr := o.generateManifest(archOut); xErr != nil {
		t.Fatalf("generateManifest failed: %v", xErr)
	}

	// ── 5. 读取并断言 manifest ──
	manifestPath := o.storage.GetManifestPath(o.versionID)
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest failed: %v", err)
	}
	var m manifestData
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal manifest failed: %v", err)
	}

	// 5.1 顶层 Meta（来自根 meta）
	if m.Meta.Title != "E2E Wiki" {
		t.Errorf("meta.title = %q, want \"E2E Wiki\" (from root meta)", m.Meta.Title)
	}
	if m.Meta.Icon != "Layers" {
		t.Errorf("meta.icon = %q, want \"Layers\"", m.Meta.Icon)
	}
	if m.ProjectName != "e2e-project" {
		t.Errorf("project_name = %q, want \"e2e-project\"", m.ProjectName)
	}
	if m.Language != "zh" {
		t.Errorf("language = %q, want \"zh\"", m.Language)
	}

	// 5.2 home = outline 首个叶子（index）
	if m.Home != "index" {
		t.Errorf("home = %q, want \"index\" (first leaf)", m.Home)
	}

	// 5.3 顶层导航 4 项，按 root meta.pages 顺序
	if len(m.Navigation) != 4 {
		t.Fatalf("expected 4 top-level nav items, got %d", len(m.Navigation))
	}
	expectedTopPaths := []string{"index", "guide", "modules", "reference"}
	for i, want := range expectedTopPaths {
		if m.Navigation[i].Path != want {
			t.Errorf("nav[%d].Path = %q, want %q", i, m.Navigation[i].Path, want)
		}
	}

	// 5.4 叶子节点 frontmatter 覆盖：index 的 Title/Description/Icon 来自 .mdx frontmatter
	navIndex := m.Navigation[0]
	if navIndex.Title != "项目首页" {
		t.Errorf("nav[0].Title = %q, want \"项目首页\" (from frontmatter)", navIndex.Title)
	}
	if navIndex.Description != "首页描述" {
		t.Errorf("nav[0].Description = %q, want \"首页描述\" (from frontmatter)", navIndex.Description)
	}
	if navIndex.Icon != "Home" {
		t.Errorf("nav[0].Icon = %q, want \"Home\" (from frontmatter)", navIndex.Icon)
	}

	// 5.5 目录节点从 metas 取属性：guide → Title="使用指南", Icon="BookOpen", DefaultOpen=true
	navGuide := m.Navigation[1]
	if navGuide.Title != "使用指南" {
		t.Errorf("nav[1].Title = %q, want \"使用指南\" (from meta)", navGuide.Title)
	}
	if navGuide.Icon != "BookOpen" {
		t.Errorf("nav[1].Icon = %q, want \"BookOpen\" (from meta)", navGuide.Icon)
	}
	if !navGuide.DefaultOpen {
		t.Errorf("nav[1].DefaultOpen = false, want true (from meta)")
	}
	if len(navGuide.Children) != 2 {
		t.Fatalf("nav[1] expected 2 children, got %d", len(navGuide.Children))
	}
	// guide 子节点 frontmatter 覆盖
	if navGuide.Children[0].Path != "guide/intro" || navGuide.Children[0].Title != "快速入门" || navGuide.Children[0].Icon != "Rocket" {
		t.Errorf("guide.children[0] = (path=%q, title=%q, icon=%q), want (\"guide/intro\", \"快速入门\", \"Rocket\")",
			navGuide.Children[0].Path, navGuide.Children[0].Title, navGuide.Children[0].Icon)
	}
	if navGuide.Children[1].Path != "guide/advanced" || navGuide.Children[1].Icon != "GraduationCap" {
		t.Errorf("guide.children[1] = (path=%q, icon=%q), want (\"guide/advanced\", \"GraduationCap\")",
			navGuide.Children[1].Path, navGuide.Children[1].Icon)
	}

	// 5.6 modules 目录：含 separator 子节点（---API 分隔---）
	navModules := m.Navigation[2]
	if navModules.Title != "核心模块" || navModules.Icon != "Boxes" || !navModules.DefaultOpen {
		t.Errorf("modules nav = (title=%q, icon=%q, defaultOpen=%v), want (\"核心模块\", \"Boxes\", true)",
			navModules.Title, navModules.Icon, navModules.DefaultOpen)
	}
	// modules.children 按 pages 顺序：auth, separator, api
	if len(navModules.Children) != 3 {
		t.Fatalf("modules expected 3 children (auth, separator, api), got %d", len(navModules.Children))
	}
	// 5.6.1 separator 节点
	sepNode := navModules.Children[1]
	if sepNode.Separator != "API 分隔" || sepNode.Title != "" || sepNode.Path != "" {
		t.Errorf("modules.children[1] = (separator=%q, title=%q, path=%q), want (\"API 分隔\", \"\", \"\")",
			sepNode.Separator, sepNode.Title, sepNode.Path)
	}
	// 5.6.2 auth 子目录（DefaultOpen=false，3 层嵌套）
	navAuth := navModules.Children[0]
	if navAuth.Path != "modules/auth" || navAuth.Title != "认证模块" || navAuth.Icon != "ShieldCheck" {
		t.Errorf("auth nav = (path=%q, title=%q, icon=%q), want (\"modules/auth\", \"认证模块\", \"ShieldCheck\")",
			navAuth.Path, navAuth.Title, navAuth.Icon)
	}
	if navAuth.DefaultOpen {
		t.Errorf("auth DefaultOpen = true, want false (from meta)")
	}
	if len(navAuth.Children) != 2 {
		t.Fatalf("auth expected 2 children, got %d", len(navAuth.Children))
	}
	// 3 层嵌套叶子：modules/auth/login
	leafLogin := navAuth.Children[0]
	if leafLogin.Path != "modules/auth/login" || leafLogin.Title != "登录模块" || leafLogin.Icon != "LogIn" || leafLogin.Description != "登录流程" {
		t.Errorf("auth.children[0] = (path=%q, title=%q, icon=%q, desc=%q), want (\"modules/auth/login\", \"登录模块\", \"LogIn\", \"登录流程\")",
			leafLogin.Path, leafLogin.Title, leafLogin.Icon, leafLogin.Description)
	}
	leafOAuth := navAuth.Children[1]
	if leafOAuth.Path != "modules/auth/oauth" || leafOAuth.Icon != "KeyRound" {
		t.Errorf("auth.children[1] = (path=%q, icon=%q), want (\"modules/auth/oauth\", \"KeyRound\")",
			leafOAuth.Path, leafOAuth.Icon)
	}
	// 5.6.3 api 子目录
	navAPI := navModules.Children[2]
	if navAPI.Path != "modules/api" || navAPI.Title != "API 模块" || navAPI.Icon != "Code" {
		t.Errorf("api nav = (path=%q, title=%q, icon=%q), want (\"modules/api\", \"API 模块\", \"Code\")",
			navAPI.Path, navAPI.Title, navAPI.Icon)
	}
	if len(navAPI.Children) != 2 {
		t.Fatalf("api expected 2 children, got %d", len(navAPI.Children))
	}
	if navAPI.Children[0].Path != "modules/api/users" || navAPI.Children[0].Icon != "Users" {
		t.Errorf("api.children[0] = (path=%q, icon=%q), want (\"modules/api/users\", \"Users\")",
			navAPI.Children[0].Path, navAPI.Children[0].Icon)
	}
	if navAPI.Children[1].Path != "modules/api/orders" || navAPI.Children[1].Icon != "ShoppingCart" {
		t.Errorf("api.children[1] = (path=%q, icon=%q), want (\"modules/api/orders\", \"ShoppingCart\")",
			navAPI.Children[1].Path, navAPI.Children[1].Icon)
	}

	// 5.7 reference 目录（DefaultOpen=false）
	navRef := m.Navigation[3]
	if navRef.Title != "参考资料" || navRef.Icon != "Library" || navRef.DefaultOpen {
		t.Errorf("reference nav = (title=%q, icon=%q, defaultOpen=%v), want (\"参考资料\", \"Library\", false)",
			navRef.Title, navRef.Icon, navRef.DefaultOpen)
	}
	if len(navRef.Children) != 1 || navRef.Children[0].Path != "reference/faq" || navRef.Children[0].Icon != "HelpCircle" {
		t.Errorf("reference.children[0] = (path=%q, icon=%q), want (\"reference/faq\", \"HelpCircle\")",
			navRef.Children[0].Path, navRef.Children[0].Icon)
	}
}
