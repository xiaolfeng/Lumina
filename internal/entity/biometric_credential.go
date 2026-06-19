package entity

import (
	"time"

	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

// BiometricCredential WebAuthn 生物特征凭证表，存储用户设备注册的通行密钥信息
type BiometricCredential struct {
	xModels.BaseEntity                                                                                        // 基础实体（ID、创建时间、更新时间）
	CredentialID   []byte     `gorm:"type:bytea;not null;uniqueIndex;comment:WebAuthn 凭证 ID" json:"-"`          // WebAuthn 凭证 ID
	PublicKey      []byte     `gorm:"type:bytea;not null;comment:公钥" json:"-"`                                 // 公钥
	AAGUID         string     `gorm:"type:varchar(64);comment:认证器型号标识" json:"aaguid"`                         // 认证器型号标识
	SignCount      uint32     `gorm:"not null;default:0;comment:签名计数器" json:"sign_count"`                       // 签名计数器
	DeviceName     string     `gorm:"type:varchar(128);comment:设备名称" json:"device_name"`                        // 设备名称
	TransportTypes string     `gorm:"type:varchar(256);comment:传输类型" json:"transport_types"`                    // 传输类型
	LastUsedAt     *time.Time `gorm:"type:timestamptz;comment:最后使用时间" json:"last_used_at,omitempty"`             // 最后使用时间
}

// GetGene 返回BiometricCredential实体的雪花算法基因编号
func (b *BiometricCredential) GetGene() xSnowflake.Gene {
	return bConst.GeneBiometricCredential
}
