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

// cacheTTLProject 项目缓存过期时间（Cache-Aside 模式）
const cacheTTLProject = 30 * time.Minute

// ProjectRepo 项目数据访问层，提供完整 CRUD 操作与 Redis 缓存层
//
// 缓存策略采用 Cache-Aside 模式：读取时优先命中缓存，未命中则回源数据库并回填缓存；
// 写入/更新时同步刷新缓存；删除时清除所有关联缓存键。
// 缓存读写委托给 ProjectCache（位于 repository/cache 子层），通过 constant.RedisKey
// 统一管理缓存键。
//
// 字段说明:
//   - db:    GORM 数据库实例，执行持久化操作
//   - cache: 项目多维度缓存管理器（ID/Name/Alias/MatchPath）
//   - log:   带命名空间的结构化日志记录器
type ProjectRepo struct {
	db    *gorm.DB
	cache *cache.ProjectCache
	log   *xLog.LogNamedLogger
}

// NewProjectRepo 创建 ProjectRepo 实例
//
// 参数说明:
//   - db:  已初始化的 GORM 数据库实例
//   - rdb: 已初始化的 Redis 客户端实例（用于构造缓存管理器）
//
// 返回值:
//   - *ProjectRepo: 配置完成的 ProjectRepo 实例指针
func NewProjectRepo(db *gorm.DB, rdb *redis.Client) *ProjectRepo {
	return &ProjectRepo{
		db: db,
		cache: &cache.ProjectCache{
			Cache: &xCache.Cache{RDB: rdb, TTL: cacheTTLProject},
		},
		log: xLog.WithName(xLog.NamedREPO, "ProjectRepo"),
	}
}

// Create 创建项目，成功后同步写入缓存
//
// 参数:
//   - ctx:     上下文对象
//   - project: 待创建的项目实体（ID 由雪花算法自动生成）
//
// 返回值:
//   - *xError.Error: 创建过程中的错误
func (r *ProjectRepo) Create(ctx context.Context, project *entity.Project) *xError.Error {
	r.log.Info(ctx, fmt.Sprintf("Create - 创建项目 [%s]", project.Name))

	if err := r.db.WithContext(ctx).Create(project).Error; err != nil {
		r.log.Warn(ctx, err.Error())
		return xError.NewError(ctx, xError.DatabaseError, "创建项目失败", false, err)
	}

	if err := r.cache.SetProject(ctx, project); err != nil {
		r.log.Warn(ctx, err.Error())
	}
	return nil
}

// GetByID 根据 ID 获取项目，优先读取缓存（Cache-First）
//
// 缓存命中时直接反序列化返回；未命中则查询数据库并回填缓存。
//
// 参数:
//   - ctx: 上下文对象
//   - id:  项目雪花 ID
//
// 返回值:
//   - *entity.Project: 查询到的项目实体
//   - *xError.Error:   查询过程中的错误
func (r *ProjectRepo) GetByID(ctx context.Context, id xSnowflake.SnowflakeID) (*entity.Project, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("GetByID - 根据ID获取项目 [%d]", id.Int64()))

	// 尝试从缓存读取
	if project, ok, _ := r.cache.GetByID(ctx, id.Int64()); ok {
		r.log.Info(ctx, fmt.Sprintf("GetByID - 缓存命中 [%d]", id.Int64()))
		return project, nil
	}

	// 缓存未命中，查询数据库
	var project entity.Project
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, xError.NewError(ctx, xError.NotFound, "项目不存在", false, nil)
		}
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询项目失败", false, err)
	}

	// 回填缓存
	if err := r.cache.SetProject(ctx, &project); err != nil {
		r.log.Warn(ctx, err.Error())
	}
	return &project, nil
}

// GetByName 根据项目名称获取项目
//
// 参数:
//   - ctx:  上下文对象
//   - name: 项目名称（唯一索引）
//
// 返回值:
//   - *entity.Project: 查询到的项目实体
//   - *xError.Error:   查询过程中的错误
func (r *ProjectRepo) GetByName(ctx context.Context, name string) (*entity.Project, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("GetByName - 根据名称获取项目 [%s]", name))

	var project entity.Project
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, xError.NewError(ctx, xError.NotFound, "项目不存在", false, nil)
		}
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询项目失败", false, err)
	}

	return &project, nil
}

