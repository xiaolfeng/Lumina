package logic

import (
	"testing"

	apiPreview "github.com/xiaolfeng/Lumina/api/preview"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
)

func TestFindPreviewEntry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		files     []apiPreview.PreviewFileResponse
		wantEntry string
	}{
		{
			name: "1. 仅 LPW 文件：优先选择 index.lpw",
			files: []apiPreview.PreviewFileResponse{
				{Filename: "other.lpw", MimeType: bConst.PreviewMimeLPW},
				{Filename: "index.lpw", MimeType: bConst.PreviewMimeLPW},
			},
			wantEntry: "index.lpw",
		},
		{
			name: "1b. 仅 LPW 文件：无 index 则选首个 lpw",
			files: []apiPreview.PreviewFileResponse{
				{Filename: "first.lpw", MimeType: bConst.PreviewMimeLPW},
				{Filename: "second.lpw", MimeType: bConst.PreviewMimeLPW},
			},
			wantEntry: "first.lpw",
		},
		{
			name: "2. LPW 与 HTML 共存：LPW 优先级高于 HTML",
			files: []apiPreview.PreviewFileResponse{
				{Filename: "index.html", MimeType: bConst.PreviewMimeHTML},
				{Filename: "document.lpw", MimeType: bConst.PreviewMimeLPW},
			},
			wantEntry: "document.lpw",
		},
		{
			name: "2b. LPW 与 HTML 共存：存在 index.lpw 优先命中",
			files: []apiPreview.PreviewFileResponse{
				{Filename: "index.html", MimeType: bConst.PreviewMimeHTML},
				{Filename: "index.lpw", MimeType: bConst.PreviewMimeLPW},
			},
			wantEntry: "index.lpw",
		},
		{
			name: "3. 仅 HTML 文件：回退选首个 HTML",
			files: []apiPreview.PreviewFileResponse{
				{Filename: "style.css", MimeType: bConst.PreviewMimeCSS},
				{Filename: "app.html", MimeType: bConst.PreviewMimeHTML},
			},
			wantEntry: "app.html",
		},
		{
			name: "3b. 仅 MDX 文件：优先 index.mdx",
			files: []apiPreview.PreviewFileResponse{
				{Filename: "overview.mdx", MimeType: bConst.PreviewMimeMDX},
				{Filename: "index.mdx", MimeType: bConst.PreviewMimeMDX},
			},
			wantEntry: "index.mdx",
		},
		{
			name: "3c. 仅 MDX 文件：无 index 优先选首个 mdx",
			files: []apiPreview.PreviewFileResponse{
				{Filename: "style.css", MimeType: bConst.PreviewMimeCSS},
				{Filename: "guide.mdx", MimeType: bConst.PreviewMimeMDX},
			},
			wantEntry: "guide.mdx",
		},
		{
			name: "3d. MDX 与 Markdown 共存：MDX 优先级高于 Markdown",
			files: []apiPreview.PreviewFileResponse{
				{Filename: "README.md", MimeType: bConst.PreviewMimeMarkdown},
				{Filename: "overview.mdx", MimeType: bConst.PreviewMimeMDX},
			},
			wantEntry: "overview.mdx",
		},
		{
			name: "4. 两者皆无：无入口",
			files: []apiPreview.PreviewFileResponse{
				{Filename: "style.css", MimeType: bConst.PreviewMimeCSS},
				{Filename: "script.js", MimeType: bConst.PreviewMimeJS},
			},
			wantEntry: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindPreviewEntry(tt.files)
			if tt.wantEntry == "" {
				if got != nil {
					t.Errorf("FindPreviewEntry() = %q, want nil", got.Filename)
				}
			} else {
				if got == nil || got.Filename != tt.wantEntry {
					t.Errorf("FindPreviewEntry() = %v, want %q", got, tt.wantEntry)
				}
			}
		})
	}
}

