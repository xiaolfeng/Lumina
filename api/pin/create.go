package pin

import xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"

// CreatePinRequest 创建Pin请求
type CreatePinRequest struct {
	Title         string                 `json:"title" label:"约束标题" binding:"required"`           // 约束标题
	Content       string                 `json:"content" label:"详细内容" binding:"required"`         // 详细内容
	Category      string                 `json:"category" label:"分类"`                             // 分类 (notice/dependency/api_change/other)
	Priority      string                 `json:"priority" label:"优先级" binding:"required"`         // 优先级 (high/medium/low)
	FromProjectID xSnowflake.SnowflakeID `json:"from_project_id" label:"来源项目ID"`                  // 来源项目ID (可选)
	ToProjectID   xSnowflake.SnowflakeID `json:"to_project_id" label:"目标项目ID" binding:"required"` // 目标项目ID
}
