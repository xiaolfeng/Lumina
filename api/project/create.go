package project

// CreateProjectRequest 创建项目请求
type CreateProjectRequest struct {
	Name        string   `json:"name" label:"项目名称" binding:"required"` // 项目名称
	AliasName   string   `json:"alias_name" label:"项目别名"`              // 项目别名
	MatchPath   []string `json:"match_path" label:"项目路径匹配列表"`          // 项目路径匹配列表
	Description string   `json:"description" label:"项目描述"`             // 项目描述
}
