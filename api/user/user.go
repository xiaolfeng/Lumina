package user

// UserInfoResponse 用户信息响应
type UserInfoResponse struct {
	Username                 string `json:"username"`                    // 用户名
	Email                    string `json:"email"`                       // 邮箱地址
	BiometricEnabled         bool   `json:"biometric_enabled"`           // 是否启用生物特征
	BiometricCredentialCount int    `json:"biometric_credential_count"` // 生物特征凭证数量
}

// UpdateProfileRequest 更新个人资料请求
type UpdateProfileRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"` // 用户名
	Email    string `json:"email" binding:"required,email"`           // 邮箱
}

// UpdatePasswordRequest 修改密码请求
type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required,min=6"` // 旧密码
	NewPassword string `json:"new_password" binding:"required,min=8"` // 新密码
}

// BiometricCredentialItem 生物特征凭证项
type BiometricCredentialItem struct {
	ID         string `json:"id"`          // 凭证 ID
	DeviceName string `json:"device_name"` // 设备名称
	AAGUID     string `json:"aaguid"`      // 认证器型号
	LastUsedAt *int64 `json:"last_used_at"` // 最后使用时间戳（nil 表示未使用）
	CreatedAt  int64  `json:"created_at"`  // 创建时间戳
}

// BiometricCredentialListResponse 凭证列表响应
type BiometricCredentialListResponse struct {
	Total int                       `json:"total"` // 总数
	Items []BiometricCredentialItem `json:"items"` // 凭证列表
}
