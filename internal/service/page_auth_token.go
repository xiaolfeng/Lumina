package service

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	xEnv "github.com/bamboo-services/bamboo-base-go/defined/env"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

// PageAuthTokenService Pages 密码门 HMAC Cookie 签名服务
type PageAuthTokenService struct {
	hmacSecret string // HMAC 签名密钥
}

var (
	pageAuthSecretOnce sync.Once
	pageAuthSecret     string
)

func ensurePageAuthSecret() string {
	pageAuthSecretOnce.Do(func() {
		secret := xEnv.GetEnvString("PAGES_HMAC_SECRET", "")
		if secret == "" {
			secret = xEnv.GetEnvString("REPOWIKI_HMAC_SECRET", "")
		}
		if secret == "" {
			buf := make([]byte, 32)
			if _, err := rand.Read(buf); err == nil {
				secret = hex.EncodeToString(buf)
			} else {
				secret = fmt.Sprintf("lumina-pages-%d", time.Now().UnixNano())
			}
		}
		pageAuthSecret = secret
	})
	return pageAuthSecret
}

// NewPageAuthTokenService 创建 PageAuthTokenService 实例
func NewPageAuthTokenService() *PageAuthTokenService {
	return &PageAuthTokenService{hmacSecret: ensurePageAuthSecret()}
}

// GenerateToken 生成 HMAC 签名 Cookie：{pageID}.{expire}.{hmac}
func (s *PageAuthTokenService) GenerateToken(pageID int64, maxAge int) (string, error) {
	if maxAge <= 0 {
		maxAge = bConst.PagesCookieMaxAge
	}
	expire := time.Now().Unix() + int64(maxAge)
	payload := fmt.Sprintf("%d.%d", pageID, expire)
	mac := hmac.New(sha256.New, []byte(s.hmacSecret))
	mac.Write([]byte(payload))
	return fmt.Sprintf("%s.%s", payload, hex.EncodeToString(mac.Sum(nil))), nil
}

// ValidateToken 校验 HMAC Cookie：格式、过期、pageID、签名
func (s *PageAuthTokenService) ValidateToken(cookieValue string, pageID int64) bool {
	parts := strings.Split(cookieValue, ".")
	if len(parts) != 3 {
		return false
	}
	tokenPageID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || tokenPageID != pageID {
		return false
	}
	expire, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || time.Now().Unix() > expire {
		return false
	}
	payload := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, []byte(s.hmacSecret))
	mac.Write([]byte(payload))
	expectedMAC := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(parts[2]), []byte(expectedMAC))
}

// CookieName 返回指定页面的密码门 Cookie 名
func CookieName(pageID int64) string {
	return fmt.Sprintf("%s%d", bConst.PagesCookieNamePrefix, pageID)
}
