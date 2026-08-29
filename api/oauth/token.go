package oauth

// TokenResponse 令牌端点成功响应（RFC 6749 §5.1）。
type TokenResponse struct {
	AccessToken  string `json:"access_token"`  // 访问令牌（lum_at_ 前缀）
	TokenType    string `json:"token_type"`    // 令牌类型，恒为 Bearer
	ExpiresIn    int64  `json:"expires_in"`    // 访问令牌有效期（秒）
	RefreshToken string `json:"refresh_token"` // 刷新令牌（lum_rt_ 前缀）
	Scope        string `json:"scope"`         // 授予作用域
}

// TokenErrorResponse 令牌端点错误响应（RFC 6749 §5.2），HTTP 400。
type TokenErrorResponse struct {
	Error            string `json:"error"`                       // 错误码（invalid_request / invalid_client / invalid_grant 等）
	ErrorDescription string `json:"error_description,omitempty"` // 错误描述
}
