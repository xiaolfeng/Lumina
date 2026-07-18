package user

import xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"

// BiometricCredentialItem 生物特征凭证项
type BiometricCredentialItem struct {
	ID         xSnowflake.SnowflakeID `json:"id"`           // 凭证 ID
	DeviceName string                 `json:"device_name"`  // 设备名称
	AAGUID     string                 `json:"aaguid"`       // 认证器型号
	LastUsedAt *int64                 `json:"last_used_at"` // 最后使用时间戳（nil 表示未使用）
	CreatedAt  int64                  `json:"created_at"`   // 创建时间戳
}

// BiometricCredentialListResponse 凭证列表响应
type BiometricCredentialListResponse struct {
	Total int                       `json:"total"` // 总数
	Items []BiometricCredentialItem `json:"items"` // 凭证列表
}
