package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// ── 辅助：在扫描结果中按相对路径查找文件 ──

// findFileByPath 在扫描结果中按路径查找文件信息，返回 (FileInfo, 是否存在)。
func findFileByPath(result *FileScanResult, target string) (FileInfo, bool) {
	for _, f := range result.Files {
		if f.Path == target {
			return f, true
		}
	}
	return FileInfo{}, false
}

// containsPath 判断扫描结果中是否包含指定路径。
func containsPath(result *FileScanResult, target string) bool {
	_, ok := findFileByPath(result, target)
	return ok
}

// containsEntryPoint 判断扫描结果入口列表中是否包含指定路径。
func containsEntryPoint(result *FileScanResult, target string) bool {
	for _, ep := range result.EntryPoints {
		if ep == target {
			return true
		}
	}
	return false
}

// ── TestFileScanner 文件扫描综合测试 ──

// TestFileScanner 验证扫描服务的目录过滤、语言检测和入口识别。
//
// 测试覆盖：
//  1. 排除目录过滤：.git/、node_modules/、vendor/、dist/ 下的文件不出现在结果中
//  2. 语言检测：main.go → Go，index.ts → TypeScript
//  3. 入口文件识别：main.go、go.mod、src/index.ts 标记为入口
//  4. 统计字段一致性：TotalFiles == len(Files)，TotalSize == 各文件大小之和
func TestFileScanner(t *testing.T) {
	repoPath := createSampleRepo(t)

	scanner := NewFileScannerService()
	result, xErr := scanner.Scan(context.Background(), repoPath)
	if xErr != nil {
		t.Fatalf("扫描失败: %v", xErr)
	}

	// ── 1. 排除目录过滤验证 ──
	excludedPaths := []string{
		"node_modules/some-pkg/index.js",
		"vendor/lib/lib.go",
		"dist/bundle.js",
	}
	for _, p := range excludedPaths {
		if containsPath(result, p) {
			t.Errorf("排除目录下的文件不应出现在扫描结果中: %s", p)
		}
	}

	// 验证应被包含的文件确实存在
	includedPaths := []string{
		"main.go", "go.mod", "package.json", "README.md",
		"internal/handler/user.go", "internal/logic/user.go",
		"src/index.ts", "src/utils.ts",
	}
	for _, p := range includedPaths {
		if !containsPath(result, p) {
			t.Errorf("应包含的文件缺失: %s", p)
		}
	}

	// ── 2. 语言检测验证 ──
	cases := []struct {
		path     string
		language string
	}{
		{"main.go", "Go"},
		{"internal/handler/user.go", "Go"},
		{"internal/logic/user.go", "Go"},
		{"src/index.ts", "TypeScript"},
		{"src/utils.ts", "TypeScript"},
		{"package.json", "JSON"},
		{"README.md", "Markdown"},
	}
	for _, c := range cases {
		fi, ok := findFileByPath(result, c.path)
		if !ok {
			t.Errorf("文件不存在: %s", c.path)
			continue
		}
		if fi.Language != c.language {
			t.Errorf("语言检测错误 [%s]: 期望 %q, 实际 %q", c.path, c.language, fi.Language)
		}
	}

	// ── 3. 入口文件识别验证 ──
	expectedEntries := []string{
		"main.go",
		"go.mod",
		"src/index.ts",
		"package.json",
	}
	for _, ep := range expectedEntries {
		if !containsEntryPoint(result, ep) {
			t.Errorf("入口文件未被识别: %s", ep)
		}
		// 入口文件 FileInfo.IsEntry 也应为 true
		fi, ok := findFileByPath(result, ep)
		if !ok {
			t.Errorf("入口文件不存在于文件列表: %s", ep)
			continue
		}
		if !fi.IsEntry {
			t.Errorf("入口文件 IsEntry 标记错误: %s", ep)
		}
	}

	// 非入口文件 IsEntry 应为 false
	nonEntries := []string{
		"internal/handler/user.go", "src/utils.ts", "README.md",
	}
	for _, p := range nonEntries {
		fi, ok := findFileByPath(result, p)
		if !ok {
			continue
		}
		if fi.IsEntry {
			t.Errorf("非入口文件被误标记为入口: %s", p)
		}
	}

	// ── 4. 统计字段一致性验证 ──
	if result.TotalFiles != len(result.Files) {
		t.Errorf("TotalFiles 不一致: 期望 %d, 实际 %d", len(result.Files), result.TotalFiles)
	}
	var expectedSize int64
	for _, f := range result.Files {
		expectedSize += f.Size
	}
	if result.TotalSize != expectedSize {
		t.Errorf("TotalSize 不一致: 期望 %d, 实际 %d", expectedSize, result.TotalSize)
	}
	// Go 文件至少 3 个（main + 2 个 internal）
	if result.LanguageStats["Go"] < 3 {
		t.Errorf("Go 文件数不足: 期望 >= 3, 实际 %d", result.LanguageStats["Go"])
	}
	// TypeScript 文件至少 2 个
	if result.LanguageStats["TypeScript"] < 2 {
		t.Errorf("TypeScript 文件数不足: 期望 >= 2, 实际 %d", result.LanguageStats["TypeScript"])
	}
}

