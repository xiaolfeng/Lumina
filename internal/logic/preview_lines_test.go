package logic

import (
	"strings"
	"testing"

	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

// TestSplitFileLines 锁定拆行契约：CRLF/CR 归一、尾随换行是终止符而非空行、空文件为 0 行
func TestSplitFileLines(t *testing.T) {
	cases := []struct {
		name     string
		content  string
		lines    []string
		trailing bool
	}{
		{"空文件", "", []string{}, false},
		{"单个换行是 1 个空行", "\n", []string{""}, true},
		{"无尾换行单行", "a", []string{"a"}, false},
		{"尾换行单行", "a\n", []string{"a"}, true},
		{"两行带尾换行", "a\nb\n", []string{"a", "b"}, true},
		{"两行无尾换行", "a\nb", []string{"a", "b"}, false},
		{"尾随换行后的空行是独立行", "a\n\n", []string{"a", ""}, true},
		{"CRLF 归一", "a\r\nb\r\n", []string{"a", "b"}, true},
		{"孤立 CR 归一", "a\rb", []string{"a", "b"}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := splitFileLines(c.content)
			if len(got.lines) != len(c.lines) {
				t.Fatalf("lines = %v, want %v", got.lines, c.lines)
			}
			for i := range c.lines {
				if got.lines[i] != c.lines[i] {
					t.Fatalf("lines = %v, want %v", got.lines, c.lines)
				}
			}
			if got.trailingNewline != c.trailing {
				t.Fatalf("trailingNewline = %v, want %v", got.trailingNewline, c.trailing)
			}
		})
	}
}

// TestJoinFileLines 锁定合并契约：尾换行按原文件保留、空行集返回空串
func TestJoinFileLines(t *testing.T) {
	cases := []struct {
		name            string
		lines           []string
		trailingNewline bool
		want            string
	}{
		{"空行集无尾换行", []string{}, false, ""},
		{"空行集带尾换行仍为空", []string{}, true, ""},
		{"单行补尾换行", []string{"a"}, true, "a\n"},
		{"单行无尾换行", []string{"a"}, false, "a"},
		{"多行连接", []string{"a", "b"}, true, "a\nb\n"},
		{"保留空行", []string{"a", ""}, true, "a\n\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := joinFileLines(c.lines, c.trailingNewline); got != c.want {
				t.Fatalf("joinFileLines = %q, want %q", got, c.want)
			}
		})
	}
}

