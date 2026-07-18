package webhook

import xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"

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
