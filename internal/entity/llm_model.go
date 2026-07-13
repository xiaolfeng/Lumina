// Package entity defines GORM database entity models.

package entity

import (
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

// LlmModel LLM 模型配置表，存储具体模型参数（关联到 Provider）
type LlmModel struct {
	xModels.BaseEntity                        // 基础实体（ID、创建时间、更新时间）
	ProviderID         xSnowflake.SnowflakeID `gorm:"type:bigint;not null;index;comment:关联Provider ID" json:"provider_id"`    // 关联Provider ID
	ModelName          string                 `gorm:"type:varchar(128);not null;comment:模型标识(如gpt-4o)" json:"model_name"`     // 模型标识
	DisplayName        string                 `gorm:"type:varchar(128);not null;comment:显示名称" json:"display_name"`            // 显示名称
	MaxTokens          int64                  `gorm:"not null;default:32000;comment:单次响应最大输出Token数" json:"max_tokens"`        // 单次响应最大输出Token数
	ContextWindow      int64                  `gorm:"not null;default:128000;comment:模型上下文窗口大小(Token)" json:"context_window"` // 模型上下文窗口大小
	Temperature        float64                `gorm:"not null;default:0.3;comment:生成温度" json:"temperature"`                   // 生成温度
	IsActive           bool                   `gorm:"not null;default:true;index;comment:是否启用" json:"is_active"`              // 是否启用
	Description        string                 `gorm:"type:varchar(255);not null;default:'';comment:描述说明" json:"description"`  // 描述说明
}

// GetGene 返回LlmModel实体的雪花算法基因编号
func (l *LlmModel) GetGene() xSnowflake.Gene {
	return bConst.GeneLlmModel
}
