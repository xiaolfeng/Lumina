package pin

// CreatePinRequest 创建Pin请求
type CreatePinRequest struct {
	Title         string `json:"title" label:"约束标题" binding:"required"`           // 约束标题
	Content       string `json:"content" label:"详细内容" binding:"required"`         // 详细内容
	Category      string `json:"category" label:"分类"`                               // 分类 (notice/dependency/api_change/other)
	Priority      string `json:"priority" label:"优先级" binding:"required"`         // 优先级 (high/medium/low)
	FromProjectID string `json:"from_project_id" label:"来源项目ID"`                  // 来源项目ID (可选)
	ToProjectID   string `json:"to_project_id" label:"目标项目ID" binding:"required"` // 目标项目ID
}

// UpdatePinRequest 更新Pin请求
type UpdatePinRequest struct {
	Priority *string `json:"priority"` // 优先级 (可选更新)
	Category *string `json:"category"` // 分类 (可选更新)
}

// PinListRequest Pin列表查询请求
type PinListRequest struct {
	ToProjectID   string `form:"to_project_id"`   // 目标项目ID筛选
	FromProjectID string `form:"from_project_id"` // 来源项目ID筛选
	Status        string `form:"status"`          // 状态筛选
	Category      string `form:"category"`        // 分类筛选
	Priority      string `form:"priority"`        // 优先级筛选
	Page          int    `form:"page"`            // 页码
	Size          int    `form:"size"`            // 每页数量
}

// PinResponse Pin响应
type PinResponse struct {
	ID            string `json:"id"`              // Pin ID (雪花ID字符串)
	Title         string `json:"title"`           // 标题
	Content       string `json:"content"`         // 内容
	Category      string `json:"category"`        // 分类
	Status        string `json:"status"`          // 状态
	Priority      string `json:"priority"`        // 优先级
	FromProjectID string `json:"from_project_id"` // 来源项目ID
	ToProjectID   string `json:"to_project_id"`   // 目标项目ID
	ConsumedAt    string `json:"consumed_at"`     // 消费时间 (空字符串表示未消费)
	CreatedAt     string `json:"created_at"`      // 创建时间
	UpdatedAt     string `json:"updated_at"`      // 更新时间
}

// PinListResponse Pin列表响应
type PinListResponse struct {
	Items []PinResponse `json:"items"` // Pin列表
	Total int64         `json:"total"` // 总数量
}
