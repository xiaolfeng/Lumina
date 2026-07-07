package entity

import (
	"time"

	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

// WebhookEvent Webhook事件表，记录Git Provider推送事件及处理状态
type WebhookEvent struct {
	xModels.BaseEntity
	ConfigID     *xSnowflake.SnowflakeID `gorm:"type:bigint;index;comment:关联RepoWikiConfig ID" json:"config_id,omitempty"`     // 关联RepoWikiConfig ID（无效Token事件可为空）
	Provider     string                  `gorm:"type:varchar(16);not null;comment:Git Provider" json:"provider"`                   // Git Provider
	EventType    string                  `gorm:"type:varchar(32);not null;comment:事件类型" json:"event_type"`                      // 事件类型
	Branch       string                  `gorm:"type:varchar(128);comment:推送分支" json:"branch,omitempty"`                         // 推送分支
	CommitBefore string                  `gorm:"type:varchar(64);comment:变更前commit" json:"commit_before,omitempty"`              // 变更前commit
	CommitAfter  string                  `gorm:"type:varchar(64);comment:变更后commit" json:"commit_after,omitempty"`               // 变更后commit
	ChangedCount int                     `gorm:"type:int;default:0;comment:变更文件数" json:"changed_count"`                          // 变更文件数
	Status       string                  `gorm:"type:varchar(16);not null;default:received;index;comment:处理状态" json:"status"`   // 处理状态
	Reason       string                  `gorm:"type:varchar(256);comment:跳过或失败原因" json:"reason,omitempty"`                     // 跳过或失败原因
	VersionID    xSnowflake.SnowflakeID  `gorm:"type:bigint;comment:关联WikiVersion ID" json:"version_id,omitempty"`              // 关联WikiVersion ID
	ResponseCode int                     `gorm:"type:int;comment:HTTP响应码" json:"response_code"`                                // HTTP响应码
	ReceivedAt   time.Time               `gorm:"type:timestamptz;not null;comment:接收时间" json:"received_at"`                     // 接收时间
	ProcessedAt  *time.Time              `gorm:"type:timestamptz;comment:处理完成时间" json:"processed_at,omitempty"`                // 处理完成时间
}

// GetGene 返回WebhookEvent实体的雪花算法基因编号
func (w *WebhookEvent) GetGene() xSnowflake.Gene {
	return bConst.GeneWebhookEvent
}
