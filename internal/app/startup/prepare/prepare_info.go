package prepare

import (
	"log/slog"

	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
)

// prepareInfo 初始化 Info 表认证种子数据
//
// 写入站主信息与系统初始化标记（认证模块内部状态，不对外暴露为设置项）。
// 使用 FirstOrCreate 保证幂等性，已存在的 key 不会被覆盖。
// auth.is-initial 默认值为 "true"，表示系统处于未初始化状态。
func (p *Prepare) prepareInfo() {
	infos := []entity.Info{
		{
			Key:         bConst.InfoKeyOwnerUsername,
			Value:       "",
			Description: "站主用户名",
		},
		{
			Key:         bConst.InfoKeyOwnerEmail,
			Value:       "",
			Description: "站主邮箱",
		},
		{
			Key:         bConst.InfoKeyOwnerPassword,
			Value:       "",
			Description: "站主密码（加密存储）",
		},
		{
			Key:         bConst.InfoKeyAuthIsInitial,
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
