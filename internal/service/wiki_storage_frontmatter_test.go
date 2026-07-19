package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// ──────────────────────────────────────────────────────────────
// TestReadPage ReadPage frontmatter 解析测试
// ──────────────────────────────────────────────────────────────

func TestReadPage(t *testing.T) {
	storage := newTestWikiStorage(t)

	tests := []struct {
		name           string
		content        string
		wantFrontmatter bool
		wantBody       string
		frontmatterKey string
		frontmatterVal interface{}
	}{
		{
			name: "有 frontmatter（标准三字段）",
			content: "---\ntitle: 入门指南\ndescription: 快速了解\nicon: BookOpen\n---\n\n# 入门指南\n\n正文内容\n",
			wantFrontmatter: true,
			wantBody:        "\n# 入门指南\n\n正文内容\n",
			frontmatterKey:  "title",
			frontmatterVal:  "入门指南",
		},
		{
			name:            "无 frontmatter（纯 Markdown）",
			content:         "# 标题\n\n这是纯正文，没有 frontmatter。\n",
			wantFrontmatter: false,
			wantBody:        "# 标题\n\n这是纯正文，没有 frontmatter。\n",
		},
		{
			name:            "未闭合 frontmatter（全文作为 Body）",
			content:         "---\ntitle: 未闭合\n这个文件没有结束分隔符\n",
			wantFrontmatter: false,
			wantBody:        "---\ntitle: 未闭合\n这个文件没有结束分隔符\n",
		},
		{
			name: "frontmatter 含中文与多行 description",
			content: "---\ntitle: 微明文档\ndescription: |\n  第一行描述\n  第二行描述\nicon: Sparkles\n---\n正文\n",
			wantFrontmatter: true,
			wantBody:        "正文\n",
			frontmatterKey:  "title",
			frontmatterVal:  "微明文档",
		},
		{
			name: "frontmatter 含 icon 字段",
			content: "---\ntitle: 测试页面\nicon: FileText\n---\nbody\n",
			wantFrontmatter: true,
			wantBody:        "body\n",
			frontmatterKey:  "icon",
			frontmatterVal:  "FileText",
		},
		{
			name:            "空文件",
			content:         "",
			wantFrontmatter: false,
			wantBody:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pagePath := filepath.Join(storage.basePath, "page_"+sanitizeFileName(tt.name)+".mdx")
			if err := os.WriteFile(pagePath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("写入测试文件失败: %v", err)
			}

			page, xErr := storage.ReadPage(pagePath)
			if xErr != nil {
				t.Fatalf("ReadPage 失败: %v", xErr)
			}

			// frontmatter 存在性
			if tt.wantFrontmatter {
				if page.Frontmatter == nil {
					t.Fatalf("期望有 frontmatter，但为 nil")
				}
				if tt.frontmatterKey != "" {
					got, ok := page.Frontmatter[tt.frontmatterKey]
					if !ok {
						t.Fatalf("frontmatter 缺少键 %q", tt.frontmatterKey)
					}
					if !equalFrontmatterValue(got, tt.frontmatterVal) {
						t.Errorf("frontmatter[%q] 不匹配: got=%v (%T) want=%v (%T)",
							tt.frontmatterKey, got, got, tt.frontmatterVal, tt.frontmatterVal)
					}
				}
			} else {
				if page.Frontmatter != nil {
					t.Errorf("期望无 frontmatter，但得到: %v", page.Frontmatter)
				}
			}

			// body
			if page.Body != tt.wantBody {
				t.Errorf("Body 不匹配:\n--- got ---\n%q\n--- want ---\n%q", page.Body, tt.wantBody)
			}

			// ModTime 应非零
			if page.ModTime.IsZero() {
				t.Errorf("ModTime 不应为零值")
			}
			if page.ModTime.After(time.Now().Add(time.Second)) {
				t.Errorf("ModTime 异常（未来时间）: %v", page.ModTime)
			}

			t.Logf("✓ %s: frontmatter=%v body_len=%d modtime=%v",
				tt.name, page.Frontmatter != nil, len(page.Body), page.ModTime)
		})
	}
}

// TestReadPage_NotFound 文件不存在时返回错误
func TestReadPage_NotFound(t *testing.T) {
	storage := newTestWikiStorage(t)
	_, xErr := storage.ReadPage(filepath.Join(storage.basePath, "nonexistent.mdx"))
	if xErr == nil {
		t.Fatal("读取不存在的文件应返回错误")
	}
	t.Logf("✓ 不存在文件正确返回错误: %v", xErr)
}

