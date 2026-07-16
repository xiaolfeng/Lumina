// Package entity defines GORM database entity models.

package entity

import (
	"time"

	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"gorm.io/datatypes"
)

// RepoWikiConfig RepoWiki配置表，与Project表1:1关联，存储Git仓库配置和Wiki生成偏好
type RepoWikiConfig struct {
	xModels.BaseEntity                        // 基础实体（ID、创建时间、更新时间）
	ProjectID          xSnowflake.SnowflakeID `gorm:"type:bigint;not null;uniqueIndex;comment:关联项目ID" json:"project_id"`                   // 关联项目ID
	GitURL             string                 `gorm:"type:varchar(512);not null;index;comment:Git仓库地址" json:"git_url"`                     // Git仓库地址
	DefaultBranch      string                 `gorm:"type:varchar(128);not null;default:main;comment:默认分支" json:"default_branch"`          // 默认分支
	LocalPath          string                 `gorm:"type:varchar(512);comment:本地克隆路径" json:"local_path"`                                  // 本地克隆路径
	DefaultLanguage    string                 `gorm:"type:varchar(16);not null;default:zh;comment:默认Wiki语言" json:"default_language"`       // 默认Wiki语言
	SSHKeyID          *xSnowflake.SnowflakeID `gorm:"type:bigint;index;comment:关联SSH密钥ID" json:"ssh_key_id,omitempty"`                            // 关联SSH密钥ID
	WikiPasswordHash   string                 `gorm:"type:varchar(128);comment:Wiki访问密码bcrypt哈希" json:"wiki_password_hash,omitempty"`      // Wiki访问密码bcrypt哈希
	Status             string                 `gorm:"type:varchar(16);not null;default:pending;index;comment:当前分析状态（快速查询用）" json:"status"` // 当前分析状态（快速查询用）
	SelectedVersionID  *xSnowflake.SnowflakeID `gorm:"type:bigint;index;comment:当前选中的Wiki版本ID" json:"selected_version_id,omitempty"`       // 当前选中的Wiki版本ID
	LastAccessedAt     *time.Time             `gorm:"type:timestamptz;index;comment:最后访问时间" json:"last_accessed_at,omitempty"`             // 最后访问时间
	WebhookToken       string                 `gorm:"type:varchar(128);uniqueIndex;comment:Webhook访问令牌" json:"webhook_token,omitempty"`      // Webhook访问令牌
	WebhookSecret      string                 `gorm:"type:varchar(128);comment:Webhook签名密钥" json:"webhook_secret,omitempty"`                // Webhook签名密钥
	WebhookBranches    datatypes.JSON         `gorm:"type:jsonb;comment:Webhook监听分支列表" json:"webhook_branches,omitempty"`                     // Webhook监听分支列表
	CustomPrompt       string                 `gorm:"type:text;comment:项目级自定义提示词" json:"custom_prompt,omitempty"`                          // 项目级自定义提示词
}

// GetGene 返回RepoWikiConfig实体的雪花算法基因编号
func (r *RepoWikiConfig) GetGene() xSnowflake.Gene {
	return bConst.GeneRepoWikiConfig
}
