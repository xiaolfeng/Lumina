package bConst

// MCP OAuth 2.1 常量：令牌前缀、作用域、端点路径与 PKCE 方法。
//
// Lumina 同时充当授权服务器与资源服务器，面向 Claude Code / ZCode / Codex
// 等 MCP 客户端提供 authorization_code + PKCE（S256）+ refresh_token 流程。
const (
	// OAuthScope MCP 访问作用域
	OAuthScope = "mcp"
	// OAuthAccessTokenPrefix 访问令牌前缀（用于与 API Key Bearer 区分）
	OAuthAccessTokenPrefix = "lum_at_"
	// OAuthRefreshTokenPrefix 刷新令牌前缀
	OAuthRefreshTokenPrefix = "lum_rt_"
	// OAuthPKCEMethodS256 仅支持的 PKCE 方法
	OAuthPKCEMethodS256 = "S256"

	// OAuthPathAuthorize 授权端点（校验后 302 到前端授权页）
	OAuthPathAuthorize = "/oauth/authorize"
	// OAuthPathToken 令牌端点
	OAuthPathToken = "/oauth/token"
	// OAuthPathRegister 动态客户端注册端点（RFC 7591）
	OAuthPathRegister = "/oauth/register"

	// OAuthWellKnownProtectedResource 受保护资源元数据（RFC 9728）
	OAuthWellKnownProtectedResource = "/.well-known/oauth-protected-resource"
	// OAuthWellKnownAuthorizationServer 授权服务器元数据（RFC 8414）
	OAuthWellKnownAuthorizationServer = "/.well-known/oauth-authorization-server"
)
