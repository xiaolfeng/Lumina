package entity

import (
	"time"

	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"

	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

// Apikey AI Agent MCP SSE 认证密钥实体
//
// 存储加密后的密钥值，单连接专用。
// 密钥原文由 crypto/rand 生成随机字节后经 base64 编码（格式: lumi_ + base64），
// 使用 bcrypt 哈希后存入 KeyHash 字段。
type Apikey struct {
	xModels.BaseEntity
	Name        string     `gorm:"not null;type:varchar(64);comment:密钥名称标识" json:"name"`                             // 密钥名称标识
	KeyHash     string     `gorm:"not null;type:varchar(255);uniqueIndex:uk_apikey_key_hash;comment:密钥加密值" json:"-"` // 密钥加密值
	KeyPrefix   string     `gorm:"not null;type:varchar(16);comment:密钥前缀" json:"key_prefix"`                         // 密钥前缀
	IsActive    bool       `gorm:"not null;default:true;comment:是否启用" json:"is_active"`                              // 是否启用
	LastUsedAt  *time.Time `gorm:"type:timestamptz;comment:最后使用时间" json:"last_used_at,omitempty"`                    // 最后使用时间
	ExpiresAt   *time.Time `gorm:"type:timestamptz;comment:过期时间" json:"expires_at,omitempty"`                        // 过期时间
	Description string     `gorm:"not null;type:varchar(255);default:'';comment:密钥用途描述" json:"description"`          // 密钥用途描述
}

// GetGene 返回 Apikey 实体的业务基因类型
func (_ *Apikey) GetGene() xSnowflake.Gene {
	return bConst.GeneApikey
}
