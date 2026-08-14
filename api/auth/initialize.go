package auth

// InitializeRequest 初始化请求
type InitializeRequest struct {
	Username   string `json:"username" label:"用户名" binding:"required,min=3,max=32"`     // 用户名
	Email      string `json:"email" label:"邮箱" binding:"required,email"`                 // 邮箱地址
	Password   string `json:"password" label:"密码" binding:"required,min=6,max=64"`       // 登录密码
	SetupToken string `json:"setup_token" label:"初始化令牌" binding:"omitempty,max=128"`     // 一次性初始化令牌（对应 XLF_INITIALIZE_TOKEN，可选）
}
