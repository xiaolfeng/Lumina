package repository

import (
	"context"
	"errors"
	"fmt"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xCache "github.com/bamboo-services/bamboo-base-go/major/cache"
	"github.com/redis/go-redis/v9"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"github.com/xiaolfeng/Lumina/internal/repository/cache"
	"gorm.io/gorm"
)

// RepoWikiConfigRepo RepoWiki 配置数据访问层，提供完整 CRUD 操作与 Redis 缓存层
//
// 缓存策略采用 Cache-Aside 模式：读取时优先命中缓存，未命中则回源数据库并回填缓存；
// 写入/更新时同步刷新缓存；删除时清除关联缓存键。
// 缓存读写委托给 RepoWikiCache（位于 repository/cache 子层），通过 constant.RedisKey
// 统一管理缓存键（Config 缓存 TTL 30 分钟）。
//
// 字段说明:
//   - db:    GORM 数据库实例，执行持久化操作
//   - cache: RepoWiki 多维度缓存管理器（Config 详情缓存）
//   - log:   带命名空间的结构化日志记录器
type RepoWikiConfigRepo struct {
	db    *gorm.DB
	cache *cache.RepoWikiCache
	log   *xLog.LogNamedLogger
}

// NewRepoWikiConfigRepo 创建 RepoWikiConfigRepo 实例
//
// 参数说明:
//   - db:  已初始化的 GORM 数据库实例
//   - rdb: 已初始化的 Redis 客户端实例（用于构造缓存管理器）
//
// 返回值:
//   - *RepoWikiConfigRepo: 配置完成的 RepoWikiConfigRepo 实例指针
func NewRepoWikiConfigRepo(db *gorm.DB, rdb *redis.Client) *RepoWikiConfigRepo {
	return &RepoWikiConfigRepo{
		db:    db,
		cache: cache.NewRepoWikiCache(&xCache.Cache{RDB: rdb}),
		log:   xLog.WithName(xLog.NamedREPO, "RepoWikiConfigRepo"),
	}
}

// Create 创建 RepoWiki 配置，成功后同步写入缓存
//
// 参数:
//   - ctx:    上下文对象
//   - config: 待创建的配置实体（ID 由雪花算法自动生成）
//
// 返回值:
//   - *entity.RepoWikiConfig: 创建后的配置实体（含生成的 ID）
//   - *xError.Error:           创建过程中的错误
func (r *RepoWikiConfigRepo) Create(ctx context.Context, config *entity.RepoWikiConfig) (*entity.RepoWikiConfig, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("Create - 创建 RepoWiki 配置 [projectID=%d]", config.ProjectID.Int64()))

	if err := r.db.WithContext(ctx).Create(config).Error; err != nil {
		r.log.Warn(ctx, err.Error())
		return nil, xError.NewError(ctx, xError.DatabaseError, "创建 RepoWiki 配置失败", false, err)
	}

	if xErr := r.cache.SetConfig(ctx, config); xErr != nil {
		r.log.Warn(ctx, xErr.Error())
	}
	return config, nil
}

// GetByID 根据 ID 获取 RepoWiki 配置，优先读取缓存（Cache-First）
//
// 缓存命中时直接反序列化返回；未命中则查询数据库并回填缓存。
//
// 参数:
//   - ctx: 上下文对象
//   - id:  配置雪花 ID
//
// 返回值:
//   - *entity.RepoWikiConfig: 查询到的配置实体
//   - *xError.Error:           查询过程中的错误
func (r *RepoWikiConfigRepo) GetByID(ctx context.Context, id xSnowflake.SnowflakeID) (*entity.RepoWikiConfig, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("GetByID - 根据 ID 获取 RepoWiki 配置 [%d]", id.Int64()))

	// 尝试从缓存读取
	if config, ok, _ := r.cache.GetConfig(ctx, id.Int64()); ok {
		r.log.Info(ctx, fmt.Sprintf("GetByID - 缓存命中 [%d]", id.Int64()))
		return config, nil
	}

	// 缓存未命中，查询数据库
	var config entity.RepoWikiConfig
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, xError.NewError(ctx, xError.NotFound, "RepoWiki 配置不存在", false, nil)
		}
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询 RepoWiki 配置失败", false, err)
	}

	// 回填缓存
	if xErr := r.cache.SetConfig(ctx, &config); xErr != nil {
		r.log.Warn(ctx, xErr.Error())
	}
	return &config, nil
}

