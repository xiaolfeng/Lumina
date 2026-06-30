package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xUtil "github.com/bamboo-services/bamboo-base-go/common/utility"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

// DownloadFileInfo 下载文件信息
type DownloadFileInfo struct {
	FilePath string `json:"filePath"` // FilePath 服务器文件系统绝对路径
	Filename string `json:"filename"` // Filename 用户可见下载文件名
	MimeType string `json:"mimeType"` // MimeType 响应 Content-Type
}

// DownloadTokenService 一次性下载令牌管理服务
// 使用 Redis 存储令牌与文件信息映射，通过 Lua 脚本保证令牌消费的原子性
type DownloadTokenService struct {
	rdb *redis.Client
}

// NewDownloadTokenService 创建 DownloadTokenService 实例
func NewDownloadTokenService(rdb *redis.Client) *DownloadTokenService {
	return &DownloadTokenService{rdb: rdb}
}

// GenerateToken 为单个文件生成一次性下载令牌
//
// 令牌格式：cs_ + 32 位无横线 hex（通过 xUtil.Security().GenerateKey() 生成）
// Redis Key 格式：<prefix>qa:download:token:<token>（通过 CacheQaDownloadToken 常量格式化）
// Redis Value：JSON 序列化的 DownloadFileInfo { filePath, filename, mimeType }
// TTL：10 分钟
//
// 参数说明:
//   - ctx: 上下文
//   - filePath: 服务器文件系统绝对路径
//   - filename: 用户可见下载文件名
//   - mimeType: 响应 Content-Type
//
// 返回值:
//   - string: 生成的下载令牌
//   - *xError.Error: 序列化失败或 Redis 写入失败时返回错误
func (s *DownloadTokenService) GenerateToken(ctx context.Context, filePath, filename, mimeType string) (string, *xError.Error) {
	token := xUtil.Security().GenerateKey()

	data, err := json.Marshal(DownloadFileInfo{
		FilePath: filePath,
		Filename: filename,
		MimeType: mimeType,
	})
	if err != nil {
		return "", xError.NewError(ctx, xError.SerializeError, "序列化下载文件信息失败", false, err)
	}

	key := bConst.CacheQaDownloadToken.Get(token).String()
	if err := s.rdb.Set(ctx, key, data, 10*time.Minute).Err(); err != nil {
		return "", xError.NewError(ctx, xError.CacheError, "写入下载令牌到 Redis 失败", false, err)
	}

	return token, nil
}

// luaConsumeToken Lua 脚本：原子 GET + DEL
//
// 执行逻辑：若 key 存在则返回 value 并立即删除；若 key 不存在则返回 nil。
// 通过 Lua 脚本保证 GET+DEL 的原子性，防止并发请求消费同一令牌。
var luaConsumeToken = redis.NewScript(`
local value = redis.call('GET', KEYS[1])
if value == false then
    return nil
end
redis.call('DEL', KEYS[1])
return value
`)

// ConsumeToken 消费令牌（下载成功后调用，立即删除令牌）
//
// 使用 Redis Lua 脚本保证 GET+DEL 原子性，防止并发下载同一令牌。
// 令牌为一次性使用，消费后立即失效。
//
// 参数说明:
//   - ctx: 上下文
//   - token: 待消费的下载令牌
//
// 返回值:
//   - *DownloadFileInfo: 令牌关联的文件信息
//   - *xError.Error: 令牌无效、已过期、已使用或数据格式异常时返回错误
func (s *DownloadTokenService) ConsumeToken(ctx context.Context, token string) (*DownloadFileInfo, *xError.Error) {
	key := bConst.CacheQaDownloadToken.Get(token).String()

	result, err := luaConsumeToken.Run(ctx, s.rdb, []string{key}).Result()
	if err != nil {
		return nil, xError.NewError(ctx, xError.NotExist, "下载令牌无效或已过期", false)
	}
	if result == nil {
		return nil, xError.NewError(ctx, xError.NotExist, "下载令牌无效或已过期", false)
	}

	dataStr, ok := result.(string)
	if !ok {
		return nil, xError.NewError(ctx, xError.DeserializeErr, "下载令牌数据格式异常", false)
	}

	var info DownloadFileInfo
	if err := json.Unmarshal([]byte(dataStr), &info); err != nil {
		return nil, xError.NewError(ctx, xError.DeserializeErr, "解析下载文件信息失败", false, err)
	}

	return &info, nil
}
