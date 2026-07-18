package apikey

import (
	"time"

	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
)

// ResetResponse API Key重置响应（包含新完整Key，仅此一次）
type ResetResponse struct {
	ID          xSnowflake.SnowflakeID `json:"id"`                   // API Key ID
	Name        string                 `json:"name"`                 // API Key名称
	Key         string                 `json:"key"`                  // 新完整密钥（仅此一次）
	KeyPrefix   string                 `json:"key_prefix"`           // 密钥前缀
	Description string                 `json:"description"`          // 描述
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"` // 过期时间
	IsActive    bool                   `json:"is_active"`            // 是否启用
	CreatedAt   string                 `json:"created_at"`           // 创建时间
}
