package apikey

import (
	"time"

	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
)

// _ 满足编译检查，确保 xModels 被正确引用
var _ = xModels.PageResponse[struct{}]{}

// CreateRequest API Key创建请求
type CreateRequest struct {
	Name        string     `json:"name" binding:"required,min=1,max=64"` // API Key名称
	Description string     `json:"description"`                           // 描述
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`                  // 过期时间（NULL=永不过期）
}

// CreateResponse API Key创建响应（包含完整Key，仅此一次）
type CreateResponse struct {
	ID          string     `json:"id"`                  // API Key ID
	Name        string     `json:"name"`                // API Key名称
	Key         string     `json:"key"`                 // 完整密钥（仅此一次）
	KeyPrefix   string     `json:"key_prefix"`          // 密钥前缀
	Description string     `json:"description"`         // 描述
	ExpiresAt   *time.Time `json:"expires_at,omitempty"` // 过期时间
	IsActive    bool       `json:"is_active"`           // 是否启用
	CreatedAt   string     `json:"created_at"`          // 创建时间（手动映射，BaseEntity.CreatedAt 的 json:"-"）
}

// UpdateRequest API Key更新请求（所有字段可选）
type UpdateRequest struct {
	Name        string     `json:"name"`                  // API Key名称
	Description *string    `json:"description"`           // 描述（指针区分零值和未传）
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`   // 过期时间
	IsActive    *bool      `json:"is_active"`              // 是否启用（指针区分零值和未传）
}

// DetailResponse API Key详情响应
type DetailResponse struct {
	ID          string     `json:"id"`                  // API Key ID
	Name        string     `json:"name"`                // API Key名称
	Key         string     `json:"key"`                 // 脱敏密钥（前8+...+后8）
	KeyPrefix   string     `json:"key_prefix"`          // 密钥前缀
	Description string     `json:"description"`         // 描述
	ExpiresAt   *time.Time `json:"expires_at,omitempty"` // 过期时间
	IsActive    bool       `json:"is_active"`           // 是否启用
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"` // 最后使用时间
	CreatedAt   string     `json:"created_at"`          // 创建时间
}

// ResetResponse API Key重置响应（包含新完整Key，仅此一次）
type ResetResponse struct {
	ID          string     `json:"id"`                  // API Key ID
	Name        string     `json:"name"`                // API Key名称
	Key         string     `json:"key"`                 // 新完整密钥（仅此一次）
	KeyPrefix   string     `json:"key_prefix"`          // 密钥前缀
	Description string     `json:"description"`         // 描述
	ExpiresAt   *time.Time `json:"expires_at,omitempty"` // 过期时间
	IsActive    bool       `json:"is_active"`           // 是否启用
	CreatedAt   string     `json:"created_at"`          // 创建时间
}

// ApikeyItem API Key列表项
type ApikeyItem struct {
	ID          string     `json:"id"`                  // API Key ID
	Name        string     `json:"name"`                // API Key名称
	Key         string     `json:"key"`                 // 脱敏密钥
	KeyPrefix   string     `json:"key_prefix"`          // 密钥前缀
	Description string     `json:"description"`         // 描述
	ExpiresAt   *time.Time `json:"expires_at,omitempty"` // 过期时间
	IsActive    bool       `json:"is_active"`           // 是否启用
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"` // 最后使用时间
	CreatedAt   string     `json:"created_at"`          // 创建时间
}
