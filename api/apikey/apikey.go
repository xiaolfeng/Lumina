package apikey

import "time"

// CreateRequest 创建 API Key 请求
type CreateRequest struct {
	Name        string     `json:"name" binding:"required,min=1,max=64"` // 密钥名称
	Description string     `json:"description"`                          // 密钥用途描述
	ExpiresAt   *time.Time `json:"expires_at"`                           // 过期时间（可选）
}

// CreateResponse 创建 API Key 响应（包含完整 Key，仅此一次）
type CreateResponse struct {
	ID          string     `json:"id"`                   // 密钥 ID
	Name        string     `json:"name"`                 // 密钥名称
	Key         string     `json:"key"`                  // 完整密钥（仅创建时返回一次）
	KeyPrefix   string     `json:"key_prefix"`           // 密钥前缀
	Description string     `json:"description"`          // 密钥用途描述
	ExpiresAt   *time.Time `json:"expires_at,omitempty"` // 过期时间
	IsActive    bool       `json:"is_active"`            // 是否启用
	CreatedAt   time.Time  `json:"created_at"`           // 创建时间
}

// UpdateRequest 更新 API Key 请求
type UpdateRequest struct {
	Name        string     `json:"name" binding:"omitempty,min=1,max=64"` // 密钥名称
	Description string     `json:"description"`                           // 密钥用途描述
	ExpiresAt   *time.Time `json:"expires_at"`                            // 过期时间
	IsActive    *bool      `json:"is_active"`                             // 是否启用
}

// ApikeyItem API Key 条目（脱敏展示，适用于列表和详情）
type ApikeyItem struct {
	ID          string     `json:"id"`                     // 密钥 ID
	Name        string     `json:"name"`                   // 密钥名称
	Key         string     `json:"key"`                    // 脱敏密钥（如 lumi_abcd****wxyz）
	KeyPrefix   string     `json:"key_prefix"`             // 密钥前缀
	Description string     `json:"description"`            // 密钥用途描述
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`   // 过期时间
	IsActive    bool       `json:"is_active"`              // 是否启用
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"` // 最后使用时间
	CreatedAt   time.Time  `json:"created_at"`             // 创建时间
}

// ListResponse API Key 列表响应
type ListResponse struct {
	Items []ApikeyItem `json:"items"` // API Key 列表
}

// DetailResponse API Key 详情响应（与 ApikeyItem 一致，脱敏）
type DetailResponse ApikeyItem

// ResetResponse 重置 API Key 响应（包含新完整 Key，仅此一次）
type ResetResponse struct {
	ID          string     `json:"id"`                   // 密钥 ID
	Name        string     `json:"name"`                 // 密钥名称
	Key         string     `json:"key"`                  // 新完整密钥（仅重置时返回一次）
	KeyPrefix   string     `json:"key_prefix"`           // 新密钥前缀
	Description string     `json:"description"`          // 密钥用途描述
	ExpiresAt   *time.Time `json:"expires_at,omitempty"` // 过期时间
	IsActive    bool       `json:"is_active"`            // 是否启用
}
