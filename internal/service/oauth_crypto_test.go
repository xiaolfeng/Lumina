package service

import (
	"crypto/sha256"
	"encoding/base64"
	"strings"
	"testing"
)

func TestOAuthRandomTokenShape(t *testing.T) {
	token, err := OAuthRandomToken()
	if err != nil {
		t.Fatalf("生成令牌失败: %v", err)
	}
	// 256 位 base64url 无填充 = 43 字符
	if len(token) != 43 || strings.ContainsAny(token, "+/=") {
		t.Fatalf("令牌格式错误: %q", token)
	}
	again, _ := OAuthRandomToken()
	if token == again {
		t.Fatal("两次生成的令牌不应相同")
	}
}

func TestOAuthVerifyPKCES256(t *testing.T) {
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	// RFC 7636 附录 B 的标准测试向量
	wantChallenge := "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"
	if !OAuthVerifyPKCES256(verifier, wantChallenge) {
		t.Fatal("合法 verifier/challenge 应通过校验")
	}
	if OAuthVerifyPKCES256("wrong-verifier", wantChallenge) {
		t.Fatal("错误 verifier 不应通过校验")
	}
	if OAuthVerifyPKCES256("", wantChallenge) || OAuthVerifyPKCES256(verifier, "") {
		t.Fatal("空参数不应通过校验")
	}

	// challenge 必须等于 S256(verifier)
	sum := sha256.Sum256([]byte(verifier))
	if got := base64.RawURLEncoding.EncodeToString(sum[:]); got != wantChallenge {
		t.Fatalf("S256(verifier) = %q, want %q", got, wantChallenge)
	}
}

func TestOAuthValidRedirectURI(t *testing.T) {
	valid := []string{
		"http://localhost:6274/callback",
		"https://client.example.com/cb",
		"http://127.0.0.1:11773/callback",
	}
	for _, uri := range valid {
		if !OAuthValidRedirectURI(uri) {
			t.Fatalf("%q 应为合法回调地址", uri)
		}
	}
	invalid := []string{
		"", "javascript:alert(1)", "ftp://x/cb", "https://", "http s://x", strings.Repeat("a", 600),
	}
	for _, uri := range invalid {
		if OAuthValidRedirectURI(uri) {
			t.Fatalf("%q 应被拒绝", uri)
		}
	}
}

func TestOAuthHashTokenStable(t *testing.T) {
	if OAuthHashToken("abc") != OAuthHashToken("abc") {
		t.Fatal("同一输入的摘要应一致")
	}
	if OAuthHashToken("abc") == OAuthHashToken("abd") {
		t.Fatal("不同输入的摘要不应一致")
	}
	if len(OAuthHashToken("abc")) != 64 {
		t.Fatal("SHA-256 摘要应为 64 位十六进制")
	}
}
