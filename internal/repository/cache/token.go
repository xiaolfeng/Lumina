package cache

import (
	"context"
	"encoding/json"
	"fmt"

	xCache "github.com/bamboo-services/bamboo-base-go/major/cache"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

// TokenInfo 令牌缓存值结构
type TokenInfo struct {
	Authenticated bool `json:"authenticated"` // Authenticated 认证状态标记
}

// AccessTokenCache AccessToken 缓存管理器 — 实现 xCache.KeyCache[string, TokenInfo] 接口
type AccessTokenCache struct {
	*xCache.Cache
}

// Get 根据 AccessToken 检索认证状态
func (c *AccessTokenCache) Get(ctx context.Context, token string) (*TokenInfo, bool, error) {
	if token == "" {
		return nil, false, fmt.Errorf("访问令牌标识为空")
	}

	val, err := c.RDB.Get(ctx, bConst.CacheAuthToken.Get(token).String()).Result()
	if err != nil {
		// Redis Nil 表示缓存未命中，不是错误
		return nil, false, nil
	}

	var info TokenInfo
	if err := json.Unmarshal([]byte(val), &info); err != nil {
		return nil, false, err
	}

	return &info, true, nil
}

// Set 将认证状态写入 AccessToken 缓存
func (c *AccessTokenCache) Set(ctx context.Context, token string, info *TokenInfo) error {
	if token == "" {
		return fmt.Errorf("访问令牌标识为空")
	}
	if info == nil {
		return nil
	}

	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("序列化认证状态失败: %w", err)
	}

	return c.RDB.Set(ctx, bConst.CacheAuthToken.Get(token).String(), data, c.TTL).Err()
}

// Exists 检查 AccessToken 是否存在于缓存中
func (c *AccessTokenCache) Exists(ctx context.Context, token string) (bool, error) {
	if token == "" {
		return false, fmt.Errorf("访问令牌标识为空")
	}

	count, err := c.RDB.Exists(ctx, bConst.CacheAuthToken.Get(token).String()).Result()
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// Delete 删除 AccessToken 对应的缓存键
func (c *AccessTokenCache) Delete(ctx context.Context, token string) error {
	if token == "" {
		return fmt.Errorf("访问令牌标识为空")
	}

	return c.RDB.Del(ctx, bConst.CacheAuthToken.Get(token).String()).Err()
}

// RefreshTokenCache RefreshToken 缓存管理器 — 实现 xCache.KeyCache[string, TokenInfo] 接口
type RefreshTokenCache struct {
	*xCache.Cache
}

// Get 根据 RefreshToken 检索认证状态
func (c *RefreshTokenCache) Get(ctx context.Context, token string) (*TokenInfo, bool, error) {
	if token == "" {
		return nil, false, fmt.Errorf("刷新令牌标识为空")
	}

	val, err := c.RDB.Get(ctx, bConst.CacheRefreshToken.Get(token).String()).Result()
	if err != nil {
		return nil, false, nil
	}

	var info TokenInfo
	if err := json.Unmarshal([]byte(val), &info); err != nil {
		return nil, false, err
	}

	return &info, true, nil
}

// Set 将认证状态写入 RefreshToken 缓存
func (c *RefreshTokenCache) Set(ctx context.Context, token string, info *TokenInfo) error {
	if token == "" {
		return fmt.Errorf("刷新令牌标识为空")
	}
	if info == nil {
		return nil
	}

	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("序列化认证状态失败: %w", err)
	}

	return c.RDB.Set(ctx, bConst.CacheRefreshToken.Get(token).String(), data, c.TTL).Err()
}

// Exists 检查 RefreshToken 是否存在于缓存中
func (c *RefreshTokenCache) Exists(ctx context.Context, token string) (bool, error) {
	if token == "" {
		return false, fmt.Errorf("刷新令牌标识为空")
	}

	count, err := c.RDB.Exists(ctx, bConst.CacheRefreshToken.Get(token).String()).Result()
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// Delete 删除 RefreshToken 对应的缓存键
func (c *RefreshTokenCache) Delete(ctx context.Context, token string) error {
	if token == "" {
		return fmt.Errorf("刷新令牌标识为空")
	}

	return c.RDB.Del(ctx, bConst.CacheRefreshToken.Get(token).String()).Err()
}
