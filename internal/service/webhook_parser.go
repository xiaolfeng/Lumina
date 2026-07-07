package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// WebhookPushEvent 统一后的 Git 推送事件
// 表示任意支持提供商的一次代码推送，字段语义与提供商无关。
type WebhookPushEvent struct {
	Provider     string   // 提供商标识（github / gitee / gitlab / gitea）
	Branch       string   // 分支名称（已移除 refs/heads/ 前缀）
	RepoURL      string   // 仓库 HTTP 克隆地址
	BeforeHash   string   // 推送前 commit hash
	AfterHash    string   // 推送后 commit hash
	ChangedFiles []string // 变更文件列表（含 added / modified / removed，去重）
}

const (
	webhookProviderGitHub = "github"
	webhookProviderGitee  = "gitee"
	webhookProviderGitLab = "gitlab"
	webhookProviderGitea  = "gitea"
)

// gitCommit 是 Git 提供商 push payload 中 commits 数组的通用字段
type gitCommit struct {
	Added    []string `json:"added"`
	Modified []string `json:"modified"`
	Removed  []string `json:"removed"`
}

// ParseWebhookPayload 解析并统一 Git webhook 推送事件。
//
// 返回值：
//   - 当事件为 push 时返回 (*WebhookPushEvent, true, nil)
//   - 当事件为非 push 时返回 (nil, false, nil)
//   - 解析失败或 provider 不支持时返回 (nil, false, error)
func ParseWebhookPayload(provider string, headers http.Header, body []byte) (*WebhookPushEvent, bool, error) {
	switch provider {
	case webhookProviderGitHub:
		if headers.Get("X-GitHub-Event") != "push" {
			return nil, false, nil
		}
		event, err := parseGitHubPush(body)
		return event, err == nil && event != nil, err
	case webhookProviderGitee:
		if headers.Get("X-Gitee-Event") != "Push Hook" {
			return nil, false, nil
		}
		event, err := parseGiteePush(body)
		return event, err == nil && event != nil, err
	case webhookProviderGitLab:
		event, err := parseGitLabPush(body)
		return event, err == nil && event != nil, err
	case webhookProviderGitea:
		if headers.Get("X-Gitea-Event") != "push" {
			return nil, false, nil
		}
		event, err := parseGiteaPush(body)
		return event, err == nil && event != nil, err
	default:
		return nil, false, errors.New("unsupported webhook provider: " + provider)
	}
}

// parseGitHubPush 解析 GitHub push payload
func parseGitHubPush(body []byte) (*WebhookPushEvent, error) {
	var payload struct {
		Ref        string `json:"ref"`
		Before     string `json:"before"`
		After      string `json:"after"`
		Repository struct {
			CloneURL string `json:"clone_url"`
		} `json:"repository"`
		Commits []gitCommit `json:"commits"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("parse GitHub push payload failed: %w", err)
	}

	return &WebhookPushEvent{
		Provider:     webhookProviderGitHub,
		Branch:       stripRef(payload.Ref),
		RepoURL:      payload.Repository.CloneURL,
		BeforeHash:   payload.Before,
		AfterHash:    payload.After,
		ChangedFiles: collectFiles(payload.Commits),
	}, nil
}

// parseGiteePush 解析 Gitee push payload
func parseGiteePush(body []byte) (*WebhookPushEvent, error) {
	var payload struct {
		Ref     string `json:"ref"`
		Before  string `json:"before"`
		After   string `json:"after"`
		Project struct {
			CloneURL string `json:"clone_url"`
		} `json:"project"`
		Repository struct {
			CloneURL string `json:"clone_url"`
		} `json:"repository"`
		Commits []gitCommit `json:"commits"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("parse Gitee push payload failed: %w", err)
	}

	repoURL := payload.Project.CloneURL
	if repoURL == "" {
		repoURL = payload.Repository.CloneURL
	}

	return &WebhookPushEvent{
		Provider:     webhookProviderGitee,
		Branch:       stripRef(payload.Ref),
		RepoURL:      repoURL,
		BeforeHash:   payload.Before,
		AfterHash:    payload.After,
		ChangedFiles: collectFiles(payload.Commits),
	}, nil
}

// parseGitLabPush 解析 GitLab push payload
func parseGitLabPush(body []byte) (*WebhookPushEvent, error) {
	var payload struct {
		ObjectKind string `json:"object_kind"`
		Ref        string `json:"ref"`
		Before     string `json:"before"`
		After      string `json:"after"`
		Project    struct {
			GitHTTPURL string `json:"git_http_url"`
		} `json:"project"`
		Commits []gitCommit `json:"commits"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("parse GitLab push payload failed: %w", err)
	}

	if payload.ObjectKind != "push" {
		return nil, nil
	}

	return &WebhookPushEvent{
		Provider:     webhookProviderGitLab,
		Branch:       stripRef(payload.Ref),
		RepoURL:      payload.Project.GitHTTPURL,
		BeforeHash:   payload.Before,
		AfterHash:    payload.After,
		ChangedFiles: collectFiles(payload.Commits),
	}, nil
}

// parseGiteaPush 解析 Gitea push payload
func parseGiteaPush(body []byte) (*WebhookPushEvent, error) {
	var payload struct {
		Ref        string `json:"ref"`
		Before     string `json:"before"`
		After      string `json:"after"`
		Repository struct {
			CloneURL string `json:"clone_url"`
		} `json:"repository"`
		Commits []gitCommit `json:"commits"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("parse Gitea push payload failed: %w", err)
	}

	return &WebhookPushEvent{
		Provider:     webhookProviderGitea,
		Branch:       stripRef(payload.Ref),
		RepoURL:      payload.Repository.CloneURL,
		BeforeHash:   payload.Before,
		AfterHash:    payload.After,
		ChangedFiles: collectFiles(payload.Commits),
	}, nil
}

// stripRef 移除 ref 中的 refs/heads/ 前缀，得到分支名
func stripRef(ref string) string {
	return strings.TrimPrefix(ref, "refs/heads/")
}

// collectFiles 从 commits 数组中提取所有变更文件并去重
// 当 commits 为空或缺失时返回空切片，不会 panic。
func collectFiles(commits []gitCommit) []string {
	seen := make(map[string]struct{}, len(commits)*3)
	files := make([]string, 0, len(commits)*3)

	for _, commit := range commits {
		for _, file := range commit.Added {
			if _, ok := seen[file]; !ok {
				seen[file] = struct{}{}
				files = append(files, file)
			}
		}
		for _, file := range commit.Modified {
			if _, ok := seen[file]; !ok {
				seen[file] = struct{}{}
				files = append(files, file)
			}
		}
		for _, file := range commit.Removed {
			if _, ok := seen[file]; !ok {
				seen[file] = struct{}{}
				files = append(files, file)
			}
		}
	}

	return files
}
