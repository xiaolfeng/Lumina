package settings

// SettingItemResponse 单个设置项响应
//
// 表示某个分类下的一个配置项，包含键值对、类型及描述。
type SettingItemResponse struct {
	Key         string `json:"key"`         // 设置项键名
	Value       string `json:"value"`       // 设置项当前值
	Type        string `json:"type"`        // 值类型(text/number/boolean等)
	Description string `json:"description"` // 设置项描述说明
}

// CategorySettingsResponse 分类设置项响应
//
// 按分类聚合的一组设置项，用于前端按模块展示。
type CategorySettingsResponse struct {
	Category string                `json:"category"` // 分类名称(如general/repo/memory)
	Items    []SettingItemResponse `json:"items"`    // 该分类下的设置项列表
}
