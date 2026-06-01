package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xCache "github.com/bamboo-services/bamboo-base-go/major/cache"
	"github.com/redis/go-redis/v9"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
)

// TokenCache Token 缓存管理器 — 封装 AT/RT 的 Redis String 操作
//
// 该类型封装了与 Redis 的交互，用于通过 Token 缓存用户实体（AT）和用户 ID（RT）。
// AT 使用 Redis String 存储 JSON 序列化的 User 实体；RT 使用 Redis String 存储 UserID 字符串。
//
// 注意: 该实现非并发安全，不建议在多 goroutine 中共享同一实例操作。
type TokenCache xCache.Cache

// SetAccessToken 将用户实体以 JSON 格式存入 Redis String
//
// 参数:
//   - ctx: 上下文对象，用于传递请求上下文
//   - tokenMD5: AccessToken 的 MD5 摘要，作为缓存键
//   - user: 要缓存的用户实体指针
//   - ttl: 缓存过期时间
//
// 返回值:
//   - error: 操作过程中发生的错误
func (c *TokenCache) SetAccessToken(ctx context.Context, tokenMD5 string, user *entity.User, ttl time.Duration) error {
	if tokenMD5 == "" {
		return fmt.Errorf("访问令牌标识为空")
	}
	if user == nil {
		return fmt.Errorf("缓存值为空")
	}

	data, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("序列化用户实体失败: %w", err)
	}

	return c.RDB.Set(ctx, bConst.CacheAuthToken.Get(tokenMD5).String(), data, ttl).Err()
}

// GetAccessToken 从 Redis String 中获取并反序列化用户实体
//
// 参数:
//   - ctx: 上下文对象，用于传递请求上下文
//   - tokenMD5: AccessToken 的 MD5 摘要，作为缓存键
//
// 返回值:
//   - *entity.User: 用户实体，未命中缓存时返回 nil
//   - error: 操作过程中发生的错误
func (c *TokenCache) GetAccessToken(ctx context.Context, tokenMD5 string) (*entity.User, error) {
	if tokenMD5 == "" {
		return nil, fmt.Errorf("访问令牌标识为空")
	}

	data, err := c.RDB.Get(ctx, bConst.CacheAuthToken.Get(tokenMD5).String()).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	var user entity.User
	if err = json.Unmarshal(data, &user); err != nil {
		return nil, fmt.Errorf("反序列化用户实体失败: %w", err)
	}

	return &user, nil
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

// SetRefreshToken 将 UserID 以字符串形式存入 Redis String
//
// 参数:
//   - ctx: 上下文对象，用于传递请求上下文
//   - rt: RefreshToken 原始值，作为缓存键
//   - userID: 用户雪花 ID
//   - ttl: 缓存过期时间
//
// 返回值:
//   - error: 操作过程中发生的错误
func (c *TokenCache) SetRefreshToken(ctx context.Context, rt string, userID xSnowflake.SnowflakeID, ttl time.Duration) error {
	if rt == "" {
		return fmt.Errorf("刷新令牌标识为空")
	}

	return c.RDB.Set(ctx, bConst.CacheRefreshToken.Get(rt).String(), userID.String(), ttl).Err()
}

// GetRefreshToken 从 Redis String 中获取并解析 UserID
//
// 参数:
//   - ctx: 上下文对象，用于传递请求上下文
//   - rt: RefreshToken 原始值，作为缓存键
//
// 返回值:
//   - xSnowflake.SnowflakeID: 用户雪花 ID，未命中缓存时返回 0
//   - error: 操作过程中发生的错误
func (c *TokenCache) GetRefreshToken(ctx context.Context, rt string) (xSnowflake.SnowflakeID, error) {
	if rt == "" {
		return 0, fmt.Errorf("刷新令牌标识为空")
	}

	value, err := c.RDB.Get(ctx, bConst.CacheRefreshToken.Get(rt).String()).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, nil
		}
		return 0, err
	}

	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("解析用户 ID 失败: %w", err)
	}

	return xSnowflake.SnowflakeID(id), nil
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
