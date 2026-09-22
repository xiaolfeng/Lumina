package auth

// LoginRequest 登录请求
type LoginRequest struct {
	Account  string `json:"account" label:"账号" binding:"required"`  // 账号（用户名或邮箱）
	Password string `json:"password" label:"密码" binding:"required"` // 登录密码
}

// TokenResponse Token 响应
type TokenResponse struct {
	AccessToken      string `json:"access_token"`       // 访问令牌
	RefreshToken     string `json:"refresh_token"`      // 刷新令牌
	ExpiresIn        int64  `json:"expires_in"`         // 访问令牌有效期（秒）
	RefreshExpiresIn int64  `json:"refresh_expires_in"` // 刷新令牌有效期（秒）
}
