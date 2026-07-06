package prepare

import (
	"log/slog"

	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
)

// prepareLlm 初始化 LLM Agent 模型分配种子数据
//
// 为每个 Agent 角色创建 llm_agent_model:{role} 键（值为空字符串，表示未配置）。
// 使用 FirstOrCreate 保证幂等性，已存在的 key 不会被覆盖。
func (p *Prepare) prepareLlm() {
	roles := []string{
		bConst.AgentRoleRepoWiki,
	}

	for _, role := range roles {
		key := bConst.LlmAgentModelKeyPrefix + role
		info := entity.Info{
			Key:         key,
			Value:       "",
			Description: "Agent 模型分配: " + role,
		}
		err := p.db.WithContext(p.ctx).
			Where("key = ?", info.Key).
			FirstOrCreate(&info).
			Error
		if err != nil {
			p.log.Warn(
				p.ctx,
				"初始化 LLM Agent 种子数据失败: "+err.Error(),
				slog.String("key", info.Key),
			)
		}
	}
}
