package prepare

import (
	"log/slog"

	"github.com/xiaolfeng/Lumina/internal/entity"
)

// prepareQa 初始化 Q&A 配置种子数据
//
// 写入 Q&A 模块运行时所需的配置项。使用 FirstOrCreate 保证幂等性，
// 已存在的 key 不会被覆盖。
func (p *Prepare) prepareQa() {
	infos := []entity.Info{
		{
			Key:         "qa.session.ttl",
			Value:       "172800",
			Description: "Session默认TTL（秒）",
		},
		{
			Key:         "runtime.domain",
			Value:       "",
			Description: "运行时域名（用于内网判定多媒体返回策略）",
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
				"初始化 Q&A 种子数据失败: "+err.Error(),
				slog.String("key", item.Key),
			)
		}
	}
}
