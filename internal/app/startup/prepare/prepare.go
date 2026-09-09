package prepare

import (
	"context"
	"fmt"

	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xCtxUtil "github.com/bamboo-services/bamboo-base-go/major/utility/context"
	"github.com/xiaolfeng/Lumina/internal/repository"
	"gorm.io/gorm"
)

type Prepare struct {
	log           *xLog.LogNamedLogger
	db            *gorm.DB
	ctx           context.Context
	workspaceRepo *repository.WorkspaceRepo
}

func New(log *xLog.LogNamedLogger, ctx context.Context) *Prepare {
	db := xCtxUtil.MustGetDB(ctx)
	rdb := xCtxUtil.MustGetRDB(ctx)
	return &Prepare{
		log:           log,
		db:            db,
		ctx:           ctx,
		workspaceRepo: repository.NewWorkspaceRepo(db, rdb),
	}
}

func (p *Prepare) Prepare() error {
	p.prepareInfo()
	if err := p.prepareWorkspace(); err != nil {
		return err
	}
	p.prepareProject()
	p.prepareQaHash()
	p.prepareLlm()
	p.prepareRepoWiki()
	p.prepareSettings()
	p.preparePreviewExpiresAt()
	return nil
}

func (p *Prepare) prepareWorkspace() error {
	defaultWorkspace, xErr := p.workspaceRepo.EnsureDefault(p.ctx)
	if xErr != nil {
		p.log.Warn(p.ctx, "保证默认空间失败: "+xErr.Error())
		return fmt.Errorf("保证默认空间失败: %s", xErr.Error())
	}
	p.log.Info(p.ctx, fmt.Sprintf("默认空间就绪 [%s]", defaultWorkspace.ID.String()))

	affected, xErr := p.workspaceRepo.BackfillProjectWorkspace(p.ctx, defaultWorkspace.ID)
	if xErr != nil {
		p.log.Warn(p.ctx, "回填项目所属空间失败: "+xErr.Error())
		return fmt.Errorf("回填项目所属空间失败: %s", xErr.Error())
	}
	if affected > 0 {
		p.log.Info(p.ctx, fmt.Sprintf("已回填项目所属空间 [%d]", affected))
	}

	if extras, xErr := p.workspaceRepo.DeduplicateDefaultFlags(p.ctx); xErr != nil {
		p.log.Warn(p.ctx, "修复重复默认空间失败: "+xErr.Error())
	} else if extras > 0 {
		p.log.Warn(p.ctx, fmt.Sprintf("发现多行默认空间，已保留最早一行并关闭其余 [%d]", extras))
	}

	if err := p.workspaceRepo.EnsureOneDefaultIndex(p.ctx); err != nil {
		p.log.Warn(p.ctx, "创建默认空间部分唯一索引失败: "+err.Error())
	}

	if err := p.workspaceRepo.SetProjectWorkspaceNotNull(p.ctx); err != nil {
		p.log.Warn(p.ctx, "设置 projects.workspace_id NOT NULL 失败: "+err.Error())
	}

	return nil
}
