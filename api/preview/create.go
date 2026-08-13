package preview

import xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"

// CreateSessionRequest 创建预览会话请求
type CreateSessionRequest struct {
	ProjectID xSnowflake.SnowflakeID `json:"project_id" label:"关联项目ID" binding:"required"` // 关联项目ID
	Title     string                 `json:"title" label:"会话标题"`                           // 会话标题 (可选)
}
