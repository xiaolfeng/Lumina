package preview

// PreviewFileEditResponse 预览文件行级编辑响应（含编辑落点区域，供调用方一次往返核对）
type PreviewFileEditResponse struct {
	PreviewFileResponse
	TotalLines  int    `json:"total_lines"`  // 编辑后总行数
	RegionStart int    `json:"region_start"` // 编辑落点区域起始行（含上下文，1 起始闭区间）
	RegionEnd   int    `json:"region_end"`   // 编辑落点区域结束行（含上下文，1 起始闭区间）
	Region      string `json:"region"`       // 编辑落点区域内容（带行号，上下文裁剪后）
}
