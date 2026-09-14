package pages

import xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"

// PageResponse 页面管理详情（不含密码哈希）
type PageResponse struct {
	ID              xSnowflake.SnowflakeID `json:"id"`                // 页面 ID
	ProjectID       xSnowflake.SnowflakeID `json:"project_id"`        // 所属项目 ID
	ProjectName     string                 `json:"project_name"`      // 所属项目名称
	Slug            string                 `json:"slug"`              // 项目内访问标识
	Title           string                 `json:"title"`             // 页面显示标题
	Description     string                 `json:"description"`       // 页面描述
	Status          string                 `json:"status"`            // published / archived
	AccessMode      string                 `json:"access_mode"`       // public / password
	LatestVersionID xSnowflake.SnowflakeID `json:"latest_version_id"` // 当前线上生效版本指针
	LatestVersion   string                 `json:"latest_version"`    // 当前线上版本号
	PageURL         string                 `json:"page_url"`          // 对外访问路径
	CreatedAt       string                 `json:"created_at"`        // 创建时间
	UpdatedAt       string                 `json:"updated_at"`        // 更新时间
}

// PageFileResponse 页面快照文件（不含正文）
type PageFileResponse struct {
	ID        xSnowflake.SnowflakeID `json:"id"`         // 文件 ID
	VersionID xSnowflake.SnowflakeID `json:"version_id"` // 所属版本 ID
	Filename  string                 `json:"filename"`   // 扁平文件名
	MimeType  string                 `json:"mime_type"`  // MIME 类型
	Size      int                    `json:"size"`       // 字节数
	CreatedAt string                 `json:"created_at"` // 创建时间
	UpdatedAt string                 `json:"updated_at"` // 更新时间
}

// PageFileContentResponse 页面文件正文（serve 接口专用）
type PageFileContentResponse struct {
	Filename string `json:"filename"`  // 文件名
	MimeType string `json:"mime_type"` // MIME 类型
	Content  string `json:"content"`   // 文件内容
}

// PagePublicMetaResponse 展示态元信息（公开或解锁后）
type PagePublicMetaResponse struct {
	Page    PageResponse        `json:"page"`    // 页面元信息
	Version PageVersionResponse `json:"version"` // 当前渲染版本
	Files   []PageFileResponse  `json:"files"`   // 文件清单（不含正文）
}
