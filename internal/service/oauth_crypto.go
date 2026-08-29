package service

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// OAuth 令牌与 PKCE 的纯函数工具集，供 logic 层编排调用。

// OAuthRandomToken 生成 256 位随机令牌的 base64url 字符串（无填充，43 字符）
func OAuthRandomToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("生成随机令牌失败: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// OAuthRandomHex 生成 n 字节随机数的十六进制字符串（2n 字符）
func OAuthRandomHex(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("生成随机串失败: %w", err)
	}
	return hex.EncodeToString(buf), nil
}

// OAuthHashToken 计算令牌的 SHA-256 十六进制摘要（缓存键不存令牌原文）
func OAuthHashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// OAuthVerifyPKCES256 校验 PKCE code_verifier 与 code_challenge（RFC 7636，仅 S256）
func OAuthVerifyPKCES256(verifier, challenge string) bool {
	if verifier == "" || challenge == "" {
		return false
	}
	sum := sha256.Sum256([]byte(verifier))
	computed := base64.RawURLEncoding.EncodeToString(sum[:])
	// 恒定时间比较，避免时序侧信道
	return subtle.ConstantTimeCompare([]byte(computed), []byte(challenge)) == 1
}

// OAuthValidRedirectURI 校验回调地址 scheme 是否允许（http/https）。
//
// 仅校验 scheme，精确匹配由 logic 层对注册列表完成；本机回环开发场景
// 允许 http（Claude Code 等客户端使用 http://localhost:<port>/callback）。
func OAuthValidRedirectURI(raw string) bool {
	if len(raw) < 12 || len(raw) > 512 {
		return false
	}
	scheme := ""
	for i := 0; i < len(raw); i++ {
		if raw[i] == ':' {
			scheme = raw[:i]
			break
		}
		if raw[i] < 'a' || raw[i] > 'z' {
			return false
		}
	}
	return scheme == "http" || scheme == "https"
}
