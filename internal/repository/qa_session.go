package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xEnv "github.com/bamboo-services/bamboo-base-go/defined/env"
	"github.com/xiaolfeng/Lumina/internal/entity"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	cacheKeySessionDetail = "qa:session:%d"   // cacheKeySessionDetail Session详情缓存
	cacheKeySessionHash   = "qa:session:hash:%s" // cacheKeySessionHash Hash→ID 缓存键
	cacheTTLSession       = 10 * time.Minute // cacheTTLSession Session缓存TTL
)

// QaSessionRepo QA会话数据访问层，提供完整 CRUD 操作与 Redis Cache-Aside 缓存
//
// 缓存策略采用 Cache-Aside 模式：读取时优先命中缓存，未命中则回源数据库并回填缓存；
// 写入时同步写入缓存；删除时清除关联缓存键。缓存键采用 ID→详情映射，TTL 统一 10 分钟。
// 额外提供 TTL 过期检查：当 Session 状态为 active 且 ExpiresAt 已过期时，自动更新为 expired。
//
// 字段说明:
//   - db:  GORM 数据库实例，执行持久化操作
//   - rdb: Redis 客户端实例，执行缓存读写
//   - log: 带命名空间的结构化日志记录器
type QaSessionRepo struct {
	db  *gorm.DB
	rdb *redis.Client
	log *xLog.LogNamedLogger
}

// NewQaSessionRepo 创建 QaSessionRepo 实例
//
// 参数说明:
//   - db:  已初始化的 GORM 数据库实例
//   - rdb: 已初始化的 Redis 客户端实例
//
// 返回值:
//   - *QaSessionRepo: 配置完成的 QaSessionRepo 实例指针
func NewQaSessionRepo(db *gorm.DB, rdb *redis.Client) *QaSessionRepo {
	return &QaSessionRepo{
		db:  db,
		rdb: rdb,
		log: xLog.WithName(xLog.NamedREPO, "QaSessionRepo"),
	}
}

// cacheKey 构建带环境前缀的 Redis 缓存键
func (r *QaSessionRepo) cacheKey(pattern string, args ...interface{}) string {
	prefix := xEnv.GetEnvString(xEnv.NoSqlPrefix, "lum:")
	return prefix + fmt.Sprintf(pattern, args...)
}

// Create 创建QA会话，成功后同步写入缓存
//
// 参数:
//   - ctx:    上下文对象
//   - session: 待创建的QA会话实体（ID 由雪花算法自动生成）
//
// 返回值:
//   - *xError.Error: 创建过程中的错误
func (r *QaSessionRepo) Create(ctx context.Context, session *entity.QaSession) *xError.Error {
	r.log.Info(ctx, fmt.Sprintf("Create - 创建QA会话 [%s]", session.Title))

	if err := r.db.WithContext(ctx).Create(session).Error; err != nil {
		r.log.Warn(ctx, err.Error())
		return xError.NewError(ctx, xError.DatabaseError, "创建QA会话失败", false, err)
	}

	r.cacheSession(ctx, session)
	// 缓存 Hash→ID 映射
	if session.Hash != "" {
		r.rdb.Set(ctx, r.cacheKey(cacheKeySessionHash, session.Hash), session.ID.String(), cacheTTLSession)
	}
	return nil
}

// GetByID 根据 ID 获取QA会话，优先读取缓存（Cache-First）
//
// 缓存命中时直接反序列化返回；未命中则查询数据库并回填缓存。
// 获取后检查会话是否已过期，若 active 且 ExpiresAt 已过则更新为 expired。
//
// 参数:
//   - ctx: 上下文对象
//   - id:  会话雪花 ID
//
// 返回值:
//   - *entity.QaSession: 查询到的QA会话实体
//   - *xError.Error:     查询过程中的错误
func (r *QaSessionRepo) GetByID(ctx context.Context, id xSnowflake.SnowflakeID) (*entity.QaSession, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("GetByID - 根据ID获取QA会话 [%d]", id.Int64()))

	// 尝试从缓存读取
	cacheData, err := r.rdb.Get(ctx, r.cacheKey(cacheKeySessionDetail, id.Int64())).Result()
	if err == nil && cacheData != "" {
		var session entity.QaSession
		if unmarshalErr := json.Unmarshal([]byte(cacheData), &session); unmarshalErr == nil {
			r.log.Info(ctx, fmt.Sprintf("GetByID - 缓存命中 [%d]", id.Int64()))
			// 缓存数据也需要检查过期
			r.expireCheck(ctx, &session)
			return &session, nil
		}
	}

	// 缓存未命中，查询数据库
	var session entity.QaSession
	if err = r.db.WithContext(ctx).Where("id = ?", id).First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, xError.NewError(ctx, xError.NotFound, "QA会话不存在", false, nil)
		}
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询QA会话失败", false, err)
	}

	// 检查过期状态
	r.expireCheck(ctx, &session)

	// 回填缓存
	r.cacheSession(ctx, &session)
	return &session, nil
}

