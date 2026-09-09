package workspace

// WorkspaceListResponse 工作空间列表响应
type WorkspaceListResponse struct {
	Items []WorkspaceResponse `json:"items"` // 空间列表
	Total int64               `json:"total"` // 总数量
}
