package oauth

// ProtectedResourceMetadata 受保护资源元数据（RFC 9728）。
//
// 作为公开裸 JSON 返回，MCP 客户端凭 401 WWW-Authenticate 中的
// resource_metadata 地址直接拉取本结构。
type ProtectedResourceMetadata struct {
	Resource               string   `json:"resource"`                         // 资源标识（MCP 端点完整 URL）
	AuthorizationServers   []string `json:"authorization_servers"`            // 授权服务器 issuer 列表
	ScopesSupported        []string `json:"scopes_supported"`                 // 支持的作用域
	BearerMethodsSupported []string `json:"bearer_methods_supported"`         // Bearer 传递方式
	ResourceDocumentation  string   `json:"resource_documentation,omitempty"` // 资源文档地址
}

// AuthorizationServerMetadata 授权服务器元数据（RFC 8414）。
type AuthorizationServerMetadata struct {
	Issuer                        string   `json:"issuer"`                                // 签发者标识（站点根地址）
	AuthorizationEndpoint         string   `json:"authorization_endpoint"`                // 授权端点
	TokenEndpoint                 string   `json:"token_endpoint"`                        // 令牌端点
	RegistrationEndpoint          string   `json:"registration_endpoint"`                 // 动态注册端点
	ResponseTypesSupported        []string `json:"response_types_supported"`              // 支持的响应类型
	GrantTypesSupported           []string `json:"grant_types_supported"`                 // 支持的授权方式
	CodeChallengeMethodsSupported []string `json:"code_challenge_methods_supported"`      // 支持的 PKCE 方法
	TokenEndpointAuthMethods      []string `json:"token_endpoint_auth_methods_supported"` // 令牌端点认证方式
	ScopesSupported               []string `json:"scopes_supported"`                      // 支持的作用域
}
