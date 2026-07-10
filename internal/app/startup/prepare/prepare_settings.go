package prepare

import (
	"log/slog"

	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
)

// prepareSettings 初始化系统设置种子数据
//
// 遍历 SettingKeyDefs，使用 FirstOrCreate 保证幂等性，
// 已存在的 key 不会被覆盖。
func (p *Prepare) prepareSettings() {
	for _, def := range bConst.SettingKeyDefs {
		item := entity.Info{
			Key:         def.Key,
			Value:       def.Default,
			Description: def.Description,
		}
		err := p.db.WithContext(p.ctx).
			Where("key = ?", item.Key).
			FirstOrCreate(&item).
			Error
		if err != nil {
			p.log.Warn(
				p.ctx,
				"初始化 Setting 种子数据失败: "+err.Error(),
				slog.String("key", item.Key),
			)
		}
	}
}
