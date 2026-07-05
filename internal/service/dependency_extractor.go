package service

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// ── FileScanResult（FileScannerService 扫描结果类型定义在 file_scanner.go）──
// ── 依赖摘要数据结构 ──

// DependencySummary 模块级依赖摘要
type DependencySummary struct {
	Modules      []ModuleInfo `json:"modules"`       // Modules 模块列表
	CoreModules  []string     `json:"core_modules"`  // CoreModules 核心模块（被依赖最多 top-5）
	TotalImports int          `json:"total_imports"` // TotalImports 总 import 数
}

// ModuleInfo 模块信息
type ModuleInfo struct {
	Name         string   `json:"name"`         // Name 模块名称（目录路径，根目录为 "."）
	Path         string   `json:"path"`         // Path 模块路径
	Dependencies []string `json:"dependencies"` // Dependencies 依赖的其他模块名称
	KeyFiles     []string `json:"key_files"`    // KeyFiles 模块内关键文件
	Interfaces   []string `json:"interfaces"`   // Interfaces 对外接口（导出符号，v1 暂不提取）
}

// ── Import 正则模式 ──

var (
	// Go import 正则
	goImportBlockRe  = regexp.MustCompile(`(?m)^\s*import\s+\(([\s\S]*?)\)`) // 多行 import 块
	goSingleImportRe = regexp.MustCompile(`(?m)^\s*import\s+"([^"]+)"`)      // 单行 import
	goImportLineRe   = regexp.MustCompile(`"([^"]+)"`)                       // import 块内的引号路径

	// Python import 正则
	pythonImportRe     = regexp.MustCompile(`(?m)^\s*import\s+(\S+)`)
	pythonFromImportRe = regexp.MustCompile(`(?m)^\s*from\s+(\S+)\s+import`)

	// JS/TS import 正则
	jsImportFromRe = regexp.MustCompile(`import\s+.*?\s+from\s+["']([^"']+)["']`)
	jsRequireRe    = regexp.MustCompile(`require\(["']([^"']+)["']\)`)
	jsImportPathRe = regexp.MustCompile(`import\s+["']([^"']+)["']`)

	// Java import 正则
	javaImportRe = regexp.MustCompile(`(?m)^\s*import\s+([\w.]+);`)

	// Rust use 正则
	rustUseRe = regexp.MustCompile(`(?m)^\s*use\s+([\w:]+)`)
)

// ── DependencyExtractorService ──

// DependencyExtractorService 依赖提取服务
//
// 扫描文本文件的 import/require 语句，构建模块级依赖摘要。
// 采用正则匹配（非语法树解析），支持 Go / Python / JS-TS / Java / Rust 五种语言。
type DependencyExtractorService struct{}

// NewDependencyExtractorService 创建 DependencyExtractorService 实例
func NewDependencyExtractorService() *DependencyExtractorService {
	return &DependencyExtractorService{}
}

// Extract 提取依赖关系，生成模块级摘要
//
// 参数：
//   - fileScanResult: 文件扫描结果（来自 FileScannerService.Scan）
//   - repoPath: 仓库本地根目录的绝对路径
//
// 返回模块级依赖摘要，包含模块列表、核心模块（top-5 被依赖最多）和总 import 数。
func (s *DependencyExtractorService) Extract(fileScanResult *FileScanResult, repoPath string) (*DependencySummary, error) {
	if fileScanResult == nil {
		return nil, fmt.Errorf("文件扫描结果不能为空")
	}

	// 1. 读取 go.mod 获取 Go 模块路径（用于判断 Go 内部 import）
	goModulePath := s.detectGoModulePath(repoPath)

	// 2. 按目录分组文件 → 定义模块（模块 = 文件所在目录相对路径）
	moduleFiles := make(map[string][]FileInfo)
	for _, file := range fileScanResult.Files {
		moduleName := dirToModule(file.Path)
		moduleFiles[moduleName] = append(moduleFiles[moduleName], file)
	}

	// 3. 收集所有内部模块路径（用于判断 import 是否为内部依赖）
	internalModules := make(map[string]bool, len(moduleFiles))
	for mod := range moduleFiles {
		internalModules[mod] = true
	}

	// 4. 解析每个文件的 import，构建模块间依赖关系
	moduleDeps := make(map[string]map[string]bool) // module → set of dep modules
	totalImports := 0

	for moduleName, files := range moduleFiles {
		if moduleDeps[moduleName] == nil {
			moduleDeps[moduleName] = make(map[string]bool)
		}
		for _, file := range files {
			content, err := os.ReadFile(filepath.Join(repoPath, filepath.FromSlash(file.Path)))
			if err != nil {
				continue // 跳过无法读取的文件
			}
			imports := s.extractImports(string(content), file.Language)
			totalImports += len(imports)
			for _, imp := range imports {
				depModule := s.resolveInternalModule(imp, file.Path, file.Language, goModulePath, internalModules)
				if depModule != "" && depModule != moduleName {
					moduleDeps[moduleName][depModule] = true
				}
			}
		}
	}

	// 5. 统计被依赖次数 → 识别核心模块（top-5）
	depCount := make(map[string]int)
	for _, deps := range moduleDeps {
		for dep := range deps {
			depCount[dep]++
		}
	}
	coreModules := topDependedModules(depCount, 5)

	// 6. 构建 ModuleInfo 列表
	modules := make([]ModuleInfo, 0, len(moduleFiles))
	for moduleName, files := range moduleFiles {
		deps := sortedKeys(moduleDeps[moduleName])
		keyFiles := make([]string, 0, len(files))
		for _, f := range files {
			keyFiles = append(keyFiles, f.Path)
		}
		sort.Strings(keyFiles)

		modules = append(modules, ModuleInfo{
			Name:         moduleName,
			Path:         moduleName,
			Dependencies: deps,
			KeyFiles:     keyFiles,
			Interfaces:   []string{},
		})
	}
	sort.Slice(modules, func(i, j int) bool {
		return modules[i].Name < modules[j].Name
	})

	return &DependencySummary{
		Modules:      modules,
		CoreModules:  coreModules,
		TotalImports: totalImports,
	}, nil
}

