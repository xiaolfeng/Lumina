package prepare

import (
	"log/slog"

	xUtil "github.com/bamboo-services/bamboo-base-go/common/utility"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"gorm.io/datatypes"
)

// prepareRepoWiki 为已存在的 RepoWikiConfig 补充 Webhook 凭据 + 选中版本迁移
func (p *Prepare) prepareRepoWiki() {
	p.backfillWebhookCredentials()
	p.migrateSelectedVersionID()
}

// backfillWebhookCredentials 为缺少 Webhook 凭据的配置补充 Token + Secret + 空分支列表。
//
// 幂等：仅处理 webhook_token 为空的记录，已存在凭据的不受影响。
func (p *Prepare) backfillWebhookCredentials() {
	var configs []entity.RepoWikiConfig
	if err := p.db.WithContext(p.ctx).
		Where("webhook_token = ? OR webhook_token IS NULL", "").
		Find(&configs).Error; err != nil {
		p.log.Warn(p.ctx, "查询缺少 Webhook 凭据的 RepoWiki 配置失败: "+err.Error())
		return
	}

	for i := range configs {
		config := &configs[i]
		config.WebhookToken = xUtil.Security().GenerateLongKey()
		config.WebhookSecret = xUtil.Security().GenerateLongKey()
		if len(config.WebhookBranches) == 0 {
			config.WebhookBranches = datatypes.JSON([]byte("[]"))
		}

		if err := p.db.WithContext(p.ctx).Save(config).Error; err != nil {
			p.log.Warn(
				p.ctx,
				"补充 RepoWiki Webhook 凭据失败: "+err.Error(),
				slog.Int64("configID", config.ID.Int64()),
			)
		}
	}
}

// migrateSelectedVersionID 为 selected_version_id 为 NULL 的配置填充最近已完成版本。
//
// 幂等：仅处理 selected_version_id IS NULL 的记录；填充后再次执行不重复处理。
// 对每个配置查找最新的 status=completed 版本（按 created_at DESC），找不到则跳过。
func (p *Prepare) migrateSelectedVersionID() {
	var configs []entity.RepoWikiConfig
	if err := p.db.WithContext(p.ctx).
		Where("selected_version_id IS NULL").
		Find(&configs).Error; err != nil {
		p.log.Warn(p.ctx, "查询未选中版本的 RepoWiki 配置失败: "+err.Error())
		return
	}

	for i := range configs {
		config := &configs[i]

		var version entity.WikiVersion
		if err := p.db.WithContext(p.ctx).
			Where("config_id = ? AND status = ?", config.ID, bConst.RepoWikiStatusCompleted).
			Order("created_at DESC").
			First(&version).Error; err != nil {
			continue // 无已完成版本，跳过
		}

		if err := p.db.WithContext(p.ctx).Model(&entity.RepoWikiConfig{}).
			Where("id = ?", config.ID).
			Update("selected_version_id", version.ID).Error; err != nil {
			p.log.Warn(
				p.ctx,
				"填充 SelectedVersionID 失败: "+err.Error(),
				slog.Int64("configID", config.ID.Int64()),
				slog.Int64("versionID", version.ID.Int64()),
			)
		}
	}
}
