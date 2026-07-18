package ssh

// CreateSshKeyRequest 创建 SSH 密钥请求
type CreateSshKeyRequest struct {
	Source      string `json:"source" label:"密钥来源" binding:"required"` // 密钥来源(generated|imported)
	Name        string `json:"name" label:"密钥名称" binding:"required"`   // 密钥名称
	Description string `json:"description,omitempty" label:"密钥描述"`     // 密钥描述
	PrivateKey  string `json:"private_key,omitempty" label:"PEM私钥"`    // PEM格式私钥（仅 imported 时必填）
}

// CreateSshKeyResponse 创建 SSH 密钥响应（创建后返回公钥+指纹，不含私钥）
type CreateSshKeyResponse struct {
	SshKeyResponse
	PublicKeyDownloadURL string `json:"public_key_download_url"` // 公钥下载地址
}
