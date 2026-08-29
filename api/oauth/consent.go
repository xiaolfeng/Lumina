package oauth

// ConsentDetail 授权页展示信息（GET /api/v1/oauth/consent 响应 data）。
type ConsentDetail struct {
	ClientName string `json:"client_name"` // 请求授权的客户端名称
	Scope      string `json:"scope"`       // 请求的作用域
}

// ConsentRequest 授权裁决请求（POST /api/v1/oauth/consent）。
type ConsentRequest struct {
	AuthorizeID string `json:"authorize_id" label:"授权请求ID" binding:"required"` // 授权请求 ID
	Approve     bool   `json:"approve" label:"是否同意"`                           // 是否同意授权
}

// ConsentRedirect 授权裁决结果（POST /api/v1/oauth/consent 响应 data）。
//
// redirect 为客户端回调地址（携带 code/state 或 error），前端直接跳转。
type ConsentRedirect struct {
	Redirect string `json:"redirect"` // 客户端回调地址
}
