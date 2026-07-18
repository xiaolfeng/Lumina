package webhook

import (
	"time"

	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
)

// WebhookEventResponse single webhook event
type WebhookEventResponse struct {
	ID           xSnowflake.SnowflakeID `json:"id"`
	ConfigID     xSnowflake.SnowflakeID `json:"config_id,omitempty"`
	Provider     string                 `json:"provider"`
	EventType    string                 `json:"event_type"`
	Branch       string                 `json:"branch,omitempty"`
	CommitBefore string                 `json:"commit_before,omitempty"`
	CommitAfter  string                 `json:"commit_after,omitempty"`
	ChangedCount int                    `json:"changed_count"`
	Status       string                 `json:"status"`
	Reason       string                 `json:"reason,omitempty"`
	VersionID    xSnowflake.SnowflakeID `json:"version_id,omitempty"`
	ResponseCode int                    `json:"response_code"`
	ReceivedAt   time.Time              `json:"received_at"`
	ProcessedAt  *time.Time             `json:"processed_at,omitempty"`
}

// WebhookEventListResponse paginated webhook events
type WebhookEventListResponse struct {
	Total int64                  `json:"total"`
	Items []WebhookEventResponse `json:"items"`
}
