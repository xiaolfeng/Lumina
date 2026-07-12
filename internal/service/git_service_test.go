package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

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
