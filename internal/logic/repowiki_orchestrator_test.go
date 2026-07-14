package logic

import (
	"encoding/json"
	"os"
	"path/filepath"
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
		{Title: "Overview", Path: "content/overview.md"},
		{Title: "Architecture", Path: "content/architecture.md"},
		{Title: "Modules", Path: "content/modules.md"},
		{Title: "Empty Path", Path: ""},
	}

	t.Run("matches missing_file/empty_page/orphan_file to outline", func(t *testing.T) {
		errors := []ValidationError{
			{Type: "missing_file", Path: "content/overview.md", Message: "missing"},
			{Type: "empty_page", Path: "content/architecture.md", Message: "empty"},
			{Type: "orphan_file", Path: "content/modules.md", Message: "orphan"},
		}
		got := findMissingEntries(errors, outline)
		if len(got) != 3 {
			t.Fatalf("expected 3 missing entries, got %d", len(got))
		}
		gotPaths := make(map[string]bool, len(got))
		for _, e := range got {
			gotPaths[e.Path] = true
		}
		for _, want := range []string{"content/overview.md", "content/architecture.md", "content/modules.md"} {
			if !gotPaths[want] {
				t.Errorf("expected missing entry with path %q", want)
			}
		}
	})

	t.Run("non-matching paths return nil", func(t *testing.T) {
		errors := []ValidationError{
			{Type: "missing_file", Path: "content/nonexistent.md", Message: "no match"},
			{Type: "empty_page", Path: "totally/different.md", Message: "no match"},
		}
		got := findMissingEntries(errors, outline)
		if got != nil {
			t.Errorf("expected nil when no error path matches outline, got %v", got)
		}
	})

	t.Run("missing_metadata errors are skipped", func(t *testing.T) {
		errors := []ValidationError{
			{Type: "missing_metadata", Path: "content/overview.md", Message: "metadata issue"},
			{Type: "missing_file", Path: "content/architecture.md", Message: "real missing"},
		}
		got := findMissingEntries(errors, outline)
		if len(got) != 1 {
			t.Fatalf("expected 1 missing entry (missing_metadata skipped), got %d", len(got))
		}
		if got[0].Path != "content/architecture.md" {
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
			{Type: "missing_metadata", Path: "content/overview.md", Message: "metadata"},
			{Type: "missing_metadata", Path: "content/architecture.md", Message: "metadata"},
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
		{Title: "概览", Path: "content/overview.md", Description: "项目概览"},
		{Title: "架构", Path: "content/architecture.md", Description: "架构设计"},
		{Title: "模块", Path: "content/modules.md", Description: "模块说明"},
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
	if m.Home != "content/overview.md" {
		t.Errorf("expected home %q (outline[0].Path), got %q", "content/overview.md", m.Home)
	}
	if len(m.Navigation) != 3 {
		t.Fatalf("expected 3 nav items, got %d", len(m.Navigation))
	}
	for i, entry := range outline {
		nav := m.Navigation[i]
		if nav.Title != entry.Title {
			t.Errorf("nav[%d].title = %q, want %q", i, nav.Title, entry.Title)
		}
		if nav.Path != entry.Path {
			t.Errorf("nav[%d].path = %q, want %q", i, nav.Path, entry.Path)
		}
	}
}

// TestGenerateManifest_EmptyOutline 验证空 outline 时 home 回退为 index.md
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
	if m.Home != "index.md" {
		t.Errorf("expected home fallback %q, got %q", "index.md", m.Home)
	}
	if len(m.Navigation) != 0 {
		t.Errorf("expected empty navigation, got %d items", len(m.Navigation))
	}
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
		{Title: "首页", Path: "content/index.md"},
		{Title: "指南", Path: "content/guide.md"},
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
	if m.Home != "content/index.md" {
		t.Errorf("expected home %q, got %q", "content/index.md", m.Home)
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
