package webhook

import (
	"time"

	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
)

// WebhookResponse is the response sent back to the Git Provider
type WebhookResponse struct {
	Status    string                 `json:"status"`               // accepted | ignored | failed
	VersionID xSnowflake.SnowflakeID `json:"version_id,omitempty"` // present when status=accepted
	Message   string                 `json:"message"`
	Reason    string                 `json:"reason,omitempty"` // present when status=ignored or failed
}

// WebhookSkipDetail provides details about why a webhook was skipped
type WebhookSkipDetail struct {
	Reason            string   `json:"reason"`
	Branch            string   `json:"branch,omitempty"`
	MonitoredBranches []string `json:"monitored_branches,omitempty"`
	EventType         string   `json:"event_type,omitempty"`
}

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
