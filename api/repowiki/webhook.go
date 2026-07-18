package repowiki

import (
	"time"

	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
)

// WebhookConfigResponse webhook 配置信息
type WebhookConfigResponse struct {
	URL           string            `json:"url"`            // 完整 Webhook URL
	Token         string            `json:"token"`          // 脱敏 Token（cs_****）
	HasSecret     bool              `json:"has_secret"`     // 是否设置 Secret
	Branches      []string          `json:"branches"`       // 监控分支列表
	ProviderGuide map[string]string `json:"provider_guide"` // 提供商 -> 配置指南
}

// UpdateWebhookBranchesRequest 更新监控分支请求
type UpdateWebhookBranchesRequest struct {
	Branches []string `json:"branches" label:"监控分支列表" binding:"required"`
}

// RegenerateWebhookResponse 重新生成凭据一次性响应
type RegenerateWebhookResponse struct {
	Token  string `json:"token"`  // 完整 Token（仅展示一次）
	Secret string `json:"secret"` // 完整 Secret（仅展示一次）
	URL    string `json:"url"`    // 完整 Webhook URL
}

// WebhookEventListResponse webhook 事件列表响应（分页）
type WebhookEventListResponse struct {
	Total int64                  `json:"total"` // 总数
	Items []WebhookEventResponse `json:"items"` // 事件列表
}

// WebhookEventResponse 单个 webhook 事件响应
type WebhookEventResponse struct {
	ID           xSnowflake.SnowflakeID `json:"id"`                      // 事件ID
	ConfigID     xSnowflake.SnowflakeID `json:"config_id,omitempty"`     // 关联配置ID
	Provider     string                 `json:"provider"`                // Git 提供商
	EventType    string                 `json:"event_type"`              // 事件类型
	Branch       string                 `json:"branch,omitempty"`        // 触发分支
	CommitBefore string                 `json:"commit_before,omitempty"` // 变更前 commit
	CommitAfter  string                 `json:"commit_after,omitempty"`  // 变更后 commit
	ChangedCount int                    `json:"changed_count"`           // 变更文件数
	Status       string                 `json:"status"`                  // 处理状态
	Reason       string                 `json:"reason,omitempty"`        // 忽略/失败原因
	VersionID    xSnowflake.SnowflakeID `json:"version_id,omitempty"`    // 关联版本ID
	ResponseCode int                    `json:"response_code"`           // 响应 HTTP 状态码
	ReceivedAt   time.Time              `json:"received_at"`             // 接收时间
	ProcessedAt  *time.Time             `json:"processed_at,omitempty"`  // 处理时间
}
