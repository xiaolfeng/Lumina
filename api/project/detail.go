package project

import xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"

// ProjectResponse 项目响应
type ProjectResponse struct {
	ID          xSnowflake.SnowflakeID `json:"id"`          // 项目ID
	Name        string                 `json:"name"`        // 项目名称
	AliasName   string                 `json:"alias_name"`  // 项目别名
	MatchPath   []string               `json:"match_path"`  // 项目路径匹配列表
	Description string                 `json:"description"` // 项目描述
	CreatedAt   string                 `json:"created_at"`  // 创建时间
	UpdatedAt   string                 `json:"updated_at"`  // 更新时间
}

// ProjectListResponse 项目列表响应
type ProjectListResponse struct {
	Items []ProjectResponse `json:"items"` // 项目列表
	Total int64             `json:"total"` // 总数量
}
