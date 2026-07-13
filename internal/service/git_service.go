package service

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	xCryptoSSH "golang.org/x/crypto/ssh"

	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
)

// ── Git 克隆服务 ──
//
// GitCloneService 封装 go-git 操作，支持公开 HTTPS 克隆和私有 SSH Key 克隆。
// 所有 Git 操作通过 go-git 纯 Go 实现完成，不依赖系统 git CLI。
//
// SSH 私钥处理：
//   - CloneRepo 接收明文 PEM 格式 SSH 私钥，直接用于认证
//   - 私钥来源由调用方（pipeline）从 SshKey 实体读取
//
// known_hosts 处理（开发阶段宽松模式）：
//   - 当前实现使用 InsecureIgnoreHostKey 跳过主机密钥验证
//   - 生产环境应切换为 NewKnownHostsCallback 读取系统 known_hosts 文件

// GitCloneService Git 仓库操作服务
type GitCloneService struct {
	log *xLog.LogNamedLogger // 专用日志记录器
}

// NewGitCloneService 创建 GitCloneService 实例
func NewGitCloneService() *GitCloneService {
	return &GitCloneService{
		log: xLog.WithName(xLog.NamedCONT, "GitCloneSvc"),
	}
}

// CloneRepo 克隆仓库到指定路径
//
// 支持两种认证模式：
//   - HTTPS 公开克隆：privateKey 为空字符串时使用匿名 HTTPS
//   - SSH Key 私有克隆：privateKey 为明文 PEM 格式私钥，直接用于认证
//
// 参数说明:
//   - ctx:        上下文（控制克隆超时）
//   - gitURL:     仓库地址（HTTPS 或 SSH 协议）
//   - branch:     要克隆的分支名（空字符串则克隆默认分支）
//   - privateKey: 明文 PEM 格式 SSH 私钥（空字符串表示公开 HTTPS 克隆）
//   - destPath:   克隆目标路径（必须不存在或为空目录）
//
// 返回值:
//   - error: 克隆过程中的错误
func (s *GitCloneService) CloneRepo(ctx context.Context, gitURL, branch, privateKey, destPath string) error {
	opts := &git.CloneOptions{
		URL:   gitURL,
		Depth: 0, // 完整克隆，支持 checkout 到任意历史 hash
	}

	// 设置分支引用（空则克隆默认分支 HEAD）
	if branch != "" {
		opts.ReferenceName = plumbing.NewBranchReferenceName(branch)
	}

	// SSH 私有仓库认证
	if privateKey != "" {
		auth, err := s.buildSSHAuth(privateKey)
		if err != nil {
			return fmt.Errorf("SSH 认证准备失败: %w", err)
		}
		opts.Auth = auth
	}

	if _, err := git.PlainCloneContext(ctx, destPath, false, opts); err != nil {
		return fmt.Errorf("克隆仓库失败 [%s]: %w", gitURL, err)
	}

	s.log.Info(ctx, "仓库克隆成功",
		slog.String("url", gitURL),
		slog.String("branch", branch),
		slog.String("dest", destPath),
	)
	return nil
}

// GetCommitHash 返回仓库 HEAD 的 commit hash（40 字符十六进制）
//
// 参数说明:
//   - repoPath: 仓库本地路径（包含 .git 目录）
//
// 返回值:
//   - string: 40 字符 commit hash
//   - error:  仓库打开失败或 HEAD 引用获取失败
func (s *GitCloneService) GetCommitHash(repoPath string) (string, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return "", fmt.Errorf("打开仓库失败 [%s]: %w", repoPath, err)
	}

	head, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("获取 HEAD 引用失败: %w", err)
	}

	return head.Hash().String(), nil
}

