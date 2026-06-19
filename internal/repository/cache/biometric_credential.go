package cache

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	xCache "github.com/bamboo-services/bamboo-base-go/major/cache"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
)

// BiometricCredentialCache 生物特征凭证多维度缓存管理器
//
// 管理三类缓存：
//   - 凭证缓存（Cache-Aside，ID + CredentialID 双维度，TTL 30min）
//   - 可用性缓存（简单 bool，TTL 30min）
//   - Challenge 临时存储（register/login 共用，TTL 60s，单次使用）
type BiometricCredentialCache struct {
	*xCache.Cache
}

// ── 凭证缓存（Cache-Aside，ID + CredentialID 双维度）──

// GetByID 根据雪花 ID 读取凭证缓存
//
// 返回值:
//   - *entity.BiometricCredential: 缓存命中的凭证实体
//   - bool:                       是否命中
//   - error:                      仅在意外错误时返回（Redis Nil 不视为错误）
func (c *BiometricCredentialCache) GetByID(ctx context.Context, id int64) (*entity.BiometricCredential, bool, error) {
	key := bConst.CacheBiometricCredentialByID.Get(id).String()
	val, err := c.RDB.Get(ctx, key).Result()
	if err != nil || val == "" {
		return nil, false, nil
	}

	var cred entity.BiometricCredential
	if err := json.Unmarshal([]byte(val), &cred); err != nil {
		return nil, false, nil
	}

	return &cred, true, nil
}

// GetByCredentialID 根据 WebAuthn CredentialID（hex 字符串）读取凭证缓存
func (c *BiometricCredentialCache) GetByCredentialID(ctx context.Context, credIDHex string) (*entity.BiometricCredential, bool, error) {
	key := bConst.CacheBiometricCredentialByCredID.Get(credIDHex).String()
	val, err := c.RDB.Get(ctx, key).Result()
	if err != nil || val == "" {
		return nil, false, nil
	}

	var cred entity.BiometricCredential
	if err := json.Unmarshal([]byte(val), &cred); err != nil {
		return nil, false, nil
	}

	return &cred, true, nil
}

// SetCredential 写入凭证缓存（同时写 ID 维度和 CredentialID 维度）
func (c *BiometricCredentialCache) SetCredential(ctx context.Context, cred *entity.BiometricCredential) error {
	if cred == nil {
		return nil
	}

	jsonData, err := json.Marshal(cred)
	if err != nil {
		return fmt.Errorf("凭证缓存序列化失败: %w", err)
	}

	// ID → 详情
	idKey := bConst.CacheBiometricCredentialByID.Get(cred.ID.Int64()).String()
	c.RDB.Set(ctx, idKey, jsonData, c.TTL)

	// CredentialID → 详情
	credIDHex := hex.EncodeToString(cred.CredentialID)
	credIDKey := bConst.CacheBiometricCredentialByCredID.Get(credIDHex).String()
	c.RDB.Set(ctx, credIDKey, jsonData, c.TTL)

	return nil
}

// DeleteCredential 清除凭证缓存（ID 和 CredentialID 两个维度都清）
func (c *BiometricCredentialCache) DeleteCredential(ctx context.Context, cred *entity.BiometricCredential) {
	if cred == nil {
		return
	}

	idKey := bConst.CacheBiometricCredentialByID.Get(cred.ID.Int64()).String()
	credIDHex := hex.EncodeToString(cred.CredentialID)
	credIDKey := bConst.CacheBiometricCredentialByCredID.Get(credIDHex).String()

	c.RDB.Del(ctx, idKey, credIDKey)
}

// ── 可用性缓存（简单 bool，TTL 30min）──

// GetAvailability 读取生物特征登录可用性缓存
//
// 返回值:
//   - bool: 可用性值（未命中时返回 false）
//   - bool: 是否命中缓存
//   - error: 仅在意外错误时返回
func (c *BiometricCredentialCache) GetAvailability(ctx context.Context) (bool, bool, error) {
	key := bConst.CacheBiometricAvailability.Get().String()
	val, err := c.RDB.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, false, nil
	}
	if err != nil {
		return false, false, nil
	}

	available, err := strconv.ParseBool(val)
	if err != nil {
		return false, false, nil
	}

	return available, true, nil
}

// SetAvailability 写入可用性缓存
func (c *BiometricCredentialCache) SetAvailability(ctx context.Context, available bool) error {
	key := bConst.CacheBiometricAvailability.Get().String()
	return c.RDB.Set(ctx, key, strconv.FormatBool(available), c.TTL).Err()
}

// ClearAvailability 清除可用性缓存
func (c *BiometricCredentialCache) ClearAvailability(ctx context.Context) {
	key := bConst.CacheBiometricAvailability.Get().String()
	c.RDB.Del(ctx, key)
}

// ── Challenge 临时存储（TTL 60s，单次使用）──

// challengeTTL Challenge 缓存的固定过期时间（60 秒），不使用主 TTL
const challengeTTL = 60 * time.Second

// challengeKey 根据 challengeType 和 sessionID 生成对应的 Redis key
func (c *BiometricCredentialCache) challengeKey(challengeType, sessionID string) (string, error) {
	switch challengeType {
	case "reg":
		return bConst.CacheBiometricChallengeRegister.Get(sessionID).String(), nil
	case "login":
		return bConst.CacheBiometricChallengeLogin.Get(sessionID).String(), nil
	default:
		return "", fmt.Errorf("unknown challenge type: %s", challengeType)
	}
}

// SetChallenge 写入 challenge（register/login 共用，通过 challengeType 区分）
//
// challengeType 为 "reg" 或 "login"；sessionID 为会话标识；data 为 WebAuthn session data 的 JSON 序列化。
func (c *BiometricCredentialCache) SetChallenge(ctx context.Context, challengeType string, sessionID string, data []byte) error {
	key, err := c.challengeKey(challengeType, sessionID)
	if err != nil {
		return err
	}
	return c.RDB.Set(ctx, key, data, challengeTTL).Err()
}

// GetChallenge 读取 challenge
//
// 返回值:
//   - []byte: challenge 数据
//   - bool:   是否命中
//   - error:  仅在意外错误时返回
func (c *BiometricCredentialCache) GetChallenge(ctx context.Context, challengeType string, sessionID string) ([]byte, bool, error) {
	key, err := c.challengeKey(challengeType, sessionID)
	if err != nil {
		return nil, false, err
	}

	val, err := c.RDB.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, nil
	}

	return []byte(val), true, nil
}

// DeleteChallenge 删除 challenge（验证后立即调用，防重放）
func (c *BiometricCredentialCache) DeleteChallenge(ctx context.Context, challengeType string, sessionID string) {
	key, err := c.challengeKey(challengeType, sessionID)
	if err != nil {
		return
	}
	c.RDB.Del(ctx, key)
}
