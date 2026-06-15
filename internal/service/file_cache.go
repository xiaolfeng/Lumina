// Package service 提供跨业务领域的通用服务（文件缓存等）。
package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	xEnv "github.com/bamboo-services/bamboo-base-go/defined/env"
	xUtil "github.com/bamboo-services/bamboo-base-go/common/utility"
)

// FileCacheService 文件缓存服务
// 负责 base64 解码后写入文件系统，存储路径为 <LUMINA_CACHE_DIR>/<session_id>/<file_uuid>
//
// 设计理念：
//   - 每个 Q&A 会话维护独立的缓存目录，便于会话结束时统一清理
//   - 文件名使用 cs_ 前缀的安全密钥，避免冲突且不可猜测
//   - 磁盘操作失败优雅返回 error，绝不 panic
type FileCacheService struct{}

// NewFileCacheService 创建 FileCacheService 实例
func NewFileCacheService() *FileCacheService {
	return &FileCacheService{}
}

// getCacheBaseDir 获取文件缓存根目录
//
// 从环境变量 LUMINA_CACHE_DIR 读取，默认 .lumina/cache
func getCacheBaseDir() string {
	return xEnv.GetEnvString("LUMINA_CACHE_DIR", ".lumina/cache")
}

// SaveBase64File 将 base64 编码的文件内容解码后写入磁盘
//
// 参数:
//   - ctx:           上下文
//   - sessionID:     会话 ID（用于构建存储路径，作为目录隔离单位）
//   - base64Content: base64 编码的文件内容
//   - filename:      原始文件名（仅用于日志，不参与存储路径构建）
//   - mimeType:      MIME 类型（仅记录，不参与存储路径构建）
//
// 返回值:
//   - filePath: 完整文件路径（<LUMINA_CACHE_DIR>/<session_id>/<file_uuid>）
//   - err:      写入失败时的错误
func (s *FileCacheService) SaveBase64File(
	ctx context.Context,
	sessionID, base64Content, filename, mimeType string,
) (string, error) {
	// 解码 base64
	data, err := base64.StdEncoding.DecodeString(base64Content)
	if err != nil {
		return "", fmt.Errorf("base64 解码失败: %w", err)
	}

	// 生成文件 UUID（cs_ + 32 位 hex），去掉 cs_ 前缀作为磁盘文件名
	fileUUID := xUtil.Security().GenerateKey()
	fileName := strings.TrimPrefix(fileUUID, "cs_")

	// 构建存储路径：<baseDir>/<sessionID>/<fileName>
	baseDir := getCacheBaseDir()
	dirPath := filepath.Join(baseDir, sessionID)
	filePath := filepath.Join(dirPath, fileName)

	// 创建会话级缓存目录（已存在则幂等）
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", fmt.Errorf("创建缓存目录失败: %w", err)
	}

	// 写入文件（0644: owner 读写, group/others 只读）
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", fmt.Errorf("写入缓存文件失败: %w", err)
	}

	return filePath, nil
}

// ReadFile 读取缓存文件内容
//
// 调用方负责关闭返回的 io.ReadCloser
func (s *FileCacheService) ReadFile(ctx context.Context, filePath string) (io.ReadCloser, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开缓存文件失败: %w", err)
	}
	return file, nil
}

// CleanSession 清理指定会话的所有缓存文件
//
// 删除 <LUMINA_CACHE_DIR>/<session_id> 目录及其下所有文件。
// 目录不存在时静默返回 nil（幂等）。
func (s *FileCacheService) CleanSession(ctx context.Context, sessionID string) error {
	baseDir := getCacheBaseDir()
	dirPath := filepath.Join(baseDir, sessionID)

	if err := os.RemoveAll(dirPath); err != nil {
		return fmt.Errorf("清理会话缓存失败: %w", err)
	}
	return nil
}
