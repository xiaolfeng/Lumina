package cache

import (
	"context"
	"encoding/json"
	"fmt"

	xCache "github.com/bamboo-services/bamboo-base-go/major/cache"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
)

// QaSessionCache QA 会话多维度缓存管理器，实现 Cache-Aside 模式
//
// 承接原先内联在 repository/qa_session.go 的缓存读写逻辑。会话缓存采用两层映射：
//   - ID → 会话 JSON 详情（主缓存，GetByID 命中）
//   - Hash → 会话 ID（哈希索引，GetByHash 命中后复用 GetByID）
//
// 与 ProjectCache 同理，多 key 维度不套单 key 的 KeyCache 接口，
// 持有 *xCache.Cache 复用 RDB/TTL。
type QaSessionCache struct {
	*xCache.Cache
}

// GetIDByHash 根据 Hash 读取会话 ID 映射缓存
//
// 返回值:
//   - parsedID: 解析后的雪花 ID（缓存值无效时返回零值）
//   - bool:     是否命中
//   - error:    仅在解析雪花 ID 失败时返回
func (c *QaSessionCache) GetIDByHash(ctx context.Context, hash string) (xSnowflake.SnowflakeID, bool, error) {
	key := bConst.CacheQaSessionIDByHash.Get(hash).String()
	val, err := c.RDB.Get(ctx, key).Result()
	if err != nil || val == "" {
		return 0, false, nil
	}

	parsedID, parseErr := xSnowflake.ParseSnowflakeID(val)
	if parseErr != nil {
		return 0, false, nil
	}

	return parsedID, true, nil
}

// GetByID 根据 ID 读取会话详情缓存
//
// 返回值:
//   - *entity.QaSession: 缓存命中的会话实体
//   - bool:               是否命中
//   - error:              仅在意外错误时返回
func (c *QaSessionCache) GetByID(ctx context.Context, id int64) (*entity.QaSession, bool, error) {
	key := bConst.CacheQaSessionByID.Get(id).String()
	val, err := c.RDB.Get(ctx, key).Result()
	if err != nil || val == "" {
		return nil, false, nil
	}

	var session entity.QaSession
	if err := json.Unmarshal([]byte(val), &session); err != nil {
		return nil, false, nil
	}

	return &session, true, nil
}

// SetSession 写入会话全维度缓存
//
// 写入两组键：ID→详情、Hash→ID（若有）。
func (c *QaSessionCache) SetSession(ctx context.Context, session *entity.QaSession) error {
	if session == nil {
		return nil
	}

	jsonData, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("会话缓存序列化失败: %w", err)
	}

	// ID → 详情
	c.RDB.Set(ctx, bConst.CacheQaSessionByID.Get(session.ID.Int64()).String(), jsonData, c.TTL)

	// Hash → ID
	if session.Hash != "" {
		c.RDB.Set(ctx, bConst.CacheQaSessionIDByHash.Get(session.Hash).String(), session.ID.String(), c.TTL)
	}

	return nil
}

// DeleteSession 清除会话全维度缓存
func (c *QaSessionCache) DeleteSession(ctx context.Context, id int64, hash string) {
	c.RDB.Del(ctx, bConst.CacheQaSessionByID.Get(id).String())
	if hash != "" {
		c.RDB.Del(ctx, bConst.CacheQaSessionIDByHash.Get(hash).String())
	}
}
