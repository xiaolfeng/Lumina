package user

// UpdatePasswordRequest 修改密码请求
type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" label:"旧密码" binding:"required,min=6"` // 旧密码
	NewPassword string `json:"new_password" label:"新密码" binding:"required,min=8"` // 新密码
}
