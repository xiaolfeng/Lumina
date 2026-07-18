package repowiki

import xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"

// CreateConfigRequest 创建 RepoWiki 配置请求
type CreateConfigRequest struct {
	RepoURL         string                  `json:"repo_url" label:"Git仓库地址" binding:"required"`  // Git仓库地址
	ProjectID       xSnowflake.SnowflakeID  `json:"project_id" label:"关联项目ID" binding:"required"` // 关联项目ID（必选）
	DefaultBranch   string                  `json:"default_branch" label:"默认分支"`                  // 默认分支（默认 main）
	DefaultLanguage string                  `json:"default_language" label:"默认Wiki语言"`            // 默认Wiki语言（默认 zh）
	SSHKeyID        *xSnowflake.SnowflakeID `json:"ssh_key_id,omitempty" label:"SSH密钥ID"`         // 关联SSH密钥ID（可选，私有仓库用）
	WikiPassword    string                  `json:"wiki_password,omitempty" label:"Wiki访问密码"`     // Wiki访问密码（可选保护）
}
