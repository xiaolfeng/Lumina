package biometric

// AvailabilityResponse 生物特征可用性响应
type AvailabilityResponse struct {
	IsAvailable bool `json:"is_available"` // 是否可用（已注册凭证）
}
