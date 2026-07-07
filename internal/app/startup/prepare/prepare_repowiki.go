package prepare

import (
	"log/slog"

	xUtil "github.com/bamboo-services/bamboo-base-go/common/utility"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"gorm.io/datatypes"
)

// prepareRepoWiki 为已存在的 RepoWikiConfig 补充 Webhook 凭据
//
// 幂等迁移：查询所有 WebhookToken 为空的配置，为每条记录生成
// Token + Secret + 空分支列表（[]）。已存在凭据的记录不受影响。
// 仅在升级到 Webhook 版本时需要执行一次。
func (p *Prepare) prepareRepoWiki() {
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
