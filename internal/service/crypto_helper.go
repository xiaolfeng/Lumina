package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"io"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
)

// ── SSH 私钥 AES-GCM 加解密 ──
//
// RepoWiki 模块需要持久化用户提供的 SSH 私钥用于私有仓库克隆。
// 私钥属于高敏感数据，必须加密存储，禁止明文落库。
//
// 加密方案：AES-256-GCM（认证加密，同时保证机密性与完整性）
//   - 密钥派生：SHA-256(secret) → 32 字节 AES-256 密钥
//   - Nonce：每次加密随机生成 12 字节，与密文拼接后 base64 编码
//   - 输出格式：base64( nonce[12] || ciphertext+tag )
//
// 密钥来源：REPOWIKI_HMAC_SECRET 环境变量（由调用方通过 xEnv 读取后传入）

// gcmNonceSize AES-GCM 标准 Nonce 长度（12 字节）
const gcmNonceSize = 12

// deriveAESKey 从任意长度密钥字符串派生 32 字节 AES-256 密钥
//
// 使用 SHA-256 单向哈希，输出固定 32 字节，直接作为 AES-256 密钥。
// 注意：此方案为项目内部使用，未采用 KDF（如 PBKDF2/Argon2）是因为
// 密钥来源为环境变量中的高熵随机串，而非低熵用户密码。
func deriveAESKey(secret string) []byte {
	h := sha256.Sum256([]byte(secret))
	return h[:] // 32 bytes
}

// EncryptSSHKey 使用 AES-GCM 加密 SSH 私钥
//
// 加密流程：
//  1. SHA-256(secret) → 32 字节 AES 密钥
//  2. AES-GCM NewCipher → GCM seal
//  3. 随机 12 字节 Nonce + 密文+Tag → base64 编码
//
// 参数说明:
//   - key:    明文 SSH 私钥（PEM 格式）
//   - secret: 加密密钥种子（通常来自 REPOWIKI_HMAC_SECRET 环境变量）
//
// 返回值:
//   - string: base64 编码的加密私钥（nonce + ciphertext）
//   - *xError.Error: 密钥为空或加密操作失败时返回
func EncryptSSHKey(key string, secret string) (string, *xError.Error) {
	ctx := context.Background()

	if secret == "" {
		return "", xError.NewError(ctx, xError.ValidationError, "SSH 密钥加密失败：加密密钥不能为空", false)
	}

	aesKey := deriveAESKey(secret)

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", xError.NewError(ctx, xError.ServerInternalError, "创建 AES cipher 失败", false, err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", xError.NewError(ctx, xError.ServerInternalError, "创建 GCM 失败", false, err)
	}

	nonce := make([]byte, gcmNonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", xError.NewError(ctx, xError.ServerInternalError, "生成 GCM nonce 失败", false, err)
	}

	// Seal: nonce || ciphertext+tag
	sealed := gcm.Seal(nonce, nonce, []byte(key), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// DecryptSSHKey 使用 AES-GCM 解密 SSH 私钥
//
// 解密流程：
//  1. base64 解码 → nonce[12] + ciphertext+tag
//  2. SHA-256(secret) → 32 字节 AES 密钥
//  3. GCM Open 验证 Tag 并还原明文
//
// 参数说明:
//   - encrypted: EncryptSSHKey 返回的 base64 编码字符串
//   - secret:    加密时使用的同一密钥种子
//
// 返回值:
//   - string: 明文 SSH 私钥（PEM 格式）
//   - *xError.Error: 密钥为空、base64 解码失败、GCM 认证失败时返回
func DecryptSSHKey(encrypted string, secret string) (string, *xError.Error) {
	ctx := context.Background()

	if secret == "" {
		return "", xError.NewError(ctx, xError.ValidationError, "SSH 密钥解密失败：加密密钥不能为空", false)
	}

	raw, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", xError.NewError(ctx, xError.DeserializeErr, "SSH 密钥 base64 解码失败", false, err)
	}

	if len(raw) < gcmNonceSize {
		return "", xError.NewError(ctx, xError.ValidationError, "SSH 密钥解密失败：密文数据过短", false)
	}

	aesKey := deriveAESKey(secret)

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", xError.NewError(ctx, xError.ServerInternalError, "创建 AES cipher 失败", false, err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", xError.NewError(ctx, xError.ServerInternalError, "创建 GCM 失败", false, err)
	}

	nonce := raw[:gcmNonceSize]
	ciphertext := raw[gcmNonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", xError.NewError(ctx, xError.ValidationError, "SSH 密钥 GCM 认证失败（密钥不匹配或数据已被篡改）", false, err)
	}

	return string(plaintext), nil
}

