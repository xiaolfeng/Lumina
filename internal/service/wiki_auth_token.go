package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	xEnv "github.com/bamboo-services/bamboo-base-go/defined/env"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

// WikiAuthTokenService Wiki 授权 HMAC Cookie 签名服务
// 负责为受保护的 Wiki 页面生成短期 HMAC 签名 Cookie，并校验其真伪与有效期。
type WikiAuthTokenService struct {
	hmacSecret string // HMAC 签名密钥
}

// NewWikiAuthTokenService 创建 WikiAuthTokenService 实例
// 从环境变量 REPOWIKI_HMAC_SECRET 读取签名密钥，默认空字符串。
func NewWikiAuthTokenService() *WikiAuthTokenService {
	return &WikiAuthTokenService{
		hmacSecret: xEnv.GetEnvString("REPOWIKI_HMAC_SECRET", ""),
	}
}

// GenerateToken 生成 HMAC 签名的 Cookie 值
// 格式：{wikiID}.{expireTimestamp}.{HMAC}
// 有效期由 bConst.RepoWikiCookieMaxAge 决定（默认 2 小时）。
//
// 参数说明:
//   - wikiID: Wiki 版本 ID
//   - password: Wiki 访问密码（保留参数，用于未来扩展或日志审计，不参与 token 生成）
//
// 返回值:
//   - string: 生成的 HMAC 签名 token
//   - error: 始终返回 nil（保留错误返回值以兼容未来扩展）
func (s *WikiAuthTokenService) GenerateToken(wikiID int64, password string) (string, error) {
	expire := time.Now().Unix() + int64(bConst.RepoWikiCookieMaxAge)

	// HMAC payload: "{wikiID}.{expire}"
	payload := fmt.Sprintf("%d.%d", wikiID, expire)

	// HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(s.hmacSecret))
	mac.Write([]byte(payload))
	hmacHex := hex.EncodeToString(mac.Sum(nil))

	// 最终 token: "{wikiID}.{expire}.{hmac}"
	token := fmt.Sprintf("%s.%s", payload, hmacHex)
	return token, nil
}

// ValidateToken 校验 HMAC Cookie 值
// 检查：格式正确、HMAC 匹配、未过期、wikiID 匹配。
//
// 参数说明:
//   - cookieValue: 客户端提交的 Cookie 值
//   - wikiID: 期望的 Wiki 版本 ID
//
// 返回值:
//   - bool: 校验通过返回 true，否则返回 false
func (s *WikiAuthTokenService) ValidateToken(cookieValue string, wikiID int64) bool {
	parts := strings.Split(cookieValue, ".")
	if len(parts) != 3 {
		return false
	}

	// 解析 wikiID
	tokenWikiID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || tokenWikiID != wikiID {
		return false
	}

	// 解析过期时间
	expire, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return false
	}
	if time.Now().Unix() > expire {
		return false // 已过期
	}

	// 校验 HMAC
	payload := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, []byte(s.hmacSecret))
	mac.Write([]byte(payload))
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(parts[2]), []byte(expectedMAC))
}

// HashPassword 使用 bcrypt 对密码进行哈希
//
// 参数说明:
//   - password: 明文密码
//
// 返回值:
//   - string: 哈希后的密码字符串
//   - error: 哈希失败时返回错误
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// VerifyPassword 校验明文密码与 bcrypt 哈希是否匹配
//
// 参数说明:
//   - password: 明文密码
//   - hash: bcrypt 哈希值
//
// 返回值:
//   - bool: 匹配返回 true，否则返回 false
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
