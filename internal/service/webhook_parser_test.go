package service

import (
	"net/http"
	"testing"
)

func TestParseWebhookPayload(t *testing.T) {
	tests := []struct {
		name            string
		provider        string
		headers         http.Header
		body            []byte
		expectedBranch  string
		expectedIsPush  bool
		expectedRepoURL string
		expectedFiles   []string
		expectError     bool
	}{
		{
			name:     "GitHub push event",
			provider: webhookProviderGitHub,
			headers:  http.Header{"X-Github-Event": []string{"push"}},
			body: []byte(`{
				"ref": "refs/heads/main",
				"before": "beforehash",
				"after": "afterhash",
				"repository": {"clone_url": "https://github.com/user/repo.git"},
				"commits": [
					{"added": ["a.txt"], "modified": ["b.txt"], "removed": ["c.txt"]}
				]
			}`),
			expectedBranch:  "main",
			expectedIsPush:  true,
			expectedRepoURL: "https://github.com/user/repo.git",
			expectedFiles:   []string{"a.txt", "b.txt", "c.txt"},
			expectError:     false,
		},
		{
			name:     "Gitee push event",
			provider: webhookProviderGitee,
			headers:  http.Header{"X-Gitee-Event": []string{"Push Hook"}},
			body: []byte(`{
				"ref": "refs/heads/develop",
				"before": "beforehash",
				"after": "afterhash",
				"project": {"clone_url": "https://gitee.com/user/repo.git"},
				"repository": {"clone_url": "https://gitee.com/user/repo-fallback.git"},
				"commits": [
					{"added": ["x.go"], "modified": ["y.go"], "removed": ["z.go"]}
				]
			}`),
			expectedBranch:  "develop",
			expectedIsPush:  true,
			expectedRepoURL: "https://gitee.com/user/repo.git",
			expectedFiles:   []string{"x.go", "y.go", "z.go"},
			expectError:     false,
		},
		{
			name:     "Gitee push event falls back to repository clone_url",
			provider: webhookProviderGitee,
			headers:  http.Header{"X-Gitee-Event": []string{"Push Hook"}},
			body: []byte(`{
				"ref": "refs/heads/master",
				"before": "beforehash",
				"after": "afterhash",
				"project": {"clone_url": ""},
				"repository": {"clone_url": "https://gitee.com/user/repo-fallback.git"},
				"commits": []
			}`),
			expectedBranch:  "master",
			expectedIsPush:  true,
			expectedRepoURL: "https://gitee.com/user/repo-fallback.git",
			expectedFiles:   []string{},
			expectError:     false,
		},
		{
			name:     "GitLab push event",
			provider: webhookProviderGitLab,
			headers:  http.Header{},
			body: []byte(`{
				"object_kind": "push",
				"ref": "refs/heads/feature/x",
				"before": "beforehash",
				"after": "afterhash",
				"project": {"git_http_url": "https://gitlab.com/user/repo.git"},
				"commits": [
					{"added": ["m.txt"], "modified": ["n.txt"], "removed": ["o.txt"]}
				]
			}`),
			expectedBranch:  "feature/x",
			expectedIsPush:  true,
			expectedRepoURL: "https://gitlab.com/user/repo.git",
			expectedFiles:   []string{"m.txt", "n.txt", "o.txt"},
			expectError:     false,
		},
		{
			name:     "GitLab non-push event",
			provider: webhookProviderGitLab,
			headers:  http.Header{},
			body: []byte(`{
				"object_kind": "merge_request",
				"ref": "refs/heads/main",
				"project": {"git_http_url": "https://gitlab.com/user/repo.git"},
				"commits": []
			}`),
			expectedBranch:  "",
			expectedIsPush:  false,
			expectedRepoURL: "",
			expectedFiles:   nil,
			expectError:     false,
		},
		{
			name:     "Gitea push event",
			provider: webhookProviderGitea,
			headers:  http.Header{"X-Gitea-Event": []string{"push"}},
			body: []byte(`{
				"ref": "refs/heads/main",
				"before": "beforehash",
				"after": "afterhash",
				"repository": {"clone_url": "https://gitea.com/user/repo.git"},
				"commits": [
					{"added": ["p.md"], "modified": ["q.md"], "removed": ["r.md"]}
				]
			}`),
			expectedBranch:  "main",
			expectedIsPush:  true,
			expectedRepoURL: "https://gitea.com/user/repo.git",
			expectedFiles:   []string{"p.md", "q.md", "r.md"},
			expectError:     false,
		},
		{
			name:     "GitHub non-push event (ping)",
			provider: webhookProviderGitHub,
			headers:  http.Header{"X-Github-Event": []string{"ping"}},
			body: []byte(`{
				"ref": "refs/heads/main",
				"repository": {"clone_url": "https://github.com/user/repo.git"}
			}`),
			expectedBranch:  "",
			expectedIsPush:  false,
			expectedRepoURL: "",
			expectedFiles:   nil,
			expectError:     false,
		},
		{
			name:     "Gitee non-push event",
			provider: webhookProviderGitee,
			headers:  http.Header{"X-Gitee-Event": []string{"Issue Hook"}},
			body: []byte(`{
				"ref": "refs/heads/main",
				"project": {"clone_url": "https://gitee.com/user/repo.git"}
			}`),
			expectedBranch:  "",
			expectedIsPush:  false,
			expectedRepoURL: "",
			expectedFiles:   nil,
			expectError:     false,
		},
		{
			name:     "Gitea non-push event",
			provider: webhookProviderGitea,
			headers:  http.Header{"X-Gitea-Event": []string{"create"}},
			body: []byte(`{
				"ref": "refs/heads/main",
				"repository": {"clone_url": "https://gitea.com/user/repo.git"}
			}`),
			expectedBranch:  "",
			expectedIsPush:  false,
			expectedRepoURL: "",
			expectedFiles:   nil,
			expectError:     false,
		},
		{
			name:     "GitHub push with empty commits",
			provider: webhookProviderGitHub,
			headers:  http.Header{"X-Github-Event": []string{"push"}},
			body: []byte(`{
				"ref": "refs/heads/main",
				"before": "beforehash",
				"after": "afterhash",
				"repository": {"clone_url": "https://github.com/user/repo.git"},
				"commits": []
			}`),
			expectedBranch:  "main",
			expectedIsPush:  true,
			expectedRepoURL: "https://github.com/user/repo.git",
			expectedFiles:   []string{},
			expectError:     false,
		},
		{
			name:     "GitHub push with invalid JSON",
			provider: webhookProviderGitHub,
			headers:  http.Header{"X-Github-Event": []string{"push"}},
			body:     []byte(`{ not valid json }`),
			expectedBranch:  "",
			expectedIsPush:  false,
			expectedRepoURL: "",
			expectedFiles:   nil,
			expectError:     true,
		},
		{
			name:     "Unsupported provider",
			provider: "bitbucket",
			headers:  http.Header{},
			body:     []byte(`{}`),
			expectedBranch:  "",
			expectedIsPush:  false,
			expectedRepoURL: "",
			expectedFiles:   nil,
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, isPush, err := ParseWebhookPayload(tt.provider, tt.headers, tt.body)

			if tt.expectError && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if isPush != tt.expectedIsPush {
				t.Fatalf("expected isPush=%v, got %v", tt.expectedIsPush, isPush)
			}
			if !tt.expectedIsPush {
				return
			}
			if event == nil {
				t.Fatalf("expected non-nil event for push")
			}
			if event.Branch != tt.expectedBranch {
				t.Fatalf("expected branch %q, got %q", tt.expectedBranch, event.Branch)
			}
			if event.RepoURL != tt.expectedRepoURL {
				t.Fatalf("expected repo URL %q, got %q", tt.expectedRepoURL, event.RepoURL)
			}
			if len(event.ChangedFiles) != len(tt.expectedFiles) {
				t.Fatalf("expected %d changed files, got %d", len(tt.expectedFiles), len(event.ChangedFiles))
			}
			for i, file := range tt.expectedFiles {
				if event.ChangedFiles[i] != file {
					t.Fatalf("expected changed file[%d]=%q, got %q", i, file, event.ChangedFiles[i])
				}
			}
		})
	}
}
