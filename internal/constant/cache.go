package bConst

import (
	"fmt"

	xEnv "github.com/bamboo-services/bamboo-base-go/defined/env"
)

// RedisKey Redis 键类型
type RedisKey string

const (
	// ── 认证缓存 ──
	CacheAuthToken    RedisKey = "auth:at:%s" // CacheAuthToken AccessToken 认证标记缓存（%s = MD5(AT)）
	CacheRefreshToken RedisKey = "auth:rt:%s" // CacheRefreshToken RefreshToken→UserID 缓存（%s = RT）

	// ── 项目缓存（Cache-Aside 三/四层映射，TTL 30 分钟）──
	CacheProjectByID          RedisKey = "project:id:%d"         // CacheProjectByID 项目 ID→详情缓存（%d = snowflake ID）
	CacheProjectIDByName      RedisKey = "project:name:%s"       // CacheProjectIDByName 项目名称→ID 映射（%s = name）
	CacheProjectIDByAlias     RedisKey = "project:alias:%s"      // CacheProjectIDByAlias 别名→ID 映射（%s = alias）
	CacheProjectIDByMatchPath RedisKey = "project:match_path:%s" // CacheProjectIDByMatchPath 路径→ID 映射（%s = match path）

	// ── QA Session 缓存（Cache-Aside ID→详情 + Hash→ID，TTL 10 分钟）──
	CacheQaSessionByID     RedisKey = "qa:session:%d"      // CacheQaSessionByID 会话 ID→详情缓存（%d = snowflake ID）
	CacheQaSessionIDByHash RedisKey = "qa:session:hash:%s" // CacheQaSessionIDByHash Hash→ID 映射（%s = 16位hash）

	// ── QA 运行时缓存 ──
	CacheQaGetAnswerRetry RedisKey = "qa:get_answer:retry:%s" // CacheQaGetAnswerRetry qa_get_answer 重试计数器（%s = sessionID）
	CacheQaDownloadToken  RedisKey = "qa:download:token:%s"   // CacheQaDownloadToken QA 一次性下载令牌缓存（%s = token）

	// ── 生物特征凭证缓存（TTL 30 分钟）──
	CacheBiometricCredentialByID     RedisKey = "biometric:cred:id:%d"     // CacheBiometricCredentialByID 凭证 ID→详情缓存（%d = snowflake ID）
	CacheBiometricCredentialByCredID RedisKey = "biometric:cred:credid:%s" // CacheBiometricCredentialByCredID WebAuthn CredentialID→详情缓存（%s = credentialID hex）
	CacheBiometricAvailability       RedisKey = "biometric:availability"   // CacheBiometricAvailability 生物特征登录可用性缓存

	// ── 生物特征 Challenge 缓存（TTL 60 秒）──
	CacheBiometricChallengeRegister RedisKey = "biometric:challenge:reg:%s"   // CacheBiometricChallengeRegister 注册 Challenge 缓存（%s = sessionID）
	CacheBiometricChallengeLogin    RedisKey = "biometric:challenge:login:%s" // CacheBiometricChallengeLogin 登录 Challenge 缓存（%s = sessionID）

	// ── SSH 密钥缓存（Cache-Aside ID→详情，TTL 30 分钟）──
	CacheSSHKeyByID RedisKey = "ssh:key:%d" // CacheSSHKeyByID SSH 密钥 ID→详情缓存（%d = snowflake ID）
)

// Get 格式化 Redis 键，自动拼接环境前缀
func (k RedisKey) Get(args ...interface{}) RedisKey {
	validKey := xEnv.GetEnvString(xEnv.NoSqlPrefix, "lum:") + string(k)
	return RedisKey(fmt.Sprintf(validKey, args...))
}

// String 将 RedisKey 转换为字符串
func (k RedisKey) String() string {
	return string(k)
}
