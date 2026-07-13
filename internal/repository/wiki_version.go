package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xCache "github.com/bamboo-services/bamboo-base-go/major/cache"
	"github.com/redis/go-redis/v9"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"github.com/xiaolfeng/Lumina/internal/repository/cache"
	"gorm.io/gorm"
)

// WikiVersionRepo Wiki 版本数据访问层，提供完整 CRUD 操作与版本状态缓存
//
// 版本状态缓存采用短 TTL（30 秒）策略，专为 Agent 端轮询分析状态优化：
//   - UpdateStatus / UpdateStage 时同步刷新版本状态缓存
//   - 外部（如 Pipeline 编排器）可直接通过 RepoWikiCache 读取状态，减少 DB 回源
//
// 字段说明:
//   - db:    GORM 数据库实例，执行持久化操作
//   - cache: RepoWiki 缓存管理器（版本状态缓存，TTL 30 秒）
//   - log:   带命名空间的结构化日志记录器
type WikiVersionRepo struct {
	db    *gorm.DB
	cache *cache.RepoWikiCache
	log   *xLog.LogNamedLogger
}

// NewWikiVersionRepo 创建 WikiVersionRepo 实例
//
// 参数说明:
//   - db:  已初始化的 GORM 数据库实例
//   - rdb: 已初始化的 Redis 客户端实例（用于构造缓存管理器）
//
// 返回值:
//   - *WikiVersionRepo: 配置完成的 WikiVersionRepo 实例指针
func NewWikiVersionRepo(db *gorm.DB, rdb *redis.Client) *WikiVersionRepo {
	return &WikiVersionRepo{
		db:    db,
		cache: cache.NewRepoWikiCache(&xCache.Cache{RDB: rdb}),
		log:   xLog.WithName(xLog.NamedREPO, "WikiVersionRepo"),
	}
}

// Create 创建 Wiki 版本记录
//
// 参数:
//   - ctx:    上下文对象
//   - version: 待创建的版本实体（ID 由雪花算法自动生成）
//
// 返回值:
//   - *entity.WikiVersion: 创建后的版本实体（含生成的 ID）
//   - *xError.Error:        创建过程中的错误
func (r *WikiVersionRepo) Create(ctx context.Context, version *entity.WikiVersion) (*entity.WikiVersion, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("Create - 创建 Wiki 版本 [configID=%d, branch=%s]", version.ConfigID.Int64(), version.Branch))

	if err := r.db.WithContext(ctx).Create(version).Error; err != nil {
		r.log.Warn(ctx, err.Error())
		return nil, xError.NewError(ctx, xError.DatabaseError, "创建 Wiki 版本失败", false, err)
	}

	// 写入版本状态缓存
	r.cache.SetVersionStatus(ctx, version.ID.Int64(), version.Status)
	return version, nil
}

// GetByID 根据 ID 获取 Wiki 版本
//
// 参数:
//   - ctx: 上下文对象
//   - id:  版本雪花 ID
//
// 返回值:
//   - *entity.WikiVersion: 查询到的版本实体
//   - *xError.Error:        查询过程中的错误
func (r *WikiVersionRepo) GetByID(ctx context.Context, id xSnowflake.SnowflakeID) (*entity.WikiVersion, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("GetByID - 根据 ID 获取 Wiki 版本 [%d]", id.Int64()))

	var version entity.WikiVersion
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&version).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, xError.NewError(ctx, xError.NotFound, "Wiki 版本不存在", false, nil)
		}
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询 Wiki 版本失败", false, err)
	}

	return &version, nil
}

// GetLatestByConfigID 获取指定配置的最新版本（按创建时间降序取第一条）
//
// 参数:
//   - ctx:      上下文对象
//   - configID: 关联的配置雪花 ID
//
// 返回值:
//   - *entity.WikiVersion: 最新的版本实体
//   - *xError.Error:        查询过程中的错误
func (r *WikiVersionRepo) GetLatestByConfigID(ctx context.Context, configID xSnowflake.SnowflakeID) (*entity.WikiVersion, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("GetLatestByConfigID - 获取配置最新版本 [configID=%d]", configID.Int64()))

	var version entity.WikiVersion
	if err := r.db.WithContext(ctx).
		Where("config_id = ?", configID).
		Order("created_at DESC").
		First(&version).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, xError.NewError(ctx, xError.NotFound, "该配置暂无版本记录", false, nil)
		}
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询最新 Wiki 版本失败", false, err)
	}

	return &version, nil
}

