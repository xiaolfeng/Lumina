package qa

import xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"

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
