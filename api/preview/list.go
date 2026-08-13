package preview

import xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"

// PreviewSessionListRequest 预览会话列表查询请求
type PreviewSessionListRequest struct {
	ProjectID xSnowflake.SnowflakeID `form:"project_id"` // 项目ID筛选
	Page      int                    `form:"page"`       // 页码
	Size      int                    `form:"size"`       // 每页数量
}

// PreviewSessionListResponse 预览会话列表响应
type PreviewSessionListResponse struct {
	Items []PreviewSessionResponse `json:"items"` // 预览会话列表
	Total int64                    `json:"total"` // 总数量
}

// PreviewFileListResponse 预览文件列表响应
type PreviewFileListResponse struct {
	Items []PreviewFileResponse `json:"items"` // 预览文件列表
}
