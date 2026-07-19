package handler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"

	apiRepowiki "github.com/xiaolfeng/Lumina/api/repowiki"
	wikiService "github.com/xiaolfeng/Lumina/internal/service"
)

// ──────────────────────────────────────────────────────────────
// TestGetWikiPage buildWikiPageResponse 单元测试
// ──────────────────────────────────────────────────────────────

// newWikiStorageForTest 构造测试用 WikiStorageService（basePath 无关紧要，ReadPage 接收完整路径）
func newWikiStorageForTest() *wikiService.WikiStorageService {
	return wikiService.NewWikiStorageService()
}

func TestGetWikiPage(t *testing.T) {
	storage := newWikiStorageForTest()

	// 构造临时 Wiki 目录
	wikiDir := t.TempDir()
	// 子目录用于嵌套路径测试
	if err := os.MkdirAll(filepath.Join(wikiDir, "modules"), 0755); err != nil {
		t.Fatalf("创建子目录失败: %v", err)
	}

	tests := []struct {
		name           string
		setupFile      string // 相对 wikiDir 的文件路径
		setupContent   string // 文件内容
		pagePath       string // 请求的页面路径（catch-all 风格，以 / 开头）
		wantErr        bool
		wantTitle      string
		wantContent    string
		wantDesc       string
		wantIcon       string
		wantLastUpdate bool // true = 期望 LastUpdated > 0
	}{
		{
			name:           "有 frontmatter（标准三字段）",
			setupFile:      "overview.mdx",
			setupContent:   "---\ntitle: 入门指南\ndescription: 快速了解项目\nicon: BookOpen\n---\n\n# 入门指南\n\n正文内容\n",
			pagePath:       "/overview",
			wantErr:        false,
			wantTitle:      "入门指南",
			wantContent:    "\n# 入门指南\n\n正文内容\n",
			wantDesc:       "快速了解项目",
			wantIcon:       "BookOpen",
			wantLastUpdate: true,
		},
		{
			name:           "无 frontmatter 时 title 回退到 basename",
			setupFile:      "modules/auth.mdx",
			setupContent:   "# 认证模块\n\n这是纯正文，没有 frontmatter。\n",
			pagePath:       "/modules/auth",
			wantErr:        false,
			wantTitle:      "auth",
			wantContent:    "# 认证模块\n\n这是纯正文，没有 frontmatter。\n",
			wantDesc:       "",
			wantIcon:       "",
			wantLastUpdate: true,
		},
		{
			name:           "frontmatter 缺 title 时回退到 basename",
			setupFile:      "guide.mdx",
			setupContent:   "---\ndescription: 仅描述\nicon: FileText\n---\n正文\n",
			pagePath:       "/guide",
			wantErr:        false,
			wantTitle:      "guide",
			wantContent:    "正文\n",
			wantDesc:       "仅描述",
			wantIcon:       "FileText",
			wantLastUpdate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 写入测试文件
			fullPath := filepath.Join(wikiDir, tt.setupFile)
			if err := os.WriteFile(fullPath, []byte(tt.setupContent), 0644); err != nil {
				t.Fatalf("写入测试文件失败: %v", err)
			}
			// 确保文件 mtime 不为 0
			fixedTime := time.Unix(1000000000, 0)
			if err := os.Chtimes(fullPath, fixedTime, fixedTime); err != nil {
				t.Fatalf("设置文件时间失败: %v", err)
			}

			resp, xErr := buildWikiPageResponse(storage, wikiDir, tt.pagePath)
			if tt.wantErr {
				if xErr == nil {
					t.Fatalf("期望错误但未返回错误")
				}
				return
			}
			if xErr != nil {
				t.Fatalf("未期望错误但返回: %v", xErr)
			}

			if resp.Title != tt.wantTitle {
				t.Errorf("Title = %q, want %q", resp.Title, tt.wantTitle)
			}
			if resp.Content != tt.wantContent {
				t.Errorf("Content = %q, want %q", resp.Content, tt.wantContent)
			}
			if resp.Description != tt.wantDesc {
				t.Errorf("Description = %q, want %q", resp.Description, tt.wantDesc)
			}
			if resp.Icon != tt.wantIcon {
				t.Errorf("Icon = %q, want %q", resp.Icon, tt.wantIcon)
			}
			if tt.wantLastUpdate && resp.LastUpdated <= 0 {
				t.Errorf("LastUpdated = %d, 期望 > 0", resp.LastUpdated)
			}
			// Path 应为 pagePath 去掉前导 /
			wantPath := strings.TrimPrefix(tt.pagePath, "/")
			if resp.Path != wantPath {
				t.Errorf("Path = %q, want %q", resp.Path, wantPath)
			}
			// Language 应为默认值
			if resp.Language == "" {
				t.Errorf("Language 不应为空")
			}
		})
	}
}

