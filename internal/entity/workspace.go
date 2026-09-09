// Package entity defines GORM database entity models.

package entity

import (
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

// Workspace 工作空间表，Project 之上的组织边界（单用户拆分生活/工作）
type Workspace struct {
	xModels.BaseEntity        // 基础实体（ID、创建时间、更新时间）
	Name               string `gorm:"type:varchar(128);not null;comment:空间名称" json:"name"`            // 空间名称
	Slug               string `gorm:"type:varchar(64);not null;uniqueIndex;comment:空间标识" json:"slug"` // 空间标识
	Description        string `gorm:"type:text;comment:空间描述" json:"description"`                      // 空间描述
	Icon               string `gorm:"type:varchar(64);not null;default:'';comment:空间图标" json:"icon"`  // 空间图标
	IsDefault          bool   `gorm:"not null;default:false;index;comment:是否为默认空间" json:"is_default"` // 是否为默认空间
}

// GetGene 返回 Workspace 实体的雪花算法基因编号
func (w *Workspace) GetGene() xSnowflake.Gene {
	return bConst.GeneWorkspace
}
