package ssh

import xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"

// SshKeyResponse SSH 密钥响应（不含 PrivateKey）
type SshKeyResponse struct {
	ID          xSnowflake.SnowflakeID `json:"id"`          // 密钥ID
	Name        string                 `json:"name"`        // 密钥名称
	Description string                 `json:"description"` // 密钥描述
	KeyType     string                 `json:"key_type"`    // 密钥类型(ed25519|rsa)
	PublicKey   string                 `json:"public_key"`  // OpenSSH格式公钥
	Fingerprint string                 `json:"fingerprint"` // SHA256指纹
	Source      string                 `json:"source"`      // 密钥来源(generated|imported)
	CreatedAt   string                 `json:"created_at"`  // 创建时间
	UpdatedAt   string                 `json:"updated_at"`  // 更新时间
}