// TestGetWikiPage_MdxNotFound 验证 .mdx 不存在时返回 404，且不会回退到 .md
func TestGetWikiPage_MdxNotFound(t *testing.T) {
	storage := newWikiStorageForTest()
	wikiDir := t.TempDir()

	// 仅写入 .md 文件（旧格式），不写 .mdx
	mdPath := filepath.Join(wikiDir, "legacy.md")
	if err := os.WriteFile(mdPath, []byte("# 旧版 .md 内容\n"), 0644); err != nil {
		t.Fatalf("写入 .md 文件失败: %v", err)
	}

	// 请求 /legacy → 期望查找 legacy.mdx → 不存在 → 404，不回退到 .md
	_, xErr := buildWikiPageResponse(storage, wikiDir, "/legacy")
	if xErr == nil {
		t.Fatal("期望 .mdx 不存在时返回错误，但未返回错误（可能错误地回退到了 .md）")
	}

	// 错误码应为 FileNotFound
	if xErr.GetErrorCode() != xError.FileNotFound {
		t.Errorf("错误码 = %v, 期望 FileNotFound", xErr.GetErrorCode())
	}
}

// TestGetWikiPage_PathTraversal 验证路径遍历拦截
func TestGetWikiPage_PathTraversal(t *testing.T) {
	storage := newWikiStorageForTest()
	wikiDir := t.TempDir()

	_, xErr := buildWikiPageResponse(storage, wikiDir, "/../../../etc/passwd")
	if xErr == nil {
		t.Fatal("期望路径遍历被拦截并返回错误")
	}
	if xErr.GetErrorCode() != xError.BadRequest {
		t.Errorf("错误码 = %v, 期望 BadRequest", xErr.GetErrorCode())
	}
}

// TestFrontmatterString 验证 frontmatterString 对各种类型的处理
func TestFrontmatterString(t *testing.T) {
	tests := []struct {
		name string
		fm   map[string]interface{}
		key  string
		want string
	}{
		{
			name: "nil map",
			fm:   nil,
			key:  "title",
			want: "",
		},
		{
			name: "缺失 key",
			fm:   map[string]interface{}{"title": "存在"},
			key:  "description",
			want: "",
		},
		{
			name: "nil 值",
			fm:   map[string]interface{}{"title": nil},
			key:  "title",
			want: "",
		},
		{
			name: "字符串值",
			fm:   map[string]interface{}{"title": "标题"},
			key:  "title",
			want: "标题",
		},
		{
			name: "整数值兜底",
			fm:   map[string]interface{}{"count": 42},
			key:  "count",
			want: "42",
		},
		{
			name: "布尔值兜底",
			fm:   map[string]interface{}{"flag": true},
			key:  "flag",
			want: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := frontmatterString(tt.fm, tt.key)
			if got != tt.want {
				t.Errorf("frontmatterString() = %q, want %q", got, tt.want)
			}
		})
	}
}

// ──────────────────────────────────────────────────────────────
// TestComputeNav computeNav 单元测试
// ──────────────────────────────────────────────────────────────

// buildTestManifest 构造测试用 manifest 导航树
//
// 结构：
//   - 入门（叶子）
//   - ---分隔线---（Separator）
//   - 模块（目录）
//     - 认证（叶子）
//     - 数据库（叶子）
//   - 指南（叶子）
//
// DFS 叶子序列（跳过 Separator）：[入门, 认证, 数据库, 指南]
func buildTestManifest() *apiRepowiki.WikiManifestResponse {
	return &apiRepowiki.WikiManifestResponse{
		Navigation: []apiRepowiki.WikiNavItem{
			{Title: "入门", Path: "intro", Icon: "Book"},
			{Separator: "分隔线"},
			{Title: "模块", Path: "modules", Icon: "Folder", Children: []apiRepowiki.WikiNavItem{
				{Title: "认证", Path: "modules/auth", Icon: "Lock"},
				{Title: "数据库", Path: "modules/db", Icon: "Database"},
			}},
			{Title: "指南", Path: "guide", Icon: "Map"},
		},
	}
}

