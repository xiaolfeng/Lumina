package entity

import (
	"time"

	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

// QaSession QA会话表，存储Agent与用户的问答会话信息
type QaSession struct {
	xModels.BaseEntity                                                                                    // 基础实体（ID、创建时间、更新时间）
	Title         string     `gorm:"type:varchar(255);not null;comment:会话标题" json:"title"`               // 会话标题
	Agent         string     `gorm:"type:varchar(255);not null;comment:Agent名称" json:"agent"`           // Agent名称
	Owner         string     `gorm:"type:varchar(255);not null;comment:创建者" json:"owner"`               // 创建者
	Type          string     `gorm:"type:varchar(16);not null;default:temporary;comment:会话类型 temporary/permanent" json:"type"` // 会话类型 temporary/permanent
	Status        string     `gorm:"type:varchar(16);not null;default:active;comment:会话状态 active/expired/deleted" json:"status"` // 会话状态 active/expired/deleted
	OnlineDevices int        `gorm:"type:int;not null;default:0;comment:在线设备数" json:"online_devices"`    // 在线设备数
	ExpiresAt     *time.Time `gorm:"type:timestamptz;comment:过期时间 永久为空" json:"expires_at,omitempty"` // 过期时间 永久为空
}

// GetGene 返回QaSession实体的雪花算法基因编号
func (q *QaSession) GetGene() xSnowflake.Gene {
	return bConst.GeneQaSession
}
