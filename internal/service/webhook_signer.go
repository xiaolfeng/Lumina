package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
)

// DetectProvider 通过请求头识别 Git webhook 提供商并返回对应的签名字段值。
//
// 识别顺序（与常见冲突风险从低到高一致）：
//  1. X-Gitee-Token   → gitee
//  2. X-Gitlab-Token  → gitlab
//  3. X-Hub-Signature-256 → github
//  4. X-Gitea-Signature → gitea
//
// 若未匹配到已知头部，返回 ("", "")。
func DetectProvider(headers http.Header) (string, string) {
	if token := headers.Get("X-Gitee-Token"); token != "" {
		return webhookProviderGitee, token
	}
	if token := headers.Get("X-Gitlab-Token"); token != "" {
		return webhookProviderGitLab, token
	}
	if sig := headers.Get("X-Hub-Signature-256"); sig != "" {
		return webhookProviderGitHub, sig
	}
	if sig := headers.Get("X-Gitea-Signature"); sig != "" {
		return webhookProviderGitea, sig
	}
	return "", ""
}

// VerifyWebhookSignature 校验 Git webhook 签名。
//
// 参数说明：
//   - provider:    提供商标识（github / gitee / gitlab / gitea）
//   - body:        webhook 原始请求体
//   - headerValue: DetectProvider 返回的头部值（签名或 token）
//   - secret:      配置的 webhook secret / token
//
// 返回值：
//   - true:  校验通过
//   - false: 校验失败、参数缺失或 provider 未知
func VerifyWebhookSignature(provider string, body []byte, headerValue string, secret string) bool {
	if headerValue == "" || secret == "" {
		return false
	}

	switch provider {
	case webhookProviderGitHub, webhookProviderGitea:
		return verifyHMACSHA256(body, headerValue, secret)
	case webhookProviderGitee, webhookProviderGitLab:
		return verifyStaticToken(headerValue, secret)
	default:
		return false
	}
}

// verifyHMACSHA256 使用 HMAC-SHA256 校验签名。
//
// 支持两种格式：
//   - GitHub: "sha256=" + hex(hmac-sha256(secret, body))
//   - Gitea:  hex(hmac-sha256(secret, body))（无前缀）
//
// 使用 hmac.Equal 进行常量时间比较，避免 timing attack。
func verifyHMACSHA256(body []byte, signature string, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))

	signature = strings.TrimPrefix(signature, "sha256=")

	return hmac.Equal([]byte(expected), []byte(signature))
}

// verifyStaticToken 对 Gitee / GitLab 的静态 token 进行常量时间比较。
// 直接使用 hmac.Equal 比较字节，防止 timing attack。
func verifyStaticToken(headerValue string, secret string) bool {
	return hmac.Equal([]byte(headerValue), []byte(secret))
}
