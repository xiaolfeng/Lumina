package ssh

// SshKeyListResponse SSH 密钥列表响应（分页）
type SshKeyListResponse struct {
	Total int64            `json:"total"` // 总数
	Items []SshKeyResponse `json:"items"` // 密钥列表
}
