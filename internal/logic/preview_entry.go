package logic

import (
	"path/filepath"
	"strings"

	apiPreview "github.com/xiaolfeng/Lumina/api/preview"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
)

// FindPreviewEntry 从 apiPreview.PreviewFileResponse 列表中选择可评审入口文件
//
// 判定依据：
// 1. 优先 index.lpw（MIME == PreviewMimeLPW 且文件名为 index.lpw）
// 2. 否则首个 MIME == PreviewMimeLPW 的文件
// 3. 否则回退现有 HTML 文件（MIME == PreviewMimeHTML）
// 4. 否则回退主组件命名的 TSX/JSX（App/Index/Main）
// 5. 否则回退首个 TSX/JSX 文件
// 6. 均无则返回 nil
func FindPreviewEntry(files []apiPreview.PreviewFileResponse) *apiPreview.PreviewFileResponse {
	var firstLpw *apiPreview.PreviewFileResponse
	var firstHTML *apiPreview.PreviewFileResponse
	var mainTSX *apiPreview.PreviewFileResponse
	var firstTSX *apiPreview.PreviewFileResponse

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
		} else if f.MimeType == bConst.PreviewMimeTSX || f.MimeType == bConst.PreviewMimeJSX {
			if isMainTSXFilename(f.Filename) && mainTSX == nil {
				mainTSX = f
			}
			if firstTSX == nil {
				firstTSX = f
			}
		}
	}

	if firstLpw != nil {
		return firstLpw
	}
	if firstHTML != nil {
		return firstHTML
	}
	if mainTSX != nil {
		return mainTSX
	}
	return firstTSX
}

// FindPreviewEntryFromPreviewFiles 从 []*entity.PreviewFile 中选择入口文件名（用于 Pages 晋升推导）。
// Q-09 修复：保留扩展名回退兜底，防止历史行 MIME 漂移导致晋升失败。
func FindPreviewEntryFromPreviewFiles(files []*entity.PreviewFile) string {
	var firstLpw string
	var firstHTML string
	var mainTSX string
	var firstTSX string

	for _, f := range files {
		if isLpwEntryFile(f.Filename, f.MimeType) {
			if strings.EqualFold(f.Filename, "index.lpw") {
				return f.Filename
			}
			if firstLpw == "" {
				firstLpw = f.Filename
			}
		} else if isHTMLEntryFile(f.Filename, f.MimeType) {
			if firstHTML == "" {
				firstHTML = f.Filename
			}
		} else if isTSXEntryFile(f.Filename, f.MimeType) {
			if isMainTSXFilename(f.Filename) && mainTSX == "" {
				mainTSX = f.Filename
			}
			if firstTSX == "" {
				firstTSX = f.Filename
			}
		}
	}

	if firstLpw != "" {
		return firstLpw
	}
	if firstHTML != "" {
		return firstHTML
	}
	if mainTSX != "" {
		return mainTSX
	}
	return firstTSX
}

// FindPreviewEntryFromPageFiles 从 []*entity.PageFile 中选择入口文件名。
// Q-09 修复：同上，保留扩展名回退兜底。
func FindPreviewEntryFromPageFiles(files []*entity.PageFile) string {
	var firstLpw string
	var firstHTML string
	var mainTSX string
	var firstTSX string

	for _, f := range files {
		if isLpwEntryFile(f.Filename, f.MimeType) {
			if strings.EqualFold(f.Filename, "index.lpw") {
				return f.Filename
			}
			if firstLpw == "" {
				firstLpw = f.Filename
			}
		} else if isHTMLEntryFile(f.Filename, f.MimeType) {
			if firstHTML == "" {
				firstHTML = f.Filename
			}
		} else if isTSXEntryFile(f.Filename, f.MimeType) {
			if isMainTSXFilename(f.Filename) && mainTSX == "" {
				mainTSX = f.Filename
			}
			if firstTSX == "" {
				firstTSX = f.Filename
			}
		}
	}

	if firstLpw != "" {
		return firstLpw
	}
	if firstHTML != "" {
		return firstHTML
	}
	if mainTSX != "" {
		return mainTSX
	}
	return firstTSX
}

// isLpwEntryFile Pages 侧 LPW 入口判定：MIME 常量或 .lpw 扩展名（历史行兜底）
func isLpwEntryFile(filename, mimeType string) bool {
	if mimeType == bConst.PreviewMimeLPW {
		return true
	}
	return strings.ToLower(filepath.Ext(filename)) == ".lpw"
}

// isHTMLEntryFile Pages 侧 HTML 入口判定：MIME 常量或 .html/.htm 扩展名（历史行兜底）
func isHTMLEntryFile(filename, mimeType string) bool {
	if mimeType == bConst.PreviewMimeHTML {
		return true
	}
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".html" || ext == ".htm"
}

// isTSXEntryFile Pages 侧 TSX 入口判定：MIME 常量或 .tsx/.jsx 扩展名（历史行兜底）
func isTSXEntryFile(filename, mimeType string) bool {
	if mimeType == bConst.PreviewMimeTSX || mimeType == bConst.PreviewMimeJSX {
		return true
	}
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".tsx" || ext == ".jsx"
}

func isMainTSXFilename(filename string) bool {
	lower := strings.ToLower(filename)
	return strings.HasPrefix(lower, "app.") || strings.HasPrefix(lower, "index.") || strings.HasPrefix(lower, "main.")
}
