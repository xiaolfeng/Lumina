package biometric

import "encoding/json"

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
