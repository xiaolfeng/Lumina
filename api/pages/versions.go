package pages

import xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"

// PageVersionResponse 页面版本摘要
type PageVersionResponse struct {
	ID              xSnowflake.SnowflakeID  `json:"id"`                          // 版本 ID
	PageID          xSnowflake.SnowflakeID  `json:"page_id"`                     // 所属页面 ID
	Version         string                  `json:"version"`                     // 语义化版本号
	Changelog       string                  `json:"changelog"`                   // 更新说明
	SourceSessionID *xSnowflake.SnowflakeID `json:"source_session_id,omitempty"` // 来源预览会话
	BaseVersionID   *xSnowflake.SnowflakeID `json:"base_version_id,omitempty"`   // 派生基准版本
	EntryFilename   string                  `json:"entry_filename"`              // 入口文件
	FileCount       int                     `json:"file_count"`                  // 文件数
	TotalSize       int64                   `json:"total_size"`                  // 快照总字节
	CreatedBy       string                  `json:"created_by"`                  // 发布者
	IsActive        bool                    `json:"is_active"`                   // 是否为当前生效版本
	CreatedAt       string                  `json:"created_at"`                  // 创建时间
}

// PageVersionListResponse 版本列表
type PageVersionListResponse struct {
	Items []PageVersionResponse `json:"items"` // 版本列表
}
