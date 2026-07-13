package llm

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
	ID          string `json:"id"`          // Provider ID
	Name        string `json:"name"`        // Provider名称
	Protocol    string `json:"protocol"`    // 协议类型
	BaseURL     string `json:"base_url"`    // 自定义API端点
	HasKey      bool   `json:"has_key"`     // 是否已设置密钥
	IsActive    bool   `json:"is_active"`   // 是否启用
	Description string `json:"description"` // 描述说明
	CreatedAt   string `json:"created_at"`  // 创建时间
	UpdatedAt   string `json:"updated_at"`  // 更新时间
}

// ProviderListItem Provider 列表项
type ProviderListItem struct {
	ID          string `json:"id"`          // Provider ID
	Name        string `json:"name"`        // Provider名称
	Protocol    string `json:"protocol"`    // 协议类型
	BaseURL     string `json:"base_url"`    // 自定义API端点
	HasKey      bool   `json:"has_key"`     // 是否已设置密钥
	IsActive    bool   `json:"is_active"`   // 是否启用
	Description string `json:"description"` // 描述说明
	CreatedAt   string `json:"created_at"`  // 创建时间
}

// CreateModelRequest 创建 LLM 模型请求
type CreateModelRequest struct {
	ProviderID    string  `json:"provider_id" label:"关联Provider ID" binding:"required"` // 关联Provider ID
	ModelName     string  `json:"model_name" label:"模型标识" binding:"required"`           // 模型标识
	DisplayName   string  `json:"display_name" label:"显示名称" binding:"required"`         // 显示名称
	MaxTokens     int64   `json:"max_tokens" label:"最大输出Token数"`                        // 单次响应最大输出Token数
	ContextWindow int64   `json:"context_window" label:"上下文窗口大小"`                       // 模型上下文窗口大小
	Temperature   float64 `json:"temperature" label:"生成温度"`                             // 生成温度
	Description   string  `json:"description" label:"描述"`                               // 描述说明
}

// UpdateModelRequest 更新 LLM 模型请求
type UpdateModelRequest struct {
	ModelName     *string  `json:"model_name"`     // 模型标识
	DisplayName   *string  `json:"display_name"`   // 显示名称
	MaxTokens     *int64   `json:"max_tokens"`     // 单次响应最大输出Token数
	ContextWindow *int64   `json:"context_window"` // 模型上下文窗口大小
	Temperature   *float64 `json:"temperature"`    // 生成温度
	IsActive      *bool    `json:"is_active"`      // 是否启用
	Description   *string  `json:"description"`    // 描述说明
}

// ModelDetailResponse 模型详情响应
type ModelDetailResponse struct {
	ID            string  `json:"id"`             // 模型 ID
	ProviderID    string  `json:"provider_id"`    // 关联Provider ID
	ModelName     string  `json:"model_name"`     // 模型标识
	DisplayName   string  `json:"display_name"`   // 显示名称
	MaxTokens     int64   `json:"max_tokens"`     // 单次响应最大输出Token数
	ContextWindow int64   `json:"context_window"` // 模型上下文窗口大小
	Temperature   float64 `json:"temperature"`    // 生成温度
	IsActive      bool    `json:"is_active"`      // 是否启用
	Description   string  `json:"description"`    // 描述说明
	CreatedAt     string  `json:"created_at"`     // 创建时间
	UpdatedAt     string  `json:"updated_at"`     // 更新时间
}

// ModelListItem 模型列表项
type ModelListItem struct {
	ID            string  `json:"id"`             // 模型 ID
	ProviderID    string  `json:"provider_id"`    // 关联Provider ID
	ModelName     string  `json:"model_name"`     // 模型标识
	DisplayName   string  `json:"display_name"`   // 显示名称
	MaxTokens     int64   `json:"max_tokens"`     // 单次响应最大输出Token数
	ContextWindow int64   `json:"context_window"` // 模型上下文窗口大小
	Temperature   float64 `json:"temperature"`    // 生成温度
	IsActive      bool    `json:"is_active"`      // 是否启用
	Description   string  `json:"description"`    // 描述说明
	CreatedAt     string  `json:"created_at"`     // 创建时间
}

// AgentModelAssignment Agent 模型分配
type AgentModelAssignment struct {
	Role      string  `json:"role"`       // Agent角色
	ModelID   *string `json:"model_id"`   // 关联模型ID(nil表示未分配)
	ModelName *string `json:"model_name"` // 模型显示名称(nil表示未分配)
}

// AgentModelAssignmentsResponse Agent 模型分配批量响应
type AgentModelAssignmentsResponse struct {
	Module      string                 `json:"module"`      // 模块标识（如 repowiki）
	Assignments []AgentModelAssignment `json:"assignments"` // 角色分配列表
}

// UpdateAgentModelRequest 更新 Agent 模型分配请求
type UpdateAgentModelRequest struct {
	ModelID string `json:"model_id" label:"模型ID" binding:"required"` // 模型ID
}
