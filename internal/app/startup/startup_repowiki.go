package startup

import (
	"context"
	"log/slog"
	"os"

	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xCtxUtil "github.com/bamboo-services/bamboo-base-go/common/utility/context"
	xEnv "github.com/bamboo-services/bamboo-base-go/defined/env"

	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/logic"
	"github.com/xiaolfeng/Lumina/internal/repository"
	"github.com/xiaolfeng/Lumina/internal/service"
)

// repoWikiInit 初始化 RepoWiki 模块：创建存储目录、构造 RepoWikiLogic 并注入 context。
//
// 该节点必须在 databaseInit 与 nosqlInit 之后执行，因为 RepoWikiLogic 依赖 db/rdb。
// 存储目录创建失败仅记录警告，不阻塞应用启动（用户可能稍后配置外部存储）。
func (r *reg) repoWikiInit(ctx context.Context) (any, error) {
	log := xLog.WithName(xLog.NamedINIT)
	log.Debug(ctx, "正在初始化 RepoWiki...")

	// 创建 RepoWiki 存储目录
	storagePath := xEnv.GetEnvString("REPOWIKI_STORAGE_PATH", "./.lumina/repowiki")
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		log.Warn(ctx, "创建 RepoWiki 存储目录失败，请检查目录权限",
			slog.String("path", storagePath),
			slog.String("err", err.Error()))
	} else {
		log.Info(ctx, "RepoWiki 存储目录已就绪",
			slog.String("path", storagePath))
	}

	// 初始化 RepoWikiLogic 并注入 context，确保 handler 与 MCP 共享同一实例
	repoWikiLogic := logic.NewRepoWikiLogic(ctx)

	// 创建 LlmResolver 并注入（持有 Repository 引用，避免循环导入）
	db := xCtxUtil.MustGetDB(ctx)
	modelRepo := repository.NewLlmModelRepo(db)
	providerRepo := repository.NewLlmProviderRepo(db)
	infoRepo := repository.NewInfoRepo(db)
	encryptSecret := xEnv.GetEnvString("LLM_ENCRYPT_SECRET", "")
	resolver := service.NewLlmResolver(modelRepo, providerRepo, infoRepo, encryptSecret)
	repoWikiLogic.SetLlmResolver(resolver)

	// 创建并注入 DocumentAssembler（Pipeline Step 5 依赖此组件生成 Wiki 文档）
	storageService := service.NewWikiStorageService()
	repoWikiLogic.SetDocumentAssembler(logic.NewDocumentAssembler(storageService, nil))

	r.ctx = context.WithValue(r.ctx, bConst.RepoWikiLogicKey, repoWikiLogic)

	log.Info(ctx, "RepoWikiLogic 初始化完成")
	return repoWikiLogic, nil
}
