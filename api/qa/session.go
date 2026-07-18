package qa

import xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"

// CreateSessionRequest 创建QA会话请求
type CreateSessionRequest struct {
	ProjectID xSnowflake.SnowflakeID `json:"project_id" label:"关联项目ID" binding:"required"` // 关联项目ID
	Title     string                 `json:"title" label:"会话标题" binding:"required"`        // 会话标题
	Agent     string                 `json:"agent" label:"Agent名称" binding:"required"`     // Agent名称
	Type      string                 `json:"type" label:"会话类型" binding:"required"`         // temporary/permanent
}

// ListSessionRequest Session列表请求
type ListSessionRequest struct {
	Page   int    `form:"page"`   // 页码
	Size   int    `form:"size"`   // 每页数量
	Status string `form:"status"` // 状态过滤 active/expired/deleted/空(全部)
	Type   string `form:"type"`   // 类型过滤 temporary/permanent/空(全部)
	Hash   string `form:"hash"`   // 哈希过滤
}

// SessionResponse Session响应
type SessionResponse struct {
	ID            xSnowflake.SnowflakeID `json:"id"`             // Session ID
	Hash          string                 `json:"hash"`           // 会话哈希标识
	Title         string                 `json:"title"`          // 会话标题
	Agent         string                 `json:"agent"`          // Agent名称
	Type          string                 `json:"type"`           // temporary/permanent
	Status        string                 `json:"status"`         // active/expired/deleted
	OnlineDevices int                    `json:"online_devices"` // 在线设备数
	ExpiresAt     string                 `json:"expires_at"`     // 过期时间（永久为空）
	CreatedAt     string                 `json:"created_at"`     // 创建时间
	UpdatedAt     string                 `json:"updated_at"`     // 更新时间
	ProjectID     xSnowflake.SnowflakeID `json:"project_id"`     // 关联项目ID
	ProjectName   string                 `json:"project_name"`   // 关联项目名称
}

// SessionDetailResponse Session详情响应（含问题列表）
type SessionDetailResponse struct {
	ID            xSnowflake.SnowflakeID    `json:"id"`             // Session ID
	Hash          string                    `json:"hash"`           // 会话哈希标识
	Title         string                    `json:"title"`          // 会话标题
	Agent         string                    `json:"agent"`          // Agent名称
	Type          string                    `json:"type"`           // temporary/permanent
	Status        string                    `json:"status"`         // active/expired/deleted
	OnlineDevices int                       `json:"online_devices"` // 在线设备数
	ExpiresAt     string                    `json:"expires_at"`     // 过期时间（永久为空）
	CreatedAt     string                    `json:"created_at"`     // 创建时间
	UpdatedAt     string                    `json:"updated_at"`     // 更新时间
	Questions     []QuestionSummaryResponse `json:"questions"`      // 问题列表
	ProjectID     xSnowflake.SnowflakeID    `json:"project_id"`     // 关联项目ID
	ProjectName   string                    `json:"project_name"`   // 关联项目名称
}

// SessionListResponse Session列表响应
type SessionListResponse struct {
	Items []SessionResponse `json:"items"` // Session列表
	Total int64             `json:"total"` // 总数量
}
