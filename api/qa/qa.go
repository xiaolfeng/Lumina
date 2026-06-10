package qa

// ListSessionRequest Session列表请求
type ListSessionRequest struct {
	Page   int    `form:"page"`   // 页码
	Size   int    `form:"size"`   // 每页数量
	Status string `form:"status"` // 状态过滤 active/expired/deleted/空(全部)
	Type   string `form:"type"`   // 类型过滤 temporary/permanent/空(全部)
}

// UpdateQaConfigRequest 更新Q&A配置
type UpdateQaConfigRequest struct {
	SessionTTL    *int    `json:"session_ttl"`    // Session TTL（秒），nil不更新
	RuntimeDomain *string `json:"runtime_domain"` // 运行时域名，nil不更新
}

// SessionResponse Session响应
type SessionResponse struct {
	ID            string `json:"id"`             // Session ID
	Title         string `json:"title"`          // 会话标题
	Agent         string `json:"agent"`          // Agent名称
	Type          string `json:"type"`           // temporary/permanent
	Status        string `json:"status"`         // active/expired/deleted
	OnlineDevices int    `json:"online_devices"` // 在线设备数
	ExpiresAt     string `json:"expires_at"`     // 过期时间（永久为空）
	CreatedAt     string `json:"created_at"`     // 创建时间
	UpdatedAt     string `json:"updated_at"`     // 更新时间
}

// SessionDetailResponse Session详情响应（含问题列表）
type SessionDetailResponse struct {
	ID            string                    `json:"id"`             // Session ID
	Title         string                    `json:"title"`          // 会话标题
	Agent         string                    `json:"agent"`          // Agent名称
	Type          string                    `json:"type"`           // temporary/permanent
	Status        string                    `json:"status"`         // active/expired/deleted
	OnlineDevices int                       `json:"online_devices"` // 在线设备数
	ExpiresAt     string                    `json:"expires_at"`     // 过期时间（永久为空）
	CreatedAt     string                    `json:"created_at"`     // 创建时间
	UpdatedAt     string                    `json:"updated_at"`     // 更新时间
	Questions     []QuestionSummaryResponse `json:"questions"`     // 问题列表
}

// QuestionSummaryResponse 问题摘要（列表用）
type QuestionSummaryResponse struct {
	ID         string `json:"id"`          // 问题ID
	Type       string `json:"type"`        // 题型
	Title      string `json:"title"`       // 标题
	Status     string `json:"status"`      // pending/answered/skipped
	CreatedAt  string `json:"created_at"`  // 创建时间
	AnsweredAt string `json:"answered_at"` // 回答时间
}

// QuestionDetailResponse 问题详情（含回答+补充）
type QuestionDetailResponse struct {
	ID          string               `json:"id"`           // 问题ID
	SessionID   string               `json:"session_id"`   // 所属Session
	Type        string               `json:"type"`         // 题型
	Title       string               `json:"title"`        // 标题
	Description string               `json:"description"`  // 描述
	Options     any                  `json:"options"`      // 选项配置
	Config      any                  `json:"config"`       // 题型特有配置
	Batch       any                  `json:"batch"`        // 分批信息
	GroupLabel  string               `json:"group_label"`  // 分组标签
	Status      string               `json:"status"`       // 状态
	Answer      any                  `json:"answer"`       // 回答数据
	Supplements []SupplementResponse `json:"supplements"`  // 关联补充内容
	CreatedAt   string               `json:"created_at"`   // 创建时间
	AnsweredAt  string               `json:"answered_at"`  // 回答时间
}

// SupplementResponse 补充内容响应
type SupplementResponse struct {
	ID          string `json:"id"`           // 补充内容ID
	TargetType  string `json:"target_type"`  // question/option
	TargetID    string `json:"target_id"`    // 关联ID
	ContentType string `json:"content_type"` // markdown/html
	Content     string `json:"content"`      // 内容
	CreatedAt   string `json:"created_at"`   // 创建时间
	UpdatedAt   string `json:"updated_at"`   // 更新时间
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
}
