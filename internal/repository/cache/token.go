package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	xCache "github.com/bamboo-services/bamboo-base-go/major/cache"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

// TokenCache Token 缓存管理器 — 封装 AT/RT 的 Redis String 操作
//
// 该类型封装了与 Redis 的交互，用于通过 Token 标识验证访问/刷新令牌的有效性。
// 单用户模式下，AT 仅存储认证状态（authenticated: true），RT 存储固定标识 "owner"。
//
// 注意: 该实现非并发安全，不建议在多 goroutine 中共享同一实例操作。
type TokenCache xCache.Cache

// SetAccessToken 将认证状态以 JSON 格式存入 Redis String
//
// 参数:
//   - ctx: 上下文对象，用于传递请求上下文
//   - tokenMD5: AccessToken 的 MD5 摘要，作为缓存键
//   - ttl: 缓存过期时间
//
// 返回值:
//   - error: 操作过程中发生的错误
func (c *TokenCache) SetAccessToken(ctx context.Context, tokenMD5 string, ttl time.Duration) error {
	if tokenMD5 == "" {
		return fmt.Errorf("访问令牌标识为空")
	}

	data, err := json.Marshal(map[string]bool{"authenticated": true})
	if err != nil {
		return fmt.Errorf("序列化认证状态失败: %w", err)
	}

	return c.RDB.Set(ctx, bConst.CacheAuthToken.Get(tokenMD5).String(), data, ttl).Err()
}

// GetAccessToken 检查 AccessToken 是否存在于 Redis 中
//
// 参数:
//   - ctx: 上下文对象，用于传递请求上下文
//   - tokenMD5: AccessToken 的 MD5 摘要，作为缓存键
//
// 返回值:
//   - bool: 令牌是否有效（true = 已认证，false = 未找到或已过期）
//   - error: 操作过程中发生的错误
func (c *TokenCache) GetAccessToken(ctx context.Context, tokenMD5 string) (bool, error) {
	if tokenMD5 == "" {
		return false, fmt.Errorf("访问令牌标识为空")
	}

	count, err := c.RDB.Exists(ctx, bConst.CacheAuthToken.Get(tokenMD5).String()).Result()
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// DeleteAccessToken 删除 AccessToken 对应的缓存键
//
// 参数:
//   - ctx: 上下文对象，用于传递请求上下文
//   - tokenMD5: AccessToken 的 MD5 摘要，作为缓存键
//
// 返回值:
//   - error: 操作过程中发生的错误
func (c *TokenCache) DeleteAccessToken(ctx context.Context, tokenMD5 string) error {
	if tokenMD5 == "" {
		return fmt.Errorf("访问令牌标识为空")
	}

	return c.RDB.Del(ctx, bConst.CacheAuthToken.Get(tokenMD5).String()).Err()
}

// SetRefreshToken 将固定标识 "owner" 存入 Redis String
//
// 参数:
//   - ctx: 上下文对象，用于传递请求上下文
//   - rt: RefreshToken 原始值，作为缓存键
//   - ttl: 缓存过期时间
//
// 返回值:
//   - error: 操作过程中发生的错误
func (c *TokenCache) SetRefreshToken(ctx context.Context, rt string, ttl time.Duration) error {
	if rt == "" {
		return fmt.Errorf("刷新令牌标识为空")
	}

	return c.RDB.Set(ctx, bConst.CacheRefreshToken.Get(rt).String(), "owner", ttl).Err()
}

// GetRefreshToken 检查 RefreshToken 是否存在于 Redis 中
//
// 参数:
//   - ctx: 上下文对象，用于传递请求上下文
//   - rt: RefreshToken 原始值，作为缓存键
//
// 返回值:
//   - bool: 令牌是否有效（true = 有效，false = 未找到或已过期）
//   - error: 操作过程中发生的错误
func (c *TokenCache) GetRefreshToken(ctx context.Context, rt string) (bool, error) {
	if rt == "" {
		return false, fmt.Errorf("刷新令牌标识为空")
	}

	count, err := c.RDB.Exists(ctx, bConst.CacheRefreshToken.Get(rt).String()).Result()
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// DeleteRefreshToken 删除 RefreshToken 对应的缓存键
//
// 参数:
//   - ctx: 上下文对象，用于传递请求上下文
//   - rt: RefreshToken 原始值，作为缓存键
//
// 返回值:
//   - error: 操作过程中发生的错误
func (c *TokenCache) DeleteRefreshToken(ctx context.Context, rt string) error {
	if rt == "" {
		return fmt.Errorf("刷新令牌标识为空")
	}

	return c.RDB.Del(ctx, bConst.CacheRefreshToken.Get(rt).String()).Err()
}