// TestApplyLineEditInsert 锁定 insert 语义：前插、末尾追加（startLine=0 与 total+1 等价）、边界拒绝
func TestApplyLineEditInsert(t *testing.T) {
	lines := []string{"a", "b", "c"}

	cases := []struct {
		name      string
		startLine int
		content   string
		want      []string
		changedS  int
		changedE  int
	}{
		{"头部前插", 1, "x", []string{"x", "a", "b", "c"}, 1, 1},
		{"中间前插多行", 2, "x\ny", []string{"a", "x", "y", "b", "c"}, 2, 3},
		{"省略 start_line 追加到末尾", 0, "x\ny", []string{"a", "b", "c", "x", "y"}, 4, 5},
		{"start_line=total+1 追加到末尾", 4, "x", []string{"a", "b", "c", "x"}, 4, 4},
		{"content 尾换行是终止符", 2, "x\n", []string{"a", "x", "b", "c"}, 2, 2},
		{"content CRLF 归一", 2, "x\r\ny", []string{"a", "x", "y", "b", "c"}, 2, 3},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, changedS, changedE, err := applyLineEdit(lines, previewLineEdit{
				operation: bConst.PreviewEditOperationInsert,
				startLine: c.startLine,
				content:   c.content,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if strings.Join(got, "|") != strings.Join(c.want, "|") {
				t.Fatalf("lines = %v, want %v", got, c.want)
			}
			if changedS != c.changedS || changedE != c.changedE {
				t.Fatalf("changed = [%d,%d], want [%d,%d]", changedS, changedE, c.changedS, c.changedE)
			}
		})
	}

	errCases := []struct {
		name      string
		startLine int
		content   string
		lines     []string
	}{
		{"content 为空", 1, "", lines},
		{"start_line 超出 total+1", 5, "x", lines},
		{"start_line 为负数", -1, "x", lines},
		{"空文件插入 start_line=2", 2, "x", []string{}},
	}
	for _, c := range errCases {
		t.Run(c.name, func(t *testing.T) {
			if _, _, _, err := applyLineEdit(c.lines, previewLineEdit{
				operation: bConst.PreviewEditOperationInsert,
				startLine: c.startLine,
				content:   c.content,
			}); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}

	t.Run("空文件省略 start_line 建立内容", func(t *testing.T) {
		got, changedS, changedE, err := applyLineEdit([]string{}, previewLineEdit{
			operation: bConst.PreviewEditOperationInsert,
			content:   "a\nb",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Join(got, "|") != "a|b" || changedS != 1 || changedE != 2 {
			t.Fatalf("got = %v [%d,%d]", got, changedS, changedE)
		}
	})
}

// TestApplyLineEditReplace 锁定 replace 语义：区间替换、空 content 等价删除、越界拒绝
func TestApplyLineEditReplace(t *testing.T) {
	lines := []string{"a", "b", "c", "d"}

	cases := []struct {
		name      string
		startLine int
		endLine   int
		content   string
		want      []string
		changedS  int
		changedE  int
	}{
		{"替换单行", 2, 2, "x", []string{"a", "x", "c", "d"}, 2, 2},
		{"多行换多行", 2, 3, "x\ny\nz", []string{"a", "x", "y", "z", "d"}, 2, 4},
		{"多行换单行", 1, 3, "x", []string{"x", "d"}, 1, 1},
		{"空 content 等价删除", 2, 3, "", []string{"a", "d"}, 2, 1},
		{"替换首行", 1, 1, "x", []string{"x", "b", "c", "d"}, 1, 1},
		{"替换末行", 4, 4, "x", []string{"a", "b", "c", "x"}, 4, 4},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, changedS, changedE, err := applyLineEdit(lines, previewLineEdit{
				operation: bConst.PreviewEditOperationReplace,
				startLine: c.startLine,
				endLine:   c.endLine,
				content:   c.content,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if strings.Join(got, "|") != strings.Join(c.want, "|") {
				t.Fatalf("lines = %v, want %v", got, c.want)
			}
			if changedS != c.changedS || changedE != c.changedE {
				t.Fatalf("changed = [%d,%d], want [%d,%d]", changedS, changedE, c.changedS, c.changedE)
			}
		})
	}

	errCases := []struct {
		name      string
		startLine int
		endLine   int
	}{
		{"start_line 为 0", 0, 2},
		{"end_line 小于 start_line", 3, 2},
		{"end_line 超出总行数", 2, 5},
	}
	for _, c := range errCases {
		t.Run(c.name, func(t *testing.T) {
			if _, _, _, err := applyTextEditErr(t, lines, bConst.PreviewEditOperationReplace, c.startLine, c.endLine, "x"); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

// TestApplyLineEditDelete 锁定 delete 语义：区间删除、整文件清空、越界拒绝
func TestApplyLineEditDelete(t *testing.T) {
	lines := []string{"a", "b", "c", "d"}

	cases := []struct {
		name      string
		startLine int
		endLine   int
		want      []string
		changedS  int
		changedE  int
	}{
		{"删除单行", 2, 2, []string{"a", "c", "d"}, 2, 1},
		{"删除区间", 2, 3, []string{"a", "d"}, 2, 1},
		{"删除首行", 1, 1, []string{"b", "c", "d"}, 1, 0},
		{"删除末行", 4, 4, []string{"a", "b", "c"}, 4, 3},
		{"删除全部", 1, 4, []string{}, 1, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, changedS, changedE, err := applyLineEdit(lines, previewLineEdit{
				operation: bConst.PreviewEditOperationDelete,
				startLine: c.startLine,
				endLine:   c.endLine,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if strings.Join(got, "|") != strings.Join(c.want, "|") {
				t.Fatalf("lines = %v, want %v", got, c.want)
			}
			if changedS != c.changedS || changedE != c.changedE {
				t.Fatalf("changed = [%d,%d], want [%d,%d]", changedS, changedE, c.changedS, c.changedE)
			}
		})
	}

	if _, _, _, err := applyTextEditErr(t, lines, bConst.PreviewEditOperationDelete, 3, 2, ""); err == nil {
		t.Fatal("delete 区间倒置应报错")
	}
	if _, _, _, err := applyTextEditErr(t, []string{}, bConst.PreviewEditOperationDelete, 1, 1, ""); err == nil {
		t.Fatal("空文件删除应报错")
	}
	if _, _, _, err := applyTextEditErr(t, lines, "append", 1, 1, "x"); err == nil {
		t.Fatal("未知 operation 应报错")
	}
}

// applyTextEditErr 以指定参数调用 applyLineEdit 并透传错误，供错误用例复用
func applyTextEditErr(t *testing.T, lines []string, operation string, startLine, endLine int, content string) ([]string, int, int, error) {
	t.Helper()
	return applyLineEdit(lines, previewLineEdit{
		operation: operation,
		startLine: startLine,
		endLine:   endLine,
		content:   content,
	})
}

// TestEditRegionWindow 锁定回显窗口：±3 行上下文、钳制边界、40 行上限、空文件零区间
func TestEditRegionWindow(t *testing.T) {
	cases := []struct {
		name       string
		totalLines int
		changedS   int
		changedE   int
		wantStart  int
		wantEnd    int
	}{
		{"空文件", 0, 1, 0, 0, 0},
		{"头部编辑钳制下界", 5, 1, 1, 1, 4},
		{"中部编辑带上下文", 10, 5, 6, 2, 9},
		{"尾部编辑钳制上界", 5, 5, 5, 2, 5},
		{"上限裁剪保留前上下文", 100, 10, 90, 7, 7 + previewEditRegionMaxLines - 1},
		{"小文件全窗", 3, 2, 2, 1, 3},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			start, end := editRegionWindow(c.totalLines, c.changedS, c.changedE)
			if start != c.wantStart || end != c.wantEnd {
				t.Fatalf("window = [%d,%d], want [%d,%d]", start, end, c.wantStart, c.wantEnd)
			}
			if c.totalLines > 0 && (start < 1 || end > c.totalLines || end < start) {
				t.Fatalf("window [%d,%d] 越界（total=%d）", start, end, c.totalLines)
			}
		})
	}
}

// TestFormatNumberedLines 锁定「行号| 文本」格式
func TestFormatNumberedLines(t *testing.T) {
	if got := formatNumberedLines(1, []string{"a", "b"}); got != "1| a\n2| b" {
		t.Fatalf("formatNumberedLines = %q", got)
	}
	if got := formatNumberedLines(41, []string{"body { margin: 0; }"}); got != "41| body { margin: 0; }" {
		t.Fatalf("formatNumberedLines = %q", got)
	}
	if got := formatNumberedLines(1, nil); got != "" {
		t.Fatalf("空行集应返回空串，got %q", got)
	}
}

// TestNumberedRegion 锁定区间切片与越界防护
func TestNumberedRegion(t *testing.T) {
	lines := []string{"a", "b", "c", "d"}
	if got := numberedRegion(lines, 2, 3); got != "2| b\n3| c" {
		t.Fatalf("numberedRegion = %q", got)
	}
	if got := numberedRegion(lines, 1, 4); !strings.HasPrefix(got, "1| a") || !strings.HasSuffix(got, "4| d") {
		t.Fatalf("numberedRegion = %q", got)
	}
	for _, c := range []struct {
		name  string
		lines []string
		start int
		end   int
	}{{"空行集", nil, 1, 1}, {"start 越界", lines, 0, 2}, {"end 越界", lines, 2, 9}, {"空区间", lines, 3, 2}} {
		t.Run(c.name, func(t *testing.T) {
			if got := numberedRegion(c.lines, c.start, c.end); got != "" {
				t.Fatalf("越界应返回空串，got %q", got)
			}
		})
	}
}
