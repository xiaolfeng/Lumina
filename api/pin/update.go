package pin

// UpdatePinRequest 更新Pin请求
type UpdatePinRequest struct {
	Priority *string `json:"priority"` // 优先级 (可选更新)
	Category *string `json:"category"` // 分类 (可选更新)
}
