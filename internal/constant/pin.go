package bConst

// Pin 分类常量
const (
	PinCategoryNotice     = "notice"     // 注意事项
	PinCategoryDependency = "dependency"  // 依赖约束
	PinCategoryAPIChange  = "api_change" // 接口变更
	PinCategoryOther       = "other"      // 其他
)

// Pin 状态常量
const (
	PinStatusPending  = "pending"  // 待消费
	PinStatusConsumed = "consumed" // 已消费
)

// Pin 优先级常量
const (
	PinPriorityHigh   = "high"   // 高
	PinPriorityMedium = "medium" // 中
	PinPriorityLow    = "low"    // 低
)
