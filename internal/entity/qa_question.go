package entity

import (
	"time"

	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"gorm.io/datatypes"
)

// QaQuestion QA问题表，存储会话中的问题及其回答
type QaQuestion struct {
	xModels.BaseEntity                                                                               // 基础实体（ID、创建时间、更新时间）
	SessionID   xSnowflake.SnowflakeID `gorm:"type:bigint;not null;index;comment:关联会话ID" json:"session_id"` // 关联会话ID
	Type        string          `gorm:"type:varchar(16);not null;comment:题型 14种之一" json:"type"`       // 题型 14种之一
	Title       string          `gorm:"type:text;not null;comment:问题标题 Markdown" json:"title"`          // 问题标题 Markdown
	Description string          `gorm:"type:text;comment:问题描述" json:"description"`                     // 问题描述
	Options     datatypes.JSON  `gorm:"type:jsonb;comment:选项配置" json:"options"`                         // 选项配置
	Config      datatypes.JSON  `gorm:"type:jsonb;comment:题型特有配置" json:"config"`                      // 题型特有配置
	Batch       datatypes.JSON  `gorm:"type:jsonb;comment:分批信息" json:"batch"`                           // 分批信息
	GroupLabel  string          `gorm:"type:varchar(255);comment:分组标签" json:"group_label"`            // 分组标签
	Status      string          `gorm:"type:varchar(16);not null;default:pending;comment:状态 pending/answered/skipped" json:"status"` // 状态 pending/answered/skipped
	Answer      datatypes.JSON  `gorm:"type:jsonb;comment:回答数据" json:"answer"`                           // 回答数据
	Media       datatypes.JSON  `gorm:"type:jsonb;comment:多媒体数据" json:"media"`                           // 多媒体数据
	AnsweredAt  *time.Time      `gorm:"type:timestamptz;comment:回答时间" json:"answered_at,omitempty"`     // 回答时间
}

// GetGene 返回QaQuestion实体的雪花算法基因编号
func (q *QaQuestion) GetGene() xSnowflake.Gene {
	return bConst.GeneQaQuestion
}
