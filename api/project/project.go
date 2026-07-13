package project

// CreateProjectRequest 创建项目请求
type CreateProjectRequest struct {
	Name        string   `json:"name" label:"项目名称" binding:"required"`       // 项目名称
	AliasName   string   `json:"alias_name" label:"项目别名"`                    // 项目别名
	MatchPath   []string `json:"match_path" label:"项目路径匹配列表"`             // 项目路径匹配列表
	Description string   `json:"description" label:"项目描述"`                   // 项目描述
}

// UpdateProjectRequest 更新项目请求
type UpdateProjectRequest struct {
	Name        string   `json:"name" label:"项目名称" binding:"required"` // 项目名称
	AliasName   string   `json:"alias_name" label:"项目别名"`              // 项目别名
	MatchPath   []string `json:"match_path" label:"项目路径匹配列表"`       // 项目路径匹配列表
	Description string   `json:"description" label:"项目描述"`             // 项目描述
}

// ProjectResponse 项目响应
type ProjectResponse struct {
	ID          string   `json:"id"`          // 项目ID
	Name        string   `json:"name"`        // 项目名称
	AliasName   string   `json:"alias_name"`  // 项目别名
	MatchPath   []string `json:"match_path"`  // 项目路径匹配列表
	Description string   `json:"description"` // 项目描述
	CreatedAt   string   `json:"created_at"`  // 创建时间
	UpdatedAt   string   `json:"updated_at"`  // 更新时间
}

// ProjectListResponse 项目列表响应
type ProjectListResponse struct {
	Items []ProjectResponse `json:"items"` // 项目列表
	Total int64             `json:"total"` // 总数量
}
