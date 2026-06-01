package entity

import (
	"time"

	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
)

// User 用户实体
type User struct {
	xModels.BaseEntity
	Username     string     `gorm:"not null;type:varchar(32);uniqueIndex:uk_user_username;comment:用户名" json:"username"` // 用户名
	Email        string     `gorm:"not null;type:varchar(128);uniqueIndex:uk_user_email;comment:邮箱地址" json:"email"`     // 邮箱地址
	PasswordHash string     `gorm:"not null;type:varchar(255);comment:密码加密值" json:"-"`                                  // 密码加密值
	IsActive     bool       `gorm:"not null;default:true;comment:是否启用" json:"is_active"`                                // 是否启用
	LastLoginAt  *time.Time `gorm:"type:timestamptz;comment:最后登录时间" json:"last_login_at,omitempty"`                     // 最后登录时间
}

// GetGene 返回 User 实体的业务基因类型
func (_ *User) GetGene() xSnowflake.Gene {
	return xSnowflake.GeneUser
}
