package service

import (
	"testing"
)

// ── Go Import 提取测试 ──

// TestExtractGoImports 验证 Go import 提取（单行 + 多行块）
func TestExtractGoImports(t *testing.T) {
	svc := NewDependencyExtractorService()

	// 合并所有 sample 片段为一段完整内容
	var content string
	for _, sample := range sampleGoImports {
		content += sample + "\n"
	}

	imports := svc.extractGoImports(content)

	expected := map[string]bool{
		"fmt":                               true,
		"strings":                           true,
		"context":                           true,
		"github.com/gin-gonic/gin":          true,
		"example.com/sample/internal/logic": true,
	}

	found := make(map[string]bool)
	for _, imp := range imports {
		found[imp] = true
	}

	for exp := range expected {
		if !found[exp] {
			t.Errorf("期望找到 import %q，但未找到。提取结果: %v", exp, imports)
		}
	}

	if len(imports) != len(expected) {
		t.Errorf("提取到 %d 个 import，期望 %d 个。提取结果: %v", len(imports), len(expected), imports)
	}
}

// ── Python Import 提取测试 ──

// TestExtractPythonImports 验证 Python import 提取
func TestExtractPythonImports(t *testing.T) {
	svc := NewDependencyExtractorService()

	var content string
	for _, sample := range samplePythonImports {
		content += sample + "\n"
	}

	imports := svc.extractPythonImports(content)

	expected := map[string]bool{
		"os":      true,
		"sys":     true,
		"flask":   true,
		"pathlib": true,
	}

	found := make(map[string]bool)
	for _, imp := range imports {
		found[imp] = true
	}

	for exp := range expected {
		if !found[exp] {
			t.Errorf("期望找到 import %q，但未找到。提取结果: %v", exp, imports)
		}
	}
}

// ── JS/TS Import 提取测试 ──

// TestExtractJSImports 验证 JS/TS import 提取（import...from / require / import "..."）
func TestExtractJSImports(t *testing.T) {
	svc := NewDependencyExtractorService()

	var content string
	for _, sample := range sampleJSImports {
		content += sample + "\n"
	}

	imports := svc.extractJSImports(content)

	expected := map[string]bool{
		"express": true,
		"react":   true,
		"axios":   true,
		"./types": true,
	}

	found := make(map[string]bool)
	for _, imp := range imports {
		found[imp] = true
	}

	for exp := range expected {
		if !found[exp] {
			t.Errorf("期望找到 import %q，但未找到。提取结果: %v", exp, imports)
		}
	}
}

// ── Java Import 提取测试 ──

// TestExtractJavaImports 验证 Java import 提取
func TestExtractJavaImports(t *testing.T) {
	svc := NewDependencyExtractorService()

	var content string
	for _, sample := range sampleJavaImports {
		content += sample + "\n"
	}

	imports := svc.extractJavaImports(content)

	expected := map[string]bool{
		"java.util.List":                  true,
		"com.example.service.UserService": true,
	}

	found := make(map[string]bool)
	for _, imp := range imports {
		found[imp] = true
	}

	for exp := range expected {
		if !found[exp] {
			t.Errorf("期望找到 import %q，但未找到。提取结果: %v", exp, imports)
		}
	}
}

// ── Rust Use 提取测试 ──

// TestExtractRustImports 验证 Rust use 提取
func TestExtractRustImports(t *testing.T) {
	svc := NewDependencyExtractorService()

	var content string
	for _, sample := range sampleRustImports {
		content += sample + "\n"
	}

	imports := svc.extractRustImports(content)

	expected := map[string]bool{
		"std::collections::HashMap": true,
		"serde::":                   true, // serde::{...} 正则捕获到 "::"
	}

	found := make(map[string]bool)
	for _, imp := range imports {
		found[imp] = true
	}

	for exp := range expected {
		if !found[exp] {
			t.Errorf("期望找到 import %q，但未找到。提取结果: %v", exp, imports)
		}
	}
}

// ── 模块级依赖摘要集成测试 ──