// ListByConfigID 按配置 ID 分页获取版本列表（按创建时间降序）
//
// 参数:
//   - ctx:      上下文对象
//   - configID: 关联的配置雪花 ID
//   - page:     页码（从 1 开始）
//   - size:     每页数量
//
// 返回值:
//   - []*entity.WikiVersion: 当前页的版本列表
//   - int64:                  符合条件的总记录数
//   - *xError.Error:          查询过程中的错误
func (r *WikiVersionRepo) ListByConfigID(ctx context.Context, configID xSnowflake.SnowflakeID, page, size int) ([]*entity.WikiVersion, int64, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("ListByConfigID - 按配置 ID 分页获取版本列表 [configID=%d, page=%d, size=%d]", configID.Int64(), page, size))

	var total int64
	if err := r.db.WithContext(ctx).Model(&entity.WikiVersion{}).Where("config_id = ?", configID).Count(&total).Error; err != nil {
		return nil, 0, xError.NewError(ctx, xError.DatabaseError, "统计 Wiki 版本数量失败", false, err)
	}

	var versions []*entity.WikiVersion
	offset := (page - 1) * size
	if err := r.db.WithContext(ctx).
		Where("config_id = ?", configID).
		Offset(offset).
		Limit(size).
		Order("created_at DESC").
		Find(&versions).Error; err != nil {
		return nil, 0, xError.NewError(ctx, xError.DatabaseError, "查询 Wiki 版本列表失败", false, err)
	}

	return versions, total, nil
}

// ListCompleted 分页查询已完成的Wiki版本（按完成时间倒序）
//
// 参数:
//   - ctx:  上下文对象
//   - page: 页码（从 1 开始）
//   - size: 每页数量
//
// 返回值:
//   - []*entity.WikiVersion: 当前页的版本列表
//   - int64:                  符合条件的总记录数
//   - *xError.Error:          查询过程中的错误
func (r *WikiVersionRepo) ListCompleted(ctx context.Context, page, size int) ([]*entity.WikiVersion, int64, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("ListCompleted - 分页查询已完成的 Wiki 版本 [page=%d, size=%d]", page, size))

	var total int64
	if err := r.db.WithContext(ctx).Model(&entity.WikiVersion{}).Where("status = ?", "completed").Count(&total).Error; err != nil {
		return nil, 0, xError.NewError(ctx, xError.DatabaseError, "统计已完成 Wiki 版本数量失败", false, err)
	}

	var versions []*entity.WikiVersion
	offset := (page - 1) * size
	if err := r.db.WithContext(ctx).
		Where("status = ?", "completed").
		Offset(offset).
		Limit(size).
		Order("completed_at DESC").
		Find(&versions).Error; err != nil {
		return nil, 0, xError.NewError(ctx, xError.DatabaseError, "查询已完成 Wiki 版本列表失败", false, err)
	}

	return versions, total, nil
}

// UpdateStatus 更新版本分析状态，同步刷新版本状态缓存
//
// 专为 Pipeline 编排器在分析阶段推进时调用（如 pending → cloning → scanning → analyzing → completed）。
// 状态变更后同步刷新缓存，使 Agent 端轮询能立即感知新状态。
//
// 参数:
//   - ctx:    上下文对象
//   - id:      版本雪花 ID
//   - status:  新状态值（参见 constant.RepoWikiStatus*）
//
// 返回值:
//   - *xError.Error: 更新过程中的错误
func (r *WikiVersionRepo) UpdateStatus(ctx context.Context, id xSnowflake.SnowflakeID, status string) *xError.Error {
	r.log.Info(ctx, fmt.Sprintf("UpdateStatus - 更新版本状态 [%d] → %s", id.Int64(), status))

	result := r.db.WithContext(ctx).
		Model(&entity.WikiVersion{}).
		Where("id = ?", id).
		Update("status", status)
	if result.Error != nil {
		r.log.Warn(ctx, result.Error.Error())
		return xError.NewError(ctx, xError.DatabaseError, "更新 Wiki 版本状态失败", false, result.Error)
	}
	if result.RowsAffected == 0 {
		return xError.NewError(ctx, xError.NotFound, "Wiki 版本不存在", false, nil)
	}

	// 刷新版本状态缓存
	r.cache.SetVersionStatus(ctx, id.Int64(), status)
	return nil
}

