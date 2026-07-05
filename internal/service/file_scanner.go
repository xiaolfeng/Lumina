// Package service 提供跨业务领域的通用服务（文件缓存、媒体回答处理等）。
package service

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

// ── 排除目录（不区分大小写）──
//
// 这些目录通常包含构建产物、依赖缓存或 IDE 配置，
// 对代码语义分析无价值且体积庞大，扫描时整体跳过。
var excludedDirs = map[string]bool{
	".git": true, "node_modules": true, "vendor": true,
	"build": true, "dist": true, "__pycache__": true,
	".idea": true, ".vscode": true, "target": true,
	"bin": true, "obj": true,
}

// ── 二进制文件扩展名 ──
//
// 图片、压缩包、可执行文件、字体、数据库等非文本资源，
// 无法进行源码级分析，扫描时直接排除。
var binaryExts = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true,
	".bmp": true, ".ico": true, ".svg": true, ".pdf": true,
	".zip": true, ".tar": true, ".gz": true, ".rar": true,
	".exe": true, ".dll": true, ".so": true, ".dylib": true,
	".class": true, ".jar": true, ".war": true,
	".mp3": true, ".mp4": true, ".avi": true, ".mov": true,
	".ttf": true, ".otf": true, ".woff": true, ".woff2": true,
	".eot": true, ".db": true, ".sqlite": true,
}

// ── 扩展名 → 语言映射 ──
//
// 覆盖主流编程与标记语言，用于 LanguageStats 统计和入口文件检测。
var extToLanguage = map[string]string{
	".go": "Go",
	".py": "Python",
	".ts": "TypeScript", ".tsx": "TypeScript",
	".js": "JavaScript", ".jsx": "JavaScript",
	".java": "Java",
	".rs":   "Rust",
	".c":    "C", ".h": "C",
	".cpp": "C++", ".cc": "C++", ".cxx": "C++", ".hpp": "C++",
	".rb":    "Ruby",
	".php":   "PHP",
	".swift": "Swift",
	".kt":    "Kotlin",
	".scala": "Scala",
	".sh":    "Shell",
	".md":    "Markdown",
	".sql":   "SQL",
	".html":  "HTML",
	".css":   "CSS",
	".json":  "JSON",
	".yaml":  "YAML", ".yml": "YAML",
	".toml": "TOML",
	".xml":  "XML",
}

// ── 入口文件名（文件名精确匹配）──
//
// 涵盖各语言的 main 文件和项目清单文件，
// 用于 EntryPoints 列表构建，帮助 Agent 快速定位项目根。
var entryPointFiles = map[string]bool{
	"main.go": true, "main.py": true, "main.rs": true, "main.java": true,
	"index.ts": true, "index.js": true, "index.tsx": true, "index.jsx": true,
	"app.py": true, "app.js": true,
	"Main.java": true,
	"go.mod":    true, "package.json": true, "Cargo.toml": true,
	"requirements.txt": true, "Gemfile": true, "pom.xml": true,
	"build.gradle": true, "Makefile": true, "CMakeLists.txt": true,
}

// FileScanResult 文件扫描结果
//
// 汇总一次仓库扫描的全部元信息，供后续依赖提取和 Agent 分析阶段消费。
// 所有字段均为值类型，可安全 JSON 序列化传递给 MCP 工具。
type FileScanResult struct {
	Files         []FileInfo     `json:"files"`          // Files 扫描通过的文件列表（已过滤排除目录/二进制/超大文件）
	LanguageStats map[string]int `json:"language_stats"` // LanguageStats 语言分布统计（语言名 → 文件数），未识别语言的文件计入 "" 键
	EntryPoints   []string       `json:"entry_points"`   // EntryPoints 入口文件相对路径列表（按扫描顺序）
	TotalFiles    int            `json:"total_files"`    // TotalFiles 扫描通过的文件总数（等于 len(Files)）
	TotalSize     int64          `json:"total_size"`     // TotalSize 扫描通过文件的总大小（字节）
}

// FileInfo 单个文件元信息
//
// 仅包含路径、大小、语言、入口标记四项元数据，
// 不读取文件内容，保证扫描开销与仓库规模线性相关。
type FileInfo struct {
	Path     string `json:"path"`     // Path 相对仓库根的路径（正斜杠分隔，跨平台一致）
	Size     int64  `json:"size"`     // Size 文件大小（字节）
	Language string `json:"language"` // Language 语言名（如 "Go"、"TypeScript"），未识别为空字符串
	IsEntry  bool   `json:"is_entry"` // IsEntry 是否为入口文件（按文件名匹配 entryPointFiles）
}

// FileScannerService 文件扫描服务
//
// 遍历代码仓库目录树，过滤排除目录、二进制文件和超大文件，
// 并完成语言识别与入口文件检测。仅读取文件元数据（os.FileInfo），
// 不读取文件内容，扫描成本为 O(文件数)。
//
// 设计理念：
//   - 固定规则过滤：v1 版本按内置白/黑名单过滤，不做 .gitignore 解析
//   - 仅元数据扫描：路径、大小、扩展名，不涉及文件内容读取
//   - 不修改仓库文件：只读遍历，扫描前后仓库状态不变
type FileScannerService struct {
	maxFileSize int64 // maxFileSize 单文件大小上限（字节），超过则跳过；默认 bConst.RepoWikiDefaultMaxFileSize
	log         *xLog.LogNamedLogger
}

