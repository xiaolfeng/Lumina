package service

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"strings"

	xCryptoSSH "golang.org/x/crypto/ssh"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
)

// ── SSH 密钥生成与导入 ──
//
// RepoWiki 模块在为私有仓库创建克隆凭证时需要两种来源的 SSH 密钥：
//   1. 用户在前端点击「生成密钥」时由后端即时生成 Ed25519 密钥对；
//   2. 用户直接粘贴已有的 PEM 私钥，后端解析后提取公钥与指纹。
//
// 本文件提供两个纯计算函数，不涉及数据库与缓存：
//   - GenerateEd25519KeyPair：生成 Ed25519 密钥对，返回 OpenSSH 公钥、PEM 私钥、SHA256 指纹
//   - ImportPrivateKey：解析 PEM 私钥，返回 OpenSSH 公钥、SHA256 指纹、密钥类型
//
// 私钥以明文 PEM 格式返回（用户决策：明文存储，不再 AES 加密）。

// GenerateEd25519KeyPair 生成 Ed25519 SSH 密钥对
//
// 生成流程：
//  1. crypto/ed25519.GenerateKey 生成原生 Ed25519 密钥对
//  2. ssh.NewPublicKey 将公钥包装为 ssh.PublicKey
//  3. ssh.MarshalAuthorizedKey 序列化为 OpenSSH authorized_keys 格式（ssh-ed25519 AAAA...）
//  4. ssh.MarshalPrivateKey 将私钥序列化为 OpenSSH PEM 块
//  5. ssh.FingerprintSHA256 计算公钥指纹
//
// 返回值：
//   - publicKey:   OpenSSH authorized_keys 格式公钥，形如 `ssh-ed25519 AAAA...`（无尾部换行）
//   - privateKey:  OpenSSH PEM 格式私钥，以 `-----BEGIN OPENSSH PRIVATE KEY-----` 开头
//   - fingerprint: SHA256 指纹，形如 `SHA256:base64`
//   - err:         生成失败时返回 *xError.Error
func GenerateEd25519KeyPair() (publicKey string, privateKey string, fingerprint string, err error) {
	ctx := context.Background()

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", "", xError.NewError(ctx, xError.ServerInternalError, "生成 Ed25519 密钥对失败", false, err)
	}

	// 公钥 → OpenSSH authorized_keys 格式
	sshPubKey, err := xCryptoSSH.NewPublicKey(pub)
	if err != nil {
		return "", "", "", xError.NewError(ctx, xError.ServerInternalError, "封装 SSH 公钥失败", false, err)
	}
	// MarshalAuthorizedKey 返回值末尾带 \n，裁剪以保证入库干净
	publicKey = strings.TrimRight(string(xCryptoSSH.MarshalAuthorizedKey(sshPubKey)), "\n")
	fingerprint = xCryptoSSH.FingerprintSHA256(sshPubKey)

	// 私钥 → OpenSSH PEM 块
	pemBlock, err := xCryptoSSH.MarshalPrivateKey(priv, "")
	if err != nil {
		return "", "", "", xError.NewError(ctx, xError.ServerInternalError, "序列化 SSH 私钥失败", false, err)
	}
	privateKey = string(pem.EncodeToMemory(pemBlock))

	return publicKey, privateKey, fingerprint, nil
}

// ImportPrivateKey 解析 PEM 格式 SSH 私钥，提取公钥和指纹
//
// 解析流程：
//  1. ssh.ParsePrivateKey 解析 PEM 私钥为 ssh.Signer
//  2. signer.PublicKey() 获取对应的 SSH 公钥
//  3. ssh.MarshalAuthorizedKey 序列化为 OpenSSH authorized_keys 格式
//  4. ssh.FingerprintSHA256 计算指纹
//  5. publicKey.Type() 获取密钥类型（如 ssh-ed25519、ssh-rsa、ecdsa-sha2-nistp256）
//
// 参数：
//   - privateKeyPEM: PEM 格式私钥字符串（支持 OpenSSH / PKCS#1 / PKCS#8 / SEC1）
//
// 返回值：
//   - publicKey:   OpenSSH authorized_keys 格式公钥（无尾部换行）
//   - fingerprint: SHA256 指纹
//   - keyType:     密钥类型字符串（如 ssh-ed25519）
//   - err:         私钥为空或解析失败时返回 *xError.Error
func ImportPrivateKey(privateKeyPEM string) (publicKey string, fingerprint string, keyType string, err error) {
	ctx := context.Background()

	if strings.TrimSpace(privateKeyPEM) == "" {
		return "", "", "", xError.NewError(ctx, xError.ValidationError, "导入 SSH 私钥失败：私钥内容不能为空", false)
	}

	signer, err := xCryptoSSH.ParsePrivateKey([]byte(privateKeyPEM))
	if err != nil {
		return "", "", "", xError.NewError(ctx, xError.ValidationError, "解析 SSH 私钥失败：PEM 格式错误或不支持的密钥类型", false, err)
	}

	sshPubKey := signer.PublicKey()
	publicKey = strings.TrimRight(string(xCryptoSSH.MarshalAuthorizedKey(sshPubKey)), "\n")
	fingerprint = xCryptoSSH.FingerprintSHA256(sshPubKey)
	keyType = sshPubKey.Type()

	return publicKey, fingerprint, keyType, nil
}
