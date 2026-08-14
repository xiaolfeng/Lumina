package dashboard

// TokenStats 令牌统计
type TokenStats struct {
	Total  int64 `json:"total"`  // 令牌总数
	Active int64 `json:"active"` // 活跃令牌数
}

// QaStats 问答会话统计
type QaStats struct {
	Total            int64 `json:"total"`             // 会话总数
	Active           int64 `json:"active"`            // 活跃会话
	Expired          int64 `json:"expired"`           // 已归档（过期）会话
	Deleted          int64 `json:"deleted"`           // 已删除会话
	PendingQuestions int64 `json:"pending_questions"` // 待回答问题总数
}

// PreviewStats 预览会话统计
type PreviewStats struct {
	Total  int64 `json:"total"`  // 预览会话总数
	Active int64 `json:"active"` // 活跃会话
	Files  int64 `json:"files"`  // 文件总数
}

// RepoWikiStats RepoWiki 统计
type RepoWikiStats struct {
	Configs    int64 `json:"configs"`    // 配置总数
	Versions   int64 `json:"versions"`   // 版本总数
	Completed  int64 `json:"completed"`  // 已完成版本数
	Generating int64 `json:"generating"` // 生成中版本数
}

// RecentPreviewItem 最近预览会话项
type RecentPreviewItem struct {
	ID        string `json:"id"`         // 会话ID
	Title     string `json:"title"`      // 会话标题
	Hash      string `json:"hash"`       // 访问哈希
	FileCount int64  `json:"file_count"` // 文件数
	Status    string `json:"status"`     // 会话状态
	UpdatedAt string `json:"updated_at"` // 更新时间
}

// OverviewResponse 看板概览响应
type OverviewResponse struct {
	Tokens         TokenStats          `json:"tokens"`          // 令牌统计
	Projects       int64               `json:"projects"`        // 项目总数
	Qa             QaStats             `json:"qa"`              // 问答统计
	Preview        PreviewStats        `json:"preview"`         // 预览统计
	RepoWiki       RepoWikiStats       `json:"repowiki"`        // RepoWiki 统计
	RecentPreviews []RecentPreviewItem `json:"recent_previews"` // 最近预览列表
}
