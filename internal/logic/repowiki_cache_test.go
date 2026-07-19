package logic

import (
	"os"
	"path/filepath"
	"testing"

	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"

	"github.com/xiaolfeng/Lumina/internal/service"
)

// TestRepoWikiLogic_CleanGitCache 验证 CleanGitCache 正确删除 Git 克隆缓存且保持幂等。
func TestRepoWikiLogic_CleanGitCache(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "repowiki-cache-test-*")
	if err != nil {
		t.Fatalf("MkdirTemp failed: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	t.Setenv("REPOWIKI_STORAGE_PATH", tmpDir)

	storage := service.NewWikiStorageService()
	l := &RepoWikiLogic{
		logic: logic{
			log: xLog.WithName(xLog.NamedLOGC, "RepoWikiLogicTest"),
		},
		svc: repowikiSvc{
			storage: storage,
		},
	}

	configID := xSnowflake.SnowflakeID(12345)
	repoPath := storage.GetRepoPath(configID.Int64())
	fakeFile := filepath.Join(repoPath, "README.md")

	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("创建假仓库目录失败: %v", err)
	}
	if err := os.WriteFile(fakeFile, []byte("# fake"), 0644); err != nil {
		t.Fatalf("创建假文件失败: %v", err)
	}

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		t.Fatal("假仓库目录应存在")
	}

	xErr := l.CleanGitCache(t.Context(), configID)
	if xErr != nil {
		t.Fatalf("CleanGitCache 应返回 nil, 得到: %v", xErr)
	}

	if _, err := os.Stat(repoPath); !os.IsNotExist(err) {
		t.Fatal("Git 克隆缓存目录应被删除")
	}

	xErr = l.CleanGitCache(t.Context(), configID)
	if xErr != nil {
		t.Fatalf("幂等调用应返回 nil, 得到: %v", xErr)
	}
}
