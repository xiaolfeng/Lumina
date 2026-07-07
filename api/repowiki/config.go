package repowiki

import "time"

// CreateConfigRequest 创建 RepoWiki 配置请求
type CreateConfigRequest struct {
	RepoURL         string `json:"repo_url" binding:"required"` // Git仓库地址
	Name            string `json:"name" binding:"required"`     // 配置名称
	ProjectID       int64  `json:"project_id"`                  // 关联项目ID（可选）
	DefaultBranch   string `json:"default_branch"`              // 默认分支（默认 main）
	DefaultLanguage string `json:"default_language"`            // 默认Wiki语言（默认 zh）
	SSHKey          string `json:"ssh_key,omitempty"`           // SSH私钥（PEM格式，私有仓库用）
	WikiPassword    string `json:"wiki_password,omitempty"`     // Wiki访问密码（可选保护）
}

// UpdateConfigRequest 更新 RepoWiki 配置请求（全部 optional）
type UpdateConfigRequest struct {
	RepoURL         *string `json:"repo_url,omitempty"`         // Git仓库地址
	Name            *string `json:"name,omitempty"`             // 配置名称
	DefaultBranch   *string `json:"default_branch,omitempty"`   // 默认分支
	DefaultLanguage *string `json:"default_language,omitempty"` // 默认Wiki语言
	SSHKey          *string `json:"ssh_key,omitempty"`          // SSH私钥（传空字符串清除）
	WikiPassword    *string `json:"wiki_password,omitempty"`    // Wiki访问密码（传空字符串清除）
}

// ConfigListResponse RepoWiki 配置列表响应（分页）
type ConfigListResponse struct {
	Total int64            `json:"total"` // 总数
	Items []ConfigResponse `json:"items"` // 配置列表
}

// ConfigResponse RepoWiki 配置响应
type ConfigResponse struct {
	ID              int64                  `json:"id"`                         // 配置ID
	ProjectID       int64                  `json:"project_id"`                 // 关联项目ID
	RepoURL         string                 `json:"repo_url"`                   // Git仓库地址
	Name            string                 `json:"name"`                       // 配置名称
	DefaultBranch   string                 `json:"default_branch"`             // 默认分支
	DefaultLanguage string                 `json:"default_language"`           // 默认Wiki语言
	Status          string                 `json:"status"`                     // 当前分析状态
	HasSSHKey       bool                   `json:"has_ssh_key"`                // 是否配置SSH密钥
	HasPassword     bool                   `json:"has_password"`               // 是否设置Wiki密码
	LastAccessedAt  *time.Time             `json:"last_accessed_at,omitempty"` // 最后访问时间
	CreatedAt       time.Time              `json:"created_at"`                 // 创建时间
	UpdatedAt       time.Time              `json:"updated_at"`                 // 更新时间
	LatestVersion   *VersionStatusResponse `json:"latest_version,omitempty"`   // 最近版本状态
}

// WebhookConfigResponse webhook 配置信息
type WebhookConfigResponse struct {
	URL           string            `json:"url"`            // 完整 Webhook URL
	Token         string            `json:"token"`          // 脱敏 Token（cs_****）
	HasSecret     bool              `json:"has_secret"`     // 是否设置 Secret
	Branches      []string          `json:"branches"`       // 监控分支列表
	ProviderGuide map[string]string `json:"provider_guide"` // 提供商 -> 配置指南
}

// UpdateWebhookBranchesRequest 更新监控分支请求
type UpdateWebhookBranchesRequest struct {
	Branches []string `json:"branches" binding:"required"`
}

// RegenerateWebhookResponse 重新生成凭据一次性响应
type RegenerateWebhookResponse struct {
	Token  string `json:"token"`  // 完整 Token（仅展示一次）
	Secret string `json:"secret"` // 完整 Secret（仅展示一次）
	URL    string `json:"url"`    // 完整 Webhook URL
}

// WebhookEventListResponse webhook 事件列表响应（分页）
type WebhookEventListResponse struct {
	Total int64                  `json:"total"` // 总数
	Items []WebhookEventResponse `json:"items"` // 事件列表
}

// WebhookEventResponse 单个 webhook 事件响应
type WebhookEventResponse struct {
	ID           int64      `json:"id"`                     // 事件ID
	ConfigID     int64      `json:"config_id,omitempty"`    // 关联配置ID
	Provider     string     `json:"provider"`               // Git 提供商
	EventType    string     `json:"event_type"`             // 事件类型
	Branch       string     `json:"branch,omitempty"`       // 触发分支
	CommitBefore string     `json:"commit_before,omitempty"` // 变更前 commit
	CommitAfter  string     `json:"commit_after,omitempty"`  // 变更后 commit
	ChangedCount int        `json:"changed_count"`          // 变更文件数
	Status       string     `json:"status"`                 // 处理状态
	Reason       string     `json:"reason,omitempty"`       // 忽略/失败原因
	VersionID    int64      `json:"version_id,omitempty"`   // 关联版本ID
	ResponseCode int        `json:"response_code"`          // 响应 HTTP 状态码
	ReceivedAt   time.Time  `json:"received_at"`            // 接收时间
	ProcessedAt  *time.Time `json:"processed_at,omitempty"` // 处理时间
}
