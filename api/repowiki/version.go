package repowiki

import (
	"time"

	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
)

// AnalyzeRequest 触发分析请求
type AnalyzeRequest struct {
	Language    string `json:"language,omitempty"`     // Wiki语言（默认使用配置的 default_language）
	Branch      string `json:"branch,omitempty"`       // 分析分支（默认使用配置的 default_branch）
	ExtraPrompt string `json:"extra_prompt,omitempty" binding:"omitempty,max=5000"` // 本次分析额外提示词（最长 5000 字符）
}

// AnalyzeResponse 触发分析响应
type AnalyzeResponse struct {
	VersionID xSnowflake.SnowflakeID `json:"version_id"` // 版本ID
	Status    string                 `json:"status"`     // 初始状态（pending）
}

// VersionStatusResponse 版本状态响应
type VersionStatusResponse struct {
	ID              xSnowflake.SnowflakeID `json:"id"`                      // 版本ID
	ConfigID        xSnowflake.SnowflakeID `json:"config_id"`               // 关联配置ID
	CommitHash      string                 `json:"commit_hash"`             // Git commit hash
	Branch          string                 `json:"branch"`                  // 分析分支
	Language        string                 `json:"language"`                // Wiki语言
	Status          string                 `json:"status"`                  // 分析状态
	CurrentStage    string                 `json:"current_stage,omitempty"` // 当前阶段
	ProgressPercent int                    `json:"progress_percent"`        // 进度百分比（0-100）
	ErrorMsg        string                 `json:"error_msg,omitempty"`     // 失败原因
	FileCount       int                    `json:"file_count"`              // 分析文件数
	TokenCount      int64                  `json:"token_count"`             // LLM token消耗
	DurationMs      int                    `json:"duration_ms"`             // 分析耗时毫秒
	StartedAt       *time.Time             `json:"started_at,omitempty"`    // 分析开始时间
	CompletedAt     *time.Time             `json:"completed_at,omitempty"`  // 分析完成时间
	CreatedAt       time.Time              `json:"created_at"`              // 创建时间
}

// VersionListResponse 版本列表响应（分页）
type VersionListResponse struct {
	Total int64                   `json:"total"` // 总数
	Items []VersionStatusResponse `json:"items"` // 版本列表
}
