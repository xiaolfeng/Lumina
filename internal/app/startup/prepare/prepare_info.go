package prepare

import (
	"log/slog"

	"github.com/xiaolfeng/Lumina/internal/entity"
)

// prepareInfo 初始化 Info 表种子数据
//
// 写入站点基础配置和站主信息。使用 FirstOrCreate 保证幂等性，
// 已存在的 key 不会被覆盖。is_initial 默认值为 "true"，表示系统处于未初始化状态。
func (p *Prepare) prepareInfo() {
	infos := []entity.Info{
		{
			Key:         "site_name",
			Value:       "Lumina",
			Description: "站点名称",
		},
		{
			Key:         "site_description",
			Value:       "赋予 AI 深度代码认知与长期记忆的知识中枢",
			Description: "站点描述",
		},
		{
			Key:         "owner_username",
			Value:       "",
			Description: "站主用户名",
		},
		{
			Key:         "owner_email",
			Value:       "",
			Description: "站主邮箱",
		},
		{
			Key:         "owner_password",
			Value:       "",
			Description: "站主密码（加密存储）",
		},
		{
			Key:         "is_initial",
			Value:       "true",
			Description: "系统是否为初始状态（true=未初始化，false=已初始化）",
		},
	}

	for _, info := range infos {
		item := info
		err := p.db.WithContext(p.ctx).
			Where("key = ?", item.Key).
			FirstOrCreate(&item).
			Error
		if err != nil {
			p.log.Warn(
				p.ctx,
				"初始化 Info 种子数据失败: "+err.Error(),
				slog.String("key", item.Key),
			)
		}
	}
}
