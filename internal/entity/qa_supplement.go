package entity

import (
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

// QaSupplement QA补充表，存储对问题或选项的补充说明
type QaSupplement struct {
	xModels.BaseEntity                        // 基础实体（ID、创建时间、更新时间）
	SessionID          xSnowflake.SnowflakeID `gorm:"type:bigint;not null;index;comment:关联会话ID" json:"session_id"`               // 关联会话ID
	TargetType         string                 `gorm:"type:varchar(16);not null;comment:关联类型 question/option" json:"target_type"` // 关联类型 question/option
	TargetID           xSnowflake.SnowflakeID `gorm:"type:bigint;not null;comment:关联ID" json:"target_id"`                        // 关联ID
	ContentType        string                 `gorm:"type:varchar(16);not null;comment:内容类型 markdown/html" json:"content_type"`  // 内容类型 markdown/html
	Content            string                 `gorm:"type:text;not null;comment:补充内容" json:"content"`                            // 补充内容
}

// GetGene 返回QaSupplement实体的雪花算法基因编号
func (q *QaSupplement) GetGene() xSnowflake.Gene {
	return bConst.GeneQaSupplement
}
