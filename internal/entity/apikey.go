// Package entity defines GORM database entity models.

package entity

import (
	"time"

	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

// Apikey API密钥表，存储API访问凭证
type Apikey struct {
	xModels.BaseEntity            // 基础实体（ID、创建时间、更新时间）
	Name               string     `gorm:"type:varchar(64);not null;comment:API Key名称" json:"name"`             // API Key名称
	KeyHash            string     `gorm:"type:varchar(255);not null;comment:密钥哈希" json:"-"`                    // 密钥哈希
	KeyPrefix          string     `gorm:"type:varchar(16);not null;comment:密钥前缀" json:"key_prefix"`            // 密钥前缀
	KeySuffix          string     `gorm:"type:varchar(16);not null;comment:密钥后缀" json:"-"`                     // 密钥后缀
	Description        string     `gorm:"type:varchar(255);not null;default:'';comment:描述" json:"description"` // 描述
	ExpiresAt          *time.Time `gorm:"type:timestamptz;comment:过期时间" json:"expires_at,omitempty"`           // 过期时间
	LastUsedAt         *time.Time `gorm:"type:timestamptz;comment:最后使用时间" json:"last_used_at,omitempty"`       // 最后使用时间
	IsActive           bool       `gorm:"type:bool;not null;default:true;comment:是否启用" json:"is_active"`       // 是否启用
}

// GetGene 返回Apikey实体的雪花算法基因编号
func (a *Apikey) GetGene() xSnowflake.Gene {
	return bConst.GeneApikey
}
