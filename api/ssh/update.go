package ssh

// UpdateSshKeyRequest 更新 SSH 密钥请求（仅 name/description 可更新）
type UpdateSshKeyRequest struct {
	Name        *string `json:"name,omitempty"`        // 密钥名称
	Description *string `json:"description,omitempty"` // 密钥描述
}
