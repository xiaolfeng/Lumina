package auth

// RefreshRequest Token 刷新请求
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" label:"刷新令牌" binding:"required"` // 刷新令牌
}