// ── TestFileScanner_ExcludeDirs 排除目录专项测试 ──

// TestFileScanner_ExcludeDirs 验证各种排除目录场景。
func TestFileScanner_ExcludeDirs(t *testing.T) {
	repoPath := createSampleRepo(t)

	// 追加额外的排除目录测试（大小写不敏感）
	extraDirs := []string{
		"BUILD/out.txt",      // build（大写）
		"Target/release.txt", // target（混合大小写）
		"obj/debug.txt",      // obj
	}
	for _, p := range extraDirs {
		fullPath := filepath.Join(repoPath, p)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("创建目录失败: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte("test"), 0644); err != nil {
			t.Fatalf("写入文件失败: %v", err)
		}
	}

	scanner := NewFileScannerService()
	result, xErr := scanner.Scan(context.Background(), repoPath)
	if xErr != nil {
		t.Fatalf("扫描失败: %v", xErr)
	}

	for _, p := range extraDirs {
		relPath := filepath.ToSlash(p)
		if containsPath(result, relPath) {
			t.Errorf("排除目录（大小写不敏感）下的文件不应出现: %s", relPath)
		}
	}
}

// ── TestFileScanner_BinaryFilter 二进制文件过滤测试 ──

// TestFileScanner_BinaryFilter 验证二进制扩展名文件被过滤。
func TestFileScanner_BinaryFilter(t *testing.T) {
	repoPath := createSampleRepo(t)

	// 创建二进制文件
	binaryFiles := []string{"logo.png", "archive.zip", "binary.exe", "font.woff2"}
	for _, f := range binaryFiles {
		fullPath := filepath.Join(repoPath, f)
		if err := os.WriteFile(fullPath, []byte("fake binary"), 0644); err != nil {
			t.Fatalf("写入文件失败: %v", err)
		}
	}

	scanner := NewFileScannerService()
	result, xErr := scanner.Scan(context.Background(), repoPath)
	if xErr != nil {
		t.Fatalf("扫描失败: %v", xErr)
	}

	for _, f := range binaryFiles {
		if containsPath(result, f) {
			t.Errorf("二进制文件不应出现在扫描结果中: %s", f)
		}
	}
}

// ── TestFileScanner_OversizedFile 超大文件过滤测试 ──

// TestFileScanner_OversizedFile 验证超过 maxFileSize 的文件被跳过。
func TestFileScanner_OversizedFile(t *testing.T) {
	repoPath := createSampleRepo(t)

	// 创建超大文件（2MB > 默认 1MB）
	bigContent := make([]byte, 2*1024*1024)
	bigPath := filepath.Join(repoPath, "huge.go")
	if err := os.WriteFile(bigPath, bigContent, 0644); err != nil {
		t.Fatalf("写入大文件失败: %v", err)
	}

	scanner := NewFileScannerService()
	result, xErr := scanner.Scan(context.Background(), repoPath)
	if xErr != nil {
		t.Fatalf("扫描失败: %v", xErr)
	}

	if containsPath(result, "huge.go") {
		t.Errorf("超大文件（2MB）不应出现在扫描结果中")
	}
}

// ── TestFileScanner_InvalidPath 无效路径测试 ──

// TestFileScanner_InvalidPath 验证对不存在路径和非目录路径的错误处理。
func TestFileScanner_InvalidPath(t *testing.T) {
	scanner := NewFileScannerService()

	// 不存在的路径
	_, xErr := scanner.Scan(context.Background(), "/nonexistent/path/that/should/not/exist")
	if xErr == nil {
		t.Error("不存在的路径应返回错误")
	}

	// 非目录路径（使用一个文件）
	filePath := filepath.Join(t.TempDir(), "afile.txt")
	if err := os.WriteFile(filePath, []byte("hi"), 0644); err != nil {
		t.Fatalf("创建文件失败: %v", err)
	}
	_, xErr = scanner.Scan(context.Background(), filePath)
	if xErr == nil {
		t.Error("非目录路径应返回错误")
	}
}

