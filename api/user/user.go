package user

// UserInfoResponse 用户信息响应
type UserInfoResponse struct {
	Username string `json:"username"` // 用户名
	Email    string `json:"email"`    // 邮箱地址
}
