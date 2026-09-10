package project

import "encoding/json"

// UpdateProjectRequest 更新项目请求
type UpdateProjectRequest struct {
	WorkspaceID *json.Number `json:"workspace_id,omitempty" label:"所属空间ID" swaggertype:"string"` // 目标空间ID，未提供时保持原空间
	Name        string       `json:"name" label:"项目名称" binding:"required"`                       // 项目名称
	AliasName   string       `json:"alias_name" label:"项目别名"`                                    // 项目别名
	MatchPath   []string     `json:"match_path" label:"项目路径匹配列表"`                                // 项目路径匹配列表
	Description string       `json:"description" label:"项目描述"`                                   // 项目描述
}