// TestModuleDependencySummary 验证模块级依赖摘要构建
//
// 使用 createSampleRepo 创建模拟仓库，手动构建 FileScanResult，
// 调用 Extract 验证模块间依赖关系和核心模块识别。
func TestModuleDependencySummary(t *testing.T) {
	repoPath := createSampleRepo(t)

	// 手动构建 FileScanResult（模拟 FileScannerService.Scan 输出，语言名使用首字母大写格式）
	scanResult := &FileScanResult{
		Files: []FileInfo{
			{Path: "main.go", Language: "Go"},
			{Path: "internal/handler/user.go", Language: "Go"},
			{Path: "internal/logic/user.go", Language: "Go"},
			{Path: "src/index.ts", Language: "TypeScript"},
			{Path: "src/utils.ts", Language: "TypeScript"},
		},
	}

	svc := NewDependencyExtractorService()
	summary, err := svc.Extract(scanResult, repoPath)
	if err != nil {
		t.Fatalf("Extract 失败: %v", err)
	}

	// ── 验证基本结构 ──
	if summary == nil {
		t.Fatal("摘要不能为 nil")
	}
	if len(summary.Modules) == 0 {
		t.Fatal("模块列表不能为空")
	}

	// 期望 4 个模块: ".", "internal/handler", "internal/logic", "src"
	expectedModuleCount := 4
	if len(summary.Modules) != expectedModuleCount {
		t.Errorf("模块数量 = %d, 期望 %d。模块列表: %v", len(summary.Modules), expectedModuleCount, moduleNames(summary.Modules))
	}

	// ── 验证模块间依赖关系 ──

	// 根模块（包含 main.go）应依赖 internal/handler
	rootModule := findModule(summary, ".")
	if rootModule == nil {
		t.Fatal("未找到根模块 \".\"")
	}
	if !containsStr(rootModule.Dependencies, "internal/handler") {
		t.Errorf("根模块应依赖 internal/handler，实际依赖: %v", rootModule.Dependencies)
	}

	// internal/handler 模块应依赖 internal/logic
	handlerModule := findModule(summary, "internal/handler")
	if handlerModule == nil {
		t.Fatal("未找到模块 internal/handler")
	}
	if !containsStr(handlerModule.Dependencies, "internal/logic") {
		t.Errorf("internal/handler 应依赖 internal/logic，实际依赖: %v", handlerModule.Dependencies)
	}

	// internal/logic 模块应无内部依赖（context / fmt 均为外部）
	logicModule := findModule(summary, "internal/logic")
	if logicModule == nil {
		t.Fatal("未找到模块 internal/logic")
	}
	if len(logicModule.Dependencies) != 0 {
		t.Errorf("internal/logic 应无内部依赖，实际依赖: %v", logicModule.Dependencies)
	}

	// ── 验证核心模块（被依赖最多 top-5）──
	// internal/handler 被根模块依赖（1 次）
	// internal/logic 被 internal/handler 依赖（1 次）
	if !containsStr(summary.CoreModules, "internal/handler") {
		t.Errorf("internal/handler 应为核心模块，CoreModules: %v", summary.CoreModules)
	}
	if !containsStr(summary.CoreModules, "internal/logic") {
		t.Errorf("internal/logic 应为核心模块，CoreModules: %v", summary.CoreModules)
	}

	// ── 验证总 import 数 ──
	// main.go: 3 imports (fmt, internal/handler, gin)
	// handler/user.go: 2 imports (gin, internal/logic)
	// logic/user.go: 2 imports (context, fmt)
	// src/index.ts: 2 imports (./utils, express)
	// src/utils.ts: 1 import (axios)
	// 合计: 10
	if summary.TotalImports != 10 {
		t.Errorf("TotalImports = %d, 期望 10", summary.TotalImports)
	}

	// ── 验证 KeyFiles ──
	if !containsStr(rootModule.KeyFiles, "main.go") {
		t.Errorf("根模块 KeyFiles 应包含 main.go，实际: %v", rootModule.KeyFiles)
	}
}

// ── 测试辅助函数 ──

func findModule(summary *DependencySummary, name string) *ModuleInfo {
	for i := range summary.Modules {
		if summary.Modules[i].Name == name {
			return &summary.Modules[i]
		}
	}
	return nil
}

func containsStr(slice []string, target string) bool {
	for _, s := range slice {
		if s == target {
			return true
		}
	}
	return false
}

func moduleNames(modules []ModuleInfo) []string {
	names := make([]string, 0, len(modules))
	for _, m := range modules {
		names = append(names, m.Name)
	}
	return names
}
