package service

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ── 辅助：用临时目录构造 saveWikiPageTool ──

func newTestSaveWikiPageTool(t *testing.T) (*saveWikiPageTool, string) {
	t.Helper()
	dir := t.TempDir()
	return &saveWikiPageTool{wikiDir: dir}, dir
}

func runSaveWikiPage(t *testing.T, tool *saveWikiPageTool, path, content string) (*struct {
	Success bool   `json:"success"`
	Path    string `json:"path"`
}, string, bool) {
	t.Helper()
	input, _ := json.Marshal(map[string]string{"path": path, "content": content})
	res, err := tool.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute 返回 error: %v", err)
	}
	if res.IsError {
		return nil, res.Content, false
	}
	var out struct {
		Success bool   `json:"success"`
		Path    string `json:"path"`
	}
	if jErr := json.Unmarshal([]byte(res.Content), &out); jErr != nil {
		t.Fatalf("解析结果失败: %v (raw=%s)", jErr, res.Content)
	}
	return &out, "", true
}

// ──────────────────────────────────────────────────────────────
// normalizeWikiPathExt 单元测试
// ──────────────────────────────────────────────────────────────

func TestNormalizeWikiPathExt(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "已是 .mdx", in: "overview.mdx", want: "overview.mdx"},
		{name: ".md 替换为 .mdx", in: "overview.md", want: "overview.mdx"},
		{name: "无扩展名追加 .mdx", in: "overview", want: "overview.mdx"},
		{name: "带目录无扩展名", in: "modules/auth", want: "modules/auth.mdx"},
		{name: "带目录 .md", in: "modules/auth.md", want: "modules/auth.mdx"},
		{name: "带目录 .mdx", in: "modules/auth.mdx", want: "modules/auth.mdx"},
		{name: "深层目录无扩展名", in: "a/b/c/page", want: "a/b/c/page.mdx"},
		{name: "其它扩展名追加 .mdx", in: "page.txt", want: "page.txt.mdx"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeWikiPathExt(tt.in)
			if got != tt.want {
				t.Fatalf("normalizeWikiPathExt(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// ──────────────────────────────────────────────────────────────
// deriveFrontmatterTitle 单元测试
// ──────────────────────────────────────────────────────────────

func TestDeriveFrontmatterTitle(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "mdx 去扩展名", in: "overview.mdx", want: "overview"},
		{name: "md 去扩展名", in: "overview.md", want: "overview"},
		{name: "无扩展名原样", in: "overview", want: "overview"},
		{name: "带目录取 basename", in: "modules/auth.mdx", want: "auth"},
		{name: "深层目录取 basename", in: "a/b/c/page.mdx", want: "page"},
		{name: "其它扩展名去除", in: "page.txt", want: "page"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deriveFrontmatterTitle(tt.in)
			if got != tt.want {
				t.Fatalf("deriveFrontmatterTitle(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// ──────────────────────────────────────────────────────────────
// ensureFrontmatter 单元测试
// ──────────────────────────────────────────────────────────────

func TestEnsureFrontmatter(t *testing.T) {
	t.Run("已有 frontmatter 原样返回", func(t *testing.T) {
		in := "---\ntitle: Foo\ndescription: bar\n---\n\nbody"
		got := ensureFrontmatter("overview.mdx", in)
		if got != in {
			t.Fatalf("expected原样返回, got=%q", got)
		}
	})

	t.Run("缺失 frontmatter 注入默认", func(t *testing.T) {
		got := ensureFrontmatter("overview.mdx", "body content")
		if !strings.HasPrefix(got, "---\n") {
			t.Fatalf("应注入 frontmatter, got=%q", got)
		}
		if !strings.Contains(got, "title: overview\n") {
			t.Fatalf("title 应为 overview, got=%q", got)
		}
		if !strings.Contains(got, "icon: FileText\n") {
			t.Fatalf("应包含 icon: FileText, got=%q", got)
		}
		if !strings.HasSuffix(got, "body content") {
			t.Fatalf("应保留原内容, got=%q", got)
		}
		if !strings.Contains(got, "---\n\nbody content") {
			t.Fatalf("frontmatter 与正文间应有空行, got=%q", got)
		}
	})

	t.Run("title 取自 basename", func(t *testing.T) {
		got := ensureFrontmatter("modules/auth.mdx", "x")
		if !strings.Contains(got, "title: auth\n") {
			t.Fatalf("title 应为 auth, got=%q", got)
		}
	})
}

// ──────────────────────────────────────────────────────────────
// saveWikiPageTool.Execute 集成测试
// ──────────────────────────────────────────────────────────────

func TestSaveWikiPage_MdxWithFrontmatter(t *testing.T) {
	tool, dir := newTestSaveWikiPageTool(t)
	content := "---\ntitle: Overview\ndescription: \nicon: FileText\n---\n\n# Hello"
	out, _, ok := runSaveWikiPage(t, tool, "overview.mdx", content)
	if !ok {
		t.Fatalf("expected success")
	}
	if out.Path != "overview.mdx" {
		t.Fatalf("path = %q, want overview.mdx", out.Path)
	}
	got, err := os.ReadFile(filepath.Join(dir, "overview.mdx"))
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}
	if string(got) != content {
		t.Fatalf("内容应原样写入, got=%q", string(got))
	}
}

func TestSaveWikiPage_MdAutoRename(t *testing.T) {
	tool, dir := newTestSaveWikiPageTool(t)
	content := "---\ntitle: Foo\ndescription: \nicon: FileText\n---\n\nbody"
	out, _, ok := runSaveWikiPage(t, tool, "overview.md", content)
	if !ok {
		t.Fatalf("expected success")
	}
	if out.Path != "overview.mdx" {
		t.Fatalf("path = %q, want overview.mdx", out.Path)
	}
	// 原始 .md 不应存在
	if _, err := os.Stat(filepath.Join(dir, "overview.md")); !os.IsNotExist(err) {
		t.Fatalf("overview.md 不应存在")
	}
	got, err := os.ReadFile(filepath.Join(dir, "overview.mdx"))
	if err != nil {
		t.Fatalf("读取 .mdx 失败: %v", err)
	}
	if string(got) != content {
		t.Fatalf("内容应原样写入, got=%q", string(got))
	}
}

func TestSaveWikiPage_NoExtAutoAppend(t *testing.T) {
	tool, dir := newTestSaveWikiPageTool(t)
	content := "---\ntitle: Overview\ndescription: \nicon: FileText\n---\n\nbody"
	out, _, ok := runSaveWikiPage(t, tool, "overview", content)
	if !ok {
		t.Fatalf("expected success")
	}
	if out.Path != "overview.mdx" {
		t.Fatalf("path = %q, want overview.mdx", out.Path)
	}
	got, err := os.ReadFile(filepath.Join(dir, "overview.mdx"))
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}
	if string(got) != content {
		t.Fatalf("内容应原样写入, got=%q", string(got))
	}
}

func TestSaveWikiPage_NoExtWithDirAutoAppend(t *testing.T) {
	tool, dir := newTestSaveWikiPageTool(t)
	content := "---\ntitle: Auth\ndescription: \nicon: FileText\n---\n\nbody"
	out, _, ok := runSaveWikiPage(t, tool, "modules/auth", content)
	if !ok {
		t.Fatalf("expected success")
	}
	if out.Path != "modules/auth.mdx" {
		t.Fatalf("path = %q, want modules/auth.mdx", out.Path)
	}
	got, err := os.ReadFile(filepath.Join(dir, "modules", "auth.mdx"))
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}
	if string(got) != content {
		t.Fatalf("内容应原样写入, got=%q", string(got))
	}
}

func TestSaveWikiPage_MdxNoFrontmatterInject(t *testing.T) {
	tool, dir := newTestSaveWikiPageTool(t)
	body := "# Hello world"
	out, _, ok := runSaveWikiPage(t, tool, "overview.mdx", body)
	if !ok {
		t.Fatalf("expected success")
	}
	if out.Path != "overview.mdx" {
		t.Fatalf("path = %q, want overview.mdx", out.Path)
	}
	got, err := os.ReadFile(filepath.Join(dir, "overview.mdx"))
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}
	gotStr := string(got)
	if !strings.HasPrefix(gotStr, "---\n") {
		t.Fatalf("应注入 frontmatter, got=%q", gotStr)
	}
	if !strings.Contains(gotStr, "title: overview\n") {
		t.Fatalf("title 应为 overview, got=%q", gotStr)
	}
	if !strings.Contains(gotStr, "icon: FileText\n") {
		t.Fatalf("应包含 icon: FileText, got=%q", gotStr)
	}
	if !strings.HasSuffix(gotStr, body) {
		t.Fatalf("应保留原内容, got=%q", gotStr)
	}
}

func TestSaveWikiPage_MetaDirRejected(t *testing.T) {
	tool, _ := newTestSaveWikiPageTool(t)
	_, msg, ok := runSaveWikiPage(t, tool, "meta/x", "content")
	if ok {
		t.Fatalf("meta/ 应被拒绝")
	}
	if !strings.Contains(msg, "meta") {
		t.Fatalf("错误信息应提及 meta, got=%q", msg)
	}
}

func TestSaveWikiPage_InjectedTitleFromBasename(t *testing.T) {
	tool, dir := newTestSaveWikiPageTool(t)
	body := "body"
	_, _, ok := runSaveWikiPage(t, tool, "modules/auth.mdx", body)
	if !ok {
		t.Fatalf("expected success")
	}
	got, err := os.ReadFile(filepath.Join(dir, "modules", "auth.mdx"))
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}
	gotStr := string(got)
	if !strings.Contains(gotStr, "title: auth\n") {
		t.Fatalf("title 应为 auth (basename 去扩展名), got=%q", gotStr)
	}
}
