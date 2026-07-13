package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xEnv "github.com/bamboo-services/bamboo-base-go/defined/env"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

// ──────────────────────────────────────────────────────────────
// WikiStorageService
// ──────────────────────────────────────────────────────────────

// WikiStorageService RepoWiki 文件系统存储管理服务
//
// 负责管理 RepoWiki 生成过程中的所有文件系统操作，包括：
//   - 路径布局管理（仓库克隆、版本数据、最终 Wiki 文档）
//   - JSON / Markdown 文件读写
//   - 目录与版本清理
//   - 路径遍历安全防护（SanitizePath）
//
// 文件系统布局：
//
//	{basePath}/
//	├── repos/{configID}/              # 克隆的 Git 仓库
//	├── versions/{versionID}/          # 版本分析数据
//	│   ├── raw/                       # 原始仓库副本
//	│   ├── passes/                    # Agent 分析 Pass 产出
//	│   ├── sessions/                  # Agent FileStore 会话目录
//	│   ├── file_scan.json             # 文件扫描结果
//	│   └── dep_summary.json           # 依赖摘要
//	└── wiki/{projectID}/zh/           # 最终 Wiki 文档
//	    └── meta/repowiki-metadata.json
//
// basePath 来自环境变量 REPOWIKI_STORAGE_PATH，默认 ./lumina/repowiki
type WikiStorageService struct {
	basePath string // 存储根目录（REPOWIKI_STORAGE_PATH）
}

// NewWikiStorageService 创建 WikiStorageService 实例
//
// 从环境变量 REPOWIKI_STORAGE_PATH 读取存储根目录，默认值 ./lumina/repowiki
func NewWikiStorageService() *WikiStorageService {
	return &WikiStorageService{
		basePath: xEnv.GetEnvString("REPOWIKI_STORAGE_PATH", "./.lumina/repowiki"),
	}
}

// ── 路径方法（纯计算，不创建目录）──

// GetRepoPath 返回仓库克隆路径
//
// → {basePath}/repos/{configID}/
func (s *WikiStorageService) GetRepoPath(configID int64) string {
	return filepath.Join(s.basePath, "repos", fmt.Sprintf("%d", configID))
}

// GetVersionPath 返回版本数据根路径
//
// → {basePath}/versions/{versionID}/
func (s *WikiStorageService) GetVersionPath(versionID int64) string {
	return filepath.Join(s.basePath, "versions", fmt.Sprintf("%d", versionID))
}

// GetRawPath 返回版本原始仓库副本路径
//
// → {versionPath}/raw/
func (s *WikiStorageService) GetRawPath(versionID int64) string {
	return filepath.Join(s.GetVersionPath(versionID), "raw")
}

// GetPassesPath 返回 Agent 分析 Pass 产出路径
//
// → {versionPath}/passes/
func (s *WikiStorageService) GetPassesPath(versionID int64) string {
	return filepath.Join(s.GetVersionPath(versionID), "passes")
}

// GetWikiPath 返回最终 Wiki 文档路径
//
// → {basePath}/wiki/{configID}/{language}/
//
// configID 为 RepoWiki 配置雪花 ID（非 Project ID、非 Version ID）。
// language 取自 bConst.RepoWikiDefaultLanguage（默认 zh）。
func (s *WikiStorageService) GetWikiPath(configID int64) string {
	return filepath.Join(s.basePath, "wiki", fmt.Sprintf("%d", configID), bConst.RepoWikiDefaultLanguage)
}

// GetFileScanPath 返回文件扫描结果路径
//
// → {versionPath}/file_scan.json
func (s *WikiStorageService) GetFileScanPath(versionID int64) string {
	return filepath.Join(s.GetVersionPath(versionID), "file_scan.json")
}

// GetDepSummaryPath 返回依赖摘要路径
//
// → {versionPath}/dep_summary.json
func (s *WikiStorageService) GetDepSummaryPath(versionID int64) string {
	return filepath.Join(s.GetVersionPath(versionID), "dep_summary.json")
}

// GetManifestPath 返回 Wiki 元数据清单路径
//
// → {wikiPath}/meta/repowiki-metadata.json
//
// configID 为 RepoWiki 配置雪花 ID。
func (s *WikiStorageService) GetManifestPath(configID int64) string {
	return filepath.Join(s.GetWikiPath(configID), "meta", "repowiki-metadata.json")
}

// GetSessionPath 返回 Agent FileStore 会话目录路径
//
// → {versionPath}/sessions/
func (s *WikiStorageService) GetSessionPath(versionID int64) string {
	return filepath.Join(s.GetVersionPath(versionID), "sessions")
}

// ── 文件 I/O ──

