package preview

import xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"

// PreviewSessionResponse 预览会话响应
type PreviewSessionResponse struct {
	ID        xSnowflake.SnowflakeID `json:"id"`         // 预览会话 ID
	ProjectID xSnowflake.SnowflakeID `json:"project_id"` // 关联项目ID
	Title     string                 `json:"title"`      // 会话标题
	Hash      string                 `json:"hash"`       // 访问哈希标识
	Status    string                 `json:"status"`     // 会话状态
	CreatedAt string                 `json:"created_at"` // 创建时间
	UpdatedAt string                 `json:"updated_at"` // 更新时间
}

// PreviewFileResponse 预览文件响应（不含 Content，文件内容经 serve 接口单独获取）
type PreviewFileResponse struct {
	ID        xSnowflake.SnowflakeID `json:"id"`         // 预览文件 ID
	SessionID xSnowflake.SnowflakeID `json:"session_id"` // 关联会话ID
	Filename  string                 `json:"filename"`   // 文件名
	MimeType  string                 `json:"mime_type"`  // MIME类型
	Size      int                    `json:"size"`       // 文件大小(字节)
	CreatedAt string                 `json:"created_at"` // 创建时间
	UpdatedAt string                 `json:"updated_at"` // 更新时间
}

// PreviewFileContentResponse 预览文件内容响应（serve 接口专用，含完整内容）
type PreviewFileContentResponse struct {
	Filename string `json:"filename"`  // 文件名
	MimeType string `json:"mime_type"` // MIME类型
	Content  string `json:"content"`   // 文件内容
}

// PreviewSessionDetailResponse 预览会话详情响应（含文件列表，公开访问用）
type PreviewSessionDetailResponse struct {
	Session PreviewSessionResponse `json:"session"` // 会话信息
	Files   []PreviewFileResponse  `json:"files"`   // 文件列表
}
