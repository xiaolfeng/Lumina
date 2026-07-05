package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// ── 辅助：用临时目录构造 WikiStorageService ──

// newTestWikiStorage 使用 RepoWikiTestHelper 的临时目录构造测试用 WikiStorageService
func newTestWikiStorage(t *testing.T) (*WikiStorageService, *RepoWikiTestHelper) {
	t.Helper()
	helper := NewRepoWikiTestHelper(t)
	storage := &WikiStorageService{basePath: helper.TempDir}
	return storage, helper
}

// ──────────────────────────────────────────────────────────────
// TestSanitizePath 路径遍历防护测试
// ──────────────────────────────────────────────────────────────

func TestSanitizePath(t *testing.T) {
	storage, _ := newTestWikiStorage(t)

	tests := []struct {
		name      string
		input     string
		wantError bool // true = 期望拦截
	}{
		// ── 正常路径 ──
		{
			name:      "正常相对路径",
			input:     "versions/123/file_scan.json",
			wantError: false,
		},
		{
			name:      "正常嵌套路径",
			input:     "wiki/42/zh/index.md",
			wantError: false,
		},
		{
			name:      "当前目录引用",
			input:     "./repos/1/data",
			wantError: false,
		},
		{
			name:      "带冗余分隔符的路径",
			input:     "versions//123///file.json",
			wantError: false,
		},
		{
			name:      "内部相对路径（正常消除）",
			input:     "versions/123/pass/../file.json",
			wantError: false,
		},

		// ── 路径遍历攻击 ──
		{
			name:      "../etc/passwd 经典遍历",
			input:     "../etc/passwd",
			wantError: true,
		},
		{
			name:      "三层向上遍历",
			input:     "../../../etc/shadow",
			wantError: true,
		},
		{
			name:      "伪装遍历（中间嵌套）",
			input:     "versions/../../../etc/passwd",
			wantError: true,
		},
		{
			name:      "深层遍历",
			input:     "wiki/1/../../../../../../etc/passwd",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := storage.SanitizePath(tt.input)
			if tt.wantError {
				if err == nil {
					t.Fatalf("期望路径被拦截，但通过了: input=%q result=%q", tt.input, result)
				}
				t.Logf("✓ 成功拦截: input=%q → %v", tt.input, err)
				return
			}
			if err != nil {
				t.Fatalf("正常路径被误拦截: input=%q err=%v", tt.input, err)
			}
			// 验证返回的是绝对路径
			if !filepath.IsAbs(result) {
				t.Errorf("结果不是绝对路径: %q", result)
			}
			t.Logf("✓ 正常通过: input=%q → %q", tt.input, result)
		})
	}
}

// TestSanitizePath_AbsoluteTraversal 绝对路径遍历拦截
func TestSanitizePath_AbsoluteTraversal(t *testing.T) {
	storage, _ := newTestWikiStorage(t)

	// 绝对路径在 basePath 之外
	absOutside := "/etc/passwd"
	_, err := storage.SanitizePath(absOutside)
	if err == nil {
		t.Fatalf("basePath 外的绝对路径应被拦截: %q", absOutside)
	}
	t.Logf("✓ 成功拦截 basePath 外的绝对路径: %q → %v", absOutside, err)
}

// ──────────────────────────────────────────────────────────────
// TestWikiStorageJSON JSON 读写测试
// ──────────────────────────────────────────────────────────────

// testJSONData JSON round-trip 测试用的数据结构
type testJSONData struct {
	Name    string   `json:"name"`    // 名称
	Count   int      `json:"count"`   // 计数
	Tags    []string `json:"tags"`    // 标签
	Active  bool     `json:"active"`  // 是否激活
	Private float64  `json:"private"` // 私有值
}