// ── Import 提取（按语言分发）──

// extractImports 根据文件语言提取 import 语句
//
// language 参数匹配 FileScannerService 输出的语言名（首字母大写格式，
// 如 "Go"、"TypeScript"）。同时兼容小写格式以增强鲁棒性。
func (s *DependencyExtractorService) extractImports(content, language string) []string {
	switch strings.ToLower(language) {
	case "go":
		return s.extractGoImports(content)
	case "python":
		return s.extractPythonImports(content)
	case "javascript", "typescript":
		return s.extractJSImports(content)
	case "java":
		return s.extractJavaImports(content)
	case "rust":
		return s.extractRustImports(content)
	default:
		return nil
	}
}

// extractGoImports 提取 Go import 语句（支持单行 import 和多行 import 块）
func (s *DependencyExtractorService) extractGoImports(content string) []string {
	var imports []string
	seen := make(map[string]bool)
	add := func(pkg string) {
		if pkg != "" && !seen[pkg] {
			imports = append(imports, pkg)
			seen[pkg] = true
		}
	}

	// 多行 import 块: import ( ... )
	for _, m := range goImportBlockRe.FindAllStringSubmatch(content, -1) {
		for _, line := range goImportLineRe.FindAllStringSubmatch(m[1], -1) {
			add(line[1])
		}
	}

	// 单行 import: import "pkg"
	for _, m := range goSingleImportRe.FindAllStringSubmatch(content, -1) {
		add(m[1])
	}

	return imports
}

// extractPythonImports 提取 Python import 语句
func (s *DependencyExtractorService) extractPythonImports(content string) []string {
	var imports []string
	seen := make(map[string]bool)
	add := func(pkg string) {
		if pkg != "" && !seen[pkg] {
			imports = append(imports, pkg)
			seen[pkg] = true
		}
	}

	for _, m := range pythonImportRe.FindAllStringSubmatch(content, -1) {
		add(m[1])
	}
	for _, m := range pythonFromImportRe.FindAllStringSubmatch(content, -1) {
		add(m[1])
	}
	return imports
}

// extractJSImports 提取 JS/TS import 语句（import...from / require / import "..."）
func (s *DependencyExtractorService) extractJSImports(content string) []string {
	var imports []string
	seen := make(map[string]bool)
	add := func(pkg string) {
		if pkg != "" && !seen[pkg] {
			imports = append(imports, pkg)
			seen[pkg] = true
		}
	}

	for _, m := range jsImportFromRe.FindAllStringSubmatch(content, -1) {
		add(m[1])
	}
	for _, m := range jsRequireRe.FindAllStringSubmatch(content, -1) {
		add(m[1])
	}
	for _, m := range jsImportPathRe.FindAllStringSubmatch(content, -1) {
		add(m[1])
	}
	return imports
}

// extractJavaImports 提取 Java import 语句
func (s *DependencyExtractorService) extractJavaImports(content string) []string {
	var imports []string
	seen := make(map[string]bool)
	for _, m := range javaImportRe.FindAllStringSubmatch(content, -1) {
		pkg := m[1]
		if !seen[pkg] {
			imports = append(imports, pkg)
			seen[pkg] = true
		}
	}
	return imports
}

