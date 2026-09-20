package pages

// UpdateAccessPolicyRequest 管理员更新页面访问策略（仅控制台）
type UpdateAccessPolicyRequest struct {
	AccessMode string `json:"access_mode" label:"访问策略" binding:"required"` // public / password
	Password   string `json:"password" label:"访问密码"`                       // password 模式必填；public 可空
}
