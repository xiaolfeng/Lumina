package settings

// UpdateCategorySettingsRequest 更新分类设置请求
//
// 批量更新某个分类下的设置项键值。
type UpdateCategorySettingsRequest struct {
	Items []struct {
		Key   string `json:"key"`   // 设置项键名
		Value string `json:"value"` // 新的设置值
	} `json:"items"` // 待更新的键值对列表
}