// GetByHash 根据 Hash 获取QA会话，优先读取 Hash→ID 缓存（Cache-First）
//
// 先通过 Hash→ID 映射缓存获取会话 ID，命中后复用 GetByID 完成完整缓存链路；
// 未命中则查询数据库并回填 Hash→ID 和 ID→详情两组缓存。
//
// 参数:
//   - ctx:  上下文对象
//   - hash: 会话哈希标识（16位字符串）
//
// 返回值:
//   - *entity.QaSession: 查询到的QA会话实体
//   - *xError.Error:     查询过程中的错误
func (r *QaSessionRepo) GetByHash(ctx context.Context, hash string) (*entity.QaSession, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("GetByHash - 根据Hash获取QA会话 [%s]", hash))

	// 1. 尝试从 Hash→ID 缓存获取会话 ID
	cachedID, err := r.rdb.Get(ctx, r.cacheKey(cacheKeySessionHash, hash)).Result()
	if err == nil && cachedID != "" {
		id, parseErr := xSnowflake.ParseSnowflakeID(cachedID)
		if parseErr == nil {
			r.log.Info(ctx, fmt.Sprintf("GetByHash - Hash缓存命中 [%s] → ID [%d]", hash, id.Int64()))
			return r.GetByID(ctx, id)
		}
	}

	// 2. 缓存未命中，查询数据库
	var session entity.QaSession
	if err = r.db.WithContext(ctx).Where("hash = ?", hash).First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, xError.NewError(ctx, xError.NotFound, "QA会话不存在", false, nil)
		}
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询QA会话失败", false, err)
	}

	// 3. 检查过期状态
	r.expireCheck(ctx, &session)

	// 4. 回填缓存（ID→详情 + Hash→ID）
	r.cacheSession(ctx, &session)
	r.rdb.Set(ctx, r.cacheKey(cacheKeySessionHash, hash), session.ID.String(), cacheTTLSession)

	return &session, nil
}

// GetByIDWithQuestions 根据 ID 获取QA会话及其关联问题列表
//
// 内部调用 GetByID 以复用缓存和过期检查逻辑，
// 然后单独查询关联的问题列表（按创建时间升序排列）。
//
// 参数:
//   - ctx: 上下文对象
//   - id:  会话雪花 ID
//
// 返回值:
//   - *entity.QaSession:  查询到的QA会话实体
//   - []*entity.QaQuestion: 关联的问题列表
//   - *xError.Error:      查询过程中的错误
func (r *QaSessionRepo) GetByIDWithQuestions(ctx context.Context, id xSnowflake.SnowflakeID) (*entity.QaSession, []*entity.QaQuestion, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("GetByIDWithQuestions - 根据ID获取QA会话及问题 [%d]", id.Int64()))

	// 复用 GetByID 获取会话（含缓存 + 过期检查）
	session, xErr := r.GetByID(ctx, id)
	if xErr != nil {
		return nil, nil, xErr
	}

	// 查询关联问题列表
	var questions []*entity.QaQuestion
	if err := r.db.WithContext(ctx).
		Where("session_id = ?", id.Int64()).
		Order("created_at ASC").
		Find(&questions).Error; err != nil {
		return nil, nil, xError.NewError(ctx, xError.DatabaseError, "查询QA问题列表失败", false, err)
	}

	return session, questions, nil
}

// List 分页获取QA会话列表（按创建时间降序），支持状态和类型过滤
//
// 参数:
//   - ctx:         上下文对象
//   - page:        页码（从 1 开始）
//   - size:        每页数量
//   - statusFilter: 状态过滤条件（空字符串表示不过滤）
//   - typeFilter:   类型过滤条件（空字符串表示不过滤）
//   - hashFilter:   哈希过滤条件（空字符串表示不过滤）
//
// 返回值:
//   - []*entity.QaSession: 当前页的会话列表
//   - int64:                符合条件的总记录数
//   - *xError.Error:        查询过程中的错误
func (r *QaSessionRepo) List(ctx context.Context, page, size int, statusFilter, typeFilter, hashFilter string) ([]*entity.QaSession, int64, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("List - 分页获取QA会话列表 [page=%d, size=%d, status=%s, type=%s, hash=%s]", page, size, statusFilter, typeFilter, hashFilter))

	// 构建基础查询
	query := r.db.WithContext(ctx).Model(&entity.QaSession{})

	// 动态追加过滤条件
	if statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}
	if typeFilter != "" {
		query = query.Where("type = ?", typeFilter)
	}
	if hashFilter != "" {
		query = query.Where("hash = ?", hashFilter)
	}

	// 统计总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, xError.NewError(ctx, xError.DatabaseError, "统计QA会话数量失败", false, err)
	}

	// 分页查询
	var sessions []*entity.QaSession
	offset := (page - 1) * size
	if err := query.
		Offset(offset).
		Limit(size).
		Order("created_at DESC").
		Find(&sessions).Error; err != nil {
		return nil, 0, xError.NewError(ctx, xError.DatabaseError, "查询QA会话列表失败", false, err)
	}

	return sessions, total, nil
}

