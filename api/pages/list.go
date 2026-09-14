package pages

import xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"

// PageListRequest 页面列表查询请求
type PageListRequest struct {
	ProjectID   xSnowflake.SnowflakeID `form:"project_id"`   // 项目ID筛选
	WorkspaceID xSnowflake.SnowflakeID `form:"workspace_id"` // 所属空间ID筛选，零值不过滤
	Page        int                    `form:"page"`         // 页码
	Size        int                    `form:"size"`         // 每页数量
}

// PageListResponse 页面列表响应
type PageListResponse struct {
	Items []PageResponse `json:"items"` // 页面列表
	Total int64          `json:"total"` // 总数量
}
