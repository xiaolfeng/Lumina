package pin

import xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"

// PinResponse Pin响应
type PinResponse struct {
	ID            xSnowflake.SnowflakeID `json:"id"`              // Pin ID (雪花ID字符串)
	Title         string                 `json:"title"`           // 标题
	Content       string                 `json:"content"`         // 内容
	Category      string                 `json:"category"`        // 分类
	Status        string                 `json:"status"`          // 状态
	Priority      string                 `json:"priority"`        // 优先级
	FromProjectID xSnowflake.SnowflakeID `json:"from_project_id"` // 来源项目ID
	ToProjectID   xSnowflake.SnowflakeID `json:"to_project_id"`   // 目标项目ID
	ConsumedAt    string                 `json:"consumed_at"`     // 消费时间 (空字符串表示未消费)
	CreatedAt     string                 `json:"created_at"`      // 创建时间
	UpdatedAt     string                 `json:"updated_at"`      // 更新时间
}
