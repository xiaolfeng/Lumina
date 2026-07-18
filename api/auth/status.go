package auth

// StatusResponse 系统状态响应
type StatusResponse struct {
	IsInitial bool `json:"is_initial"` // 系统是否为初始状态（true=未初始化）
}