// NewFileScannerService 创建 FileScannerService 实例
//
// 默认最大文件大小为 bConst.RepoWikiDefaultMaxFileSize（1MB）。
func NewFileScannerService() *FileScannerService {
	return &FileScannerService{
		maxFileSize: bConst.RepoWikiDefaultMaxFileSize,
		log:         xLog.WithName(xLog.NamedCONT, "FileScanner"),
	}
}

// Scan 遍历仓库目录树，返回扫描结果
//
// 使用 filepath.Walk 自顶向下遍历 repoPath：
//  1. 目录检查：任一路径组件命中 excludedDirs（不区分大小写）则整体 SkipDir
//  2. 文件过滤：二进制扩展名 → 跳过；size > maxFileSize → 跳过
//  3. 语言检测：按扩展名查 extToLanguage，未命中则 Language 为空字符串
//  4. 入口检测：按文件名查 entryPointFiles
//  5. 路径归一化：相对路径统一使用正斜杠（'/'），保证跨平台一致
//
// 参数说明:
//   - ctx: 上下文
//   - repoPath: 仓库根目录的绝对或相对路径
//
// 返回值:
//   - *FileScanResult: 扫描结果（Files 按扫描顺序，可能为空切片）
//   - *xError.Error: 路径不存在或 Walk 过程中遇到的 I/O 错误
func (s *FileScannerService) Scan(ctx context.Context, repoPath string) (*FileScanResult, *xError.Error) {
	// 路径存在性校验，给出比 Walk 原始错误更清晰的失败原因
	info, err := os.Stat(repoPath)
	if err != nil {
		return nil, xError.NewError(ctx, xError.ServerInternalError, xError.ErrMessage("仓库路径不可访问: "+repoPath), true, err)
	}
	if !info.IsDir() {
		return nil, xError.NewError(ctx, xError.ServerInternalError, xError.ErrMessage("仓库路径不是目录: "+repoPath), true, nil)
	}

	result := &FileScanResult{
		Files:         make([]FileInfo, 0, 64),
		LanguageStats: make(map[string]int),
		EntryPoints:   make([]string, 0),
	}

	walkErr := filepath.Walk(repoPath, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			// Walk 自身遇到的访问错误（权限/符号链接断裂等），记录后跳过该节点继续
			s.log.Warn(ctx, "扫描节点访问失败，已跳过", slog.String("path", path), slog.String("err", err.Error()))
			return nil
		}

		// 计算相对仓库根的路径；根节点本身 rel 为 "."
		rel, relErr := filepath.Rel(repoPath, path)
		if relErr != nil {
			// 理论上不会失败（path 一定在 repoPath 下），防御性记录
			s.log.Warn(ctx, "相对路径计算失败，已跳过", slog.String("path", path), slog.String("err", relErr.Error()))
			return nil
		}

		// ── 目录处理：检查任一路径组件是否命中排除目录 ──
		if fi.IsDir() {
			if rel == "." {
				return nil // 仓库根目录自身不检查
			}
			if isExcludedDir(rel) {
				return filepath.SkipDir
			}
			return nil
		}

		// ── 文件处理 ──

		// 二进制文件按扩展名排除
		if isBinaryExt(filepath.Ext(path)) {
			return nil
		}
		// 超大文件跳过
		if fi.Size() > s.maxFileSize {
			return nil
		}

		// 路径归一化为正斜杠（Windows 兼容）
		relPath := filepath.ToSlash(rel)
		fileName := filepath.Base(path)
		language := detectLanguage(filepath.Ext(fileName))
		isEntry := isEntryPoint(fileName)

		result.Files = append(result.Files, FileInfo{
			Path:     relPath,
			Size:     fi.Size(),
			Language: language,
			IsEntry:  isEntry,
		})
		result.LanguageStats[language]++
		result.TotalSize += fi.Size()
		if isEntry {
			result.EntryPoints = append(result.EntryPoints, relPath)
		}
		return nil
	})
	if walkErr != nil {
		return nil, xError.NewError(ctx, xError.ServerInternalError, xError.ErrMessage("仓库目录遍历失败: "+repoPath), true, walkErr)
	}

	result.TotalFiles = len(result.Files)
	return result, nil
}

// ── 内部辅助函数 ──

// isExcludedDir 判断相对路径中是否存在任一组件命中 excludedDirs（不区分大小写）。
//
// 例如 rel="src/.git/config" 时，".git" 命中，返回 true。
// 该函数仅在目录节点被调用（rel != "."），但实现上对文件路径同样安全。
func isExcludedDir(rel string) bool {
	rel = filepath.ToSlash(rel)
	for part := range strings.SplitSeq(rel, "/") {
		if excludedDirs[strings.ToLower(part)] {
			return true
		}
	}
	return false
}

// isBinaryExt 判断扩展名是否属于二进制文件（大小写不敏感）。
func isBinaryExt(ext string) bool {
	return binaryExts[strings.ToLower(ext)]
}

// detectLanguage 按扩展名返回语言名，未识别返回空字符串。
func detectLanguage(ext string) string {
	return extToLanguage[strings.ToLower(ext)]
}

// isEntryPoint 判断文件名是否为入口/清单文件。
func isEntryPoint(fileName string) bool {
	return entryPointFiles[fileName]
}
