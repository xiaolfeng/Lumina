package repository

import (
	"context"
	"fmt"
	"time"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xCache "github.com/bamboo-services/bamboo-base-go/major/cache"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"github.com/xiaolfeng/Lumina/internal/repository/cache"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// cacheTTLSession QA 会话缓存 TTL（Cache-Aside 模式）
const cacheTTLSession = 10 * time.Minute

// QaSessionRepo QA会话数据访问层，提供完整 CRUD 操作与 Redis Cache-Aside 缓存
//
// 缓存策略采用 Cache-Aside 模式：读取时优先命中缓存，未命中则回源数据库并回填缓存；
// 写入时同步写入缓存；删除时清除关联缓存键。缓存读写委托给 QaSessionCache
// （位于 repository/cache 子层），通过 constant.RedisKey 统一管理缓存键。
// 额外提供 TTL 过期检查：当 Session 状态为 active 且 ExpiresAt 已过期时，自动更新为 expired。
//
// 字段说明:
//   - db:    GORM 数据库实例，执行持久化操作
//   - cache: 会话多维度缓存管理器（ID→详情 + Hash→ID）
//   - log:   带命名空间的结构化日志记录器
type QaSessionRepo struct {
	db    *gorm.DB
	cache *cache.QaSessionCache
	log   *xLog.LogNamedLogger
}

// NewQaSessionRepo 创建 QaSessionRepo 实例
//
// 参数说明:
//   - db:  已初始化的 GORM 数据库实例
//   - rdb: 已初始化的 Redis 客户端实例（用于构造缓存管理器）
//
// 返回值:
//   - *QaSessionRepo: 配置完成的 QaSessionRepo 实例指针
func NewQaSessionRepo(db *gorm.DB, rdb *redis.Client) *QaSessionRepo {
	return &QaSessionRepo{
		db: db,
		cache: &cache.QaSessionCache{
			Cache: &xCache.Cache{RDB: rdb, TTL: cacheTTLSession},
		},
		log: xLog.WithName(xLog.NamedREPO, "QaSessionRepo"),
	}
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

	if err := r.cache.SetSession(ctx, session); err != nil {
		r.log.Warn(ctx, err.Error())
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
	if session, ok, _ := r.cache.GetByID(ctx, id.Int64()); ok {
		r.log.Info(ctx, fmt.Sprintf("GetByID - 缓存命中 [%d]", id.Int64()))
		// 缓存数据也需要检查过期
		r.expireCheck(ctx, session)
		return session, nil
	}

	// 缓存未命中，查询数据库
	var session entity.QaSession
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, xError.NewError(ctx, xError.NotFound, "QA会话不存在", false, nil)
		}
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询QA会话失败", false, err)
	}

	// 检查过期状态
	r.expireCheck(ctx, &session)

	// 回填缓存
	if err := r.cache.SetSession(ctx, &session); err != nil {
		r.log.Warn(ctx, err.Error())
	}
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
	if cachedID, ok, _ := r.cache.GetIDByHash(ctx, hash); ok {
		r.log.Info(ctx, fmt.Sprintf("GetByHash - Hash缓存命中 [%s] → ID [%d]", hash, cachedID.Int64()))
		return r.GetByID(ctx, cachedID)
	}

	// 2. 缓存未命中，查询数据库
	var session entity.QaSession
	if err := r.db.WithContext(ctx).Where("hash = ?", hash).First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, xError.NewError(ctx, xError.NotFound, "QA会话不存在", false, nil)
		}
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询QA会话失败", false, err)
	}

	// 3. 检查过期状态
	r.expireCheck(ctx, &session)

	// 4. 回填缓存（ID→详情 + Hash→ID）
	if err := r.cache.SetSession(ctx, &session); err != nil {
		r.log.Warn(ctx, err.Error())
	}

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
		Order("created_at DESC").
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
	r.cache.DeleteSession(ctx, id.Int64(), session.Hash)
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
		if err := r.cache.SetSession(ctx, session); err != nil {
			r.log.Warn(ctx, err.Error())
		}
	}
}

// ClearCache 清除指定会话的 Redis 缓存（公开方法，供 Logic 层在归档/状态变更后调用）
//
// 与 Delete 的缓存清理区别：ClearCache 自行查询会话 Hash（若缓存未命中则回源 DB），
// 然后清除 ID→详情 和 Hash→ID 两组缓存键。适用于 ArchiveSession 等不经过 Delete 流程
// 但需要刷新缓存的场景。
//
// 参数:
//   - ctx: 上下文对象
//   - id:  会话雪花 ID
//
// 返回值:
//   - *xError.Error: 查询 Hash 过程中的错误（缓存删除失败仅记录日志，不返回错误）
func (r *QaSessionRepo) ClearCache(ctx context.Context, id xSnowflake.SnowflakeID) *xError.Error {
	r.log.Info(ctx, fmt.Sprintf("ClearCache - 清除会话缓存 [%d]", id.Int64()))

	// 尝试从数据库获取 Hash（用于清理 Hash→ID 映射缓存）
	// 不走 GetByID 缓存链路，避免回填刚刚要删除的缓存
	var session entity.QaSession
	if err := r.db.WithContext(ctx).Select("hash").Where("id = ?", id).First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 会话不存在，仅清理 ID 详情缓存（Hash 未知无法清理）
			r.cache.DeleteSession(ctx, id.Int64(), "")
			return nil
		}
		return xError.NewError(ctx, xError.DatabaseError, "查询会话Hash失败", false, err)
	}

	r.cache.DeleteSession(ctx, id.Int64(), session.Hash)
	return nil
}