func TestFindPreviewEntryFromPreviewFiles(t *testing.T) {
	t.Parallel()

	files := []*entity.PreviewFile{
		{Filename: "style.css", MimeType: bConst.PreviewMimeCSS},
		{Filename: "page.html", MimeType: bConst.PreviewMimeHTML},
		{Filename: "index.lpw", MimeType: bConst.PreviewMimeLPW},
	}

	got := FindPreviewEntryFromPreviewFiles(files)
	if got != "index.lpw" {
		t.Errorf("FindPreviewEntryFromPreviewFiles() = %q, want index.lpw", got)
	}

	// 仅 HTML
	htmlFiles := []*entity.PreviewFile{
		{Filename: "index.html", MimeType: bConst.PreviewMimeHTML},
	}
	if got := FindPreviewEntryFromPreviewFiles(htmlFiles); got != "index.html" {
		t.Errorf("FindPreviewEntryFromPreviewFiles(html) = %q, want index.html", got)
	}

	// 仅 MDX
	mdxFiles := []*entity.PreviewFile{
		{Filename: "overview.mdx", MimeType: bConst.PreviewMimeMDX},
	}
	if got := FindPreviewEntryFromPreviewFiles(mdxFiles); got != "overview.mdx" {
		t.Errorf("FindPreviewEntryFromPreviewFiles(mdx) = %q, want overview.mdx", got)
	}

	// 无入口
	noneFiles := []*entity.PreviewFile{
		{Filename: "main.go", MimeType: bConst.PreviewMimePlain},
	}
	if got := FindPreviewEntryFromPreviewFiles(noneFiles); got != "" {
		t.Errorf("FindPreviewEntryFromPreviewFiles(none) = %q, want empty", got)
	}
}

// TestFindPreviewEntryExtensionFallback Q-09 回归：历史行 MIME 漂移时，
// Pages 侧入口判定必须回退到扩展名匹配（MCP 侧 FindPreviewEntry 仍保持严格 MIME）。
func TestFindPreviewEntryExtensionFallback(t *testing.T) {
	t.Parallel()

	// 1. PreviewFile：裸 "text/html"（无 charset）+ index.html 文件名 → 扩展名兜底命中
	legacyHTML := []*entity.PreviewFile{
		{Filename: "index.html", MimeType: "text/html"},
	}
	if got := FindPreviewEntryFromPreviewFiles(legacyHTML); got != "index.html" {
		t.Errorf("legacy bare text/html + index.html should match by extension, got %q", got)
	}

	// 2. PreviewFile：MIME 漂移为 text/plain 的 .htm 文件 → 扩展名兜底命中
	legacyHTM := []*entity.PreviewFile{
		{Filename: "home.htm", MimeType: bConst.PreviewMimePlain},
	}
	if got := FindPreviewEntryFromPreviewFiles(legacyHTM); got != "home.htm" {
		t.Errorf("plain-mime home.htm should match by extension, got %q", got)
	}

	// 3. PageFile：.lpw 扩展名但 MIME 漂移 → 扩展名兜底命中
	legacyLpw := []*entity.PageFile{
		{Filename: "doc.lpw", MimeType: bConst.PreviewMimeJSON},
	}
	if got := FindPreviewEntryFromPageFiles(legacyLpw); got != "doc.lpw" {
		t.Errorf("json-mime doc.lpw should match by extension, got %q", got)
	}

	// 3b. PreviewFile：.mdx 扩展名但 MIME 漂移为 text/plain → 扩展名兜底命中
	legacyMdx := []*entity.PreviewFile{
		{Filename: "overview.mdx", MimeType: bConst.PreviewMimePlain},
	}
	if got := FindPreviewEntryFromPreviewFiles(legacyMdx); got != "overview.mdx" {
		t.Errorf("plain-mime overview.mdx should match by extension, got %q", got)
	}

	// 4. MCP 侧 FindPreviewEntry 保持严格：裸 "text/html" 不命中（既有契约由 preview_tools_test 钉死）
	bare := []apiPreview.PreviewFileResponse{
		{Filename: "index.html", MimeType: "text/html"},
	}
	if entry := FindPreviewEntry(bare); entry != nil {
		t.Errorf("MCP FindPreviewEntry must stay strict on bare text/html, got %q", entry.Filename)
	}
}
