package pages

// UnlockPageRequest 公开解锁密码门
type UnlockPageRequest struct {
	Password string `json:"password" label:"访问密码" binding:"required"` // 访问密码
}

// PageAuthCheckResponse 密码门状态
type PageAuthCheckResponse struct {
	Authenticated    bool `json:"authenticated"`     // Cookie 是否有效
	PasswordRequired bool `json:"password_required"` // 是否需要密码
}
