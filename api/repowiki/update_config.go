package repowiki

import xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"

// UpdateConfigRequest 更新 RepoWiki 配置请求（全部 optional）
type UpdateConfigRequest struct {
	RepoURL         *string                 `json:"repo_url,omitempty"`         // Git仓库地址
	DefaultBranch   *string                 `json:"default_branch,omitempty"`   // 默认分支
	DefaultLanguage *string                 `json:"default_language,omitempty"` // 默认Wiki语言
	SSHKeyID        *xSnowflake.SnowflakeID `json:"ssh_key_id,omitempty"`       // 关联SSH密钥ID（传 0 或 nil 清除关联）
	WikiPassword    *string                 `json:"wiki_password,omitempty"`    // Wiki访问密码（传空字符串清除）
	CustomPrompt    *string                 `json:"custom_prompt,omitempty" binding:"omitempty,max=10000"` // 项目级自定义提示词（最长 10000 字符）
}

// UpdateSelectedVersionRequest 切换当前选中的 Wiki 版本
type UpdateSelectedVersionRequest struct {
	VersionID xSnowflake.SnowflakeID `json:"version_id" label:"Wiki版本ID" binding:"required"` // Wiki版本ID（必须属于该配置且状态为 completed）
}
