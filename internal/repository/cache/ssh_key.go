package cache

import (
	"context"
	"encoding/json"
	"time"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xCache "github.com/bamboo-services/bamboo-base-go/major/cache"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
)

const cacheTTLSshKey = 30 * time.Minute

// sshKeyCacheData SshKey 缓存序列化结构
// entity.SshKey.PrivateKey 为 json:"-"，直接 json.Marshal 会跳过该字段导致缓存丢失私钥；
// 通过嵌入 *entity.SshKey 并重声明 PrivateKey 字段（正常 json tag），外层字段遮蔽内层的 json:"-"
// 字段，实现完整序列化（含私钥）。反序列化后需手动回填 PrivateKey 到 entity。
type sshKeyCacheData struct {
	*entity.SshKey
	PrivateKey string `json:"private_key"`
}

// SshKeyCache SSH 密钥缓存管理器，实现 Cache-Aside 模式
//
// 缓存 ID → SshKey JSON 详情（含私钥），TTL 30 分钟，与 RepoWikiConfigCache 同策略。
// 持有 *xCache.Cache 复用 RDB 连接。
type SshKeyCache struct {
	*xCache.Cache
}

// NewSshKeyCache 创建 SshKeyCache 实例
func NewSshKeyCache(cache *xCache.Cache) *SshKeyCache {
	cache.TTL = cacheTTLSshKey
	return &SshKeyCache{Cache: cache}
}

// GetConfig 根据 ID 读取 SSH 密钥详情缓存
//
// 返回值:
//   - *entity.SshKey: 缓存命中的密钥实体（含 PrivateKey）
//   - bool:            是否命中
//   - *xError.Error:    仅在意外错误时返回（Redis Nil 不视为错误）
func (c *SshKeyCache) GetConfig(ctx context.Context, sshKeyID int64) (*entity.SshKey, bool, *xError.Error) {
	key := bConst.CacheSSHKeyByID.Get(sshKeyID).String()
	val, err := c.RDB.Get(ctx, key).Result()
	if err != nil || val == "" {
		return nil, false, nil
	}

	var data sshKeyCacheData
	data.SshKey = &entity.SshKey{}
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, false, nil
	}

	data.SshKey.PrivateKey = data.PrivateKey
	return data.SshKey, true, nil
}

// SetConfig 写入 SSH 密钥缓存（含 PrivateKey）
func (c *SshKeyCache) SetConfig(ctx context.Context, sshKey *entity.SshKey) *xError.Error {
	if sshKey == nil {
		return nil
	}

	data := sshKeyCacheData{
		SshKey:     sshKey,
		PrivateKey: sshKey.PrivateKey,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return xError.NewError(ctx, xError.SerializeError, "SSH 密钥缓存序列化失败", false, err)
	}

	c.RDB.Set(ctx, bConst.CacheSSHKeyByID.Get(sshKey.ID.Int64()).String(), jsonData, c.TTL)
	return nil
}

// DeleteConfig 清除指定 ID 的 SSH 密钥缓存
func (c *SshKeyCache) DeleteConfig(ctx context.Context, sshKeyID int64) {
	c.RDB.Del(ctx, bConst.CacheSSHKeyByID.Get(sshKeyID).String())
}
