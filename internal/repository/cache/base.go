package cache

import (
	"time"

	"github.com/redis/go-redis/v9"
)

// Base 缓存基础依赖，持有 Redis 客户端与默认过期时间。
//
// bamboo-base-go v1.2.3 起移除了原 xCache.Cache 结构体，改为 Manager 门面 + 泛型接口
// （KeyCache / HashCache / SetCache / ListCache）。但本项目缓存多为多 key 维度
// （如 Project 的 ID/Name/Alias/MatchPath 四层映射、BiometricCredential 的
// ID/CredentialID 双维度），与单 key 的泛型接口不匹配，需直接操作 Redis 客户端。
// 故保留此轻量结构复用 RDB / TTL，语义等价于旧版 xCache.Cache。
//
// 单 key 的简单 KV 缓存可优先考虑迁移到 xCache.KeyCache 泛型接口。
type Base struct {
	RDB *redis.Client // Redis 客户端
	TTL time.Duration // 默认过期时间，0 表示永不过期
}
