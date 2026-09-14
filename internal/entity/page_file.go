package entity

import (
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

// PageFile 页面快照文件表，扁平单层文件名，绑定不可变版本
type PageFile struct {
	xModels.BaseEntity                        // 基础实体（ID、创建时间、更新时间）
	VersionID          xSnowflake.SnowflakeID `gorm:"type:bigint;not null;uniqueIndex:uk_version_file;index;comment:所属快照版本ID" json:"version_id"` // 所属快照版本ID
	Filename           string                 `gorm:"type:varchar(255);not null;uniqueIndex:uk_version_file;comment:文件名(扁平单层)" json:"filename"`  // 文件名(扁平单层)
	MimeType           string                 `gorm:"type:varchar(128);not null;comment:MIME类型" json:"mime_type"`                                // MIME类型
	Size               int                    `gorm:"type:int;not null;default:0;comment:文件大小(字节)" json:"size"`                                  // 文件大小(字节)
	Content            string                 `gorm:"type:text;not null;comment:文件内容" json:"content"`                                            // 文件内容
}

// GetGene 返回 PageFile 实体的雪花算法基因编号
func (p *PageFile) GetGene() xSnowflake.Gene {
	return bConst.GenePageFile
}
