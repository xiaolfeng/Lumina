package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

// TestWikiAuthToken 测试 Wiki HMAC Cookie token 的生成与校验
func TestWikiAuthToken(t *testing.T) {
	service := NewWikiAuthTokenService()
	wikiID := int64(12345)
	password := "my-secret-password"

	t.Run("valid token round trip", func(t *testing.T) {
		token, err := service.GenerateToken(wikiID, password)
		if err != nil {
			t.Fatalf("GenerateToken failed: %v", err)
		}
		if token == "" {
			t.Fatal("expected non-empty token")
		}
		if !service.ValidateToken(token, wikiID) {
			t.Fatal("expected valid token to pass validation")
		}
	})

	t.Run("expired token", func(t *testing.T) {
		// 构造一个已过期的 token：过期时间戳为 1 小时前
		expire := time.Now().Unix() - 3600
		payload := fmt.Sprintf("%d.%d", wikiID, expire)
		mac := hmac.New(sha256.New, []byte(service.hmacSecret))
		mac.Write([]byte(payload))
		hmacHex := hex.EncodeToString(mac.Sum(nil))
		expiredToken := fmt.Sprintf("%s.%s", payload, hmacHex)

		if service.ValidateToken(expiredToken, wikiID) {
			t.Fatal("expected expired token to fail validation")
		}
	})

	t.Run("tampered hmac", func(t *testing.T) {
		token, err := service.GenerateToken(wikiID, password)
		if err != nil {
			t.Fatalf("GenerateToken failed: %v", err)
		}
		parts := strings.Split(token, ".")
		if len(parts) != 3 {
			t.Fatalf("expected token format wikiID.expire.hmac, got %s", token)
		}
		// 篡改 HMAC 部分
		parts[2] = strings.Repeat("0", len(parts[2]))
		tamperedToken := strings.Join(parts, ".")

		if service.ValidateToken(tamperedToken, wikiID) {
			t.Fatal("expected tampered HMAC token to fail validation")
		}
	})

	t.Run("mismatched wikiID", func(t *testing.T) {
		token, err := service.GenerateToken(wikiID, password)
		if err != nil {
			t.Fatalf("GenerateToken failed: %v", err)
		}
		if service.ValidateToken(token, wikiID+1) {
			t.Fatal("expected token with mismatched wikiID to fail validation")
		}
	})

	t.Run("malformed token", func(t *testing.T) {
		malformedTokens := []string{
			"",
			"12345",
			"12345.expire",
			"12345.expire.hmac.extra",
			"not-a-number.expire.hmac",
		}
		for _, token := range malformedTokens {
			if service.ValidateToken(token, wikiID) {
				t.Fatalf("expected malformed token %q to fail validation", token)
			}
		}
	})

	t.Run("token contains expected max age", func(t *testing.T) {
		before := time.Now().Unix()
		token, err := service.GenerateToken(wikiID, password)
		if err != nil {
			t.Fatalf("GenerateToken failed: %v", err)
		}
		after := time.Now().Unix()

		parts := strings.Split(token, ".")
		if len(parts) != 3 {
			t.Fatalf("expected token format wikiID.expire.hmac, got %s", token)
		}
		expire, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			t.Fatalf("expected numeric expire timestamp, got %s", parts[1])
		}
		expectedMin := before + int64(bConst.RepoWikiCookieMaxAge) - 1
		expectedMax := after + int64(bConst.RepoWikiCookieMaxAge) + 1
		if expire < expectedMin || expire > expectedMax {
			t.Fatalf("expire timestamp %d out of expected range [%d, %d]", expire, expectedMin, expectedMax)
		}
	})
}

// TestWikiPasswordHash 测试 bcrypt 密码哈希与校验
func TestWikiPasswordHash(t *testing.T) {
	t.Run("correct password verifies", func(t *testing.T) {
		password := "wiki-secure-pass"
		hash, err := HashPassword(password)
		if err != nil {
			t.Fatalf("HashPassword failed: %v", err)
		}
		if hash == "" || hash == password {
			t.Fatal("expected non-empty hash different from password")
		}
		if !VerifyPassword(password, hash) {
			t.Fatal("expected correct password to verify")
		}
	})

	t.Run("wrong password fails", func(t *testing.T) {
		password := "wiki-secure-pass"
		hash, err := HashPassword(password)
		if err != nil {
			t.Fatalf("HashPassword failed: %v", err)
		}
		if VerifyPassword("wrong-password", hash) {
			t.Fatal("expected wrong password to fail verification")
		}
	})
}