func TestWikiStorageJSON(t *testing.T) {
	storage, _ := newTestWikiStorage(t)

	original := testJSONData{
		Name:    "Lumina RepoWiki",
		Count:   42,
		Tags:    []string{"go", "wiki", "llm"},
		Active:  true,
		Private: 3.14159,
	}

	jsonPath := filepath.Join(storage.basePath, "versions", "999", "file_scan.json")

	// ── WriteJSON ──
	t.Run("WriteJSON", func(t *testing.T) {
		if xErr := storage.WriteJSON(jsonPath, original); xErr != nil {
			t.Fatalf("WriteJSON 失败: %v", xErr)
		}

		// 确认文件存在
		info, err := os.Stat(jsonPath)
		if err != nil {
			t.Fatalf("文件不存在: %v", err)
		}
		if info.Size() == 0 {
			t.Fatal("文件大小为 0")
		}
		t.Logf("✓ JSON 写入成功: %s (%d bytes)", jsonPath, info.Size())
	})

	// ── ReadJSON ──
	t.Run("ReadJSON", func(t *testing.T) {
		var loaded testJSONData
		if xErr := storage.ReadJSON(jsonPath, &loaded); xErr != nil {
			t.Fatalf("ReadJSON 失败: %v", xErr)
		}

		// 验证 round-trip 数据一致性
		if loaded.Name != original.Name {
			t.Errorf("Name 不匹配: got=%q want=%q", loaded.Name, original.Name)
		}
		if loaded.Count != original.Count {
			t.Errorf("Count 不匹配: got=%d want=%d", loaded.Count, original.Count)
		}
		if len(loaded.Tags) != len(original.Tags) {
			t.Errorf("Tags 长度不匹配: got=%d want=%d", len(loaded.Tags), len(original.Tags))
		}
		if loaded.Active != original.Active {
			t.Errorf("Active 不匹配: got=%v want=%v", loaded.Active, original.Active)
		}
		if loaded.Private != original.Private {
			t.Errorf("Private 不匹配: got=%f want=%f", loaded.Private, original.Private)
		}
		t.Logf("✓ JSON round-trip 数据一致: %+v", loaded)
	})

	// ── JSON 缩进验证 ──
	t.Run("JSONIndent", func(t *testing.T) {
		raw, err := os.ReadFile(jsonPath)
		if err != nil {
			t.Fatalf("读取文件失败: %v", err)
		}
		// 确认是格式化的 JSON（包含换行符 + 缩进）
		if string(raw) == "" {
			t.Fatal("文件为空")
		}
		var verify map[string]interface{}
		if err := json.Unmarshal(raw, &verify); err != nil {
			t.Fatalf("文件内容不是有效 JSON: %v", err)
		}
		// 检查包含缩进空白（MarshalIndent 用 2 空格）
		if !contains(string(raw), "  ") {
			t.Error("JSON 未使用缩进格式")
		}
		t.Logf("✓ JSON 缩进格式正确")
	})

	// ── ReadJSON 不存在文件 ──
	t.Run("ReadJSONNotFound", func(t *testing.T) {
		var data testJSONData
		xErr := storage.ReadJSON("/nonexistent/path/missing.json", &data)
		if xErr == nil {
			t.Fatal("读取不存在的文件应返回错误")
		}
		t.Logf("✓ 不存在文件正确返回错误: %v", xErr)
	})
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || indexOf(s, substr) >= 0)
}

