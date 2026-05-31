package bConst

import (
	"fmt"

	xEnv "github.com/bamboo-services/bamboo-base-go/defined/env"
)

// RedisKey Redis 键类型
type RedisKey string

const (
	CacheAuthToken    RedisKey = "auth:at:%s" // CacheAuthToken AccessToken→User 缓存（%s = MD5(AT)）
	CacheRefreshToken RedisKey = "auth:rt:%s" // CacheRefreshToken RefreshToken→UserID 缓存（%s = RT）
)

// Get 格式化 Redis 键，自动拼接环境前缀
func (k RedisKey) Get(args ...interface{}) RedisKey {
	validKey := xEnv.GetEnvString(xEnv.NoSqlPrefix, "lum:") + string(k)
	return RedisKey(fmt.Sprintf(validKey, args...))
}

// String 将 RedisKey 转换为字符串
func (k RedisKey) String() string {
	return string(k)
}
