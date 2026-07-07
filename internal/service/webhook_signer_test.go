package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"testing"
)

func computeGitHubSignature(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func computeGiteaSignature(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func TestDetectProvider(t *testing.T) {
	tests := []struct {
		name                string
		headers             http.Header
		expectedProvider    string
		expectedHeaderValue string
	}{
		{
			name:                "GitHub header",
			headers:             http.Header{"X-Hub-Signature-256": []string{"sha256=abc123"}},
			expectedProvider:    webhookProviderGitHub,
			expectedHeaderValue: "sha256=abc123",
		},
		{
			name:                "Gitee header",
			headers:             http.Header{"X-Gitee-Token": []string{"gitee-token-123"}},
			expectedProvider:    webhookProviderGitee,
			expectedHeaderValue: "gitee-token-123",
		},
		{
			name:                "GitLab header",
			headers:             http.Header{"X-Gitlab-Token": []string{"gitlab-token-123"}},
			expectedProvider:    webhookProviderGitLab,
			expectedHeaderValue: "gitlab-token-123",
		},
		{
			name:                "Gitea header",
			headers:             http.Header{"X-Gitea-Signature": []string{"gitea-sig-123"}},
			expectedProvider:    webhookProviderGitea,
			expectedHeaderValue: "gitea-sig-123",
		},
		{
			name:                "No known headers",
			headers:             http.Header{"X-Unknown": []string{"something"}},
			expectedProvider:    "",
			expectedHeaderValue: "",
		},
		{
			name:                "Empty headers",
			headers:             http.Header{},
			expectedProvider:    "",
			expectedHeaderValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, headerValue := DetectProvider(tt.headers)
			if provider != tt.expectedProvider {
				t.Fatalf("expected provider %q, got %q", tt.expectedProvider, provider)
			}
			if headerValue != tt.expectedHeaderValue {
				t.Fatalf("expected header value %q, got %q", tt.expectedHeaderValue, headerValue)
			}
		})
	}
}

func TestVerifyWebhookSignature(t *testing.T) {
	body := []byte(`{"ref":"refs/heads/main"}`)
	secret := "webhook-secret"

	tests := []struct {
		name        string
		provider    string
		body        []byte
		headerValue string
		secret      string
		expected    bool
	}{
		{
			name:        "GitHub valid signature",
			provider:    webhookProviderGitHub,
			body:        body,
			headerValue: computeGitHubSignature(body, secret),
			secret:      secret,
			expected:    true,
		},
		{
			name:        "Gitea valid signature",
			provider:    webhookProviderGitea,
			body:        body,
			headerValue: computeGiteaSignature(body, secret),
			secret:      secret,
			expected:    true,
		},
		{
			name:        "Gitee valid static token",
			provider:    webhookProviderGitee,
			body:        body,
			headerValue: secret,
			secret:      secret,
			expected:    true,
		},
		{
			name:        "GitLab valid static token",
			provider:    webhookProviderGitLab,
			body:        body,
			headerValue: secret,
			secret:      secret,
			expected:    true,
		},
		{
			name:        "GitHub wrong secret",
			provider:    webhookProviderGitHub,
			body:        body,
			headerValue: computeGitHubSignature(body, "wrong-secret"),
			secret:      secret,
			expected:    false,
		},
		{
			name:        "Gitea wrong secret",
			provider:    webhookProviderGitea,
			body:        body,
			headerValue: computeGiteaSignature(body, "wrong-secret"),
			secret:      secret,
			expected:    false,
		},
		{
			name:        "Gitee wrong token",
			provider:    webhookProviderGitee,
			body:        body,
			headerValue: "wrong-token",
			secret:      secret,
			expected:    false,
		},
		{
			name:        "GitLab wrong token",
			provider:    webhookProviderGitLab,
			body:        body,
			headerValue: "wrong-token",
			secret:      secret,
			expected:    false,
		},
		{
			name:        "Empty header value",
			provider:    webhookProviderGitHub,
			body:        body,
			headerValue: "",
			secret:      secret,
			expected:    false,
		},
		{
			name:        "Empty secret",
			provider:    webhookProviderGitHub,
			body:        body,
			headerValue: computeGitHubSignature(body, secret),
			secret:      "",
			expected:    false,
		},
		{
			name:        "Unknown provider",
			provider:    "bitbucket",
			body:        body,
			headerValue: "sig",
			secret:      secret,
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VerifyWebhookSignature(tt.provider, tt.body, tt.headerValue, tt.secret)
			if result != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
