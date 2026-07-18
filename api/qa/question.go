package qa

import xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"

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
