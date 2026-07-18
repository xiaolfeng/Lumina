package qa

// UpdateQaConfigRequest 更新Q&A配置
type UpdateQaConfigRequest struct {
	SessionTTL    *int    `json:"session_ttl"`    // Session TTL（秒），nil不更新
	RuntimeDomain *string `json:"runtime_domain"` // 运行时域名，nil不更新
	PollSlice     *int    `json:"poll_slice"`     // qa_get_answer 单次阻塞上限（秒），nil不更新
	MaxRetries    *int    `json:"max_retries"`    // qa_get_answer 最大重试次数（达到后返回STOPPED），nil不更新
}

// QaConfigResponse Q&A配置响应
type QaConfigResponse struct {
	SessionTTL    int    `json:"session_ttl"`    // Session TTL（秒）
	RuntimeDomain string `json:"runtime_domain"` // 运行时域名
	PollSlice     int    `json:"poll_slice"`     // qa_get_answer 单次阻塞上限（秒）
	MaxRetries    int    `json:"max_retries"`    // qa_get_answer 最大重试次数
}
