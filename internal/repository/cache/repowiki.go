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

// cacheTTLRepoWikiConfig RepoWiki 配置缓存过期时间（Cache-Aside 模式，与项目缓存保持一致）
const cacheTTLRepoWikiConfig = 30 * time.Minute

// cacheTTLRepoWikiVersionStatus 版本状态缓存过期时间（短 TTL，专为 Agent 轮询优化）
const cacheTTLRepoWikiVersionStatus = 30 * time.Second

// RepoWikiCache RepoWiki 多维度缓存管理器，实现 Cache-Aside 模式
//
// 缓存分两组：
//   - Config 缓存：ID → 配置 JSON 详情，TTL 30 分钟，与 ProjectCache 同策略
//   - 版本状态缓存：versionID → status 字符串，TTL 30 秒，专为 Agent 轮询状态优化（短 TTL 保证状态准实时）
//
// 持有 *xCache.Cache 复用 RDB；Config 缓存使用 c.TTL（30min），
// 版本状态缓存使用独立的 cacheTTLRepoWikiVersionStatus（30s）。
type RepoWikiCache struct {
	*xCache.Cache
}

// NewRepoWikiCache 创建 RepoWikiCache 实例
func NewRepoWikiCache(cache *xCache.Cache) *RepoWikiCache {
	cache.TTL = cacheTTLRepoWikiConfig
	return &RepoWikiCache{Cache: cache}
}

// ── Config 缓存（TTL 30 分钟）──

// GetConfig 根据 ID 读取 RepoWiki 配置详情缓存
//
// 返回值:
//   - *entity.RepoWikiConfig: 缓存命中的配置实体
//   - bool:                    是否命中
//   - *xError.Error:            仅在意外错误时返回（Redis Nil 不视为错误）
func (c *RepoWikiCache) GetConfig(ctx context.Context, configID int64) (*entity.RepoWikiConfig, bool, *xError.Error) {
	key := bConst.CacheRepoWikiConfigByID.Get(configID).String()
	val, err := c.RDB.Get(ctx, key).Result()
	if err != nil || val == "" {
		return nil, false, nil
	}

	var config entity.RepoWikiConfig
	if err := json.Unmarshal([]byte(val), &config); err != nil {
		return nil, false, nil
	}

	return &config, true, nil
}

// SetConfig 写入 RepoWiki 配置缓存
func (c *RepoWikiCache) SetConfig(ctx context.Context, config *entity.RepoWikiConfig) *xError.Error {
	if config == nil {
		return nil
	}

	jsonData, err := json.Marshal(config)
	if err != nil {
		return xError.NewError(ctx, xError.SerializeError, "RepoWiki 配置缓存序列化失败", false, err)
	}

	c.RDB.Set(ctx, bConst.CacheRepoWikiConfigByID.Get(config.ID.Int64()).String(), jsonData, c.TTL)
	return nil
}

// DeleteConfig 清除指定 ID 的配置缓存
func (c *RepoWikiCache) DeleteConfig(ctx context.Context, configID int64) {
	c.RDB.Del(ctx, bConst.CacheRepoWikiConfigByID.Get(configID).String())
}

// ── 版本状态缓存（TTL 30 秒，轮询优化）──

// GetVersionStatus 读取版本分析状态缓存
//
// 返回值:
//   - string:       缓存命中的状态字符串
//   - bool:         是否命中
func (c *RepoWikiCache) GetVersionStatus(ctx context.Context, versionID int64) (string, bool) {
	key := bConst.CacheRepoWikiVersionStatus.Get(versionID).String()
	val, err := c.RDB.Get(ctx, key).Result()
	if err != nil || val == "" {
		return "", false
	}
	return val, true
}

// SetVersionStatus 写入版本分析状态缓存（TTL 30 秒）
func (c *RepoWikiCache) SetVersionStatus(ctx context.Context, versionID int64, status string) {
	c.RDB.Set(ctx, bConst.CacheRepoWikiVersionStatus.Get(versionID).String(), status, cacheTTLRepoWikiVersionStatus)
}

// DeleteVersionStatus 清除版本状态缓存（状态变更后主动失效）
func (c *RepoWikiCache) DeleteVersionStatus(ctx context.Context, versionID int64) {
	c.RDB.Del(ctx, bConst.CacheRepoWikiVersionStatus.Get(versionID).String())
}