// UpdateStage 更新版本当前阶段
//
// 专为 Pipeline 编排器在阶段切换时调用（如 scan → dep_extract → pass1 → pass2 → ...）。
//
// 参数:
//   - ctx:   上下文对象
//   - id:     版本雪花 ID
//   - stage:  新阶段值（参见 constant.RepoWikiStage*）
//
// 返回值:
//   - *xError.Error: 更新过程中的错误
func (r *WikiVersionRepo) UpdateStage(ctx context.Context, id xSnowflake.SnowflakeID, stage string) *xError.Error {
	r.log.Info(ctx, fmt.Sprintf("UpdateStage - 更新版本阶段 [%d] → %s", id.Int64(), stage))

	result := r.db.WithContext(ctx).
		Model(&entity.WikiVersion{}).
		Where("id = ?", id).
		Update("current_stage", stage)
	if result.Error != nil {
		r.log.Warn(ctx, result.Error.Error())
		return xError.NewError(ctx, xError.DatabaseError, "更新 Wiki 版本阶段失败", false, result.Error)
	}
	if result.RowsAffected == 0 {
		return xError.NewError(ctx, xError.NotFound, "Wiki 版本不存在", false, nil)
	}

	return nil
}

// Update 更新 Wiki 版本完整实体
//
// 注意：若更新涉及 status 字段变更，缓存不会自动失效（因 Save 全量更新无法感知字段差异）。
// 若需同步缓存状态，请显式调用 UpdateStatus。
//
// 参数:
//   - ctx:     上下文对象
//   - version: 待更新的版本实体（需包含完整字段）
//
// 返回值:
//   - *xError.Error: 更新过程中的错误
func (r *WikiVersionRepo) Update(ctx context.Context, version *entity.WikiVersion) *xError.Error {
	r.log.Info(ctx, fmt.Sprintf("Update - 更新 Wiki 版本 [id=%d]", version.ID.Int64()))

	if err := r.db.WithContext(ctx).Save(version).Error; err != nil {
		r.log.Warn(ctx, err.Error())
		return xError.NewError(ctx, xError.DatabaseError, "更新 Wiki 版本失败", false, err)
	}

	// 同步刷新版本状态缓存（确保缓存与 DB 一致）
	r.cache.SetVersionStatus(ctx, version.ID.Int64(), version.Status)
	return nil
}

// Delete 删除 Wiki 版本记录，成功后清除版本状态缓存
//
// 参数:
//   - ctx: 上下文对象
//   - id:  待删除的版本雪花 ID
//
// 返回值:
//   - *xError.Error: 删除过程中的错误
func (r *WikiVersionRepo) Delete(ctx context.Context, id xSnowflake.SnowflakeID) *xError.Error {
	r.log.Info(ctx, fmt.Sprintf("Delete - 删除 Wiki 版本 [%d]", id.Int64()))

	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.WikiVersion{})
	if result.Error != nil {
		r.log.Warn(ctx, result.Error.Error())
		return xError.NewError(ctx, xError.DatabaseError, "删除 Wiki 版本失败", false, result.Error)
	}
	if result.RowsAffected == 0 {
		return xError.NewError(ctx, xError.NotFound, "Wiki 版本不存在", false, nil)
	}

	// 清除版本状态缓存
	r.cache.DeleteVersionStatus(ctx, id.Int64())
	return nil
}

// GetStaleTasks 查询所有非终态且超过超时阈值的版本记录
//
// 非终态状态：pending, cloning, scanning, analyzing, assembling
// 终态状态：completed, failed, cancelled（不查询）
//
// 参数:
//   - ctx:        上下文
//   - timeoutSec: 超时阈值（秒），查询 updated_at < NOW() - timeoutSec 的记录
//   - maxRetry:   最大重试次数，查询 retry_count <= maxRetry 的记录（含等于，使 RetryStaleTask 能标记超限任务为 failed）
// 返回值:
//   - []*entity.WikiVersion: 超时的非终态版本列表
//   - *xError.Error: 查询错误
func (r *WikiVersionRepo) GetStaleTasks(ctx context.Context, timeoutSec int, maxRetry int) ([]*entity.WikiVersion, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("GetStaleTasks - 查询超时非终态任务 [timeoutSec=%d, maxRetry=%d]", timeoutSec, maxRetry))

	nonTerminalStatuses := []string{"pending", "cloning", "scanning", "analyzing", "assembling"}
	threshold := time.Now().Add(-time.Duration(timeoutSec) * time.Second)

	var versions []*entity.WikiVersion
	if err := r.db.WithContext(ctx).
		Where("status IN ? AND updated_at < ? AND retry_count <= ?", nonTerminalStatuses, threshold, maxRetry).
		Find(&versions).Error; err != nil {
		r.log.Warn(ctx, err.Error())
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询超时任务失败", false, err)
	}

	return versions, nil
}
