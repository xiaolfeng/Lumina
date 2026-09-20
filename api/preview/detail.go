package preview

import xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"

// PreviewSessionResponse 预览会话响应
type PreviewSessionResponse struct {
	ID              xSnowflake.SnowflakeID  `json:"id"`                          // 预览会话 ID
	ProjectID       xSnowflake.SnowflakeID  `json:"project_id"`                  // 关联项目ID
	Title           string                  `json:"title"`                       // 会话标题
	Hash            string                  `json:"hash"`                        // 访问哈希标识
	Status          string                  `json:"status"`                      // 会话状态
	FileCount       int64                   `json:"file_count"`                  // 文件数量
	ExpiresAt       string                  `json:"expires_at"`                  // 过期时间
	SourcePageID    *xSnowflake.SnowflakeID `json:"source_page_id,omitempty"`    // Fork 来源页面
	SourcePageSlug  string                  `json:"source_page_slug,omitempty"`  // Fork 来源页面 slug
	SourceVersionID *xSnowflake.SnowflakeID `json:"source_version_id,omitempty"` // Fork 基准版本
	CreatedAt       string                  `json:"created_at"`                  // 创建时间
	UpdatedAt       string                  `json:"updated_at"`                  // 更新时间
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

// PreviewFileLinesResponse 预览文件行级读取响应（MCP 行区间读取用；区间模式下内容带行号）
type PreviewFileLinesResponse struct {
	Filename   string `json:"filename"`    // 文件名
	MimeType   string `json:"mime_type"`   // MIME类型
	Size       int    `json:"size"`        // 文件大小(字节)
	TotalLines int    `json:"total_lines"` // 总行数
	StartLine  int    `json:"start_line"`  // 返回内容起始行（1 起始；空文件为 0）
	EndLine    int    `json:"end_line"`    // 返回内容结束行（闭区间；空文件为 0）
	Content    string `json:"content"`     // 文件内容（区间模式按「行号| 文本」格式，全量模式为原始内容）
}

// PreviewSessionDetailResponse 预览会话详情响应（含文件列表，公开访问用）
type PreviewSessionDetailResponse struct {
	Session PreviewSessionResponse `json:"session"` // 会话信息
	Files   []PreviewFileResponse  `json:"files"`   // 文件列表
}

// PreviewFileDetailResponse 预览文件详情响应（含关联会话哈希，供 preview supplement 渲染解析 serve 地址）
type PreviewFileDetailResponse struct {
	PreviewFileResponse
	SessionHash string `json:"session_hash"` // 关联会话访问哈希
}
