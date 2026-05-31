package repository

import (
	"context"
	"time"

	"github.com/xiaolfeng/Lumina/internal/entity"
	"github.com/xiaolfeng/Lumina/internal/repository/cache"
	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	"github.com/redis/go-redis/v9"
)

// TokenRepo Token 缓存仓储，负责管理 AccessToken 和 RefreshToken 的缓存读写
//
// 该类型为纯缓存仓储，无数据库依赖。通过封装 TokenCache 提供统一的数据访问接口，
// 供上层业务逻辑层使用。AT 缓存默认 2 小时过期，RT 缓存默认 14 天过期。
//
// 字段说明:
//   - atCache: AccessToken 缓存管理器，存储用户实体
//   - rtCache: RefreshToken 缓存管理器，存储用户 ID
//   - log: 带命名空间的结构化日志记录器
type TokenRepo struct {
	atCache *cache.TokenCache
	rtCache *cache.TokenCache
	log     *xLog.LogNamedLogger
}

// NewTokenRepo 初始化并返回一个 TokenRepo 仓储实例
//
// 该工厂函数通过组装 Redis 客户端和日志记录器，构建一个具备 AT/RT 缓存管理的
// TokenRepo 仓储对象。AT 缓存默认 2 小时过期，RT 缓存默认 14 天过期。
//
// 参数说明:
//   - rdb: 已初始化的 Redis 客户端实例，用于构建缓存策略
//
// 返回值:
//   - *TokenRepo: 配置完成的 TokenRepo 仓储实例指针，可直接用于业务逻辑层
func NewTokenRepo(rdb *redis.Client) *TokenRepo {
	return &TokenRepo{
		atCache: &cache.TokenCache{
			RDB: rdb,
			TTL: 2 * time.Hour,
		},
		rtCache: &cache.TokenCache{
			RDB: rdb,
			TTL: 14 * 24 * time.Hour,
		},
		log: xLog.WithName(xLog.NamedREPO, "TokenRepo"),
	}
}

// SetAccessToken 将用户实体写入 AccessToken 摘要对应的缓存
//
// 该方法通过 tokenMD5 作为键，将完整的用户实体序列化为 JSON 写入 Redis String 缓存。
// 缓存写入失败仅记录警告日志，不影响上层业务返回。
//
// 参数:
//   - ctx: 上下文对象，用于传递请求上下文
//   - tokenMD5: AccessToken 的 MD5 摘要，作为缓存键
//   - user: 要缓存的用户实体指针
//
// 返回值:
//   - *xError.Error: 缓存写入过程中的错误
func (r *TokenRepo) SetAccessToken(ctx context.Context, tokenMD5 string, user *entity.User) *xError.Error {
	r.log.Info(ctx, "SetAccessToken - 写入 AccessToken 用户缓存")

	if err := r.atCache.SetAccessToken(ctx, tokenMD5, user, r.atCache.TTL); err != nil {
		r.log.Warn(ctx, err.Error())
		return xError.NewError(ctx, xError.CacheError, "写入 AccessToken 用户缓存失败", true, err)
	}
	return nil
}

// GetAccessToken 从缓存中获取指定 AccessToken 摘要对应的用户实体
//
// 该方法通过 tokenMD5 作为键，从 Redis String 缓存中读取 JSON 反序列化的用户实体。
// 缓存未命中时返回 (nil, false, nil)，调用方可据此判断是否需要回退到数据库获取。
//
// 参数:
//   - ctx: 上下文对象，用于传递请求上下文
//   - tokenMD5: AccessToken 的 MD5 摘要，作为缓存键
//
// 返回值:
//   - *entity.User: 命中的用户实体，未命中时返回 nil
//   - bool: 是否命中缓存（true 表示命中，false 表示未命中）
//   - *xError.Error: 缓存读取过程中的错误
func (r *TokenRepo) GetAccessToken(ctx context.Context, tokenMD5 string) (*entity.User, bool, *xError.Error) {
	r.log.Info(ctx, "GetAccessToken - 从缓存获取 AccessToken 用户信息")

	user, err := r.atCache.GetAccessToken(ctx, tokenMD5)
	if err != nil {
		return nil, false, xError.NewError(ctx, xError.CacheError, "获取 AccessToken 用户缓存失败", true, err)
	}
	if user != nil {
		return user, true, nil
	}
	return nil, false, nil
}

