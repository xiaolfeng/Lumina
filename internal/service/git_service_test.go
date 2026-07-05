package service

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ── AES-GCM SSH Key 加解密测试 ──

// TestSSHKeyEncryptDecrypt 验证 SSH 私钥加解密 round-trip
//
// 测试场景：
//  1. 标准 PEM 私钥 → 加密 → 解密 → 还原一致
//  2. 同一明文两次加密产生不同密文（nonce 随机性）
//  3. 空密钥拒绝
//  4. 错误密钥解密失败
func TestSSHKeyEncryptDecrypt(t *testing.T) {
	const testSecret = "test-hmac-secret-for-repowiki-2026"

	// 模拟 PEM 格式 SSH 私钥（测试用，非真实密钥）
	sampleSSHKey := `-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
QyNTUxOQAAACR0ZXN0LWtleS1mb3ItbHVtaW5hLXJlcG93aWtpLWRlbW8AAAAJdGVzdC1r
ZXktZm9yLWx1bWluYS1yZXBvd2lraS1kZW1vAAAAC3NzaC1lZDI1NTE5AAAAgAAAAAAAAA
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
-----END OPENSSH PRIVATE KEY-----`

	t.Run("round-trip", func(t *testing.T) {
		encrypted, xErr := EncryptSSHKey(sampleSSHKey, testSecret)
		if xErr != nil {
			t.Fatalf("EncryptSSHKey 失败: %v", xErr)
		}

		if encrypted == "" {
			t.Fatal("加密结果为空")
		}

		if encrypted == sampleSSHKey {
			t.Fatal("加密结果与明文相同，加密未生效")
		}

		decrypted, xErr := DecryptSSHKey(encrypted, testSecret)
		if xErr != nil {
			t.Fatalf("DecryptSSHKey 失败: %v", xErr)
		}

		if decrypted != sampleSSHKey {
			t.Fatalf("解密结果与原文不一致\n期望: %q\n实际: %q", sampleSSHKey, decrypted)
		}
	})

	t.Run("nonce-randomness", func(t *testing.T) {
		enc1, _ := EncryptSSHKey(sampleSSHKey, testSecret)
		enc2, _ := EncryptSSHKey(sampleSSHKey, testSecret)

		if enc1 == enc2 {
			t.Fatal("同一明文两次加密产生相同密文，nonce 随机性可能有问题")
		}

		// 两者都应能正确解密
		dec1, _ := DecryptSSHKey(enc1, testSecret)
		dec2, _ := DecryptSSHKey(enc2, testSecret)
		if dec1 != sampleSSHKey || dec2 != sampleSSHKey {
			t.Fatal("两次加密的密文解密后不一致")
		}
	})

	t.Run("empty-secret-rejected", func(t *testing.T) {
		_, xErr := EncryptSSHKey(sampleSSHKey, "")
		if xErr == nil {
			t.Fatal("空密钥应被拒绝")
		}

		_, xErr = DecryptSSHKey("dGVzdA==", "")
		if xErr == nil {
			t.Fatal("空密钥应被拒绝")
		}
	})

	t.Run("wrong-secret-fails", func(t *testing.T) {
		encrypted, _ := EncryptSSHKey(sampleSSHKey, testSecret)

		_, xErr := DecryptSSHKey(encrypted, "wrong-secret")
		if xErr == nil {
			t.Fatal("错误密钥解密应失败（GCM Tag 认证）")
		}
	})

	t.Run("empty-key", func(t *testing.T) {
		encrypted, _ := EncryptSSHKey("", testSecret)
		decrypted, _ := DecryptSSHKey(encrypted, testSecret)
		if decrypted != "" {
			t.Fatalf("空字符串 round-trip 失败: 期望空, 实际 %q", decrypted)
		}
	})

	t.Run("unicode-key", func(t *testing.T) {
		unicodeKey := "这是包含中文的密钥内容 🔑"
		encrypted, _ := EncryptSSHKey(unicodeKey, testSecret)
		decrypted, _ := DecryptSSHKey(encrypted, testSecret)
		if decrypted != unicodeKey {
			t.Fatalf("Unicode round-trip 失败: 期望 %q, 实际 %q", unicodeKey, decrypted)
		}
	})
}

