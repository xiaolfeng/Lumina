package project

import xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"

// ProjectListRequest 项目列表查询请求
type ProjectListRequest struct {
	Page        int                    `form:"page"`         // 页码
	Size        int                    `form:"size"`         // 每页数量
	WorkspaceID xSnowflake.SnowflakeID `form:"workspace_id"` // 所属空间ID筛选，零值不过滤
}
