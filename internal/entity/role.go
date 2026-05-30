package entity

import (
	"fmt"
	"regexp"
	"time"

	"gorm.io/gorm"
)

type RoleName string

const (
	RoleOwner  RoleName = "OWNER"
	RoleAdmin  RoleName = "ADMIN"
	RoleMember RoleName = "MEMBER"
)

func (r RoleName) String() string {
	return string(r)
}

var roleNamePattern = regexp.MustCompile(`^[A-Z_]{2,32}$`)

type Role struct {
	Name        RoleName  `gorm:"primaryKey;type:varchar(32);comment:角色名称" json:"name"`
	DisplayName string    `gorm:"not null;type:varchar(64);comment:角色显示名" json:"display_name"`
	Description string    `gorm:"not null;type:varchar(255);comment:角色描述" json:"description"`
	CreatedAt   time.Time `gorm:"not null;type:timestamptz;autoCreateTime:milli;comment:创建时间" json:"-"`
}

func (r *Role) BeforeCreate(_ *gorm.DB) error {
	if roleNamePattern.MatchString(r.Name.String()) {
		return nil
	}
	return fmt.Errorf("invalid role name: %s", r.Name)
}
