package repowiki

import (
	"time"

	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	"github.com/xiaolfeng/Lumina/api/project"
)

// ConfigResponse RepoWiki 配置响应
type ConfigResponse struct {
	ID                xSnowflake.SnowflakeID  `json:"id"`                            // 配置ID
	ProjectID         xSnowflake.SnowflakeID  `json:"project_id"`                    // 关联项目ID
	ProjectInfo       *project.ProjectResponse `json:"project,omitempty"`            // 关联项目信息
	RepoURL           string                  `json:"repo_url"`                      // Git仓库地址
	CustomPrompt      string                  `json:"custom_prompt,omitempty"`       // 项目级自定义提示词
	DefaultBranch     string                  `json:"default_branch"`                // 默认分支
	DefaultLanguage   string                  `json:"default_language"`              // 默认Wiki语言
	Status            string                  `json:"status"`                        // 当前分析状态
	SSHKeyID          *xSnowflake.SnowflakeID `json:"ssh_key_id,omitempty"`          // 关联SSH密钥ID（字符串形式的雪花ID，nil 表示未关联）
	HasPassword       bool                    `json:"has_password"`                  // 是否设置Wiki密码
	SelectedVersionID *xSnowflake.SnowflakeID `json:"selected_version_id,omitempty"` // 当前选中的Wiki版本ID（nil 表示尚未生成或未选择）
	LastAccessedAt    *time.Time              `json:"last_accessed_at,omitempty"`    // 最后访问时间
	CreatedAt         time.Time               `json:"created_at"`                    // 创建时间
	UpdatedAt         time.Time               `json:"updated_at"`                    // 更新时间
	LatestVersion     *VersionStatusResponse  `json:"latest_version,omitempty"`      // 最近版本状态
}

// ConfigListResponse RepoWiki 配置列表响应（分页）
type ConfigListResponse struct {
	Total int64            `json:"total"` // 总数
	Items []ConfigResponse `json:"items"` // 配置列表
}
