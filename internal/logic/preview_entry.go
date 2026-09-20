package logic

import (
	"strings"

	apiPreview "github.com/xiaolfeng/Lumina/api/preview"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
)

// FindPreviewEntry 从 apiPreview.PreviewFileResponse 列表中选择可评审入口文件
//
// 判定严格依据 MIME 常量（PreviewMimeLPW 与 PreviewMimeHTML），避免裸字符串与扩展名漂移：
// 1. 优先 index.lpw（MIME == PreviewMimeLPW 且文件名为 index.lpw）
// 2. 否则首个 MIME == PreviewMimeLPW 的文件
// 3. 否则回退现有 HTML 文件（MIME == PreviewMimeHTML）
// 4. 均无则返回 nil
func FindPreviewEntry(files []apiPreview.PreviewFileResponse) *apiPreview.PreviewFileResponse {
	var firstLpw *apiPreview.PreviewFileResponse
	var firstHTML *apiPreview.PreviewFileResponse

	for i := range files {
		f := &files[i]
		if f.MimeType == bConst.PreviewMimeLPW {
			if strings.EqualFold(f.Filename, "index.lpw") {
				return f
			}
			if firstLpw == nil {
				firstLpw = f
			}
		} else if f.MimeType == bConst.PreviewMimeHTML {
			if firstHTML == nil {
				firstHTML = f
			}
		}
	}

	if firstLpw != nil {
		return firstLpw
	}
	return firstHTML
}

// FindPreviewEntryFromPreviewFiles 从 []*entity.PreviewFile 中选择入口文件名（用于 Pages 晋升推导）
func FindPreviewEntryFromPreviewFiles(files []*entity.PreviewFile) string {
	var firstLpw string
	var firstHTML string

	for _, f := range files {
		if f.MimeType == bConst.PreviewMimeLPW {
			if strings.EqualFold(f.Filename, "index.lpw") {
				return f.Filename
			}
			if firstLpw == "" {
				firstLpw = f.Filename
			}
		} else if f.MimeType == bConst.PreviewMimeHTML {
			if firstHTML == "" {
				firstHTML = f.Filename
			}
		}
	}

	if firstLpw != "" {
		return firstLpw
	}
	return firstHTML
}

// FindPreviewEntryFromPageFiles 从 []*entity.PageFile 中选择入口文件名
func FindPreviewEntryFromPageFiles(files []*entity.PageFile) string {
	var firstLpw string
	var firstHTML string

	for _, f := range files {
		if f.MimeType == bConst.PreviewMimeLPW {
			if strings.EqualFold(f.Filename, "index.lpw") {
				return f.Filename
			}
			if firstLpw == "" {
				firstLpw = f.Filename
			}
		} else if f.MimeType == bConst.PreviewMimeHTML {
			if firstHTML == "" {
				firstHTML = f.Filename
			}
		}
	}

	if firstLpw != "" {
		return firstLpw
	}
	return firstHTML
}