// TestReadPage_ModTime ModTime 反映文件实际修改时间
func TestReadPage_ModTime(t *testing.T) {
	storage := newTestWikiStorage(t)
	pagePath := filepath.Join(storage.basePath, "modtime_test.mdx")
	if err := os.WriteFile(pagePath, []byte("# hi\n"), 0644); err != nil {
		t.Fatalf("写入文件失败: %v", err)
	}

	// 修改文件 mtime 为过去某时刻
	pastTime := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(pagePath, pastTime, pastTime); err != nil {
		t.Fatalf("修改 mtime 失败: %v", err)
	}

	page, xErr := storage.ReadPage(pagePath)
	if xErr != nil {
		t.Fatalf("ReadPage 失败: %v", xErr)
	}

	// 允许 1 秒误差
	diff := page.ModTime.Sub(pastTime)
	if diff < -time.Second || diff > time.Second {
		t.Errorf("ModTime 不匹配: got=%v want≈%v (diff=%v)", page.ModTime, pastTime, diff)
	}
	t.Logf("✓ ModTime 正确: %v", page.ModTime)
}

// ──────────────────────────────────────────────────────────────
// TestReadMetaJSON meta.json 读取测试
// ──────────────────────────────────────────────────────────────

func TestReadMetaJSON(t *testing.T) {
	storage := newTestWikiStorage(t)

	t.Run("meta.json 存在", func(t *testing.T) {
		dirPath := filepath.Join(storage.basePath, "wiki", "section")
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			t.Fatalf("创建目录失败: %v", err)
		}
		meta := map[string]interface{}{
			"title":       "模块概览",
			"icon":        "Boxes",
			"pages":       []interface{}{"overview", "auth"},
			"description": "section 描述",
		}
		metaBytes, _ := json.MarshalIndent(meta, "", "  ")
		if err := os.WriteFile(filepath.Join(dirPath, "meta.json"), metaBytes, 0644); err != nil {
			t.Fatalf("写入 meta.json 失败: %v", err)
		}

		got, xErr := storage.ReadMetaJSON(dirPath)
		if xErr != nil {
			t.Fatalf("ReadMetaJSON 失败: %v", xErr)
		}
		if got == nil {
			t.Fatal("期望返回非 nil map")
		}
		if v, ok := got["title"].(string); !ok || v != "模块概览" {
			t.Errorf("title 不匹配: got=%v", got["title"])
		}
		if v, ok := got["icon"].(string); !ok || v != "Boxes" {
			t.Errorf("icon 不匹配: got=%v", got["icon"])
		}
		t.Logf("✓ meta.json 读取成功: %v", got)
	})

	t.Run("meta.json 不存在（返回 nil,nil）", func(t *testing.T) {
		dirPath := filepath.Join(storage.basePath, "empty_section")
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			t.Fatalf("创建目录失败: %v", err)
		}
		got, xErr := storage.ReadMetaJSON(dirPath)
		if xErr != nil {
			t.Fatalf("meta.json 不存在时应返回 nil,nil，但得到错误: %v", xErr)
		}
		if got != nil {
			t.Errorf("meta.json 不存在时应返回 nil map，但得到: %v", got)
		}
		t.Logf("✓ meta.json 不存在时正确返回 nil,nil")
	})

	t.Run("目录本身不存在（返回 nil,nil）", func(t *testing.T) {
		got, xErr := storage.ReadMetaJSON(filepath.Join(storage.basePath, "never_existed"))
		if xErr != nil {
			t.Fatalf("目录不存在时应返回 nil,nil，但得到错误: %v", xErr)
		}
		if got != nil {
			t.Errorf("目录不存在时应返回 nil map，但得到: %v", got)
		}
		t.Logf("✓ 目录不存在时正确返回 nil,nil")
	})
}

// ──────────────────────────────────────────────────────────────
// 辅助函数
// ──────────────────────────────────────────────────────────────

// sanitizeFileName 将测试用例名转为安全的文件名片段
func sanitizeFileName(name string) string {
	out := make([]byte, 0, len(name))
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			out = append(out, byte(r))
		default:
			out = append(out, '_')
		}
	}
	return string(out)
}

// equalFrontmatterValue 比较 frontmatter 值（处理 yaml.v3 把字符串解析为 string 的情况）
func equalFrontmatterValue(got, want interface{}) bool {
	// yaml.v3 把纯字符串解析为 string，多行 | 解析为 string（含换行）
	if gotStr, ok := got.(string); ok {
		if wantStr, ok2 := want.(string); ok2 {
			return gotStr == wantStr
		}
	}
	return got == want
}
