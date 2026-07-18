package user

// UserInfoResponse 用户信息响应
type UserInfoResponse struct {
	Username                 string `json:"username"`                   // 用户名
	Email                    string `json:"email"`                      // 邮箱地址
	BiometricEnabled         bool   `json:"biometric_enabled"`          // 是否启用生物特征
	BiometricCredentialCount int    `json:"biometric_credential_count"` // 生物特征凭证数量
}