// DeleteAccessToken 删除 AccessToken 摘要对应的缓存
//
// 参数:
//   - ctx: 上下文对象，用于传递请求上下文
//   - tokenMD5: AccessToken 的 MD5 摘要，作为缓存键
//
// 返回值:
//   - *xError.Error: 缓存删除过程中的错误
func (r *TokenRepo) DeleteAccessToken(ctx context.Context, tokenMD5 string) *xError.Error {
	r.log.Info(ctx, "DeleteAccessToken - 删除 AccessToken 用户缓存")

	if err := r.atCache.DeleteAccessToken(ctx, tokenMD5); err != nil {
		r.log.Warn(ctx, err.Error())
		return xError.NewError(ctx, xError.CacheError, "删除 AccessToken 用户缓存失败", true, err)
	}
	return nil
}

// SetRefreshToken 将用户 ID 写入 RefreshToken 对应的缓存
//
// 该方法通过 RT 作为键，将用户雪花 ID 以字符串形式写入 Redis String 缓存。
//
// 参数:
//   - ctx: 上下文对象，用于传递请求上下文
//   - rt: RefreshToken 原始值，作为缓存键
//   - userID: 用户雪花 ID
//
// 返回值:
//   - *xError.Error: 缓存写入过程中的错误
func (r *TokenRepo) SetRefreshToken(ctx context.Context, rt string, userID xSnowflake.SnowflakeID) *xError.Error {
	r.log.Info(ctx, "SetRefreshToken - 写入 RefreshToken 缓存")

	if err := r.rtCache.SetRefreshToken(ctx, rt, userID, r.rtCache.TTL); err != nil {
		r.log.Warn(ctx, err.Error())
		return xError.NewError(ctx, xError.CacheError, "写入 RefreshToken 缓存失败", true, err)
	}
	return nil
}

// GetRefreshToken 从缓存中获取指定 RefreshToken 对应的用户 ID
//
// 该方法通过 RT 作为键，从 Redis String 缓存中读取并解析用户雪花 ID。
// 缓存未命中时返回 (0, false, nil)，调用方可据此判断是否需要回退到数据库获取。
//
// 参数:
//   - ctx: 上下文对象，用于传递请求上下文
//   - rt: RefreshToken 原始值，作为缓存键
//
// 返回值:
//   - xSnowflake.SnowflakeID: 用户雪花 ID，未命中时返回 0
//   - bool: 是否命中缓存（true 表示命中，false 表示未命中）
//   - *xError.Error: 缓存读取过程中的错误
func (r *TokenRepo) GetRefreshToken(ctx context.Context, rt string) (xSnowflake.SnowflakeID, bool, *xError.Error) {
	r.log.Info(ctx, "GetRefreshToken - 从缓存获取 RefreshToken 用户 ID")

	userID, err := r.rtCache.GetRefreshToken(ctx, rt)
	if err != nil {
		return 0, false, xError.NewError(ctx, xError.CacheError, "获取 RefreshToken 缓存失败", true, err)
	}
	if userID != 0 {
		return userID, true, nil
	}
	return 0, false, nil
}

// DeleteRefreshToken 删除 RefreshToken 对应的缓存
//
// 参数:
//   - ctx: 上下文对象，用于传递请求上下文
//   - rt: RefreshToken 原始值，作为缓存键
//
// 返回值:
//   - *xError.Error: 缓存删除过程中的错误
func (r *TokenRepo) DeleteRefreshToken(ctx context.Context, rt string) *xError.Error {
	r.log.Info(ctx, "DeleteRefreshToken - 删除 RefreshToken 缓存")

	if err := r.rtCache.DeleteRefreshToken(ctx, rt); err != nil {
		r.log.Warn(ctx, err.Error())
		return xError.NewError(ctx, xError.CacheError, "删除 RefreshToken 缓存失败", true, err)
	}
	return nil
}
