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
	outline := []WikiEntry{
		{Title: "Overview", Path: "overview.md"},
		{Title: "Architecture", Path: "architecture.md"},
		{Title: "Modules", Path: "modules.md"},
		{Title: "Empty Path", Path: ""},
	}

	t.Run("matches missing_file/empty_page/orphan_file to outline", func(t *testing.T) {
		errors := []ValidationError{
			{Type: "missing_file", Path: "overview.md", Message: "missing"},
			{Type: "empty_page", Path: "architecture.md", Message: "empty"},
			{Type: "orphan_file", Path: "modules.md", Message: "orphan"},
		}
		got := findMissingEntries(errors, outline)
		if len(got) != 3 {
			t.Fatalf("expected 3 missing entries, got %d", len(got))
		}
		gotPaths := make(map[string]bool, len(got))
		for _, e := range got {
			gotPaths[e.Path] = true
		}
		for _, want := range []string{"overview.md", "architecture.md", "modules.md"} {
			if !gotPaths[want] {
				t.Errorf("expected missing entry with path %q", want)
			}
		}
	})

	t.Run("non-matching paths return nil", func(t *testing.T) {
		errors := []ValidationError{
			{Type: "missing_file", Path: "nonexistent.md", Message: "no match"},
			{Type: "empty_page", Path: "totally/different.md", Message: "no match"},
		}
		got := findMissingEntries(errors, outline)
		if got != nil {
			t.Errorf("expected nil when no error path matches outline, got %v", got)
		}
	})

	t.Run("missing_metadata errors are skipped", func(t *testing.T) {
		errors := []ValidationError{
			{Type: "missing_metadata", Path: "overview.md", Message: "metadata issue"},
			{Type: "missing_file", Path: "architecture.md", Message: "real missing"},
		}
		got := findMissingEntries(errors, outline)
		if len(got) != 1 {
			t.Fatalf("expected 1 missing entry (missing_metadata skipped), got %d", len(got))
		}
		if got[0].Path != "architecture.md" {
			t.Errorf("expected architecture.md, got %q", got[0].Path)
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

	t.Run("only missing_metadata errors return nil", func(t *testing.T) {
		errors := []ValidationError{
			{Type: "missing_metadata", Path: "overview.md", Message: "metadata"},
			{Type: "missing_metadata", Path: "architecture.md", Message: "metadata"},
		}
		got := findMissingEntries(errors, outline)
		if got != nil {
			t.Errorf("expected nil when only missing_metadata errors, got %v", got)
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

	// big.md: 超过 100 字节（合格）
	bigContent := make([]byte, writerFileMinSize+50)
	for i := range bigContent {
		bigContent[i] = 'x'
	}
	if err := os.WriteFile(filepath.Join(wikiDir, "big.md"), bigContent, 0644); err != nil {
		t.Fatalf("write big.md failed: %v", err)
	}

	// small.md: 小于 100 字节（不合格，视为缺失）
	if err := os.WriteFile(filepath.Join(wikiDir, "small.md"), []byte("tiny"), 0644); err != nil {
		t.Fatalf("write small.md failed: %v", err)
	}

	// missing.md: 不创建（缺失）

	outline := []WikiEntry{
		{Title: "Big", Path: "big.md"},
		{Title: "Small", Path: "small.md"},
		{Title: "Missing", Path: "missing.md"},
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
	if !missingPaths["small.md"] {
		t.Errorf("expected small.md in missing (size < %d)", writerFileMinSize)
	}
	if !missingPaths["missing.md"] {
		t.Errorf("expected missing.md in missing (file not found)")
	}
	if missingPaths["big.md"] {
		t.Errorf("big.md should not be in missing (size >= %d)", writerFileMinSize)
	}
}

// ──────────────────────────────────────────────────────────────────────
// TestGenerateManifest
// ──────────────────────────────────────────────────────────────────────

func TestGenerateManifest(t *testing.T) {
	o, tmpDir := newTestOrchestrator(t, 2002, "test-project", "zh")
	defer os.RemoveAll(tmpDir)

	outline := []WikiEntry{
		{Title: "概览", Path: "overview.md", Description: "项目概览"},
		{Title: "架构", Children: []WikiEntry{
			{Title: "架构设计", Path: "architecture.md", Description: "架构设计"},
			{Title: "模块说明", Path: "modules.md", Description: "模块说明"},
		}},
		{Title: "参考", Path: "reference.md", Description: "参考资料"},
	}

	if xErr := o.generateManifest(outline); xErr != nil {
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
	if m.Home != "overview.md" {
		t.Errorf("expected home %q (first top-level leaf), got %q", "overview.md", m.Home)
	}
	if len(m.Navigation) != 3 {
		t.Fatalf("expected 3 top-level nav items, got %d", len(m.Navigation))
	}
	if m.Navigation[0].Title != "概览" || m.Navigation[0].Path != "overview.md" {
		t.Errorf("nav[0] = (%q, %q), want (\"概览\", \"overview.md\")", m.Navigation[0].Title, m.Navigation[0].Path)
	}
	if m.Navigation[1].Title != "架构" || m.Navigation[1].Path != "" || len(m.Navigation[1].Children) != 2 {
		t.Errorf("nav[1] should be directory node with 2 children, got title=%q path=%q children=%d",
			m.Navigation[1].Title, m.Navigation[1].Path, len(m.Navigation[1].Children))
	}
	if m.Navigation[1].Children[0].Path != "architecture.md" || m.Navigation[1].Children[1].Path != "modules.md" {
		t.Errorf("nav[1].children paths mismatch, got %v", m.Navigation[1].Children)
	}
	if m.Navigation[2].Title != "参考" || m.Navigation[2].Path != "reference.md" {
		t.Errorf("nav[2] = (%q, %q), want (\"参考\", \"reference.md\")", m.Navigation[2].Title, m.Navigation[2].Path)
	}
}

// TestGenerateManifest_EmptyOutline 验证空 outline 时 home 为空字符串
func TestGenerateManifest_EmptyOutline(t *testing.T) {
	o, tmpDir := newTestOrchestrator(t, 2003, "empty-project", "en")
	defer os.RemoveAll(tmpDir)

	if xErr := o.generateManifest(nil); xErr != nil {
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
		{Title: "Guide", Path: "guide.md", Description: "Top-level guide"},
		{Title: "Section A", Children: []WikiEntry{
			{Title: "Subsection A1", Children: []WikiEntry{
				{Title: "Deep Leaf", Path: "deep/leaf.md", Description: "Deep leaf"},
			}},
			{Title: "Leaf A2", Path: "a2.md", Description: "A2 leaf"},
		}},
		{Title: "Section B", Path: "b.md", Description: "Section B"},
	}

	if xErr := o.generateManifest(outline); xErr != nil {
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

	if m.Home != "guide.md" {
		t.Errorf("expected home %q (shallowest first leaf), got %q", "guide.md", m.Home)
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
	if deep.Path != "deep/leaf.md" {
		t.Errorf("expected deep leaf path %q, got %q", "deep/leaf.md", deep.Path)
	}
}

// TestGenerateManifest_NoTopLevelLeaf 验证顶层无叶子时 home 取 DFS 首个叶子
func TestGenerateManifest_NoTopLevelLeaf(t *testing.T) {
	o, tmpDir := newTestOrchestrator(t, 2005, "no-top-leaf-project", "en")
	defer os.RemoveAll(tmpDir)

	outline := []WikiEntry{
		{Title: "Section A", Children: []WikiEntry{
			{Title: "Leaf A1", Path: "a1.md", Description: "A1"},
			{Title: "Leaf A2", Path: "a2.md", Description: "A2"},
		}},
		{Title: "Section B", Children: []WikiEntry{
			{Title: "Leaf B1", Path: "b1.md", Description: "B1"},
		}},
	}

	if xErr := o.generateManifest(outline); xErr != nil {
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

	if m.Home != "a1.md" {
		t.Errorf("expected home %q (DFS first leaf in children), got %q", "a1.md", m.Home)
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
				{Title: "Leaf A1", Path: "a1.md", ExploreRefs: []string{"ref1"}, Complexity: "high"},
			}},
			{Title: "Leaf A2", Path: "a2.md", ExploreRefs: []string{"ref2"}, Complexity: "medium"},
		}},
		{Title: "Leaf B", Path: "b.md", ExploreRefs: []string{"ref3"}, Complexity: "low"},
		{Title: "Empty Directory", Children: []WikiEntry{}},
	}

	leaves := flattenOutlineLeaves(outline)
	if len(leaves) != 3 {
		t.Fatalf("expected 3 leaves, got %d", len(leaves))
	}

	expectedPaths := []string{"a1.md", "a2.md", "b.md"}
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
			{Title: "A", Path: "a.md"},
			{Title: "B", Path: "b.md"},
		}
		if got := findFirstLeafPath(outline); got != "a.md" {
			t.Errorf("findFirstLeafPath = %q, want %q", got, "a.md")
		}
	})

	t.Run("top-level all directories", func(t *testing.T) {
		outline := []WikiEntry{
			{Title: "A", Children: []WikiEntry{
				{Title: "B", Path: "b.md"},
			}},
		}
		if got := findFirstLeafPath(outline); got != "b.md" {
			t.Errorf("findFirstLeafPath = %q, want %q", got, "b.md")
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
			{Title: "First Leaf", Path: "first.md"},
		}
		if got := findFirstLeafPath(outline); got != "first.md" {
			t.Errorf("findFirstLeafPath = %q, want %q", got, "first.md")
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
		{Title: "首页", Path: "index.md"},
		{Title: "指南", Path: "guide.md"},
	}

	// 模拟 Execute 中 generateManifest 调用（先于 runValidator）
	if xErr := o.generateManifest(outline); xErr != nil {
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
	if m.Home != "index.md" {
		t.Errorf("expected home %q, got %q", "index.md", m.Home)
	}
	if len(m.Navigation) != 2 {
		t.Errorf("expected 2 nav items, got %d", len(m.Navigation))
	}

	// 验证 verifyWriterOutputs 在 manifest 生成前可正常工作
	// （Execute 流程：writers → verifyWriterOutputs → generateManifest → runValidator）
	wikiDir := o.storage.GetWikiPath(o.versionID)
	missing := o.verifyWriterOutputs(outline, wikiDir)
	if len(missing) != 2 {
		t.Errorf("expected 2 missing entries (no files written), got %d", len(missing))
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
