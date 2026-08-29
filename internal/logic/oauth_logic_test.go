package logic

import (
	"encoding/json"
	"testing"

	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

const oauthTestBase = "https://lumina.example"

func TestOAuthAuthorizationServerMetadataShape(t *testing.T) {
	meta := OAuthAuthorizationServerMetadata(oauthTestBase + "/")
	if meta.Issuer != oauthTestBase {
		t.Fatalf("issuer = %q, want %q（应去掉尾斜杠）", meta.Issuer, oauthTestBase)
	}
	if meta.AuthorizationEndpoint != oauthTestBase+bConst.OAuthPathAuthorize {
		t.Fatalf("authorization_endpoint = %q", meta.AuthorizationEndpoint)
	}
	if meta.TokenEndpoint != oauthTestBase+bConst.OAuthPathToken {
		t.Fatalf("token_endpoint = %q", meta.TokenEndpoint)
	}
	if meta.RegistrationEndpoint != oauthTestBase+bConst.OAuthPathRegister {
		t.Fatalf("registration_endpoint = %q", meta.RegistrationEndpoint)
	}
	if len(meta.CodeChallengeMethodsSupported) != 1 || meta.CodeChallengeMethodsSupported[0] != bConst.OAuthPKCEMethodS256 {
		t.Fatalf("code_challenge_methods = %v", meta.CodeChallengeMethodsSupported)
	}
	data, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}
	for _, key := range []string{"issuer", "authorization_endpoint", "token_endpoint", "registration_endpoint"} {
		if _, ok := raw[key]; !ok {
			t.Fatalf("元数据缺少字段 %s", key)
		}
	}
}

func TestOAuthProtectedResourceMetadataShape(t *testing.T) {
	meta := OAuthProtectedResourceMetadata(oauthTestBase)
	if meta.Resource != oauthTestBase+bConst.AIPluginMCPPath {
		t.Fatalf("resource = %q, want MCP 端点", meta.Resource)
	}
	if len(meta.AuthorizationServers) != 1 || meta.AuthorizationServers[0] != oauthTestBase {
		t.Fatalf("authorization_servers = %v", meta.AuthorizationServers)
	}
}

func TestOAuthResourceMatches(t *testing.T) {
	cases := []struct {
		expected, actual string
		want             bool
	}{
		{"https://a.example/api/v1/mcp", "https://a.example/api/v1/mcp", true},
		{"https://A.example/api/v1/mcp", "https://a.example/api/v1/mcp", true},
		{"https://a.example/api/v1/mcp", "https://a.example:443/api/v1/mcp", true},
		{"https://a.example/api/v1/mcp", "http://a.example/api/v1/mcp", false},
		{"https://a.example/api/v1/mcp", "https://b.example/api/v1/mcp", false},
		{"https://a.example/api/v1/mcp", "https://a.example/api/v1/other", false},
		{"not a url", "not a url", true},
		{"", "https://a.example", false},
	}
	for _, c := range cases {
		if got := oauthResourceMatches(c.expected, c.actual); got != c.want {
			t.Fatalf("oauthResourceMatches(%q, %q) = %v, want %v", c.expected, c.actual, got, c.want)
		}
	}
}

func TestOAuthRedirectWith(t *testing.T) {
	if got := oauthRedirectWith("http://localhost:1/cb", "code-1", "state-1", ""); got != "http://localhost:1/cb?code=code-1&state=state-1" {
		t.Fatalf("redirect = %q", got)
	}
	if got := oauthRedirectWith("http://localhost:1/cb", "", "", "access_denied"); got != "http://localhost:1/cb?error=access_denied" {
		t.Fatalf("redirect = %q", got)
	}
	if got := oauthRedirectWith("http://localhost:1/cb?x=1", "c", "", ""); got != "http://localhost:1/cb?x=1&code=c" {
		t.Fatalf("redirect = %q", got)
	}
}

func TestOAuthRedirectURIMarshalRoundtrip(t *testing.T) {
	uris := []string{"http://localhost:6274/callback", "https://c.example/cb"}
	raw := marshalOAuthRedirectURIs(uris)
	if got := oauthUnmarshalRedirectURIs(raw); len(got) != 2 || got[0] != uris[0] || got[1] != uris[1] {
		t.Fatalf("roundtrip = %v, want %v", got, uris)
	}
}
