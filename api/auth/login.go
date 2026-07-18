package auth

// LoginRequest 登录请求
type LoginRequest struct {
	Account  string `json:"account" label:"账号" binding:"required"`  // 账号（用户名或邮箱）
	Password string `json:"password" label:"密码" binding:"required"` // 登录密码
}

// TokenResponse Token 响应
type TokenResponse struct {
	AccessToken  string `json:"access_token"`  // 访问令牌
	RefreshToken string `json:"refresh_token"` // 刷新令牌
	ExpiresIn    int64  `json:"expires_in"`    // 过期时间（秒）
}
