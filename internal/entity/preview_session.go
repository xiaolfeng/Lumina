package entity

import (
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

// PreviewSession 预览会话表，存储前端可视化预览的工作区会话信息
type PreviewSession struct {
	xModels.BaseEntity                        // 基础实体（ID、创建时间、更新时间）
	ProjectID          xSnowflake.SnowflakeID `gorm:"type:bigint;not null;index;comment:关联项目ID" json:"project_id"`                        // 关联项目ID
	Title              string                 `gorm:"type:varchar(255);not null;comment:会话标题" json:"title"`                               // 会话标题
	Hash               string                 `gorm:"type:char(32);uniqueIndex;not null;comment:访问哈希标识" json:"hash"`                      // 访问哈希标识
	Status             string                 `gorm:"type:varchar(16);not null;default:active;comment:会话状态 active/deleted" json:"status"` // 会话状态 active/deleted
}

// GetGene 返回PreviewSession实体的雪花算法基因编号
func (p *PreviewSession) GetGene() xSnowflake.Gene {
	return bConst.GenePreviewSession
}
