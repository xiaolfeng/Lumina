// Package entity defines GORM database entity models.

package entity

import (
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

// Project 项目表，存储用户项目基本信息
type Project struct {
	xModels.BaseEntity                                                                                    // 基础实体（ID、创建时间、更新时间）
	Name        string   `gorm:"type:varchar(128);not null;uniqueIndex;comment:项目名称" json:"name"`    // 项目名称
	AliasName   string   `gorm:"type:varchar(255);comment:项目别名" json:"alias_name"`                 // 项目别名
	MatchPath   []string `gorm:"type:json;serializer:json;comment:项目路径匹配列表" json:"match_path"`     // 项目路径匹配列表
	Description string   `gorm:"type:text;comment:项目描述" json:"description"`                            // 项目描述
}

// GetGene 返回Project实体的雪花算法基因编号
func (p *Project) GetGene() xSnowflake.Gene {
	return bConst.GeneProject
}
