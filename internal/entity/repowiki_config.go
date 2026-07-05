// Package entity defines GORM database entity models.

package entity

import (
	"time"

	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

// RepoWikiConfig RepoWiki配置表，与Project表1:1关联，存储Git仓库配置和Wiki生成偏好
type RepoWikiConfig struct {
	xModels.BaseEntity                        // 基础实体（ID、创建时间、更新时间）
	ProjectID          xSnowflake.SnowflakeID `gorm:"type:bigint;not null;uniqueIndex;comment:关联项目ID" json:"project_id"`                   // 关联项目ID
	Name               string                 `gorm:"type:varchar(128);not null;comment:配置名称" json:"name"`                                 // 配置名称
	GitURL             string                 `gorm:"type:varchar(512);not null;index;comment:Git仓库地址" json:"git_url"`                     // Git仓库地址
	DefaultBranch      string                 `gorm:"type:varchar(128);not null;default:main;comment:默认分支" json:"default_branch"`          // 默认分支
	LocalPath          string                 `gorm:"type:varchar(512);comment:本地克隆路径" json:"local_path"`                                  // 本地克隆路径
	DefaultLanguage    string                 `gorm:"type:varchar(16);not null;default:zh;comment:默认Wiki语言" json:"default_language"`       // 默认Wiki语言
	SSHKeyEncrypted    string                 `gorm:"type:text;comment:AES加密的SSH私钥" json:"ssh_key_encrypted,omitempty"`                    // AES加密的SSH私钥
	WikiPasswordHash   string                 `gorm:"type:varchar(128);comment:Wiki访问密码bcrypt哈希" json:"wiki_password_hash,omitempty"`      // Wiki访问密码bcrypt哈希
	Status             string                 `gorm:"type:varchar(16);not null;default:pending;index;comment:当前分析状态（快速查询用）" json:"status"` // 当前分析状态（快速查询用）
	LastAccessedAt     *time.Time             `gorm:"type:timestamptz;index;comment:最后访问时间" json:"last_accessed_at,omitempty"`             // 最后访问时间
}

// GetGene 返回RepoWikiConfig实体的雪花算法基因编号
func (r *RepoWikiConfig) GetGene() xSnowflake.Gene {
	return bConst.GeneRepoWikiConfig
}
