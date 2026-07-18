package apikey

import "time"

// UpdateRequest API Key更新请求（所有字段可选）
type UpdateRequest struct {
	Name        string     `json:"name"`                 // API Key名称
	Description *string    `json:"description"`          // 描述（指针区分零值和未传）
	ExpiresAt   *time.Time `json:"expires_at,omitempty"` // 过期时间
	IsActive    *bool      `json:"is_active"`            // 是否启用（指针区分零值和未传）
}
