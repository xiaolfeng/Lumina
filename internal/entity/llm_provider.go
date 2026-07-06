// Package entity defines GORM database entity models.

package entity

import (
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

// LlmProvider LLM Provider 配置表，存储大模型供应商连接信息
type LlmProvider struct {
	xModels.BaseEntity              // 基础实体（ID、创建时间、更新时间）
	Name               string       `gorm:"type:varchar(128);not null;uniqueIndex;comment:Provider名称" json:"name"`                // Provider名称
	Protocol           string       `gorm:"type:varchar(32);not null;comment:协议类型(openai/anthropic)" json:"protocol"`           // 协议类型
	BaseURL            string       `gorm:"type:varchar(512);not null;default:'';comment:自定义API端点" json:"base_url"`             // 自定义API端点
	APIKeyEncrypted    string       `gorm:"type:text;not null;comment:AES-GCM加密的API密钥" json:"-"`                               // AES-GCM加密的API密钥
	IsActive           bool         `gorm:"not null;default:true;index;comment:是否启用" json:"is_active"`                          // 是否启用
	Description        string       `gorm:"type:varchar(255);not null;default:'';comment:描述说明" json:"description"`               // 描述说明
}

// GetGene 返回LlmProvider实体的雪花算法基因编号
func (l *LlmProvider) GetGene() xSnowflake.Gene {
	return bConst.GeneLlmProvider
}