// ── TestEntryPointDetection 入口文件检测专项测试 ──

// TestEntryPointDetection 验证 isEntryPoint 函数对各语言入口/清单文件的识别。
func TestEntryPointDetection(t *testing.T) {
	cases := []struct {
		name    string
		isEntry bool
	}{
		// Go
		{"main.go", true},
		{"go.mod", true},
		// Python
		{"main.py", true},
		{"app.py", true},
		{"requirements.txt", true},
		// Rust
		{"main.rs", true},
		{"Cargo.toml", true},
		// JavaScript/TypeScript
		{"index.ts", true},
		{"index.js", true},
		{"index.tsx", true},
		{"index.jsx", true},
		{"app.js", true},
		{"package.json", true},
		// Java
		{"main.java", true},
		{"Main.java", true},
		{"pom.xml", true},
		{"build.gradle", true},
		// Ruby
		{"Gemfile", true},
		// 构建工具
		{"Makefile", true},
		{"CMakeLists.txt", true},
		// 非入口文件
		{"handler.go", false},
		{"utils.ts", false},
		{"helper.py", false},
		{"README.md", false},
		{"config.yaml", false},
		{"test.js", false},
		{"Main.go", false}, // 大小写敏感：Main.go 不是 main.go
		{"main_go", false},
		{"", false},
	}

	for _, c := range cases {
		got := isEntryPoint(c.name)
		if got != c.isEntry {
			t.Errorf("isEntryPoint(%q) = %v, 期望 %v", c.name, got, c.isEntry)
		}
	}
}

// ── TestEntryPointDetection_ScanIntegration 入口检测与扫描集成测试 ──

// TestEntryPointDetection_ScanIntegration 验证扫描结果中入口文件的标记与独立检测一致。
func TestEntryPointDetection_ScanIntegration(t *testing.T) {
	repoPath := createSampleRepo(t)

	scanner := NewFileScannerService()
	result, xErr := scanner.Scan(context.Background(), repoPath)
	if xErr != nil {
		t.Fatalf("扫描失败: %v", xErr)
	}

	// 遍历所有文件，验证 IsEntry 标记与 isEntryPoint(文件名) 一致
	for _, f := range result.Files {
		fileName := filepath.Base(f.Path)
		expected := isEntryPoint(fileName)
		if f.IsEntry != expected {
			t.Errorf("入口标记不一致 [%s]: FileInfo.IsEntry=%v, isEntryPoint(%q)=%v",
				f.Path, f.IsEntry, fileName, expected)
		}
		// IsEntry 为 true 的文件也应出现在 EntryPoints 列表中
		if f.IsEntry && !containsEntryPoint(result, f.Path) {
			t.Errorf("IsEntry=true 但未出现在 EntryPoints 列表: %s", f.Path)
		}
	}
}

// ── TestDetectLanguage 语言检测单元测试 ──

// TestDetectLanguage 验证 detectLanguage 对各种扩展名的识别。
func TestDetectLanguage(t *testing.T) {
	cases := []struct {
		ext      string
		language string
	}{
		{".go", "Go"},
		{".py", "Python"},
		{".ts", "TypeScript"},
		{".tsx", "TypeScript"},
		{".js", "JavaScript"},
		{".jsx", "JavaScript"},
		{".java", "Java"},
		{".rs", "Rust"},
		{".c", "C"},
		{".h", "C"},
		{".cpp", "C++"},
		{".rb", "Ruby"},
		{".php", "PHP"},
		{".swift", "Swift"},
		{".kt", "Kotlin"},
		{".scala", "Scala"},
		{".sh", "Shell"},
		{".md", "Markdown"},
		{".sql", "SQL"},
		{".html", "HTML"},
		{".css", "CSS"},
		{".json", "JSON"},
		{".yaml", "YAML"},
		{".yml", "YAML"},
		{".toml", "TOML"},
		{".xml", "XML"},
		// 大小写不敏感
		{".GO", "Go"},
		{".TS", "TypeScript"},
		// 未识别
		{".mod", ""},
		{".unknown", ""},
		{"", ""},
	}

	for _, c := range cases {
		got := detectLanguage(c.ext)
		if got != c.language {
			t.Errorf("detectLanguage(%q) = %q, 期望 %q", c.ext, got, c.language)
		}
	}
}
