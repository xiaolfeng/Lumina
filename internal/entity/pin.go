package entity

import (
	"time"

	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

// Pin Pin表，存储跨项目依赖约束与点对点推送信息
type Pin struct {
	xModels.BaseEntity                                                                                    // 基础实体（ID、创建时间、更新时间）
	FromProjectID xSnowflake.SnowflakeID `gorm:"not null;index;comment:来源项目ID" json:"from_project_id"` // 来源项目ID
	ToProjectID   xSnowflake.SnowflakeID `gorm:"not null;index;comment:目标项目ID" json:"to_project_id"`   // 目标项目ID
	Title        string                 `gorm:"type:varchar(255);not null;comment:约束标题" json:"title"`                 // 约束标题
	Content      string                 `gorm:"type:text;not null;comment:详细内容" json:"content"`                             // 详细内容
	Category     string                 `gorm:"type:varchar(16);default:notice;comment:分类" json:"category"`                // 分类
	Status       string                 `gorm:"type:varchar(16);default:pending;index;comment:状态" json:"status"`           // 状态
	Priority     string                 `gorm:"type:varchar(16);default:medium;comment:优先级" json:"priority"`             // 优先级
	ConsumedAt   *time.Time             `gorm:"type:timestamptz;comment:消费时间" json:"consumed_at,omitempty"` // 消费时间
}

// GetGene 返回Pin实体的雪花算法基因编号
func (p *Pin) GetGene() xSnowflake.Gene {
	return bConst.GenePin
}
