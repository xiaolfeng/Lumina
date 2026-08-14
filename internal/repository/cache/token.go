package cache

import (
	"context"
	"encoding/json"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

// TokenInfo 令牌缓存值结构
type TokenInfo struct {
	Authenticated bool `json:"authenticated"` // Authenticated 认证状态标记
}

// AccessTokenCache AccessToken 缓存管理器
type AccessTokenCache struct {
	*Base
}

// Get 根据 AccessToken 检索认证状态
func (c *AccessTokenCache) Get(ctx context.Context, token string) (*TokenInfo, bool, *xError.Error) {
	if token == "" {
		return nil, false, xError.NewError(ctx, xError.BadRequest, "访问令牌标识为空", false)
	}

	val, err := c.RDB.Get(ctx, bConst.CacheAuthToken.Get(token).String()).Result()
	if err != nil {
		// Redis Nil 表示缓存未命中，不是错误
		return nil, false, nil
	}

	var info TokenInfo
	if err := json.Unmarshal([]byte(val), &info); err != nil {
		return nil, false, xError.NewError(ctx, xError.DeserializeErr, "反序列化认证状态失败", false, err)
	}

	return &info, true, nil
}

// Set 将认证状态写入 AccessToken 缓存
func (c *AccessTokenCache) Set(ctx context.Context, token string, info *TokenInfo) *xError.Error {
	if token == "" {
		return xError.NewError(ctx, xError.BadRequest, "访问令牌标识为空", false)
	}
	if info == nil {
		return nil
	}

	data, err := json.Marshal(info)
	if err != nil {
		return xError.NewError(ctx, xError.SerializeError, "序列化认证状态失败", false, err)
	}

	if err := c.RDB.Set(ctx, bConst.CacheAuthToken.Get(token).String(), data, c.TTL).Err(); err != nil {
		return xError.NewError(ctx, xError.CacheError, "写入 AccessToken 缓存失败", false, err)
	}
	return nil
}

// Exists 检查 AccessToken 是否存在于缓存中
func (c *AccessTokenCache) Exists(ctx context.Context, token string) (bool, *xError.Error) {
	if token == "" {
		return false, xError.NewError(ctx, xError.BadRequest, "访问令牌标识为空", false)
	}

	count, err := c.RDB.Exists(ctx, bConst.CacheAuthToken.Get(token).String()).Result()
	if err != nil {
		return false, xError.NewError(ctx, xError.CacheError, "检查 AccessToken 缓存失败", false, err)
	}

	return count > 0, nil
}

// Delete 删除 AccessToken 对应的缓存键
func (c *AccessTokenCache) Delete(ctx context.Context, token string) *xError.Error {
	if token == "" {
		return xError.NewError(ctx, xError.BadRequest, "访问令牌标识为空", false)
	}

	if err := c.RDB.Del(ctx, bConst.CacheAuthToken.Get(token).String()).Err(); err != nil {
		return xError.NewError(ctx, xError.CacheError, "删除 AccessToken 缓存失败", false, err)
	}
	return nil
}

// RefreshTokenCache RefreshToken 缓存管理器
type RefreshTokenCache struct {
	*Base
}

// Get 根据 RefreshToken 检索认证状态
func (c *RefreshTokenCache) Get(ctx context.Context, token string) (*TokenInfo, bool, *xError.Error) {
	if token == "" {
		return nil, false, xError.NewError(ctx, xError.BadRequest, "刷新令牌标识为空", false)
	}

	val, err := c.RDB.Get(ctx, bConst.CacheRefreshToken.Get(token).String()).Result()
	if err != nil {
		return nil, false, nil
	}

	var info TokenInfo
	if err := json.Unmarshal([]byte(val), &info); err != nil {
		return nil, false, xError.NewError(ctx, xError.DeserializeErr, "反序列化认证状态失败", false, err)
	}

	return &info, true, nil
}

// Set 将认证状态写入 RefreshToken 缓存
func (c *RefreshTokenCache) Set(ctx context.Context, token string, info *TokenInfo) *xError.Error {
	if token == "" {
		return xError.NewError(ctx, xError.BadRequest, "刷新令牌标识为空", false)
	}
	if info == nil {
		return nil
	}

	data, err := json.Marshal(info)
	if err != nil {
		return xError.NewError(ctx, xError.SerializeError, "序列化认证状态失败", false, err)
	}

	if err := c.RDB.Set(ctx, bConst.CacheRefreshToken.Get(token).String(), data, c.TTL).Err(); err != nil {
		return xError.NewError(ctx, xError.CacheError, "写入 RefreshToken 缓存失败", false, err)
	}
	return nil
}

// Exists 检查 RefreshToken 是否存在于缓存中
func (c *RefreshTokenCache) Exists(ctx context.Context, token string) (bool, *xError.Error) {
	if token == "" {
		return false, xError.NewError(ctx, xError.BadRequest, "刷新令牌标识为空", false)
	}

	count, err := c.RDB.Exists(ctx, bConst.CacheRefreshToken.Get(token).String()).Result()
	if err != nil {
		return false, xError.NewError(ctx, xError.CacheError, "检查 RefreshToken 缓存失败", false, err)
	}

	return count > 0, nil
}

// Delete 删除 RefreshToken 对应的缓存键
func (c *RefreshTokenCache) Delete(ctx context.Context, token string) *xError.Error {
	if token == "" {
		return xError.NewError(ctx, xError.BadRequest, "刷新令牌标识为空", false)
	}

	if err := c.RDB.Del(ctx, bConst.CacheRefreshToken.Get(token).String()).Err(); err != nil {
		return xError.NewError(ctx, xError.CacheError, "删除 RefreshToken 缓存失败", false, err)
	}
	return nil
}

// consumeRefreshTokenLua 原子消费脚本：GET 命中则 DEL 并返回 1，否则返回 0。
// 兼容所有 Redis 版本（GETDEL 需 Redis 6.2+，Lua 脚本无版本限制）。
const consumeRefreshTokenLua = `
local v = redis.call('GET', KEYS[1])
if v then
  redis.call('DEL', KEYS[1])
  return 1
end
return 0
`

// Consume 原子地消费（GET+DEL Lua）RefreshToken，用于令牌旋转防止并发重复兑换。
//
// 返回 true 表示成功消费；false 表示 token 不存在或已被并发消费。
func (c *RefreshTokenCache) Consume(ctx context.Context, token string) (bool, *xError.Error) {
	if token == "" {
		return false, xError.NewError(ctx, xError.BadRequest, "刷新令牌标识为空", false)
	}

	result, err := c.RDB.Eval(ctx, consumeRefreshTokenLua, []string{bConst.CacheRefreshToken.Get(token).String()}).Int()
	if err != nil {
		return false, xError.NewError(ctx, xError.CacheError, "消费 RefreshToken 缓存失败", false, err)
	}
	return result == 1, nil
}
