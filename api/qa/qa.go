package qa

import xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"

// ListSessionRequest Session列表请求
type ListSessionRequest struct {
	Page   int    `form:"page"`   // 页码
	Size   int    `form:"size"`   // 每页数量
	Status string `form:"status"` // 状态过滤 active/expired/deleted/空(全部)
	Type   string `form:"type"`   // 类型过滤 temporary/permanent/空(全部)
	Hash   string `form:"hash"`   // 哈希过滤
}

// UpdateQaConfigRequest 更新Q&A配置
type UpdateQaConfigRequest struct {
	SessionTTL    *int    `json:"session_ttl"`    // Session TTL（秒），nil不更新
	RuntimeDomain *string `json:"runtime_domain"` // 运行时域名，nil不更新
	PollSlice     *int    `json:"poll_slice"`     // qa_get_answer 单次阻塞上限（秒），nil不更新
	MaxRetries    *int    `json:"max_retries"`    // qa_get_answer 最大重试次数（达到后返回STOPPED），nil不更新
}

// CreateSessionRequest 创建QA会话请求
type CreateSessionRequest struct {
	ProjectID xSnowflake.SnowflakeID `json:"project_id" label:"关联项目ID" binding:"required"` // 关联项目ID
	Title     string                 `json:"title" label:"会话标题" binding:"required"`        // 会话标题
	Agent     string                 `json:"agent" label:"Agent名称" binding:"required"`     // Agent名称
	Type      string                 `json:"type" label:"会话类型" binding:"required"`         // temporary/permanent
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

// QuestionSummaryResponse 问题摘要（列表用）
//
// 用于会话详情接口返回完整历史问答（含回答内容、选项、补充等），
// 支持前端 Interact 页面刷新后恢复历史问答。
type QuestionSummaryResponse struct {
	ID          xSnowflake.SnowflakeID `json:"id"`          // 问题ID
	Type        string                 `json:"type"`        // 题型
	Title       string                 `json:"title"`       // 标题
	Options     any                    `json:"options"`     // 选项配置
	Config      any                    `json:"config"`      // 题型特有配置
	Batch       any                    `json:"batch"`       // 分批信息
	GroupLabel  string                 `json:"group_label"` // 分组标签
	Supplement  bool                   `json:"supplement"`  // 是否携带补充内容
	Status      string                 `json:"status"`      // pending/answered/skipped
	Answer      any                    `json:"answer"`      // 回答数据
	Media       any                    `json:"media"`       // 多媒体数据
	Supplements []SupplementResponse   `json:"supplements"` // 关联补充内容
	CreatedAt   string                 `json:"created_at"`  // 创建时间
	AnsweredAt  string                 `json:"answered_at"` // 回答时间
}

// QuestionDetailResponse 问题详情（含回答+补充）
type QuestionDetailResponse struct {
	ID          xSnowflake.SnowflakeID `json:"id"`          // 问题ID
	SessionID   xSnowflake.SnowflakeID `json:"session_id"`  // 所属Session
	Type        string                 `json:"type"`        // 题型
	Title       string                 `json:"title"`       // 标题
	Description string                 `json:"description"` // 描述
	Options     any                    `json:"options"`     // 选项配置
	Config      any                    `json:"config"`      // 题型特有配置
	Batch       any                    `json:"batch"`       // 分批信息
	GroupLabel  string                 `json:"group_label"` // 分组标签
	Status      string                 `json:"status"`      // 状态
	Answer      any                    `json:"answer"`      // 回答数据
	Supplements []SupplementResponse   `json:"supplements"` // 关联补充内容
	CreatedAt   string                 `json:"created_at"`  // 创建时间
	AnsweredAt  string                 `json:"answered_at"` // 回答时间
}

// SupplementResponse 补充内容响应
type SupplementResponse struct {
	ID          xSnowflake.SnowflakeID `json:"id"`           // 补充内容ID
	TargetType  string                 `json:"target_type"`  // question/option
	TargetID    xSnowflake.SnowflakeID `json:"target_id"`    // 关联ID
	ContentType string                 `json:"content_type"` // markdown/html
	Content     string                 `json:"content"`      // 内容
	CreatedAt   string                 `json:"created_at"`   // 创建时间
	UpdatedAt   string                 `json:"updated_at"`   // 更新时间
}

// SessionListResponse Session列表响应
type SessionListResponse struct {
	Items []SessionResponse `json:"items"` // Session列表
	Total int64             `json:"total"` // 总数量
}

// QaConfigResponse Q&A配置响应
type QaConfigResponse struct {
	SessionTTL    int    `json:"session_ttl"`    // Session TTL（秒）
	RuntimeDomain string `json:"runtime_domain"` // 运行时域名
	PollSlice     int    `json:"poll_slice"`     // qa_get_answer 单次阻塞上限（秒）
	MaxRetries    int    `json:"max_retries"`    // qa_get_answer 最大重试次数
}
