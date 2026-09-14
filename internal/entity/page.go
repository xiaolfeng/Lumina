package entity

import (
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

// Page 持久化即时页面表，绑定所属项目，通过 LatestVersionID 指向当前线上生效快照
type Page struct {
	xModels.BaseEntity                        // 基础实体（ID、创建时间、更新时间）
	ProjectID          xSnowflake.SnowflakeID `gorm:"type:bigint;not null;uniqueIndex:uk_project_slug;index;comment:所属项目ID" json:"project_id"`         // 所属项目ID
	Slug               string                 `gorm:"type:varchar(64);not null;uniqueIndex:uk_project_slug;comment:项目内访问标识" json:"slug"`               // 项目内访问标识
	Title              string                 `gorm:"type:varchar(255);not null;comment:页面显示标题" json:"title"`                                          // 页面显示标题
	Description        string                 `gorm:"type:text;comment:页面描述" json:"description"`                                                       // 页面描述
	Status             string                 `gorm:"type:varchar(16);not null;default:published;index;comment:页面状态 published/archived" json:"status"` // 页面状态 published/archived
	AccessMode         string                 `gorm:"type:varchar(16);not null;default:public;comment:访问权限策略 public/password" json:"access_mode"`      // 访问权限策略 public/password
	PasswordHash       string                 `gorm:"type:varchar(128);comment:访问密码哈希" json:"-"`                                                       // 访问密码哈希
	LatestVersionID    xSnowflake.SnowflakeID `gorm:"type:bigint;not null;default:0;index;comment:当前线上生效版本指针" json:"latest_version_id"`                // 当前线上生效版本指针
}

// GetGene 返回 Page 实体的雪花算法基因编号
func (p *Page) GetGene() xSnowflake.Gene {
	return bConst.GenePage
}
