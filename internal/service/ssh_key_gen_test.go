package service

import (
	"strings"
	"testing"
)

// TestGenerateEd25519KeyPair 验证 Ed25519 密钥对生成的输出格式
//
// 断言点：
//   - 公钥以 "ssh-ed25519 " 开头（OpenSSH authorized_keys 格式）
//   - 私钥以 "-----BEGIN" 开头（PEM 格式）
//   - 指纹以 "SHA256:" 开头
//   - 公钥、私钥、指纹均非空
//   - 生成的密钥可被 ImportPrivateKey 回解析（生成 ↔ 导入对称性）
func TestGenerateEd25519KeyPair(t *testing.T) {
	publicKey, privateKey, fingerprint, err := GenerateEd25519KeyPair()
	if err != nil {
		t.Fatalf("GenerateEd25519KeyPair() 返回错误: %v", err)
	}

	if publicKey == "" {
		t.Fatal("publicKey 不能为空")
	}
	if privateKey == "" {
		t.Fatal("privateKey 不能为空")
	}
	if fingerprint == "" {
		t.Fatal("fingerprint 不能为空")
	}

	if !strings.HasPrefix(publicKey, "ssh-ed25519 ") {
		t.Errorf("publicKey 应以 'ssh-ed25519 ' 开头，实际: %q", publicKey)
	}
	if !strings.HasPrefix(privateKey, "-----BEGIN") {
		t.Errorf("privateKey 应以 '-----BEGIN' 开头，实际前 20 字符: %q", privateKey[:20])
	}
	if !strings.HasPrefix(fingerprint, "SHA256:") {
		t.Errorf("fingerprint 应以 'SHA256:' 开头，实际: %q", fingerprint)
	}

	// 验证私钥 PEM 块完整（含 BEGIN 和 END）
	if !strings.Contains(privateKey, "-----END") {
		t.Error("privateKey 应包含 '-----END' 结束标记")
	}
	if !strings.Contains(privateKey, "OPENSSH PRIVATE KEY") {
		t.Error("privateKey 应为 OpenSSH 格式（含 'OPENSSH PRIVATE KEY' 标识）")
	}

	// 公钥不应包含换行（保证入库干净）
	if strings.Contains(publicKey, "\n") {
		t.Errorf("publicKey 不应包含换行符，实际: %q", publicKey)
	}
}

// TestImportPrivateKey 验证从 PEM 私钥导入并提取公钥与指纹
//
// 断言点：
//   - 用 GenerateEd25519KeyPair 生成的私钥导入，应成功
//   - 导入后的指纹应与生成时的指纹一致（同一密钥对）
//   - keyType 应为 "ssh-ed25519"
//   - 导入后的公钥应与生成时的公钥一致
//   - 空私钥应返回错误
func TestImportPrivateKey(t *testing.T) {
	// 生成一对密钥用于测试
	expectedPub, privateKey, expectedFingerprint, err := GenerateEd25519KeyPair()
	if err != nil {
		t.Fatalf("前置 GenerateEd25519KeyPair() 失败: %v", err)
	}

	publicKey, fingerprint, keyType, err := ImportPrivateKey(privateKey)
	if err != nil {
		t.Fatalf("ImportPrivateKey() 返回错误: %v", err)
	}

	if publicKey != expectedPub {
		t.Errorf("公钥不匹配\n生成: %q\n导入: %q", expectedPub, publicKey)
	}
	if fingerprint != expectedFingerprint {
		t.Errorf("指纹不匹配\n生成: %q\n导入: %q", expectedFingerprint, fingerprint)
	}
	if keyType != "ssh-ed25519" {
		t.Errorf("keyType 应为 'ssh-ed25519'，实际: %q", keyType)
	}
}

// TestImportPrivateKey_Empty 验证空私钥入参的错误处理
func TestImportPrivateKey_Empty(t *testing.T) {
	_, _, _, err := ImportPrivateKey("")
	if err == nil {
		t.Fatal("空私钥应返回错误，实际返回 nil")
	}

	_, _, _, err = ImportPrivateKey("   \n\t  ")
	if err == nil {
		t.Fatal("纯空白私钥应返回错误，实际返回 nil")
	}
}

// TestImportPrivateKey_Invalid 验证无效私钥入参的错误处理
func TestImportPrivateKey_Invalid(t *testing.T) {
	_, _, _, err := ImportPrivateKey("not a valid pem private key")
	if err == nil {
		t.Fatal("无效私钥应返回错误，实际返回 nil")
	}
}

// TestGenerateEd25519KeyPair_Uniqueness 验证多次生成密钥对的唯一性
//
// Ed25519 随机生成，连续两次生成的指纹必须不同，否则说明随机源失效。
func TestGenerateEd25519KeyPair_Uniqueness(t *testing.T) {
	_, _, fp1, err := GenerateEd25519KeyPair()
	if err != nil {
		t.Fatalf("第一次 GenerateEd25519KeyPair() 失败: %v", err)
	}
	_, _, fp2, err := GenerateEd25519KeyPair()
	if err != nil {
		t.Fatalf("第二次 GenerateEd25519KeyPair() 失败: %v", err)
	}
	if fp1 == fp2 {
		t.Errorf("两次生成的指纹相同，随机源可能失效: %q", fp1)
	}
}
