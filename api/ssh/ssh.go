package ssh

// CreateSshKeyRequest 创建 SSH 密钥请求
type CreateSshKeyRequest struct {
	Source      string `json:"source" label:"密钥来源" binding:"required"`       // 密钥来源(generated|imported)
	Name        string `json:"name" label:"密钥名称" binding:"required"`         // 密钥名称
	Description string `json:"description,omitempty" label:"密钥描述"`           // 密钥描述
	PrivateKey  string `json:"private_key,omitempty" label:"PEM私钥"`           // PEM格式私钥（仅 imported 时必填）
}

// UpdateSshKeyRequest 更新 SSH 密钥请求（仅 name/description 可更新）
type UpdateSshKeyRequest struct {
	Name        *string `json:"name,omitempty"`        // 密钥名称
	Description *string `json:"description,omitempty"` // 密钥描述
}

// SshKeyResponse SSH 密钥响应（不含 PrivateKey）
type SshKeyResponse struct {
	ID          string `json:"id"`          // 密钥ID
	Name        string `json:"name"`        // 密钥名称
	Description string `json:"description"` // 密钥描述
	KeyType     string `json:"key_type"`    // 密钥类型(ed25519|rsa)
	PublicKey   string `json:"public_key"`  // OpenSSH格式公钥
	Fingerprint string `json:"fingerprint"` // SHA256指纹
	Source      string `json:"source"`      // 密钥来源(generated|imported)
	CreatedAt   string `json:"created_at"`  // 创建时间
	UpdatedAt   string `json:"updated_at"`  // 更新时间
}

// SshKeyListResponse SSH 密钥列表响应（分页）
type SshKeyListResponse struct {
	Total int64            `json:"total"` // 总数
	Items []SshKeyResponse `json:"items"` // 密钥列表
}

// CreateSshKeyResponse 创建 SSH 密钥响应（创建后返回公钥+指纹，不含私钥）
type CreateSshKeyResponse struct {
	SshKeyResponse
	PublicKeyDownloadURL string `json:"public_key_download_url"` // 公钥下载地址
}