// Delete 硬删除QA会话及关联数据（QaSupplement → QaQuestion → QaSession 级联删除），成功后清除缓存
//
// 参数:
//   - ctx: 上下文对象
//   - id:  待删除的会话雪花 ID
//
// 返回值:
//   - *xError.Error: 删除过程中的错误
func (r *QaSessionRepo) Delete(ctx context.Context, id xSnowflake.SnowflakeID) *xError.Error {
	r.log.Info(ctx, fmt.Sprintf("Delete - 硬删除QA会话及关联数据 [%d]", id.Int64()))

	// 先查找会话是否存在（用于获取 Hash 清理缓存）
	var session entity.QaSession
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return xError.NewError(ctx, xError.NotFound, "QA会话不存在", false, nil)
		}
		return xError.NewError(ctx, xError.DatabaseError, "查询待删除QA会话失败", false, err)
	}

	// 1. 级联删除 QaSupplement
	r.db.WithContext(ctx).Where("session_id = ?", id).Delete(&entity.QaSupplement{})
	// 2. 级联删除 QaQuestion
	r.db.WithContext(ctx).Where("session_id = ?", id).Delete(&entity.QaQuestion{})
	// 3. 硬删除 QaSession（Unscoped 跳过软删除）
	result := r.db.WithContext(ctx).Unscoped().Where("id = ?", id).Delete(&entity.QaSession{})
	if result.Error != nil {
		return xError.NewError(ctx, xError.DatabaseError, "硬删除QA会话失败", false, result.Error)
	}
	if result.RowsAffected == 0 {
		return xError.NewError(ctx, xError.NotFound, "QA会话不存在", false, nil)
	}

	// 4. 清除所有关联缓存（详情 + Hash→ID）
	r.clearSessionCache(ctx, id, session.Hash)
	return nil
}

// UpdateStatus 更新会话状态
//
// 参数:
//   - ctx:    上下文对象
//   - id:     会话雪花 ID
//   - status: 目标状态值
//
// 返回值:
//   - *xError.Error: 更新过程中的错误
func (r *QaSessionRepo) UpdateStatus(ctx context.Context, id xSnowflake.SnowflakeID, status string) *xError.Error {
	r.log.Info(ctx, fmt.Sprintf("UpdateStatus - 更新会话状态 [id=%d, status=%s]", id.Int64(), status))
	if err := r.db.WithContext(ctx).Model(&entity.QaSession{}).Where("id = ?", id).Update("status", status).Error; err != nil {
		return xError.NewError(ctx, xError.DatabaseError, "更新会话状态失败", false, err)
	}
	return nil
}

// expireCheck 检查会话是否已过期，若 active 且 ExpiresAt 已过则更新为 expired
//
// 参数:
//   - ctx:    上下文对象
//   - session: 待检查的会话实体（可能会被就地修改）
func (r *QaSessionRepo) expireCheck(ctx context.Context, session *entity.QaSession) {
	if session.Status == "active" && session.ExpiresAt != nil && time.Now().After(*session.ExpiresAt) {
		r.log.Info(ctx, fmt.Sprintf("expireCheck - 会话已过期，更新状态 [%d]", session.ID.Int64()))

		session.Status = "expired"
		if err := r.db.WithContext(ctx).
			Model(&entity.QaSession{}).
			Where("id = ?", session.ID.Int64()).
			Update("status", "expired").Error; err != nil {
			r.log.Warn(ctx, fmt.Sprintf("expireCheck - 更新过期状态失败: %s", err.Error()))
		}

		// 更新缓存中的过期状态
		r.cacheSession(ctx, session)
	}
}

// cacheSession 将QA会话详情写入 Redis 缓存
//
// 写入一组缓存键：
//   - ID → 会话 JSON 详情（TTL 10 分钟）
func (r *QaSessionRepo) cacheSession(ctx context.Context, session *entity.QaSession) {
	jsonData, err := json.Marshal(session)
	if err != nil {
		r.log.Warn(ctx, fmt.Sprintf("缓存序列化失败: %s", err.Error()))
		return
	}

	// ID → 详情
	r.rdb.Set(ctx, r.cacheKey(cacheKeySessionDetail, session.ID.Int64()), jsonData, cacheTTLSession)
}

// clearSessionCache 清除QA会话关联的 Redis 缓存
//
// 清除范围：ID 详情键 + Hash→ID 映射键
func (r *QaSessionRepo) clearSessionCache(ctx context.Context, id xSnowflake.SnowflakeID, hash string) {
	r.rdb.Del(ctx, r.cacheKey(cacheKeySessionDetail, id.Int64()))
	if hash != "" {
		r.rdb.Del(ctx, r.cacheKey(cacheKeySessionHash, hash))
	}
}
