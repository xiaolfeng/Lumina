package workspace

import xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"

// WorkspaceResponse 工作空间响应
type WorkspaceResponse struct {
	ID          xSnowflake.SnowflakeID `json:"id"`          // 空间ID
	Name        string                 `json:"name"`        // 空间名称
	Slug        string                 `json:"slug"`        // 空间标识
	Description string                 `json:"description"` // 空间描述
	Icon        string                 `json:"icon"`        // 空间图标
	IsDefault   bool                   `json:"is_default"`  // 是否为默认空间
	CreatedAt   string                 `json:"created_at"`  // 创建时间
	UpdatedAt   string                 `json:"updated_at"`  // 更新时间
}

// DeleteWorkspaceResponse 删除工作空间响应
type DeleteWorkspaceResponse struct {
	MovedProjectCount int64 `json:"moved_project_count"` // 迁到默认空间的项目数
}
