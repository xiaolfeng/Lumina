package llm

import xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"

// CreateModelRequest 创建 LLM 模型请求
type CreateModelRequest struct {
	ProviderID    xSnowflake.SnowflakeID `json:"provider_id" label:"关联Provider ID" binding:"required"` // 关联Provider ID
	ModelName     string                 `json:"model_name" label:"模型标识" binding:"required"`           // 模型标识
	DisplayName   string                 `json:"display_name" label:"显示名称" binding:"required"`         // 显示名称
	MaxTokens     int64                  `json:"max_tokens" label:"最大输出Token数"`                        // 单次响应最大输出Token数
	ContextWindow int64                  `json:"context_window" label:"上下文窗口大小"`                       // 模型上下文窗口大小
	Temperature   float64                `json:"temperature" label:"生成温度"`                             // 生成温度
	ThinkingEffort string                `json:"thinking_effort" label:"思考强度"`                          // 思考强度(none/low/medium/high,空=不启用)
	Description   string                 `json:"description" label:"描述"`                               // 描述说明
}

// UpdateModelRequest 更新 LLM 模型请求
type UpdateModelRequest struct {
	ModelName      *string  `json:"model_name"`      // 模型标识
	DisplayName    *string  `json:"display_name"`    // 显示名称
	MaxTokens      *int64   `json:"max_tokens"`      // 单次响应最大输出Token数
	ContextWindow  *int64   `json:"context_window"`  // 模型上下文窗口大小
	Temperature    *float64 `json:"temperature"`     // 生成温度
	ThinkingEffort *string  `json:"thinking_effort"` // 思考强度(none/low/medium/high,空=不启用)
	IsActive       *bool    `json:"is_active"`       // 是否启用
	Description    *string  `json:"description"`     // 描述说明
}

// ModelDetailResponse 模型详情响应
type ModelDetailResponse struct {
	ID             xSnowflake.SnowflakeID `json:"id"`              // 模型 ID
	ProviderID     xSnowflake.SnowflakeID `json:"provider_id"`     // 关联Provider ID
	ModelName      string                 `json:"model_name"`      // 模型标识
	DisplayName    string                 `json:"display_name"`    // 显示名称
	MaxTokens      int64                  `json:"max_tokens"`      // 单次响应最大输出Token数
	ContextWindow  int64                  `json:"context_window"`  // 模型上下文窗口大小
	Temperature    float64                `json:"temperature"`     // 生成温度
	ThinkingEffort string                 `json:"thinking_effort"` // 思考强度
	IsActive       bool                   `json:"is_active"`       // 是否启用
	Description    string                 `json:"description"`     // 描述说明
	CreatedAt      string                 `json:"created_at"`      // 创建时间
	UpdatedAt      string                 `json:"updated_at"`      // 更新时间
}

// ModelListItem 模型列表项
type ModelListItem struct {
	ID             xSnowflake.SnowflakeID `json:"id"`              // 模型 ID
	ProviderID     xSnowflake.SnowflakeID `json:"provider_id"`     // 关联Provider ID
	ModelName      string                 `json:"model_name"`      // 模型标识
	DisplayName    string                 `json:"display_name"`    // 显示名称
	MaxTokens      int64                  `json:"max_tokens"`      // 单次响应最大输出Token数
	ContextWindow  int64                  `json:"context_window"`  // 模型上下文窗口大小
	Temperature    float64                `json:"temperature"`     // 生成温度
	ThinkingEffort string                 `json:"thinking_effort"` // 思考强度
	IsActive       bool                   `json:"is_active"`       // 是否启用
	Description    string                 `json:"description"`     // 描述说明
	CreatedAt      string                 `json:"created_at"`      // 创建时间
}
