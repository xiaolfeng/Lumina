package pages

// SwitchActiveResponse 切换生效版本结果
type SwitchActiveResponse struct {
	Page    PageResponse        `json:"page"`    // 更新后的页面
	Version PageVersionResponse `json:"version"` // 新的生效版本
}
