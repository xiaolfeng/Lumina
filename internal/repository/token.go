package repository

import (
	"context"
	"time"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xCache "github.com/bamboo-services/bamboo-base-go/major/cache"
	"github.com/redis/go-redis/v9"
	"github.com/xiaolfeng/Lumina/internal/repository/cache"
)

// TokenRepo Token 缓存仓储，负责管理 AccessToken 和 RefreshToken 的缓存读写
//
// 该类型为纯缓存仓储，无数据库依赖。通过封装 AccessTokenCache 和 RefreshTokenCache
// 提供统一的数据访问接口，供上层业务逻辑层使用。
// AT 缓存默认 2 小时过期，RT 缓存默认 14 天过期。
// 采用单用户模式：AT 仅存储认证状态标记，RT 存储认证状态标记。
//
// 字段说明:
//   - atCache: AccessToken 缓存管理器
//   - rtCache: RefreshToken 缓存管理器
//   - log:     带命名空间的结构化日志记录器
type TokenRepo struct {
	atCache *cache.AccessTokenCache
	rtCache *cache.RefreshTokenCache
	log     *xLog.LogNamedLogger
}

// NewTokenRepo 初始化并返回一个 TokenRepo 仓储实例
//
// 该工厂函数通过组装 Redis 客户端和日志记录器，构建具备 AT/RT 缓存管理的
// TokenRepo 仓储对象。AT 缓存默认 2 小时过期，RT 缓存默认 14 天过期。
//
// 参数说明:
//   - rdb: 已初始化的 Redis 客户端实例，用于构建缓存策略
//
// 返回值:
//   - *TokenRepo: 配置完成的 TokenRepo 仓储实例指针，可直接用于业务逻辑层
func NewTokenRepo(rdb *redis.Client) *TokenRepo {
	return &TokenRepo{
		atCache: &cache.AccessTokenCache{
			Cache: &xCache.Cache{RDB: rdb, TTL: 2 * time.Hour},
		},
		rtCache: &cache.RefreshTokenCache{
			Cache: &xCache.Cache{RDB: rdb, TTL: 14 * 24 * time.Hour},
		},
		log: xLog.WithName(xLog.NamedREPO, "TokenRepo"),
	}
}

// SetAccessToken 将认证状态写入 AccessToken 缓存
//
// 参数:
//   - ctx: 上下文对象，用于传递请求上下文
//   - token: AccessToken 原始值，作为缓存键
//
// 返回值:
//   - *xError.Error: 缓存写入过程中的错误
func (r *TokenRepo) SetAccessToken(ctx context.Context, token string) *xError.Error {
	r.log.Info(ctx, "SetAccessToken - 写入 AccessToken 缓存")

	if xErr := r.atCache.Set(ctx, token, &cache.TokenInfo{Authenticated: true}); xErr != nil {
		r.log.Warn(ctx, xErr.Error())
		return xErr
	}
	return nil
}

// GetAccessToken 从缓存中检查指定 AccessToken 是否存在
//
// 参数:
//   - ctx: 上下文对象，用于传递请求上下文
//   - token: AccessToken 原始值，作为缓存键
//
// 返回值:
//   - bool: 是否命中缓存（true 表示命中，false 表示未命中）
//   - *xError.Error: 缓存读取过程中的错误
func (r *TokenRepo) GetAccessToken(ctx context.Context, token string) (bool, *xError.Error) {
	r.log.Info(ctx, "GetAccessToken - 从缓存获取 AccessToken 状态")

	found, xErr := r.atCache.Exists(ctx, token)
	if xErr != nil {
		return false, xErr
	}
	return found, nil
}

// DeleteAccessToken 删除 AccessToken 对应的缓存
//
// 参数:
//   - ctx: 上下文对象，用于传递请求上下文
//   - token: AccessToken 原始值，作为缓存键
//
// 返回值:
//   - *xError.Error: 缓存删除过程中的错误
func (r *TokenRepo) DeleteAccessToken(ctx context.Context, token string) *xError.Error {
	r.log.Info(ctx, "DeleteAccessToken - 删除 AccessToken 缓存")

	if xErr := r.atCache.Delete(ctx, token); xErr != nil {
		r.log.Warn(ctx, xErr.Error())
		return xErr
	}
	return nil
}

// SetRefreshToken 将 RefreshToken 写入缓存
//
// 参数:
//   - ctx: 上下文对象，用于传递请求上下文
//   - token: RefreshToken 原始值，作为缓存键
//
// 返回值:
//   - *xError.Error: 缓存写入过程中的错误
func (r *TokenRepo) SetRefreshToken(ctx context.Context, token string) *xError.Error {
	r.log.Info(ctx, "SetRefreshToken - 写入 RefreshToken 缓存")

	if xErr := r.rtCache.Set(ctx, token, &cache.TokenInfo{Authenticated: true}); xErr != nil {
		r.log.Warn(ctx, xErr.Error())
		return xErr
	}
	return nil
}

// GetRefreshToken 从缓存中检查指定 RefreshToken 是否存在
//
// 参数:
//   - ctx: 上下文对象，用于传递请求上下文
//   - token: RefreshToken 原始值，作为缓存键
//
// 返回值:
//   - bool: 是否命中缓存（true 表示命中，false 表示未命中）
//   - *xError.Error: 缓存读取过程中的错误
func (r *TokenRepo) GetRefreshToken(ctx context.Context, token string) (bool, *xError.Error) {
	r.log.Info(ctx, "GetRefreshToken - 从缓存获取 RefreshToken 状态")

	found, xErr := r.rtCache.Exists(ctx, token)
	if xErr != nil {
		return false, xErr
	}
	return found, nil
}

// DeleteRefreshToken 删除 RefreshToken 对应的缓存
//
// 参数:
//   - ctx: 上下文对象，用于传递请求上下文
//   - token: RefreshToken 原始值，作为缓存键
//
// 返回值:
//   - *xError.Error: 缓存删除过程中的错误
func (r *TokenRepo) DeleteRefreshToken(ctx context.Context, token string) *xError.Error {
	r.log.Info(ctx, "DeleteRefreshToken - 删除 RefreshToken 缓存")

	if xErr := r.rtCache.Delete(ctx, token); xErr != nil {
		r.log.Warn(ctx, xErr.Error())
		return xErr
	}
	return nil
}
