package auth

// InitializeRequest 初始化请求
type InitializeRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"` // 用户名
	Email    string `json:"email" binding:"required,email"`           // 邮箱地址
	Password string `json:"password" binding:"required,min=6,max=64"` // 登录密码
}

// LoginRequest 登录请求
type LoginRequest struct {
	Account  string `json:"account" binding:"required"`  // 账号（用户名或邮箱）
	Password string `json:"password" binding:"required"` // 登录密码
}

// RefreshRequest Token 刷新请求
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"` // 刷新令牌
}

// TokenResponse Token 响应
type TokenResponse struct {
	AccessToken  string `json:"access_token"`  // 访问令牌
	RefreshToken string `json:"refresh_token"` // 刷新令牌
	ExpiresIn    int64  `json:"expires_in"`    // 过期时间（秒）
}

// StatusResponse 系统状态响应
type StatusResponse struct {
	IsInitial bool `json:"is_initial"` // 系统是否为初始状态（true=未初始化）
}
