package pages

import (
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	apiPreview "github.com/xiaolfeng/Lumina/api/preview"
)

// ForkPageRequest 从页面版本派生新预览会话
type ForkPageRequest struct {
	VersionID xSnowflake.SnowflakeID `json:"version_id"` // 可选，默认当前生效版本
}

// ForkPageResponse Fork 结果
type ForkPageResponse struct {
	Session    apiPreview.PreviewSessionResponse `json:"session"`     // 新预览会话
	PreviewURL string                            `json:"preview_url"` // 工作台路径
}
