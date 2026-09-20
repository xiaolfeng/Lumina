package preview

import xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"

// PromoteSessionRequest 将预览会话晋升为 Pages 快照的请求（不含密码字段）
type PromoteSessionRequest struct {
	Slug            string `json:"slug" label:"页面标识" binding:"max=64"`            // 项目内访问标识；Fork 发布新版本时可空
	Title           string `json:"title" label:"页面标题" binding:"required,max=255"` // 页面显示标题
	Description     string `json:"description" label:"页面描述"`                      // 页面描述
	Version         string `json:"version" label:"版本号" binding:"max=32"`          // 语义化版本号，空则自动递增
	Changelog       string `json:"changelog" label:"更新说明"`                        // 版本更新说明
	SetAsActive     bool   `json:"set_as_active"`                                 // 是否立即设为线上生效版本
	ConfirmConflict bool   `json:"confirm_conflict"`                              // 线上指针已前进时确认覆盖
}

// PromoteConflictResponse 晋升并发冲突详情
type PromoteConflictResponse struct {
	CurrentVersion   string                 `json:"current_version"`    // 当前线上版本号
	CurrentVersionID xSnowflake.SnowflakeID `json:"current_version_id"` // 当前线上版本 ID
	CurrentChangelog string                 `json:"current_changelog"`  // 当前线上版本说明
	CurrentCreatedBy string                 `json:"current_created_by"` // 当前线上版本发布者
	CurrentCreatedAt string                 `json:"current_created_at"` // 当前线上版本发布时间
	SourceVersionID  xSnowflake.SnowflakeID `json:"source_version_id"`  // 本次草稿基于的版本 ID
	Message          string                 `json:"message"`            // 冲突提示文案
}
