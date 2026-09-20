package entity

import (
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

// PageVersion 页面不可变快照版本，晋升时深拷贝文件并记录溯源
type PageVersion struct {
	xModels.BaseEntity                         // 基础实体（ID、创建时间、更新时间）
	PageID             xSnowflake.SnowflakeID  `gorm:"type:bigint;not null;index;comment:所属页面ID" json:"page_id"`              // 所属页面ID
	Version            string                  `gorm:"type:varchar(32);not null;comment:语义化版本号" json:"version"`               // 语义化版本号
	Changelog          string                  `gorm:"type:text;comment:版本更新说明" json:"changelog"`                             // 版本更新说明
	SourceSessionID    *xSnowflake.SnowflakeID `gorm:"type:bigint;index;comment:来源预览会话ID" json:"source_session_id,omitempty"` // 来源预览会话ID
	BaseVersionID      *xSnowflake.SnowflakeID `gorm:"type:bigint;index;comment:派生基准版本ID" json:"base_version_id,omitempty"`   // 派生基准版本ID
	EntryFilename      string                  `gorm:"type:varchar(255);not null;comment:默认入口文件名" json:"entry_filename"`      // 默认入口文件名
	FileCount          int                     `gorm:"type:int;not null;default:0;comment:快照内文件总数" json:"file_count"`         // 快照内文件总数
	TotalSize          int64                   `gorm:"type:bigint;not null;default:0;comment:快照总字节数" json:"total_size"`       // 快照总字节数
	CreatedBy          string                  `gorm:"type:varchar(64);not null;default:'';comment:发布操作者" json:"created_by"`  // 发布操作者
}

// GetGene 返回 PageVersion 实体的雪花算法基因编号
func (p *PageVersion) GetGene() xSnowflake.Gene {
	return bConst.GenePageVersion
}
