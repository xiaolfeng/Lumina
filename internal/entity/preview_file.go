package entity

import (
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

// PreviewFile 预览文件表，存储预览会话内的前端文件（扁平单层，可跨文件相对引用）
type PreviewFile struct {
	xModels.BaseEntity                        // 基础实体（ID、创建时间、更新时间）
	SessionID          xSnowflake.SnowflakeID `gorm:"type:bigint;not null;index;uniqueIndex:uk_session_file;comment:关联会话ID" json:"session_id"`  // 关联会话ID
	Filename           string                 `gorm:"type:varchar(255);not null;uniqueIndex:uk_session_file;comment:文件名(扁平单层)" json:"filename"` // 文件名(扁平单层)
	MimeType           string                 `gorm:"type:varchar(64);not null;comment:MIME类型" json:"mime_type"`                                // MIME类型
	Content            string                 `gorm:"type:text;not null;comment:文件内容" json:"content"`                                           // 文件内容
	Size               int                    `gorm:"type:int;not null;default:0;comment:文件大小(字节)" json:"size"`                                 // 文件大小(字节)
}

// GetGene 返回PreviewFile实体的雪花算法基因编号
func (p *PreviewFile) GetGene() xSnowflake.Gene {
	return bConst.GenePreviewFile
}
