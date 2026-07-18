package apikey

import (
	"time"

	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
)

// ApikeyItem API Key列表项
type ApikeyItem struct {
	ID          xSnowflake.SnowflakeID `json:"id"`                     // API Key ID
	Name        string                 `json:"name"`                   // API Key名称
	Key         string                 `json:"key"`                    // 脱敏密钥
	KeyPrefix   string                 `json:"key_prefix"`             // 密钥前缀
	Description string                 `json:"description"`            // 描述
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`   // 过期时间
	IsActive    bool                   `json:"is_active"`              // 是否启用
	LastUsedAt  *time.Time             `json:"last_used_at,omitempty"` // 最后使用时间
	CreatedAt   string                 `json:"created_at"`             // 创建时间
}
