package biometric

import "encoding/json"

// AvailabilityResponse 生物特征可用性响应
type AvailabilityResponse struct {
	IsAvailable bool `json:"is_available"` // 是否可用（已注册凭证）
}

// RegisterStartRequest 注册开始请求
type RegisterStartRequest struct {
	DeviceName string `json:"device_name" label:"设备名称" binding:"required,max=128"` // 设备名称
}

// RegisterStartResponse 注册开始响应（传递给浏览器 navigator.credentials.create）
type RegisterStartResponse struct {
	SessionToken string          `json:"session_token"`                // 会话令牌（关联 Redis challenge）
	Options      json.RawMessage `json:"options" swaggertype:"object"` // WebAuthn PublicKeyCredentialCreationOptions（透传 JSON）
}

// RegisterFinishRequest 注册完成请求（navigator.credentials.create 返回值）
type RegisterFinishRequest struct {
	SessionToken string          `json:"session_token" label:"会话令牌" binding:"required"`                   // 会话令牌
	DeviceName   string          `json:"device_name" label:"设备名称" binding:"required,max=128"`             // 设备名称
	Credential   json.RawMessage `json:"credential" label:"凭证数据" binding:"required" swaggertype:"object"` // CredentialAttestation JSON
}

// RegisterFinishResponse 注册完成响应
type RegisterFinishResponse struct {
	Success bool   `json:"success"` // 是否成功
	Message string `json:"message"` // 消息
}

// LoginStartResponse 登录开始响应（传递给浏览器 navigator.credentials.get）
type LoginStartResponse struct {
	SessionToken string          `json:"session_token"`                // 会话令牌
	Options      json.RawMessage `json:"options" swaggertype:"object"` // WebAuthn PublicKeyCredentialRequestOptions
}

// LoginFinishRequest 登录完成请求
type LoginFinishRequest struct {
	SessionToken string          `json:"session_token" label:"会话令牌" binding:"required"`                   // 会话令牌
	Credential   json.RawMessage `json:"credential" label:"凭证数据" binding:"required" swaggertype:"object"` // CredentialAssertion JSON
}

// LoginFinishResponse 登录完成响应
type LoginFinishResponse struct {
	AccessToken  string `json:"access_token"`  // 访问令牌
	RefreshToken string `json:"refresh_token"` // 刷新令牌
	ExpiresAt    int64  `json:"expires_at"`    // 过期时间戳
}
