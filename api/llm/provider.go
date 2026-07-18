package llm

import xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"

// CreateProviderRequest 创建 LLM Provider 请求
type CreateProviderRequest struct {
	Name        string `json:"name" label:"Provider名称" binding:"required"` // Provider名称
	Protocol    string `json:"protocol" label:"协议类型" binding:"required"`   // 协议类型(openai/anthropic)
	BaseURL     string `json:"base_url" label:"API端点"`                     // 自定义API端点
	APIKey      string `json:"api_key" label:"API密钥" binding:"required"`   // API密钥(明文，后端加密存储)
	Description string `json:"description" label:"描述"`                     // 描述说明
}

// UpdateProviderRequest 更新 LLM Provider 请求
type UpdateProviderRequest struct {
	Name        *string `json:"name"`        // Provider名称
	Protocol    *string `json:"protocol"`    // 协议类型
	BaseURL     *string `json:"base_url"`    // 自定义API端点
	APIKey      *string `json:"api_key"`     // API密钥(空时不更新)
	IsActive    *bool   `json:"is_active"`   // 是否启用
	Description *string `json:"description"` // 描述说明
}

// ProviderDetailResponse Provider 详情响应
type ProviderDetailResponse struct {
	ID          xSnowflake.SnowflakeID `json:"id"`          // Provider ID
	Name        string                 `json:"name"`        // Provider名称
	Protocol    string                 `json:"protocol"`    // 协议类型
	BaseURL     string                 `json:"base_url"`    // 自定义API端点
	HasKey      bool                   `json:"has_key"`     // 是否已设置密钥
	IsActive    bool                   `json:"is_active"`   // 是否启用
	Description string                 `json:"description"` // 描述说明
	CreatedAt   string                 `json:"created_at"`  // 创建时间
	UpdatedAt   string                 `json:"updated_at"`  // 更新时间
}

// ProviderListItem Provider 列表项
type ProviderListItem struct {
	ID          xSnowflake.SnowflakeID `json:"id"`          // Provider ID
	Name        string                 `json:"name"`        // Provider名称
	Protocol    string                 `json:"protocol"`    // 协议类型
	BaseURL     string                 `json:"base_url"`    // 自定义API端点
	HasKey      bool                   `json:"has_key"`     // 是否已设置密钥
	IsActive    bool                   `json:"is_active"`   // 是否启用
	Description string                 `json:"description"` // 描述说明
	CreatedAt   string                 `json:"created_at"`  // 创建时间
}
