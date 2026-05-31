package common

// BaseResponse 通用 API 响应结构（Swagger 文档用）
//
// 对应 bamboo-base-go 的 xBase.BaseResponse，
// 本结构仅用于 Swagger 文档生成，业务代码不应直接使用。
type BaseResponse struct {
	Context      string      `json:"context"`
	Output       string      `json:"output"`
	Code         uint        `json:"code"`
	Message      string      `json:"message"`
	ErrorMessage *string     `json:"error_message,omitempty"`
	Overhead     int64       `json:"overhead,omitempty"`
	Data         interface{} `json:"data,omitempty"`
}
