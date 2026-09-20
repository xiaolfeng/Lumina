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

	// 无入口
	noneFiles := []*entity.PreviewFile{
		{Filename: "main.go", MimeType: bConst.PreviewMimePlain},
	}
	if got := FindPreviewEntryFromPreviewFiles(noneFiles); got != "" {
		t.Errorf("FindPreviewEntryFromPreviewFiles(none) = %q, want empty", got)
	}
}