// indexOf 简单字符串搜索
func indexOf(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// ──────────────────────────────────────────────────────────────
// TestWikiStorageMarkdown Markdown 读写测试
// ──────────────────────────────────────────────────────────────

func TestWikiStorageMarkdown(t *testing.T) {
	storage, _ := newTestWikiStorage(t)

	original := "# Lumina 项目\n\n## 概述\n这是一个 **Wiki** 文档测试。\n\n- 项目一\n- 项目二\n- 项目三\n\n```go\nfunc main() {\n    fmt.Println(\"Hello, Lumina!\")\n}\n```\n"

	mdPath := filepath.Join(storage.basePath, "wiki", "777", "zh", "index.md")

	// ── WriteMarkdown ──
	if xErr := storage.WriteMarkdown(mdPath, original); xErr != nil {
		t.Fatalf("WriteMarkdown 失败: %v", xErr)
	}

	// 确认文件存在
	if _, err := os.Stat(mdPath); err != nil {
		t.Fatalf("Markdown 文件不存在: %v", err)
	}
	t.Logf("✓ Markdown 写入成功: %s", mdPath)

	// ── ReadMarkdown ──
	loaded, xErr := storage.ReadMarkdown(mdPath)
	if xErr != nil {
		t.Fatalf("ReadMarkdown 失败: %v", xErr)
	}
	if loaded != original {
		t.Errorf("Markdown round-trip 数据不一致:\n--- got ---\n%s\n--- want ---\n%s", loaded, original)
	}
	t.Logf("✓ Markdown round-trip 数据一致（%d chars）", len(loaded))
}

// ──────────────────────────────────────────────────────────────
// TestCleanVersion 版本清理测试
// ──────────────────────────────────────────────────────────────

func TestCleanVersion(t *testing.T) {
	storage, _ := newTestWikiStorage(t)

	versionID := int64(888)

	// 创建版本目录结构
	versionPath := storage.GetVersionPath(versionID)
	dirs := []string{
		filepath.Join(versionPath, "raw"),
		filepath.Join(versionPath, "passes"),
		filepath.Join(versionPath, "sessions"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatalf("创建目录失败 %s: %v", d, err)
		}
	}
	// 写入一些测试文件
	files := []string{
		filepath.Join(versionPath, "file_scan.json"),
		filepath.Join(versionPath, "dep_summary.json"),
		filepath.Join(versionPath, "passes", "pass1.md"),
		filepath.Join(versionPath, "raw", "main.go"),
	}
	for _, f := range files {
		if xErr := storage.WriteMarkdown(f, "test content"); xErr != nil {
			t.Fatalf("写入文件失败 %s: %v", f, xErr)
		}
	}

	// 确认目录存在
	if _, err := os.Stat(versionPath); os.IsNotExist(err) {
		t.Fatal("版本目录创建后不存在")
	}

	// ── CleanVersion ──
	if xErr := storage.CleanVersion(versionID); xErr != nil {
		t.Fatalf("CleanVersion 失败: %v", xErr)
	}

	// 确认目录已被删除
	if _, err := os.Stat(versionPath); !os.IsNotExist(err) {
		t.Fatalf("版本目录清理后仍存在: %s", versionPath)
	}
	t.Logf("✓ 版本目录清理成功: %s", versionPath)

	// ── CleanVersion 幂等（目录不存在时不报错）──
	if xErr := storage.CleanVersion(versionID); xErr != nil {
		t.Fatalf("CleanVersion 不幂等（目录不存在时不应报错）: %v", xErr)
	}
	t.Logf("✓ CleanVersion 幂等性验证通过")
}

// TestCleanRepo 仓库清理测试
func TestCleanRepo(t *testing.T) {
	storage, _ := newTestWikiStorage(t)

	configID := int64(666)

	// 创建仓库目录
	repoPath := storage.GetRepoPath(configID)
	if err := os.MkdirAll(filepath.Join(repoPath, ".git"), 0755); err != nil {
		t.Fatalf("创建仓库目录失败: %v", err)
	}

	// ── CleanRepo ──
	if xErr := storage.CleanRepo(configID); xErr != nil {
		t.Fatalf("CleanRepo 失败: %v", xErr)
	}
	if _, err := os.Stat(repoPath); !os.IsNotExist(err) {
		t.Fatalf("仓库目录清理后仍存在: %s", repoPath)
	}
	t.Logf("✓ 仓库目录清理成功: %s", repoPath)
}

// ──────────────────────────────────────────────────────────────
// TestWikiStoragePaths 路径方法测试
// ──────────────────────────────────────────────────────────────

func TestWikiStoragePaths(t *testing.T) {
	storage, _ := newTestWikiStorage(t)

	t.Run("GetRepoPath", func(t *testing.T) {
		got := storage.GetRepoPath(100)
		wantSuffix := filepath.Join("repos", "100")
		if !endsWith(got, wantSuffix) {
			t.Errorf("GetRepoPath 后缀错误: got=%q want suffix=%q", got, wantSuffix)
		}
	})

	t.Run("GetVersionPath", func(t *testing.T) {
		got := storage.GetVersionPath(200)
		wantSuffix := filepath.Join("versions", "200")
		if !endsWith(got, wantSuffix) {
			t.Errorf("GetVersionPath 后缀错误: got=%q want suffix=%q", got, wantSuffix)
		}
	})

	t.Run("GetRawPath", func(t *testing.T) {
		got := storage.GetRawPath(300)
		wantSuffix := filepath.Join("versions", "300", "raw")
		if !endsWith(got, wantSuffix) {
			t.Errorf("GetRawPath 后缀错误: got=%q want suffix=%q", got, wantSuffix)
		}
	})

	t.Run("GetPassesPath", func(t *testing.T) {
		got := storage.GetPassesPath(301)
		wantSuffix := filepath.Join("versions", "301", "passes")
		if !endsWith(got, wantSuffix) {
			t.Errorf("GetPassesPath 后缀错误: got=%q want suffix=%q", got, wantSuffix)
		}
	})

	t.Run("GetWikiPath", func(t *testing.T) {
		got := storage.GetWikiPath(400)
		wantSuffix := filepath.Join("wiki", "400", "zh")
		if !endsWith(got, wantSuffix) {
			t.Errorf("GetWikiPath 后缀错误: got=%q want suffix=%q", got, wantSuffix)
		}
	})

	t.Run("GetFileScanPath", func(t *testing.T) {
		got := storage.GetFileScanPath(500)
		wantSuffix := filepath.Join("versions", "500", "file_scan.json")
		if !endsWith(got, wantSuffix) {
			t.Errorf("GetFileScanPath 后缀错误: got=%q want suffix=%q", got, wantSuffix)
		}
	})

	t.Run("GetDepSummaryPath", func(t *testing.T) {
		got := storage.GetDepSummaryPath(501)
		wantSuffix := filepath.Join("versions", "501", "dep_summary.json")
		if !endsWith(got, wantSuffix) {
			t.Errorf("GetDepSummaryPath 后缀错误: got=%q want suffix=%q", got, wantSuffix)
		}
	})

	t.Run("GetManifestPath", func(t *testing.T) {
		got := storage.GetManifestPath(600)
		wantSuffix := filepath.Join("wiki", "600", "zh", "meta", "repowiki-metadata.json")
		if !endsWith(got, wantSuffix) {
			t.Errorf("GetManifestPath 后缀错误: got=%q want suffix=%q", got, wantSuffix)
		}
	})

	t.Run("GetSessionPath", func(t *testing.T) {
		got := storage.GetSessionPath(700)
		wantSuffix := filepath.Join("versions", "700", "sessions")
		if !endsWith(got, wantSuffix) {
			t.Errorf("GetSessionPath 后缀错误: got=%q want suffix=%q", got, wantSuffix)
		}
	})
}

// endsWith 检查路径是否以指定后缀结尾
func endsWith(path, suffix string) bool {
	return len(path) >= len(suffix) && path[len(path)-len(suffix):] == suffix
}

// ──────────────────────────────────────────────────────────────
// TestEnsureDir 目录创建测试
// ──────────────────────────────────────────────────────────────

func TestEnsureDir(t *testing.T) {
	storage, _ := newTestWikiStorage(t)

	dirPath := filepath.Join(storage.basePath, "deeply", "nested", "dir")

	// 创建前不存在
	if _, err := os.Stat(dirPath); !os.IsNotExist(err) {
		t.Fatal("目录不应提前存在")
	}

	// EnsureDir
	if err := storage.EnsureDir(dirPath); err != nil {
		t.Fatalf("EnsureDir 失败: %v", err)
	}

	// 创建后存在
	info, err := os.Stat(dirPath)
	if err != nil {
		t.Fatalf("目录创建后不存在: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("路径不是目录")
	}

	// 幂等：再次调用不应报错
	if err := storage.EnsureDir(dirPath); err != nil {
		t.Fatalf("EnsureDir 不幂等: %v", err)
	}
	t.Logf("✓ EnsureDir 递归创建 + 幂等性验证通过: %s", dirPath)
}