// extractRustImports 提取 Rust use 语句
func (s *DependencyExtractorService) extractRustImports(content string) []string {
	var imports []string
	seen := make(map[string]bool)
	for _, m := range rustUseRe.FindAllStringSubmatch(content, -1) {
		pkg := m[1]
		if !seen[pkg] {
			imports = append(imports, pkg)
			seen[pkg] = true
		}
	}
	return imports
}

// ── 内部模块解析 ──

// resolveInternalModule 判断 import 是否指向仓库内部模块，返回模块名或空字符串
func (s *DependencyExtractorService) resolveInternalModule(
	importPath string,
	fileRelPath string,
	language string,
	goModulePath string,
	internalModules map[string]bool,
) string {
	switch strings.ToLower(language) {
	case "go":
		return resolveGoModule(importPath, goModulePath, internalModules)
	case "javascript", "typescript":
		return resolveRelativeModule(importPath, fileRelPath, internalModules)
	case "python":
		return resolveRelativeModule(importPath, fileRelPath, internalModules)
	default:
		return ""
	}
}

// resolveGoModule 解析 Go import 路径，判断是否为内部模块
//
// importPath 形如 "example.com/sample/internal/handler"，
// 去掉 goModulePath 前缀后得到 "internal/handler"，
// 在 internalModules 中查找匹配（支持子包向上查找父模块）。
func resolveGoModule(importPath, goModulePath string, internalModules map[string]bool) string {
	if goModulePath == "" {
		return ""
	}

	var modPath string
	switch {
	case importPath == goModulePath:
		modPath = "."
	case strings.HasPrefix(importPath, goModulePath+"/"):
		modPath = strings.TrimPrefix(importPath, goModulePath+"/")
	default:
		return "" // 外部依赖
	}

	// 精确匹配
	if internalModules[modPath] {
		return modPath
	}
	// 子包向上查找父模块（modPath 可能是 "internal/handler/auth"，而模块是 "internal/handler"）
	parts := strings.Split(modPath, "/")
	for i := len(parts) - 1; i > 0; i-- {
		parent := strings.Join(parts[:i], "/")
		if internalModules[parent] {
			return parent
		}
	}
	return ""
}

// resolveRelativeModule 解析相对路径 import（JS/TS/Python），判断是否为内部模块
//
// importPath 形如 "./utils" 或 "../components/Button"，
// 基于 fileRelPath 的目录解析为绝对模块路径后，在 internalModules 中查找。
func resolveRelativeModule(importPath, fileRelPath string, internalModules map[string]bool) string {
	// 仅处理以 "." 开头的相对路径
	if !strings.HasPrefix(importPath, ".") {
		return ""
	}

	fileDir := filepath.ToSlash(filepath.Dir(fileRelPath))
	resolved := filepath.ToSlash(filepath.Clean(filepath.Join(fileDir, importPath)))
	// 去除文件扩展名（import 路径通常不带扩展名）
	resolved = strings.TrimSuffix(resolved, filepath.Ext(resolved))

	// 精确匹配
	if internalModules[resolved] {
		return resolved
	}
	// 向上查找父模块
	parts := strings.Split(resolved, "/")
	for i := len(parts) - 1; i > 0; i-- {
		parent := strings.Join(parts[:i], "/")
		if internalModules[parent] {
			return parent
		}
	}
	return ""
}

// ── 辅助函数 ──

// detectGoModulePath 从 go.mod 中解析 Go 模块路径
func (s *DependencyExtractorService) detectGoModulePath(repoPath string) string {
	content, err := os.ReadFile(filepath.Join(repoPath, "go.mod"))
	if err != nil {
		return ""
	}
	for line := range strings.SplitSeq(string(content), "\n") {
		line = strings.TrimSpace(line)
		if after, found := strings.CutPrefix(line, "module "); found {
			return strings.TrimSpace(after)
		}
	}
	return ""
}

// dirToModule 从文件相对路径提取模块名（= 目录路径，根目录为 "."）
func dirToModule(relPath string) string {
	dir := filepath.ToSlash(filepath.Dir(relPath))
	if dir == "." {
		return "."
	}
	return dir
}

// topDependedModules 返回被依赖次数最多的 top-N 模块
func topDependedModules(depCount map[string]int, n int) []string {
	type kv struct {
		key string
		cnt int
	}
	items := make([]kv, 0, len(depCount))
	for k, v := range depCount {
		items = append(items, kv{k, v})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].cnt != items[j].cnt {
			return items[i].cnt > items[j].cnt // 次数降序
		}
		return items[i].key < items[j].key // 同次数按名称升序
	})
	if len(items) > n {
		items = items[:n]
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		result = append(result, item.key)
	}
	return result
}

// sortedKeys 返回 map key 的排序切片
func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
