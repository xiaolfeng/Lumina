package logic

import (
	"context"
	"io"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	"github.com/xiaolfeng/Lumina/internal/service"
)

// ConsumeDownloadToken 消费一次性下载令牌，返回令牌关联的文件信息
//
// 令牌通过 Redis Lua 脚本原子 GET+DEL 消费，保证一次性使用。
// 令牌无效、已过期或已使用时返回 NotExist 错误。
//
// 参数说明:
//   - ctx:   上下文
//   - token: 待消费的下载令牌
//
// 返回值:
//   - *service.DownloadFileInfo: 令牌关联的文件信息（路径/文件名/MIME）
//   - *xError.Error: 令牌无效或数据异常时返回错误
func (l *QaLogic) ConsumeDownloadToken(ctx context.Context, token string) (*service.DownloadFileInfo, *xError.Error) {
	l.log.Info(ctx, "ConsumeDownloadToken - 消费下载令牌")

	info, xErr := l.repo.downloadToken.ConsumeToken(ctx, token)
	if xErr != nil {
		return nil, xErr
	}

	return info, nil
}

// OpenDownloadFile 打开待下载的缓存文件，返回可读流
//
// 调用方负责关闭返回的 io.ReadCloser。
// 文件不存在或打开失败时返回 FileReadError 错误。
//
// 参数说明:
//   - ctx:      上下文
//   - filePath: 服务器文件系统绝对路径
//
// 返回值:
//   - io.ReadCloser: 文件读取流
//   - *xError.Error: 文件打开失败时返回错误
func (l *QaLogic) OpenDownloadFile(ctx context.Context, filePath string) (io.ReadCloser, *xError.Error) {
	l.log.Info(ctx, "OpenDownloadFile - 打开下载文件")

	reader, xErr := l.repo.fileCache.ReadFile(ctx, filePath)
	if xErr != nil {
		return nil, xErr
	}

	return reader, nil
}
