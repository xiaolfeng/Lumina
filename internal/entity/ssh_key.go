// Package entity defines GORM database entity models.

package entity

import (
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

// SshKey SSH密钥表，存储用于Git克隆的SSH密钥对
type SshKey struct {
	xModels.BaseEntity        // 基础实体（ID、创建时间、更新时间）
	Name               string `gorm:"type:varchar(128);not null;uniqueIndex;comment:密钥名称" json:"name"`                            // 密钥名称
	Description        string `gorm:"type:text;comment:密钥描述" json:"description"`                                                  // 密钥描述
	KeyType            string `gorm:"type:varchar(16);not null;comment:密钥类型(ed25519|rsa)" json:"key_type"`                        // 密钥类型(ed25519|rsa)
	PublicKey          string `gorm:"type:text;not null;comment:OpenSSH格式公钥" json:"public_key"`                                   // OpenSSH格式公钥
	PrivateKey         string `gorm:"type:text;not null;comment:PEM格式私钥(明文存储,API永不返回)" json:"-"`                                          // PEM格式私钥(明文存储,API永不返回)
	Fingerprint        string `gorm:"type:varchar(128);not null;index;comment:SHA256指纹" json:"fingerprint"`                       // SHA256指纹
	Source             string `gorm:"type:varchar(16);not null;default:generated;comment:密钥来源(generated|imported)" json:"source"` // 密钥来源(generated|imported)
}

// GetGene 返回SshKey实体的雪花算法基因编号
func (s *SshKey) GetGene() xSnowflake.Gene {
	return bConst.GeneSSHKey
}
