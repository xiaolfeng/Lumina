package user

// UpdateProfileRequest 更新个人资料请求
type UpdateProfileRequest struct {
	Username string `json:"username" label:"用户名" binding:"required,min=3,max=32"` // 用户名
	Email    string `json:"email" label:"邮箱" binding:"required,email"`            // 邮箱
}
