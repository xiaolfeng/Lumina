package pin

import xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"

// PinListRequest Pin列表查询请求
type PinListRequest struct {
	ToProjectID   xSnowflake.SnowflakeID `form:"to_project_id"`   // 目标项目ID筛选
	FromProjectID xSnowflake.SnowflakeID `form:"from_project_id"` // 来源项目ID筛选
	Status        string                 `form:"status"`          // 状态筛选
	Category      string                 `form:"category"`        // 分类筛选
	Priority      string                 `form:"priority"`        // 优先级筛选
	Page          int                    `form:"page"`            // 页码
	Size          int                    `form:"size"`            // 每页数量
}

// PinListResponse Pin列表响应
type PinListResponse struct {
	Items []PinResponse `json:"items"` // Pin列表
	Total int64         `json:"total"` // 总数量
}
