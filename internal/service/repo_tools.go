// Package service RepoWiki Agent 只读工具集。
//
// 本文件提供 file_read 和 file_search 两个工具，用于 Agent 在仓库路径范围内
// 安全地读取文件和搜索文件。所有路径访问均限制在 repoPath 目录内，
// 防止路径遍历攻击。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bamboo-services/bamboo-agent/tool"
)

const (
	// fileReadMaxChars 是 file_read 工具返回内容的最大字符数，
	// 超过部分会被截断并追加 "...[truncated]" 标记。
	fileReadMaxChars = 100 * 1024
	// fileSearchMaxResults 是 file_search 工具返回的最大文件数量。
	fileSearchMaxResults = 50
)

// ──────────────────────────────────────────────────────────────────────
// file_read 工具
// ──────────────────────────────────────────────────────────────────────

// NewFileReadTool 创建限定在 repoPath 下的 file_read 工具。
//
// 参数说明:
//   - repoPath: 仓库根目录绝对路径，工具只能读取该目录下的文件
func NewFileReadTool(repoPath string) tool.Tool {
	return &fileReadTool{repoPath: repoPath}
}

// fileReadTool file_read 工具实现
type fileReadTool struct {
	repoPath string // 仓库根目录路径
}

// Info 返回工具元信息
func (t *fileReadTool) Info() tool.ToolInfo {
	return tool.ToolInfo{
		Name:        "file_read",
		Description: "读取仓库中指定文件的内容，返回文件文本。路径相对于仓库根目录。",
		Parameters: tool.InputSchema{
			Type: "object",
			Properties: map[string]tool.PropertyDef{
				"path": {
					Type:        "string",
					Description: "相对于仓库根目录的文件路径，例如 README.md 或 src/main.go",
				},
			},
			Required: []string{"path"},
		},
		MaxResultSizeChars: fileReadMaxChars,
	}
}

// Execute 执行文件读取
func (t *fileReadTool) Execute(ctx context.Context, input json.RawMessage) (*tool.ToolResult, error) {
	var args struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(input, &args); err != nil {
		return &tool.ToolResult{Content: fmt.Sprintf("参数解析失败: %v", err), IsError: true}, nil
	}

	if strings.TrimSpace(args.Path) == "" {
		return &tool.ToolResult{Content: "path 不能为空", IsError: true}, nil
	}

	safePath, err := resolveSafeRepoPath(t.repoPath, args.Path)
	if err != nil {
		return &tool.ToolResult{Content: err.Error(), IsError: true}, nil
	}

	data, err := os.ReadFile(safePath)
	if err != nil {
		return &tool.ToolResult{Content: fmt.Sprintf("读取文件失败: %v", err), IsError: true}, nil
	}

	content := string(data)
	if len(content) > fileReadMaxChars {
		content = content[:fileReadMaxChars] + "...[truncated]"
	}

	return &tool.ToolResult{Content: content, IsError: false}, nil
}

// ──────────────────────────────────────────────────────────────────────
// file_search 工具
// ──────────────────────────────────────────────────────────────────────

// NewFileSearchTool 创建限定在 repoPath 下的 file_search 工具。
//
// 参数说明:
//   - repoPath: 仓库根目录绝对路径，工具只在该目录下搜索文件
func NewFileSearchTool(repoPath string) tool.Tool {
	return &fileSearchTool{repoPath: repoPath}
}

// fileSearchTool file_search 工具实现
type fileSearchTool struct {
	repoPath string // 仓库根目录路径
}

// Info 返回工具元信息
func (t *fileSearchTool) Info() tool.ToolInfo {
	return tool.ToolInfo{
		Name:        "file_search",
		Description: "在仓库中搜索文件，返回匹配的文件路径列表（相对于仓库根目录）。支持按文件名子串匹配，忽略大小写。",
		Parameters: tool.InputSchema{
			Type: "object",
			Properties: map[string]tool.PropertyDef{
				"pattern": {
					Type:        "string",
					Description: "要匹配的文件名关键词（支持子串匹配，忽略大小写），例如 main.go 或 README",
				},
			},
			Required: []string{"pattern"},
		},
	}
}

// Execute 执行文件搜索
func (t *fileSearchTool) Execute(ctx context.Context, input json.RawMessage) (*tool.ToolResult, error) {
	var args struct {
		Pattern string `json:"pattern"`
	}
	if err := json.Unmarshal(input, &args); err != nil {
		return &tool.ToolResult{Content: fmt.Sprintf("参数解析失败: %v", err), IsError: true}, nil
	}

	pattern := strings.ToLower(strings.TrimSpace(args.Pattern))
	if pattern == "" {
		return &tool.ToolResult{Content: "pattern 不能为空", IsError: true}, nil
	}

	var matches []string
	err := filepath.Walk(t.repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// 单个文件错误不影响整体搜索，记录并跳过
			return nil
		}
		if info.IsDir() {
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		if strings.Contains(strings.ToLower(info.Name()), pattern) {
			rel, relErr := filepath.Rel(t.repoPath, path)
			if relErr == nil {
				matches = append(matches, rel)
			}
		}
		return nil
	})
	if err != nil {
		return &tool.ToolResult{Content: fmt.Sprintf("搜索失败: %v", err), IsError: true}, nil
	}

	if len(matches) > fileSearchMaxResults {
		matches = matches[:fileSearchMaxResults]
	}

	resultJSON, _ := json.Marshal(matches)
	return &tool.ToolResult{Content: string(resultJSON), IsError: false}, nil
}

// ──────────────────────────────────────────────────────────────────────
// 路径安全辅助函数
// ──────────────────────────────────────────────────────────────────────

// resolveSafeRepoPath 将相对路径解析为仓库内的绝对安全路径，防止路径遍历。
//
// 规则:
//   - 清理路径中的 .、.. 等元素
//   - 拒绝绝对路径输入
//   - 解析后必须位于 repoPath 目录内
//
// 返回:
//   - string: 安全的绝对路径
//   - error: 路径非法或解析失败
func resolveSafeRepoPath(repoPath, relPath string) (string, error) {
	// 清理输入路径，移除 ../ 等潜在遍历元素
	cleanRel := filepath.Clean(relPath)
	if filepath.IsAbs(cleanRel) {
		return "", fmt.Errorf("路径不能为绝对路径: %s", relPath)
	}

	fullPath := filepath.Join(repoPath, cleanRel)

	absRepo, err := filepath.Abs(repoPath)
	if err != nil {
		return "", fmt.Errorf("无法解析仓库路径: %v", err)
	}
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("无法解析文件路径: %v", err)
	}

	// 确保 absPath 在 absRepo 内部（允许 absPath == absRepo 仅当 relPath 为 "."）
	if absPath != absRepo {
		if !strings.HasPrefix(absPath, absRepo+string(filepath.Separator)) {
			return "", fmt.Errorf("路径越界: %s", relPath)
		}
	}

	return absPath, nil
}
