package auth

// InitializeRequest 初始化请求
type InitializeRequest struct {
	Username string `json:"username" label:"用户名" binding:"required,min=3,max=32"` // 用户名
	Email    string `json:"email" label:"邮箱" binding:"required,email"`             // 邮箱地址
	Password string `json:"password" label:"密码" binding:"required,min=6,max=64"`   // 登录密码
}