// WriteJSON 将数据序列化为 JSON（2 空格缩进）并写入指定路径
//
// 自动创建缺失的父目录。路径不做 SanitizePath 校验（调用方使用内部路径方法时已安全）。
func (s *WikiStorageService) WriteJSON(path string, data interface{}) *xError.Error {
	// 确保父目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return xError.NewError(context.Background(), xError.ServerInternalError,
			xError.ErrMessage("创建目录失败 "+dir), false, err)
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return xError.NewError(context.Background(), xError.ServerInternalError,
			xError.ErrMessage("JSON 序列化失败"), false, err)
	}

	if err := os.WriteFile(path, jsonBytes, 0644); err != nil {
		return xError.NewError(context.Background(), xError.ServerInternalError,
			xError.ErrMessage("写入 JSON 文件失败 "+path), false, err)
	}
	return nil
}

// ReadJSON 从指定路径读取 JSON 文件并反序列化到 target
func (s *WikiStorageService) ReadJSON(path string, target interface{}) *xError.Error {
	data, err := os.ReadFile(path)
	if err != nil {
		return xError.NewError(context.Background(), xError.FileNotFound,
			xError.ErrMessage("读取 JSON 文件失败 "+path), false, err)
	}

	if err := json.Unmarshal(data, target); err != nil {
		return xError.NewError(context.Background(), xError.ServerInternalError,
			xError.ErrMessage("JSON 反序列化失败 "+path), false, err)
	}
	return nil
}

// WriteMarkdown 将 Markdown 内容写入指定路径
//
// 自动创建缺失的父目录
func (s *WikiStorageService) WriteMarkdown(path string, content string) *xError.Error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return xError.NewError(context.Background(), xError.ServerInternalError,
			xError.ErrMessage("创建目录失败 "+dir), false, err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return xError.NewError(context.Background(), xError.ServerInternalError,
			xError.ErrMessage("写入 Markdown 文件失败 "+path), false, err)
	}
	return nil
}

// ReadMarkdown 从指定路径读取 Markdown 文件内容
func (s *WikiStorageService) ReadMarkdown(path string) (string, *xError.Error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", xError.NewError(context.Background(), xError.FileNotFound,
			xError.ErrMessage("读取 Markdown 文件失败 "+path), false, err)
	}
	return string(data), nil
}

// ── 清理 ──

// CleanVersion 清理指定版本的所有文件系统数据
//
// 删除 {basePath}/versions/{versionID}/ 目录（含 raw/passes/sessions/ 等全部子目录）。
// 目录不存在时静默返回 nil（幂等）。
func (s *WikiStorageService) CleanVersion(versionID int64) *xError.Error {
	versionPath := s.GetVersionPath(versionID)
	if err := os.RemoveAll(versionPath); err != nil {
		return xError.NewError(nil, xError.FileDeleteError,
			xError.ErrMessage("清理版本目录失败 "+versionPath), false, err)
	}
	return nil
}

// CleanRepo 清理指定配置对应的克隆仓库
//
// 删除 {basePath}/repos/{configID}/ 目录。
// 目录不存在时静默返回 nil（幂等）。
func (s *WikiStorageService) CleanRepo(configID int64) *xError.Error {
	repoPath := s.GetRepoPath(configID)
	if err := os.RemoveAll(repoPath); err != nil {
		return xError.NewError(nil, xError.FileDeleteError,
			xError.ErrMessage("清理仓库目录失败 "+repoPath), false, err)
	}
	return nil
}

// ── 安全 ──

// SanitizePath 对用户提供的路径进行安全校验，防止路径遍历攻击
//
// 校验流程：
//  1. filepath.Clean 清理路径中的冗余分隔符和 `.`
//  2. 检测 `..` 组件 — 清理后仍残留 `..` 说明尝试向上跳出，直接拒绝
//  3. 相对路径拼接到 basePath 之下
//  4. 通过 filepath.Rel 确认最终路径在 basePath 范围内
//
// 返回值:
//   - string: 经过安全校验的绝对路径
//   - error:  路径包含遍历或超出 basePath 范围时返回错误
func (s *WikiStorageService) SanitizePath(path string) (string, *xError.Error) {
	cleaned := filepath.Clean(path)

	if strings.Contains(cleaned, "..") {
		return "", xError.NewError(context.Background(), xError.ServerInternalError,
			xError.ErrMessage("检测到路径遍历: "+path), false, nil)
	}

	if !filepath.IsAbs(cleaned) {
		cleaned = filepath.Join(s.basePath, cleaned)
	}

	rel, err := filepath.Rel(s.basePath, cleaned)
	if err != nil {
		return "", xError.NewError(context.Background(), xError.ServerInternalError,
			xError.ErrMessage("路径超出存储根目录: "+path), false, err)
	}
	if strings.HasPrefix(rel, "..") {
		return "", xError.NewError(context.Background(), xError.ServerInternalError,
			xError.ErrMessage("路径超出存储根目录: "+path), false, nil)
	}

	return cleaned, nil
}

// ── 初始化 ──

// EnsureDir 确保指定路径的目录存在，递归创建（幂等）
func (s *WikiStorageService) EnsureDir(path string) *xError.Error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return xError.NewError(context.Background(), xError.ServerInternalError,
			xError.ErrMessage("创建目录失败: "+path), false, err)
	}
	return nil
}
