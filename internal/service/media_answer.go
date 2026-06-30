// Package service 提供跨业务领域的通用服务（文件缓存、媒体回答处理等）。
package service

import (
	"context"
	"log/slog"

	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
)

// MediaAnswerService 媒体回答处理服务
//
// 负责 image/file 题型回答的 base64 解码落盘与回答数据清洗，
// 将 WebSocket 层的文件持久化细节隔离到 service 层，保持 handler 精简。
//
// 设计理念：
//   - ProcessMediaAnswer：image/file 题型遍历 images/files 数组，将 content 字段（base64）
//     解码落盘并替换为 filePath，落盘失败的单个文件跳过并记录日志，不影响其他文件
//   - SanitizeAnswer：剥离 image/file 中体积过大的 base64 content/preview 字段，
//     持久化到 DB 只需保留 filename + mimeType + filePath
type MediaAnswerService struct {
	fileCache *FileCacheService    // 文件缓存服务（base64 解码 + 磁盘写入）
	log       *xLog.LogNamedLogger // 专用日志记录器
}

// NewMediaAnswerService 创建 MediaAnswerService 实例
//
// 内部持有 FileCacheService 单例和专用日志记录器，调用方无需额外注入。
func NewMediaAnswerService() *MediaAnswerService {
	return &MediaAnswerService{
		fileCache: NewFileCacheService(),
		log:       xLog.WithName(xLog.NamedCONT, "MediaAnswerSvc"),
	}
}

// ProcessMediaAnswer 处理 image/file 类型回答，将 base64 content 解码写入文件系统并替换为 filePath。
//
// image 题型遍历 images 数组，file 题型遍历 files 数组。
// 每个元素的 content 字段（base64）被解码落盘，然后替换为 filePath 字段。
// 落盘失败的文件跳过并记录日志，不影响其他文件处理。
// 返回更新后的 answerData；若 answerData 不是预期 map 结构则原样返回。
//
// 参数说明:
//   - ctx:       上下文
//   - sessionID: 会话 ID（用于构建文件存储路径，作为目录隔离单位）
//   - answerData: 原始回答数据（期望为 map[string]interface{}）
//   - qType:     问题类型（"image" 或 "file"）
//
// 返回值:
//   - any: 处理后的回答数据（content 字段已替换为 filePath）
func (s *MediaAnswerService) ProcessMediaAnswer(ctx context.Context, sessionID string, answerData any, qType string) any {
	m, ok := answerData.(map[string]interface{})
	if !ok {
		return answerData
	}

	fieldKey := "images"
	if qType == "file" {
		fieldKey = "files"
	}

	items, ok := m[fieldKey].([]interface{})
	if !ok {
		return answerData
	}

	for i, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		content, hasContent := itemMap["content"].(string)
		if !hasContent || content == "" {
			continue
		}

		filename, _ := itemMap["filename"].(string)
		if filename == "" {
			filename, _ = itemMap["name"].(string)
		}
		mimeType, _ := itemMap["mimeType"].(string)

		filePath, xErr := s.fileCache.SaveBase64File(ctx, sessionID, content, filename, mimeType)
		if xErr != nil {
			s.log.Warn(ctx, "文件保存失败",
				slog.String("filename", filename),
				slog.String("error", xErr.Error()),
			)
			continue
		}

		// 移除 base64 原始数据，写入磁盘路径
		delete(itemMap, "content")
		itemMap["filePath"] = filePath
		items[i] = itemMap
	}

	m[fieldKey] = items
	return m
}

// SanitizeAnswer 清洗回答数据，剥离 image/file 中体积过大的 base64 content/preview 字段。
//
// 图片/文件题的回答结构为 { images: [{ filename, mimeType, content }] }，
// 其中 content 是 base64 编码的完整文件数据。持久化到 DB 只需保留 filename + mimeType，
// content 会通过独立的文件存储或 Media 字段处理，不应留在 answer JSON 中。
//
// 参数说明:
//   - answer: 原始回答数据（期望为 map[string]interface{}）
//
// 返回值:
//   - any: 清洗后的回答数据（content/preview 字段已移除）
func (s *MediaAnswerService) SanitizeAnswer(answer any) any {
	m, ok := answer.(map[string]interface{})
	if !ok {
		return answer
	}

	for _, key := range []string{"images", "files"} {
		if items, ok := m[key].([]interface{}); ok {
			cleaned := make([]interface{}, 0, len(items))
			for _, item := range items {
				if itemMap, ok := item.(map[string]interface{}); ok {
					delete(itemMap, "content")
					delete(itemMap, "preview")
					cleaned = append(cleaned, itemMap)
				}
			}
			m[key] = cleaned
		}
	}

	return m
}