func TestComputeNav(t *testing.T) {
	manifest := buildTestManifest()

	tests := []struct {
		name            string
		currentPagePath string
		wantPrev        *apiRepowiki.WikiNavRef
		wantNext        *apiRepowiki.WikiNavRef
		wantBreadcrumb  []apiRepowiki.WikiNavRef
	}{
		{
			name:            "首页无 prev，有 next",
			currentPagePath: "intro",
			wantPrev:        nil,
			wantNext:        &apiRepowiki.WikiNavRef{Title: "认证", Path: "modules/auth", Icon: "Lock"},
			wantBreadcrumb:  []apiRepowiki.WikiNavRef{{Title: "入门", Path: "intro", Icon: "Book"}},
		},
		{
			name:            "末页有 prev，无 next",
			currentPagePath: "guide",
			wantPrev:        &apiRepowiki.WikiNavRef{Title: "数据库", Path: "modules/db", Icon: "Database"},
			wantNext:        nil,
			wantBreadcrumb:  []apiRepowiki.WikiNavRef{{Title: "指南", Path: "guide", Icon: "Map"}},
		},
		{
			name:            "中间页有 prev 和 next",
			currentPagePath: "modules/auth",
			wantPrev:        &apiRepowiki.WikiNavRef{Title: "入门", Path: "intro", Icon: "Book"},
			wantNext:        &apiRepowiki.WikiNavRef{Title: "数据库", Path: "modules/db", Icon: "Database"},
			wantBreadcrumb: []apiRepowiki.WikiNavRef{
				{Title: "模块", Path: "modules", Icon: "Folder"},
				{Title: "认证", Path: "modules/auth", Icon: "Lock"},
			},
		},
		{
			name:            "嵌套目录第二页 breadcrumb 正确",
			currentPagePath: "modules/db",
			wantPrev:        &apiRepowiki.WikiNavRef{Title: "认证", Path: "modules/auth", Icon: "Lock"},
			wantNext:        &apiRepowiki.WikiNavRef{Title: "指南", Path: "guide", Icon: "Map"},
			wantBreadcrumb: []apiRepowiki.WikiNavRef{
				{Title: "模块", Path: "modules", Icon: "Folder"},
				{Title: "数据库", Path: "modules/db", Icon: "Database"},
			},
		},
		{
			name:            "Separator 被跳过（intro 的 next 是 modules/auth 而非 Separator）",
			currentPagePath: "intro",
			wantPrev:        nil,
			wantNext:        &apiRepowiki.WikiNavRef{Title: "认证", Path: "modules/auth", Icon: "Lock"},
			wantBreadcrumb:  []apiRepowiki.WikiNavRef{{Title: "入门", Path: "intro", Icon: "Book"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prev, next, breadcrumb := computeNav(manifest, tt.currentPagePath)

			if !navRefEqual(prev, tt.wantPrev) {
				t.Errorf("prev = %v, want %v", navRefFormat(prev), navRefFormat(tt.wantPrev))
			}
			if !navRefEqual(next, tt.wantNext) {
				t.Errorf("next = %v, want %v", navRefFormat(next), navRefFormat(tt.wantNext))
			}
			if !navRefSliceEqual(breadcrumb, tt.wantBreadcrumb) {
				t.Errorf("breadcrumb = %v, want %v", breadcrumb, tt.wantBreadcrumb)
			}
		})
	}
}

// TestComputeNav_NotFound 当前页不在导航树中时返回全 nil
func TestComputeNav_NotFound(t *testing.T) {
	manifest := buildTestManifest()
	prev, next, breadcrumb := computeNav(manifest, "nonexistent/path")
	if prev != nil || next != nil || breadcrumb != nil {
		t.Errorf("期望全 nil，得到 prev=%v next=%v breadcrumb=%v", prev, next, breadcrumb)
	}
}

// TestComputeNav_NilManifest nil manifest 或空 path 返回全 nil
func TestComputeNav_NilManifest(t *testing.T) {
	prev, next, breadcrumb := computeNav(nil, "intro")
	if prev != nil || next != nil || breadcrumb != nil {
		t.Errorf("nil manifest 期望全 nil")
	}

	manifest := buildTestManifest()
	prev, next, breadcrumb = computeNav(manifest, "")
	if prev != nil || next != nil || breadcrumb != nil {
		t.Errorf("空 path 期望全 nil")
	}
}

// navRefEqual 比较两个 *WikiNavRef（nil 视为相等当且仅当双方都 nil）
func navRefEqual(a, b *apiRepowiki.WikiNavRef) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Title == b.Title && a.Path == b.Path && a.Icon == b.Icon
}

// navRefSliceEqual 比较两个 WikiNavRef 切片
func navRefSliceEqual(a, b []apiRepowiki.WikiNavRef) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Title != b[i].Title || a[i].Path != b[i].Path || a[i].Icon != b[i].Icon {
			return false
		}
	}
	return true
}

// navRefFormat 格式化 *WikiNavRef 用于错误输出
func navRefFormat(r *apiRepowiki.WikiNavRef) string {
	if r == nil {
		return "<nil>"
	}
	return "{" + r.Title + "," + r.Path + "," + r.Icon + "}"
}
