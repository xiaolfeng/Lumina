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

func TestPageAuthToken(t *testing.T) {
	svc := NewPageAuthTokenService()
	pageID := int64(98765)

	t.Run("valid token round trip", func(t *testing.T) {
		token, err := svc.GenerateToken(pageID, bConst.PagesCookieMaxAge)
		if err != nil {
			t.Fatalf("GenerateToken failed: %v", err)
		}
		if !svc.ValidateToken(token, pageID) {
			t.Fatal("expected valid token to pass")
		}
	})

	t.Run("expired token", func(t *testing.T) {
		expire := time.Now().Unix() - 3600
		payload := fmt.Sprintf("%d.%d", pageID, expire)
		mac := hmac.New(sha256.New, []byte(svc.hmacSecret))
		mac.Write([]byte(payload))
		expired := fmt.Sprintf("%s.%s", payload, hex.EncodeToString(mac.Sum(nil)))
		if svc.ValidateToken(expired, pageID) {
			t.Fatal("expected expired token to fail")
		}
	})

	t.Run("mismatched page id", func(t *testing.T) {
		token, err := svc.GenerateToken(pageID, bConst.PagesCookieMaxAge)
		if err != nil {
			t.Fatalf("GenerateToken failed: %v", err)
		}
		if svc.ValidateToken(token, pageID+1) {
			t.Fatal("expected mismatched page id to fail")
		}
	})

	t.Run("cookie name is page scoped", func(t *testing.T) {
		if got := CookieName(pageID); got != fmt.Sprintf("%s%d", bConst.PagesCookieNamePrefix, pageID) {
			t.Fatalf("CookieName() = %q", got)
		}
	})

	t.Run("default max age", func(t *testing.T) {
		before := time.Now().Unix()
		token, err := svc.GenerateToken(pageID, 0)
		if err != nil {
			t.Fatalf("GenerateToken failed: %v", err)
		}
		after := time.Now().Unix()
		parts := strings.Split(token, ".")
		expire, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			t.Fatalf("expire parse: %v", err)
		}
		min := before + int64(bConst.PagesCookieMaxAge) - 1
		max := after + int64(bConst.PagesCookieMaxAge) + 1
		if expire < min || expire > max {
			t.Fatalf("expire %d out of [%d, %d]", expire, min, max)
		}
	})
}