// ── LLM API Key AES-GCM 加解密 ──
//
// LLM 模块需要持久化用户提供的 Provider API Key 用于调用大模型服务。
// API Key 属于高敏感数据，必须加密存储，禁止明文落库。
//
// 加密方案与 SSH 私钥完全一致（AES-256-GCM + SHA-256 派生 + base64），
// 仅函数名和密钥来源不同。
// 密钥来源：LLM_ENCRYPT_SECRET 环境变量（由调用方通过 xEnv 读取后传入）

// EncryptAPIKey 使用 AES-GCM 加密 LLM API Key
//
// 加密流程：
//  1. SHA-256(secret) → 32 字节 AES 密钥
//  2. AES-GCM NewCipher → GCM seal
//  3. 随机 12 字节 Nonce + 密文+Tag → base64 编码
//
// 参数说明:
//   - key:    明文 API Key
//   - secret: 加密密钥种子（通常来自 LLM_ENCRYPT_SECRET 环境变量）
//
// 返回值:
//   - string: base64 编码的加密 API Key（nonce + ciphertext）
//   - *xError.Error: 密钥为空或加密操作失败时返回
func EncryptAPIKey(key string, secret string) (string, *xError.Error) {
	ctx := context.Background()

	if secret == "" {
		return "", xError.NewError(ctx, xError.ValidationError, "API Key 加密失败：加密密钥不能为空", false)
	}

	aesKey := deriveAESKey(secret)

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", xError.NewError(ctx, xError.ServerInternalError, "创建 AES cipher 失败", false, err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", xError.NewError(ctx, xError.ServerInternalError, "创建 GCM 失败", false, err)
	}

	nonce := make([]byte, gcmNonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", xError.NewError(ctx, xError.ServerInternalError, "生成 GCM nonce 失败", false, err)
	}

	// Seal: nonce || ciphertext+tag
	sealed := gcm.Seal(nonce, nonce, []byte(key), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// DecryptAPIKey 使用 AES-GCM 解密 LLM API Key
//
// 解密流程：
//  1. base64 解码 → nonce[12] + ciphertext+tag
//  2. SHA-256(secret) → 32 字节 AES 密钥
//  3. GCM Open 验证 Tag 并还原明文
//
// 参数说明:
//   - encrypted: EncryptAPIKey 返回的 base64 编码字符串
//   - secret:    加密时使用的同一密钥种子
//
// 返回值:
//   - string: 明文 API Key
//   - *xError.Error: 密钥为空、base64 解码失败、GCM 认证失败时返回
func DecryptAPIKey(encrypted string, secret string) (string, *xError.Error) {
	ctx := context.Background()

	if secret == "" {
		return "", xError.NewError(ctx, xError.ValidationError, "API Key 解密失败：加密密钥不能为空", false)
	}

	raw, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", xError.NewError(ctx, xError.DeserializeErr, "API Key base64 解码失败", false, err)
	}

	if len(raw) < gcmNonceSize {
		return "", xError.NewError(ctx, xError.ValidationError, "API Key 解密失败：密文数据过短", false)
	}

	aesKey := deriveAESKey(secret)

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", xError.NewError(ctx, xError.ServerInternalError, "创建 AES cipher 失败", false, err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", xError.NewError(ctx, xError.ServerInternalError, "创建 GCM 失败", false, err)
	}

	nonce := raw[:gcmNonceSize]
	ciphertext := raw[gcmNonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", xError.NewError(ctx, xError.ValidationError, "API Key GCM 认证失败（密钥不匹配或数据已被篡改）", false, err)
	}

	return string(plaintext), nil
}
