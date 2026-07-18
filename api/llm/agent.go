package llm

import xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"

// AgentModelAssignment Agent 模型分配
type AgentModelAssignment struct {
	Role      string                  `json:"role"`       // Agent角色
	ModelID   *xSnowflake.SnowflakeID `json:"model_id"`   // 关联模型ID(nil表示未分配)
	ModelName *string                 `json:"model_name"` // 模型显示名称(nil表示未分配)
}

// AgentModelAssignmentsResponse Agent 模型分配批量响应
type AgentModelAssignmentsResponse struct {
	Module      string                 `json:"module"`      // 模块标识（如 repowiki）
	Assignments []AgentModelAssignment `json:"assignments"` // 角色分配列表
}

// UpdateAgentModelRequest 更新 Agent 模型分配请求
type UpdateAgentModelRequest struct {
	ModelID xSnowflake.SnowflakeID `json:"model_id" label:"模型ID" binding:"required"` // 模型ID
}
