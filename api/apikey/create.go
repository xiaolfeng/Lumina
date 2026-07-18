package apikey

import (
	"time"

	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
)

// CreateRequest API Key创建请求
type CreateRequest struct {
	Name        string     `json:"name" label:"API Key名称" binding:"required,min=1,max=64"` // API Key名称
	Description string     `json:"description" label:"描述"`                                 // 描述
	ExpiresAt   *time.Time `json:"expires_at,omitempty" label:"过期时间"`                      // 过期时间（NULL=永不过期）
}

// CreateResponse API Key创建响应（包含完整Key，仅此一次）
type CreateResponse struct {
	ID          xSnowflake.SnowflakeID `json:"id"`                   // API Key ID
	Name        string                 `json:"name"`                 // API Key名称
	Key         string                 `json:"key"`                  // 完整密钥（仅此一次）
	KeyPrefix   string                 `json:"key_prefix"`           // 密钥前缀
	Description string                 `json:"description"`          // 描述
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"` // 过期时间
	IsActive    bool                   `json:"is_active"`            // 是否启用
	CreatedAt   string                 `json:"created_at"`           // 创建时间（手动映射，BaseEntity.CreatedAt 的 json:"-"）
}
0