// List 分页获取项目列表（按创建时间降序）
//
// 参数:
//   - ctx:  上下文对象
//   - page: 页码（从 1 开始）
//   - size: 每页数量
//
// 返回值:
//   - []*entity.Project: 当前页的项目列表
//   - int64:             符合条件的总记录数
//   - *xError.Error:     查询过程中的错误
func (r *ProjectRepo) List(ctx context.Context, page, size int) ([]*entity.Project, int64, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("List - 分页获取项目列表 [page=%d, size=%d]", page, size))

	var total int64
	if err := r.db.WithContext(ctx).Model(&entity.Project{}).Count(&total).Error; err != nil {
		return nil, 0, xError.NewError(ctx, xError.DatabaseError, "统计项目数量失败", false, err)
	}

	var projects []*entity.Project
	offset := (page - 1) * size
	if err := r.db.WithContext(ctx).
		Offset(offset).
		Limit(size).
		Order("created_at DESC").
		Find(&projects).Error; err != nil {
		return nil, 0, xError.NewError(ctx, xError.DatabaseError, "查询项目列表失败", false, err)
	}

	return projects, total, nil
}

// Update 更新项目，成功后刷新缓存
//
// 参数:
//   - ctx:     上下文对象
//   - project: 待更新的项目实体（需包含完整字段）
//
// 返回值:
//   - *xError.Error: 更新过程中的错误
func (r *ProjectRepo) Update(ctx context.Context, project *entity.Project) *xError.Error {
	r.log.Info(ctx, fmt.Sprintf("Update - 更新项目 [%s]", project.Name))

	if err := r.db.WithContext(ctx).Save(project).Error; err != nil {
		r.log.Warn(ctx, err.Error())
		return xError.NewError(ctx, xError.DatabaseError, "更新项目失败", false, err)
	}

	// 刷新缓存
	if err := r.cache.SetProject(ctx, project); err != nil {
		r.log.Warn(ctx, err.Error())
	}
	return nil
}

// Delete 删除项目，成功后清除所有关联缓存
//
// 参数:
//   - ctx: 上下文对象
//   - id:  待删除的项目雪花 ID
//
// 返回值:
//   - *xError.Error: 删除过程中的错误
func (r *ProjectRepo) Delete(ctx context.Context, id xSnowflake.SnowflakeID) *xError.Error {
	r.log.Info(ctx, fmt.Sprintf("Delete - 删除项目 [%d]", id.Int64()))

	// 先获取项目信息以便清除缓存
	var project entity.Project
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return xError.NewError(ctx, xError.NotFound, "项目不存在", false, nil)
		}
		return xError.NewError(ctx, xError.DatabaseError, "查询待删除项目失败", false, err)
	}

	// 执行删除
	if err := r.db.WithContext(ctx).Delete(&entity.Project{}, id).Error; err != nil {
		r.log.Warn(ctx, err.Error())
		return xError.NewError(ctx, xError.DatabaseError, "删除项目失败", false, err)
	}

	// 清除缓存
	r.cache.DeleteProject(ctx, &project)
	return nil
}

// FindByAliasName 根据别名查询项目
//
// 使用 PostgreSQL JSON 包含查询（@>）匹配 AliasName 数组中的指定别名。
//
// 参数:
//   - ctx:  上下文对象
//   - alias: 待匹配的项目别名
//
// 返回值:
//   - *entity.Project: 查询到的项目实体
//   - *xError.Error:   查询过程中的错误
func (r *ProjectRepo) FindByAliasName(ctx context.Context, alias string) (*entity.Project, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("FindByAliasName - 根据别名查询项目 [%s]", alias))

	var project entity.Project
	if err := r.db.WithContext(ctx).
		Where("alias_name @> ?", fmt.Sprintf(`["%s"]`, alias)).
		First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, xError.NewError(ctx, xError.NotFound, "项目不存在", false, nil)
		}
		return nil, xError.NewError(ctx, xError.DatabaseError, "根据别名查询项目失败", false, err)
	}

	return &project, nil
}

// FindByMatchPath 根据路径匹配查询项目
//
// 使用 PostgreSQL JSON 展开匹配：检查 MatchPath 数组中是否有任一元素是
// 查询路径的前缀（即项目路径是查询路径的父目录或匹配路径）。
// 例如：MatchPath=["/home/user/Lumina"] 可以匹配 "/home/user/Lumina/src/main.go"
//
// 参数:
//   - ctx:  上下文对象
//   - path: 待匹配的路径
//
// 返回值:
//   - *entity.Project: 第一个匹配的项目实体
//   - *xError.Error:   查询过程中的错误
func (r *ProjectRepo) FindByMatchPath(ctx context.Context, path string) (*entity.Project, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("FindByMatchPath - 根据路径匹配项目 [%s]", path))

	var project entity.Project
	if err := r.db.WithContext(ctx).
		Where("EXISTS (SELECT 1 FROM jsonb_array_elements_text(match_path) AS elem WHERE ? LIKE (elem || '%'))", path).
		First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, xError.NewError(ctx, xError.NotFound, "项目不存在", false, nil)
		}
		return nil, xError.NewError(ctx, xError.DatabaseError, "根据路径匹配项目失败", false, err)
	}

	return &project, nil
}
