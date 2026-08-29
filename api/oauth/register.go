package oauth

// RegisterRequest 动态客户端注册请求（RFC 7591）。
//
// MCP 客户端（Claude Code / ZCode / Codex 等）在首次连接时自动提交；
// Lumina 单用户场景下注册为公共客户端，token_endpoint_auth_method 恒为 none。
type RegisterRequest struct {
	ClientName              string   `json:"client_name"`                            // 客户端名称
	RedirectURIs            []string `json:"redirect_uris" binding:"required,min=1"` // 回调地址列表（精确匹配）
	GrantTypes              []string `json:"grant_types"`                            // 请求的授权方式（忽略，恒为 authorization_code）
	ResponseTypes           []string `json:"response_types"`                         // 请求的响应类型（忽略，恒为 code）
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`             // 令牌端点认证方式（忽略，恒为 none）
	Scope                   string   `json:"scope"`                                  // 请求作用域（忽略，恒为 mcp）
}

// RegisterResponse 动态客户端注册响应。
type RegisterResponse struct {
	ClientID                string   `json:"client_id"`                  // 客户端 ID（雪花 ID 十进制字符串）
	ClientIDIssuedAt        int64    `json:"client_id_issued_at"`        // 签发时间（Unix 秒）
	ClientName              string   `json:"client_name,omitempty"`      // 客户端名称
	RedirectURIs            []string `json:"redirect_uris"`              // 回调地址列表
	GrantTypes              []string `json:"grant_types"`                // 授权方式
	ResponseTypes           []string `json:"response_types"`             // 响应类型
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"` // 令牌端点认证方式
	Scope                   string   `json:"scope"`                      // 授予作用域
}