// GetByProjectID 根据项目 ID 获取 RepoWiki 配置
//
// 参数:
//   - ctx:       上下文对象
//   - projectID: 关联的项目雪花 ID
//
// 返回值:
//   - *entity.RepoWikiConfig: 查询到的配置实体
//   - *xError.Error:           查询过程中的错误
func (r *RepoWikiConfigRepo) GetByProjectID(ctx context.Context, projectID xSnowflake.SnowflakeID) (*entity.RepoWikiConfig, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("GetByProjectID - 根据项目 ID 获取 RepoWiki 配置 [%d]", projectID.Int64()))

	var config entity.RepoWikiConfig
	if err := r.db.WithContext(ctx).Where("project_id = ?", projectID).First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, xError.NewError(ctx, xError.NotFound, "该项目未配置 RepoWiki", false, nil)
		}
		return nil, xError.NewError(ctx, xError.DatabaseError, "根据项目 ID 查询 RepoWiki 配置失败", false, err)
	}

	return &config, nil
}

// List 分页获取 RepoWiki 配置列表（按创建时间降序）
//
// 参数:
//   - ctx:  上下文对象
//   - page: 页码（从 1 开始）
//   - size: 每页数量
//
// 返回值:
//   - []*entity.RepoWikiConfig: 当前页的配置列表
//   - int64:                     符合条件的总记录数
//   - *xError.Error:             查询过程中的错误
func (r *RepoWikiConfigRepo) List(ctx context.Context, page, size int) ([]*entity.RepoWikiConfig, int64, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("List - 分页获取 RepoWiki 配置列表 [page=%d, size=%d]", page, size))

	var total int64
	if err := r.db.WithContext(ctx).Model(&entity.RepoWikiConfig{}).Count(&total).Error; err != nil {
		return nil, 0, xError.NewError(ctx, xError.DatabaseError, "统计 RepoWiki 配置数量失败", false, err)
	}

	var configs []*entity.RepoWikiConfig
	offset := (page - 1) * size
	if err := r.db.WithContext(ctx).
		Offset(offset).
		Limit(size).
		Order("created_at DESC").
		Find(&configs).Error; err != nil {
		return nil, 0, xError.NewError(ctx, xError.DatabaseError, "查询 RepoWiki 配置列表失败", false, err)
	}

	return configs, total, nil
}

// Update 更新 RepoWiki 配置，成功后刷新缓存
//
// 参数:
//   - ctx:    上下文对象
//   - config: 待更新的配置实体（需包含完整字段）
//
// 返回值:
//   - *xError.Error: 更新过程中的错误
func (r *RepoWikiConfigRepo) Update(ctx context.Context, config *entity.RepoWikiConfig) *xError.Error {
	r.log.Info(ctx, fmt.Sprintf("Update - 更新 RepoWiki 配置 [id=%d]", config.ID.Int64()))

	if err := r.db.WithContext(ctx).Save(config).Error; err != nil {
		r.log.Warn(ctx, err.Error())
		return xError.NewError(ctx, xError.DatabaseError, "更新 RepoWiki 配置失败", false, err)
	}

	// 刷新缓存
	if xErr := r.cache.SetConfig(ctx, config); xErr != nil {
		r.log.Warn(ctx, xErr.Error())
	}
	return nil
}

// Delete 删除 RepoWiki 配置，成功后清除关联缓存
//
// 参数:
//   - ctx: 上下文对象
//   - id:  待删除的配置雪花 ID
//
// 返回值:
//   - *xError.Error: 删除过程中的错误
func (r *RepoWikiConfigRepo) Delete(ctx context.Context, id xSnowflake.SnowflakeID) *xError.Error {
	r.log.Info(ctx, fmt.Sprintf("Delete - 删除 RepoWiki 配置 [%d]", id.Int64()))

	// 执行删除
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.RepoWikiConfig{})
	if result.Error != nil {
		r.log.Warn(ctx, result.Error.Error())
		return xError.NewError(ctx, xError.DatabaseError, "删除 RepoWiki 配置失败", false, result.Error)
	}
	if result.RowsAffected == 0 {
		return xError.NewError(ctx, xError.NotFound, "RepoWiki 配置不存在", false, nil)
	}

	// 清除缓存
	r.cache.DeleteConfig(ctx, id.Int64())
	return nil
}
