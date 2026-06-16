package cache

import (
	"context"
	"time"

	xCache "github.com/bamboo-services/bamboo-base-go/major/cache"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

// QaRetryCache qa_get_answer 重试计数器缓存管理器
//
// 承接原先散落在 logic 层的 l.rdb.Incr/Expire/Del 调用，将重试计数器语义
// 封装到 cache 子层。该计数器是纯 Redis INCR 计数器，与 KeyCache 接口
// （GET/SET/DEL）语义不匹配，因此不实现 KeyCache，仅持有 *xCache.Cache
// 复用 RDB 连接。
//
// 计数语义：首次 INCR（结果==1）时自动设置 TTL，后续递增不重置 TTL；
// 消费成功或会话结束时通过 Reset 删除计数器。
type QaRetryCache struct {
	*xCache.Cache
}

// Increment 递增指定会话的重试计数器并返回递增后的值
//
// 首次递增（返回值为 1）时自动设置 TTL；后续递增不延长 TTL。
// TTL <= 0 时兜底使用 48 小时。
//
// 参数:
//   - ctx:       上下文对象
//   - sessionID: 会话 ID（用于构建缓存键）
//   - ttl:       首次设置的过期时间
//
// 返回值:
//   - int64: 递增后的计数值
//   - error: Redis 操作失败时返回错误
func (c *QaRetryCache) Increment(ctx context.Context, sessionID string, ttl time.Duration) (int64, error) {
	if sessionID == "" {
		return 0, errRetrySessionEmpty
	}

	key := bConst.CacheQaGetAnswerRetry.Get(sessionID).String()

	count, err := c.RDB.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	// 首次递增时配置过期时间
	if count == 1 {
		if ttl <= 0 {
			ttl = 48 * time.Hour
		}
		_ = c.RDB.Expire(ctx, key, ttl).Err()
	}

	return count, nil
}

// Reset 重置（删除）指定会话的重试计数器
//
// 消费成功后会话归档/取消时调用。删除失败仅返回 error，由调用方记录日志。
//
// 参数:
//   - ctx:       上下文对象
//   - sessionID: 会话 ID
//
// 返回值:
//   - error: Redis 操作失败时返回错误
func (c *QaRetryCache) Reset(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return errRetrySessionEmpty
	}

	key := bConst.CacheQaGetAnswerRetry.Get(sessionID).String()
	return c.RDB.Del(ctx, key).Err()
}

// errRetrySessionEmpty 会话 ID 为空时的哨兵错误
var errRetrySessionEmpty = &retryCacheError{"重试计数器的会话ID为空"}

// retryCacheError 重试缓存错误类型
type retryCacheError struct{ msg string }

func (e *retryCacheError) Error() string { return e.msg }
