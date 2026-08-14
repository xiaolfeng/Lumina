package logic

import (
	"context"
	"time"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xCtxUtil "github.com/bamboo-services/bamboo-base-go/major/utility/context"
	apiDashboard "github.com/xiaolfeng/Lumina/api/dashboard"
	"github.com/xiaolfeng/Lumina/internal/repository"
)

// dashboardRepo 看板模块依赖的仓储集合
type dashboardRepo struct {
	dashboard *repository.DashboardRepo
}

// DashboardLogic 看板业务逻辑层，负责跨模块聚合统计编排。
type DashboardLogic struct {
	logic
	repo dashboardRepo
}

// NewDashboardLogic 创建看板业务逻辑层实例
func NewDashboardLogic(ctx context.Context) *DashboardLogic {
	db := xCtxUtil.MustGetDB(ctx)

	return &DashboardLogic{
		logic: logic{
			log: xLog.WithName(xLog.NamedLOGC, "DashboardLogic"),
		},
		repo: dashboardRepo{
			dashboard: repository.NewDashboardRepo(db),
		},
	}
}

// Overview 聚合各模块统计，返回看板概览数据。
func (l *DashboardLogic) Overview(ctx context.Context) (*apiDashboard.OverviewResponse, *xError.Error) {
	l.log.Info(ctx, "Overview - 聚合看板概览统计")

	resp := &apiDashboard.OverviewResponse{
		RecentPreviews: make([]apiDashboard.RecentPreviewItem, 0),
	}

	// 令牌统计
	tokenTotal, tokenActive, xErr := l.repo.dashboard.TokenCounts(ctx)
	if xErr != nil {
		return nil, xErr
	}
	resp.Tokens = apiDashboard.TokenStats{Total: tokenTotal, Active: tokenActive}

	// 项目总数
	projectTotal, xErr := l.repo.dashboard.ProjectCount(ctx)
	if xErr != nil {
		return nil, xErr
	}
	resp.Projects = projectTotal

	// 问答会话统计
	qaTotal, qaActive, qaExpired, qaDeleted, qaPending, xErr := l.repo.dashboard.QaSessionCounts(ctx)
	if xErr != nil {
		return nil, xErr
	}
	resp.Qa = apiDashboard.QaStats{
		Total:            qaTotal,
		Active:           qaActive,
		Expired:          qaExpired,
		Deleted:          qaDeleted,
		PendingQuestions: qaPending,
	}

	// 预览会话统计
	previewTotal, previewActive, previewFiles, xErr := l.repo.dashboard.PreviewCounts(ctx)
	if xErr != nil {
		return nil, xErr
	}
	resp.Preview = apiDashboard.PreviewStats{
		Total:  previewTotal,
		Active: previewActive,
		Files:  previewFiles,
	}

	// RepoWiki 统计
	configs, versions, completed, generating, xErr := l.repo.dashboard.RepoWikiCounts(ctx)
	if xErr != nil {
		return nil, xErr
	}
	resp.RepoWiki = apiDashboard.RepoWikiStats{
		Configs:    configs,
		Versions:   versions,
		Completed:  completed,
		Generating: generating,
	}

	// 最近预览列表
	rows, xErr := l.repo.dashboard.RecentPreviews(ctx, 5)
	if xErr != nil {
		return nil, xErr
	}
	for _, row := range rows {
		resp.RecentPreviews = append(resp.RecentPreviews, apiDashboard.RecentPreviewItem{
			ID:        xSnowflake.SnowflakeID(row.ID).String(),
			Title:     row.Title,
			Hash:      row.Hash,
			FileCount: row.FileCount,
			Status:    row.Status,
			UpdatedAt: row.UpdatedAt.Format(time.RFC3339),
		})
	}

	return resp, nil
}
