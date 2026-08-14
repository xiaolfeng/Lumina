package repository

import (
	"context"
	"fmt"
	"time"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"gorm.io/gorm"
)

// DashboardRepo 看板数据访问层，提供跨模块的聚合统计查询。
//
// 看板概览为只读聚合场景，直接对数据库执行 COUNT / 分组统计，
// 不引入缓存（统计结果时效性优先，且各模块已有独立缓存层）。
//
// 字段说明:
//   - db:  GORM 数据库实例，执行聚合查询
//   - log: 带命名空间的结构化日志记录器
type DashboardRepo struct {
	db  *gorm.DB
	log *xLog.LogNamedLogger
}

// NewDashboardRepo 创建 DashboardRepo 实例
//
// 参数说明:
//   - db: 已初始化的 GORM 数据库实例
//
// 返回值:
//   - *DashboardRepo: 配置完成的 DashboardRepo 实例指针
func NewDashboardRepo(db *gorm.DB) *DashboardRepo {
	return &DashboardRepo{
		db:  db,
		log: xLog.WithName(xLog.NamedREPO, "DashboardRepo"),
	}
}

// count 统计指定模型的记录总数，支持可选过滤条件。
func (r *DashboardRepo) count(ctx context.Context, model any, where string, args ...any) (int64, *xError.Error) {
	var total int64
	query := r.db.WithContext(ctx).Model(model)
	if where != "" {
		query = query.Where(where, args...)
	}
	if err := query.Count(&total).Error; err != nil {
		return 0, xError.NewError(ctx, xError.DatabaseError, "统计记录数量失败", false, err)
	}
	return total, nil
}

// TokenCounts 统计令牌总数与活跃令牌数。
func (r *DashboardRepo) TokenCounts(ctx context.Context) (total, active int64, xErr *xError.Error) {
	total, xErr = r.count(ctx, &entity.Apikey{}, "")
	if xErr != nil {
		return 0, 0, xErr
	}
	active, xErr = r.count(ctx, &entity.Apikey{}, "is_active = ?", true)
	if xErr != nil {
		return 0, 0, xErr
	}
	return total, active, nil
}

// ProjectCount 统计项目总数。
func (r *DashboardRepo) ProjectCount(ctx context.Context) (int64, *xError.Error) {
	return r.count(ctx, &entity.Project{}, "")
}

// QaSessionCounts 统计问答会话总数、各状态分布及待回答问题总数。
func (r *DashboardRepo) QaSessionCounts(ctx context.Context) (total, active, expired, deleted, pending int64, xErr *xError.Error) {
	total, xErr = r.count(ctx, &entity.QaSession{}, "")
	if xErr != nil {
		return 0, 0, 0, 0, 0, xErr
	}
	active, xErr = r.count(ctx, &entity.QaSession{}, "status = ?", "active")
	if xErr != nil {
		return 0, 0, 0, 0, 0, xErr
	}
	expired, xErr = r.count(ctx, &entity.QaSession{}, "status = ?", "expired")
	if xErr != nil {
		return 0, 0, 0, 0, 0, xErr
	}
	deleted, xErr = r.count(ctx, &entity.QaSession{}, "status = ?", "deleted")
	if xErr != nil {
		return 0, 0, 0, 0, 0, xErr
	}
	pending, xErr = r.count(ctx, &entity.QaQuestion{}, "status = ?", "pending")
	if xErr != nil {
		return 0, 0, 0, 0, 0, xErr
	}
	return total, active, expired, deleted, pending, nil
}

// PreviewCounts 统计预览会话总数、活跃会话数与文件总数。
func (r *DashboardRepo) PreviewCounts(ctx context.Context) (total, active, files int64, xErr *xError.Error) {
	total, xErr = r.count(ctx, &entity.PreviewSession{}, "")
	if xErr != nil {
		return 0, 0, 0, xErr
	}
	active, xErr = r.count(ctx, &entity.PreviewSession{}, "status = ?", bConst.PreviewSessionStatusActive)
	if xErr != nil {
		return 0, 0, 0, xErr
	}
	files, xErr = r.count(ctx, &entity.PreviewFile{}, "")
	if xErr != nil {
		return 0, 0, 0, xErr
	}
	return total, active, files, nil
}

// RepoWikiCounts 统计 RepoWiki 配置、版本总数及完成/生成中分布。
func (r *DashboardRepo) RepoWikiCounts(ctx context.Context) (configs, versions, completed, generating int64, xErr *xError.Error) {
	configs, xErr = r.count(ctx, &entity.RepoWikiConfig{}, "")
	if xErr != nil {
		return 0, 0, 0, 0, xErr
	}
	versions, xErr = r.count(ctx, &entity.WikiVersion{}, "")
	if xErr != nil {
		return 0, 0, 0, 0, xErr
	}
	completed, xErr = r.count(ctx, &entity.WikiVersion{}, "status = ?", bConst.RepoWikiStatusCompleted)
	if xErr != nil {
		return 0, 0, 0, 0, xErr
	}
	// 生成中 = 所有非终态（pending/cloning/scanning/analyzing/assembling）
	generating, xErr = r.count(ctx, &entity.WikiVersion{},
		"status NOT IN ?", []string{
			bConst.RepoWikiStatusCompleted,
			bConst.RepoWikiStatusFailed,
			bConst.RepoWikiStatusCancelled,
		})
	if xErr != nil {
		return 0, 0, 0, 0, xErr
	}
	return configs, versions, completed, generating, nil
}

// recentPreviewRow 最近预览聚合查询的中间结果。
type recentPreviewRow struct {
	ID        int64     `gorm:"column:id"`
	Title     string    `gorm:"column:title"`
	Hash      string    `gorm:"column:hash"`
	Status    string    `gorm:"column:status"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
	FileCount int64     `gorm:"column:file_count"`
}

// RecentPreviews 查询最近 limit 个活跃预览会话及其文件数（按更新时间降序）。
func (r *DashboardRepo) RecentPreviews(ctx context.Context, limit int) ([]recentPreviewRow, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("RecentPreviews - 查询最近预览会话 [limit=%d]", limit))

	// 通过 GORM 命名策略解析真实表名（含框架表前缀），避免硬编码
	sessionTable := r.db.NamingStrategy.TableName("PreviewSession")
	fileTable := r.db.NamingStrategy.TableName("PreviewFile")

	// 原生 SQL：GROUP BY 需包含全部非聚合列，避免 GORM 对 s.id 别名加引号导致
	// PostgreSQL 将 "s"."id" 误解析为 schema.table 的问题
	sql := fmt.Sprintf(
		`SELECT s.id, s.title, s.hash, s.status, s.updated_at, COUNT(f.id) AS file_count
		 FROM %s AS s
		 LEFT JOIN %s AS f ON f.session_id = s.id
		 WHERE s.status = ?
		 GROUP BY s.id, s.title, s.hash, s.status, s.updated_at
		 ORDER BY s.updated_at DESC
		 LIMIT ?`,
		sessionTable, fileTable,
	)

	rows := make([]recentPreviewRow, 0)
	if err := r.db.WithContext(ctx).
		Raw(sql, bConst.PreviewSessionStatusActive, limit).
		Scan(&rows).Error; err != nil {
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询最近预览会话失败", false, err)
	}
	return rows, nil
}