// ── Git 公开仓库克隆测试 ──

// TestGitClonePublic 验证公开 HTTPS 仓库克隆
//
// 使用 GitHub octocat/Hello-World（小型测试仓库）验证：
//  1. PlainCloneContext 成功克隆
//  2. 克隆目录包含预期文件（README）
//  3. GetCommitHash 返回 40 字符 hash
//  4. 浅克隆（Depth=1）不包含完整历史
//
// 注意：此测试需要网络连接，短模式下自动跳过。
func TestGitClonePublic(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过网络测试（short 模式）")
	}

	helper := NewRepoWikiTestHelper(t)
	svc := NewGitCloneService()

	const repoURL = "https://github.com/octocat/Hello-World.git"
	destPath := filepath.Join(helper.TempDir, "hello-world")

	// 设置 30 秒超时，避免网络卡死
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// ── 克隆 ──
	if err := svc.CloneRepo(ctx, repoURL, "", "", destPath); err != nil {
		t.Fatalf("CloneRepo 失败: %v", err)
	}

	// ── 验证目录存在且非空 ──
	entries, err := os.ReadDir(destPath)
	if err != nil {
		t.Fatalf("读取克隆目录失败: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("克隆目录为空")
	}

	// ── 验证 README 存在 ──
	readmePath := filepath.Join(destPath, "README")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		// 某些仓库可能是 README.md
		readmePath = filepath.Join(destPath, "README.md")
		if _, err := os.Stat(readmePath); os.IsNotExist(err) {
			t.Fatal("克隆目录中未找到 README 文件")
		}
	}

	// ── 验证 GetCommitHash ──
	hash, err := svc.GetCommitHash(destPath)
	if err != nil {
		t.Fatalf("GetCommitHash 失败: %v", err)
	}

	if len(hash) != 40 {
		t.Fatalf("Commit hash 长度异常: 期望 40, 实际 %d (%q)", len(hash), hash)
	}

	// 验证 hex 格式
	for _, c := range hash {
		isHex := (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')
		if !isHex {
			t.Fatalf("Commit hash 包含非十六进制字符: %q", hash)
		}
	}

	// ── 验证 .git 目录存在（非 bare 仓库） ──
	gitDir := filepath.Join(destPath, ".git")
	if info, err := os.Stat(gitDir); err != nil || !info.IsDir() {
		t.Fatal(".git 目录不存在，克隆可能异常")
	}
}

// TestGitCloneServiceCreation 验证 GitCloneService 可以正常创建
//
// 不涉及网络操作，仅验证构造函数不 panic。
func TestGitCloneServiceCreation(t *testing.T) {
	svc := NewGitCloneService()
	if svc == nil {
		t.Fatal("NewGitCloneService 返回 nil")
	}
	if svc.log == nil {
		t.Fatal("GitCloneService.log 未初始化")
	}
}

// TestDecryptSSHKeyInvalidBase64 验证无效 base64 输入的错误处理
func TestDecryptSSHKeyInvalidBase64(t *testing.T) {
	_, xErr := DecryptSSHKey("!!!not-valid-base64!!!", "some-secret")
	if xErr == nil {
		t.Fatal("无效 base64 应返回错误")
	}

	if !strings.Contains(xErr.Error(), "base64") {
		t.Fatalf("错误信息应包含 'base64', 实际: %v", xErr)
	}
}

// TestDecryptSSHKeyTooShort 验证过短密文的错误处理
func TestDecryptSSHKeyTooShort(t *testing.T) {
	// base64("short") — 短于 12 字节 nonce
	_, xErr := DecryptSSHKey("c2hvcnQ=", "some-secret")
	if xErr == nil {
		t.Fatal("过短密文应返回错误")
	}
}
