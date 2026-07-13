package prepare

import (
	"errors"
	"log/slog"

	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"gorm.io/gorm"
)

// prepareLlm 初始化 LLM Agent 模型分配种子数据
//
// 为每个 Agent 角色创建 llm_agent_model:{role} 键（值为空字符串，表示未配置）。
// 使用 FirstOrCreate 保证幂等性，已存在的 key 不会被覆盖。
func (p *Prepare) prepareLlm() {
	roles := bConst.AgentRolesRepoWiki

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

	p.migrateOldAgentModelKey()
}

// migrateOldAgentModelKey 将旧键 llm_agent_model:repowiki 的值迁移到 llm_agent_model:repowiki:coordinator
//
// 迁移规则：旧键不存在或值为空时不处理；新键已存在且非空时不覆盖；否则将旧值复制到新键。
// 旧键保留不删除，作为 coordinator 的兼容别名。
func (p *Prepare) migrateOldAgentModelKey() {
	const legacyRepoWikiRole = "repowiki"
	oldKey := bConst.LlmAgentModelKeyPrefix + legacyRepoWikiRole
	newKey := bConst.LlmAgentModelKeyPrefix + bConst.AgentRoleRepoWikiCoordinator

	var oldInfo entity.Info
	if err := p.db.WithContext(p.ctx).Where("key = ?", oldKey).First(&oldInfo).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return
		}
		p.log.Warn(p.ctx, "读取旧 LLM Agent 模型键失败: "+err.Error(), slog.String("key", oldKey))
		return
	}
	if oldInfo.Value == "" {
		return
	}

	var newInfo entity.Info
	err := p.db.WithContext(p.ctx).Where("key = ?", newKey).First(&newInfo).Error
	if err == nil && newInfo.Value != "" {
		return
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		p.log.Warn(p.ctx, "读取新 LLM Agent 模型键失败: "+err.Error(), slog.String("key", newKey))
		return
	}

	target := entity.Info{
		Key:         newKey,
		Value:       oldInfo.Value,
		Description: "Agent 模型分配: " + bConst.AgentRoleRepoWikiCoordinator,
	}
	if err := p.db.WithContext(p.ctx).
		Where("key = ?", newKey).
		Assign(target).
		FirstOrCreate(&target).Error; err != nil {
		p.log.Warn(
			p.ctx,
			"迁移旧 LLM Agent 模型键失败: "+err.Error(),
			slog.String("oldKey", oldKey),
			slog.String("newKey", newKey),
		)
	}
}