// GetChangedFiles 返回两个 commit 之间的变更文件列表
//
// 通过 go-git Patch 计算差异，收集所有新增、修改、删除的文件路径。
//
// 参数说明:
//   - repoPath: 仓库本地路径
//   - oldHash:  旧 commit hash（40 字符）
//   - newHash:  新 commit hash（40 字符）
//
// 返回值:
//   - []string: 变更文件路径列表
//   - error:    仓库打开、commit 查找、Patch 计算失败
func (s *GitCloneService) GetChangedFiles(repoPath, oldHash, newHash string) ([]string, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("打开仓库失败 [%s]: %w", repoPath, err)
	}

	oldCommit, err := repo.CommitObject(plumbing.NewHash(oldHash))
	if err != nil {
		return nil, fmt.Errorf("查找旧 commit 失败 [%s]: %w", oldHash, err)
	}

	newCommit, err := repo.CommitObject(plumbing.NewHash(newHash))
	if err != nil {
		return nil, fmt.Errorf("查找新 commit 失败 [%s]: %w", newHash, err)
	}

	patch, err := oldCommit.Patch(newCommit)
	if err != nil {
		return nil, fmt.Errorf("计算 Patch 失败: %w", err)
	}

	// 收集所有变更文件路径（去重）
	seen := make(map[string]bool)
	var files []string
	for _, fp := range patch.FilePatches() {
		from, to := fp.Files()
		switch {
		case to != nil: // 新增或修改
			if !seen[to.Path()] {
				files = append(files, to.Path())
				seen[to.Path()] = true
			}
		case from != nil: // 删除
			if !seen[from.Path()] {
				files = append(files, from.Path())
				seen[from.Path()] = true
			}
		}
	}

	return files, nil
}

// FetchAndCheckout 拉取最新代码并切换到指定 commit hash
//
// 先 fetch 远程更新，再 checkout 到指定 commit hash（强制）。
// commitHash 为空时 checkout 到 branch 的 HEAD。
func (s *GitCloneService) FetchAndCheckout(ctx context.Context, repoPath, branch, commitHash, privateKey string) error {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("打开仓库失败 [%s]: %w", repoPath, err)
	}

	// 构建 fetch 选项
	fetchOpts := &git.FetchOptions{
		RemoteName: "origin",
		Force:      true,
	}
	if privateKey != "" {
		auth, err := s.buildSSHAuth(privateKey)
		if err != nil {
			return fmt.Errorf("SSH 认证准备失败: %w", err)
		}
		fetchOpts.Auth = auth
	}

	// fetch 远程更新
	if err := repo.Fetch(fetchOpts); err != nil {
		if err != git.NoErrAlreadyUpToDate {
			return fmt.Errorf("fetch 远程更新失败: %w", err)
		}
	}

	// 获取 worktree
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("获取 Worktree 失败: %w", err)
	}

	// checkout 到指定 commit hash 或分支
	checkoutOpts := &git.CheckoutOptions{Force: true}
	if commitHash != "" {
		checkoutOpts.Hash = plumbing.NewHash(commitHash)
	} else if branch != "" {
		checkoutOpts.Branch = plumbing.NewBranchReferenceName(branch)
	}
	if err := worktree.Checkout(checkoutOpts); err != nil {
		return fmt.Errorf("checkout 失败: %w", err)
	}

	s.log.Info(ctx, "仓库 fetch 并 checkout 成功",
		slog.String("repoPath", repoPath),
		slog.String("branch", branch),
		slog.String("commitHash", commitHash),
	)
	return nil
}

// EnsureCloned 确保仓库已克隆到指定路径（幂等）
//
// 如 destPath 已有 .git 目录则跳过克隆直接返回（调用方后续调 FetchAndCheckout 更新）。
// 否则执行完整克隆。
func (s *GitCloneService) EnsureCloned(ctx context.Context, gitURL, branch, privateKey, destPath string) error {
	gitDir := filepath.Join(destPath, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		s.log.Info(ctx, "仓库已存在，跳过克隆",
			slog.String("destPath", destPath),
		)
		return nil
	}

	return s.CloneRepo(ctx, gitURL, branch, privateKey, destPath)
}

// buildSSHAuth 从明文 PEM 私钥构建 go-git SSH 认证方法
//
// 使用 ssh.NewPublicKeys 构建 PublicKeys 认证器
// 使用 InsecureIgnoreHostKey 跳过 known_hosts 验证（开发阶段宽松模式）
//
// 参数说明:
//   - privateKey: 明文 PEM 格式 SSH 私钥
//
// 返回值:
//   - *gitssh.PublicKeys: go-git SSH 认证方法（实现 transport.AuthMethod 接口）
//   - error: 私钥解析失败
func (s *GitCloneService) buildSSHAuth(privateKey string) (*gitssh.PublicKeys, error) {
	// go-git ssh.NewPublicKeys 接受 PEM 格式私钥字节数组
	auth, err := gitssh.NewPublicKeys("git", []byte(privateKey), "")
	if err != nil {
		return nil, fmt.Errorf("解析 SSH 私钥失败: %w", err)
	}

	// 开发阶段宽松模式：跳过主机密钥验证
	// 生产环境应替换为 gitssh.NewKnownHostsCallback() 读取 ~/.ssh/known_hosts
	auth.HostKeyCallback = xCryptoSSH.InsecureIgnoreHostKey()

	return auth, nil
}
