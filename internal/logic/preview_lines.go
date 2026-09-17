package logic

import (
	"errors"
	"fmt"
	"strings"

	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

// 行级编辑回显的上下文与体积约束（变更主体前后各保留 3 行，区域整体上限 40 行）
const (
	previewEditContextLines   = 3
	previewEditRegionMaxLines = 40
)

// previewFileLines 拆行结果：行切片与原文件是否以换行符结尾
//
// 行语义契约（edit 与 get 共用，单测锁定）：
//   - 换行统一按 LF 处理：CRLF 与孤立 CR 读取时归一化，写回统一用 LF 连接；
//   - 尾随换行是终止符而非空行："a\nb\n" 为 2 行，"a\n\n" 为 2 行（第 2 行为空行）；
//   - 空文件为 0 行；行号一律 1 起始、闭区间。
type previewFileLines struct {
	lines           []string
	trailingNewline bool
}

// splitFileLines 将文件内容拆为行（CRLF/CR 归一为 LF，结尾单个换行视为终止符而非空行）
func splitFileLines(content string) previewFileLines {
	if content == "" {
		return previewFileLines{lines: []string{}}
	}
	normalized := strings.ReplaceAll(strings.ReplaceAll(content, "\r\n", "\n"), "\r", "\n")
	trailing := strings.HasSuffix(normalized, "\n")
	if trailing {
		normalized = normalized[:len(normalized)-1]
	}
	return previewFileLines{lines: strings.Split(normalized, "\n"), trailingNewline: trailing}
}

// joinFileLines 将行切片合回文件内容（按原文件的结尾换行符原样保留；空行集返回空串）
func joinFileLines(lines []string, trailingNewline bool) string {
	if len(lines) == 0 {
		return ""
	}
	joined := strings.Join(lines, "\n")
	if trailingNewline {
		joined += "\n"
	}
	return joined
}

// previewLineEdit 行级编辑请求（行号 1 起始闭区间；startLine 为 0 表示未指定，仅 insert 允许并追加到末尾）
type previewLineEdit struct {
	operation string
	startLine int
	endLine   int
	content   string
}

// applyLineEdit 对行切片应用行级编辑，返回新行切片与变更主体区间（新切片中 1 起始闭区间）
//
// 变更主体区间覆盖 insert/replace 写入的新行；delete 操作主体为空区间（start 为原删除起点、end 为 start-1）。
// 编辑内容的行拆分与文件一致（CRLF 归一、结尾单个换行视为终止符）。
func applyLineEdit(lines []string, edit previewLineEdit) ([]string, int, int, error) {
	total := len(lines)
	switch edit.operation {
	case bConst.PreviewEditOperationInsert:
		if edit.content == "" {
			return nil, 0, 0, errors.New("insert 操作的 content 不能为空")
		}
		pos := edit.startLine
		if pos == 0 {
			pos = total + 1
		}
		if pos < 1 || pos > total+1 {
			return nil, 0, 0, fmt.Errorf("insert 的 start_line 无效：需要 1 ≤ start_line ≤ %d（当前 %d）", total+1, pos)
		}
		contentLines := splitFileLines(edit.content).lines
		next := make([]string, 0, total+len(contentLines))
		next = append(next, lines[:pos-1]...)
		next = append(next, contentLines...)
		next = append(next, lines[pos-1:]...)
		return next, pos, pos + len(contentLines) - 1, nil
	case bConst.PreviewEditOperationReplace:
		if err := validateClosedRange(total, edit.startLine, edit.endLine); err != nil {
			return nil, 0, 0, fmt.Errorf("replace %s", err.Error())
		}
		contentLines := splitFileLines(edit.content).lines
		next := make([]string, 0, total+len(contentLines))
		next = append(next, lines[:edit.startLine-1]...)
		next = append(next, contentLines...)
		next = append(next, lines[edit.endLine:]...)
		return next, edit.startLine, edit.startLine + len(contentLines) - 1, nil
	case bConst.PreviewEditOperationDelete:
		if err := validateClosedRange(total, edit.startLine, edit.endLine); err != nil {
			return nil, 0, 0, fmt.Errorf("delete %s", err.Error())
		}
		next := make([]string, 0, total)
		next = append(next, lines[:edit.startLine-1]...)
		next = append(next, lines[edit.endLine:]...)
		return next, edit.startLine, edit.startLine - 1, nil
	default:
		return nil, 0, 0, fmt.Errorf("不支持的 operation: %s（可选 insert/replace/delete）", edit.operation)
	}
}

// validateClosedRange 校验闭区间行号在 [1, total] 内且 start ≤ end
func validateClosedRange(total, start, end int) error {
	if start < 1 || end < start || end > total {
		return fmt.Errorf("行区间无效：需要 1 ≤ start_line ≤ end_line ≤ %d（当前 start=%d, end=%d）", total, start, end)
	}
	return nil
}

// editRegionWindow 计算编辑落点回显窗口（变更主体 ±previewEditContextLines，钳制到有效范围并按上限裁剪）
//
// 空文件返回 (0, 0)；窗口超上限时以「保留变更主体起点之前的上下文」为锚裁剪。
func editRegionWindow(totalLines, changedStart, changedEnd int) (int, int) {
	if totalLines == 0 {
		return 0, 0
	}
	start := changedStart - previewEditContextLines
	if start < 1 {
		start = 1
	}
	end := changedEnd + previewEditContextLines
	if end > totalLines {
		end = totalLines
	}
	if end < start {
		end = start
	}
	if end-start+1 > previewEditRegionMaxLines {
		start = changedStart - previewEditContextLines
		if start < 1 {
			start = 1
		}
		end = start + previewEditRegionMaxLines - 1
		if end > totalLines {
			end = totalLines
		}
	}
	return start, end
}

// formatNumberedLines 将行切片格式化为带行号文本（「%d| %s」，startLine 为首行行号，1 起始）
func formatNumberedLines(startLine int, lines []string) string {
	var builder strings.Builder
	for i, line := range lines {
		if i > 0 {
			builder.WriteByte('\n')
		}
		fmt.Fprintf(&builder, "%d| %s", startLine+i, line)
	}
	return builder.String()
}

// numberedRegion 取行切片的 [start, end] 闭区间并格式化为带行号文本（越界或空集返回空串）
func numberedRegion(lines []string, start, end int) string {
	if len(lines) == 0 || start < 1 || end < start || end > len(lines) {
		return ""
	}
	return formatNumberedLines(start, lines[start-1:end])
}
