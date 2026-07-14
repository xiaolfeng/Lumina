// Package logic RepoWiki 分析管道编排器。
//
// AnalysisPipeline 串联 RepoWiki 的完整分析流程：
//
//	cloning → orchestrating（SubAgentOrchestrator 5 阶段 Wiki 生成）→ completed
//
// Pipeline 只负责 Git 准备、状态机驱动和产出路径登记；
// 真正的 LLM 编排由 SubAgentOrchestrator 完成。
// 任意步骤失败将版本标记为 failed 并终止管道。
//
// Pipeline 由 RepoWikiLogic.AnalyzeRepo 在后台 goroutine 中启动，
// 通过指针引用 Logic 持有的全部 service 和 repository 依赖。
package logic

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"

	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
)

// ──────────────────────────────────────────────────────────────────────
// AnalysisPipeline
// ──────────────────────────────────────────────────────────────────────

// AnalysisPipeline 分析管道编排器
//
// 职责：
//   - Git 仓库幂等克隆 / 增量更新（复用 config 级克隆目录）
//   - 驱动 SubAgentOrchestrator 执行 5 阶段 Wiki 生成
//   - 每一步更新版本状态（status + current_stage + error_msg）
//   - 完成后登记产出路径并清理旧版 config 级 wiki 目录
//   - Context 取消时优雅停止并标记为 cancelled
//
// 非职责：
//   - 不负责版本记录创建（由 RepoWikiLogic.AnalyzeRepo 完成）
//   - 不负责并发控制（由 RepoWikiLogic.semaphore 管理）
//   - 不负责 LLM 角色配置解析（由 RepoWikiLogic.resolveOrchestrator 完成）
type AnalysisPipeline struct {
	logic           *RepoWikiLogic     // 引用 Logic 以访问 service / repository
	log             *xLog.LogNamedLogger
	orchestrator    *SubAgentOrchestrator // 子 Agent 编排引擎（5 阶段 Wiki 生成）
	llmProviderName string                // LLM Provider 协议名称（写入版本记录，取 coordinator 角色）
	llmModelName    string                // LLM 模型名称（写入版本记录，取 coordinator 角色）
}

// NewAnalysisPipeline 创建 AnalysisPipeline 实例
//
// 参数说明:
//   - l:            RepoWikiLogic 引用（Pipeline 通过它访问全部依赖）
//   - log:          日志记录器（建议复用 RepoWikiLogic 的 log 实例）
//   - orchestrator: SubAgentOrchestrator 实例（由 AnalyzeRepo 解析 5 角色配置后构建）
//   - providerName: LLM Provider 协议名称（写入 WikiVersion.LLMProvider，取 coordinator 角色）
//   - modelName:    LLM 模型名称（写入 WikiVersion.LLMModel，取 coordinator 角色）
func NewAnalysisPipeline(l *RepoWikiLogic, log *xLog.LogNamedLogger, orchestrator *SubAgentOrchestrator, providerName, modelName string) *AnalysisPipeline {
	return &AnalysisPipeline{
		logic:           l,
		log:             log,
		orchestrator:    orchestrator,
		llmProviderName: providerName,
		llmModelName:    modelName,
	}
}

// ──────────────────────────────────────────────────────────────────────
// Execute
// ──────────────────────────────────────────────────────────────────────

// Execute 执行完整分析流程
//
// 分析步骤：
//  1. **Cloning** — EnsureCloned（幂等）+ FetchAndCheckout（拉取最新）+ GetCommitHash
//  2. **Orchestrating** — SubAgentOrchestrator.Execute 驱动 overview/explore/architect/writer/validator 5 阶段
//  3. **Completed** — 登记产出路径、自动指向新版本、清理旧版 wiki 目录
//
// 错误处理：
//   - 任意步骤失败 → updateStatus(failed) + 返回 *xError.Error
//   - Context 取消 → updateStatus(cancelled) + 返回取消错误
//   - 总超时（RepoWikiPipelineTimeoutMin，默认 60 分钟）由 Execute 顶层 context 控制
func (p *AnalysisPipeline) Execute(
	ctx context.Context,
	config *entity.RepoWikiConfig,
	version *entity.WikiVersion,
) *xError.Error {
	// 总超时控制：RepoWikiPipelineTimeoutMin（默认 60 分钟）
	ctx, cancel := context.WithTimeout(ctx, time.Duration(bConst.RepoWikiPipelineTimeoutMin)*time.Minute)
	defer cancel()

	startedAt := time.Now()
	versionID := version.ID.Int64()

	p.log.Info(ctx, "分析管道启动",
		slog.Int64("versionID", versionID),
		slog.String("gitURL", config.GitURL),
		slog.String("branch", version.Branch),
		slog.Int("timeoutMin", bConst.RepoWikiPipelineTimeoutMin))

	version.LLMProvider = p.llmProviderName
	version.LLMModel = p.llmModelName

	// 辅助函数：更新版本和配置状态（使用独立 context 防止管道取消影响 DB 写入）
	updateStatus := func(status, stage, errMsg string) {
		bgCtx := context.Background()

		version.Status = status
		if stage != "" {
			version.CurrentStage = stage
		}
		if errMsg != "" {
			version.ErrorMsg = errMsg
		}
		if status == bConst.RepoWikiStatusCompleted {
			now := time.Now()
			version.CompletedAt = &now
			version.DurationMs = int(now.Sub(startedAt).Milliseconds())
		}
		if status == bConst.RepoWikiStatusCloning && version.StartedAt == nil {
			now := time.Now()
			version.StartedAt = &now
		}

		if uErr := p.logic.repo.version.Update(bgCtx, version); uErr != nil {
			p.log.Error(bgCtx, "更新版本状态失败",
				slog.Int64("versionID", versionID),
				slog.String("targetStatus", status),
				slog.String("err", uErr.Error()))
		}

		// 同步更新配置的冗余 status 字段（列表页快速展示用）
		config.Status = status
		if uErr := p.logic.repo.config.Update(bgCtx, config); uErr != nil {
			p.log.Error(bgCtx, "更新配置状态失败",
				slog.Int64("configID", config.ID.Int64()),
				slog.String("err", uErr.Error()))
		}
	}

	// ════════════════════════════════════════════════════════════════
	// Step 1: Git 克隆 / 增量更新
	// ════════════════════════════════════════════════════════════════
	updateStatus(bConst.RepoWikiStatusCloning, "", "")

	// 复用 config 级克隆目录（repos/{configID}/），跨版本增量更新
	repoPath := p.logic.svc.storage.GetRepoPath(config.ID.Int64())
	cloneBranch := config.DefaultBranch
	if cloneBranch == "" {
		cloneBranch = version.Branch
	}

	p.log.Info(ctx, "Step 1/3: 准备仓库（幂等克隆 + 增量更新）",
		slog.Int64("versionID", versionID),
		slog.String("gitURL", config.GitURL),
		slog.String("branch", cloneBranch),
		slog.String("repoPath", repoPath))

	// 查询关联的 SSH 密钥明文私钥（SSHKeyID 为空时使用 HTTPS 匿名克隆）
	var privateKey string
	if config.SSHKeyID != nil {
		sshKey, found, xErr := p.logic.repo.sshKey.GetByID(ctx, *config.SSHKeyID)
		if xErr != nil {
			p.log.Error(ctx, "查询 SSH 密钥失败",
				slog.Int64("versionID", versionID),
				slog.Int64("sshKeyID", config.SSHKeyID.Int64()),
				slog.String("err", xErr.Error()))
			updateStatus(bConst.RepoWikiStatusFailed, "", xErr.Error())
			return xErr
		}
		if !found {
			p.log.Error(ctx, "关联的 SSH 密钥不存在",
				slog.Int64("versionID", versionID),
				slog.Int64("sshKeyID", config.SSHKeyID.Int64()))
			errMsg := fmt.Sprintf("关联的 SSH 密钥不存在 [sshKeyID=%d]", config.SSHKeyID.Int64())
			updateStatus(bConst.RepoWikiStatusFailed, "", errMsg)
			return xError.NewError(ctx, xError.NotFound, xError.ErrMessage(errMsg), false, nil)
		}
		privateKey = sshKey.PrivateKey
	}

	// 幂等克隆：已存在则跳过
	if err := p.logic.svc.git.EnsureCloned(ctx, config.GitURL, cloneBranch, privateKey, repoPath); err != nil {
		p.log.Error(ctx, "仓库克隆失败",
			slog.Int64("versionID", versionID),
			slog.String("err", err.Error()))
		updateStatus(bConst.RepoWikiStatusFailed, "", err.Error())
		return xError.NewError(ctx, xError.ServerInternalError,
			xError.ErrMessage("仓库克隆失败: "+err.Error()), false, err)
	}

	// 增量更新：fetch + checkout 到 branch HEAD（commitHash 留空，拉取最新）
	if err := p.logic.svc.git.FetchAndCheckout(ctx, repoPath, cloneBranch, "", privateKey); err != nil {
		p.log.Error(ctx, "仓库增量更新失败",
			slog.Int64("versionID", versionID),
			slog.String("err", err.Error()))
		updateStatus(bConst.RepoWikiStatusFailed, "", err.Error())
		return xError.NewError(ctx, xError.ServerInternalError,
			xError.ErrMessage("仓库增量更新失败: "+err.Error()), false, err)
	}

	// 获取 commit hash（失败不阻塞流程，记录为 "unknown"）
	commitHash, hashErr := p.logic.svc.git.GetCommitHash(repoPath)
	if hashErr != nil {
		p.log.Warn(ctx, "获取 commit hash 失败，使用 'unknown'",
			slog.Int64("versionID", versionID),
			slog.String("err", hashErr.Error()))
		commitHash = "unknown"
	}
	version.CommitHash = commitHash

	// Context 取消检查
	if ctx.Err() != nil {
		updateStatus(bConst.RepoWikiStatusCancelled, "", ctx.Err().Error())
		return xError.NewError(ctx, xError.ServerInternalError,
			xError.ErrMessage("分析任务已取消"), false, ctx.Err())
	}

	// ════════════════════════════════════════════════════════════════
	// Step 2: SubAgentOrchestrator 执行（5 阶段 Wiki 生成）
	// ════════════════════════════════════════════════════════════════
	p.log.Info(ctx, "Step 2/3: 启动 SubAgentOrchestrator 编排",
		slog.Int64("versionID", versionID),
		slog.String("repoPath", repoPath))

	if oErr := p.orchestrator.Execute(ctx, func(stage string) {
		// Orchestrator 回调：更新 current_stage（status 保持 orchestrating/analyzing 语义）
		updateStatus(bConst.RepoWikiStatusAnalyzing, stage, "")
	}); oErr != nil {
		p.log.Error(ctx, "SubAgentOrchestrator 执行失败",
			slog.Int64("versionID", versionID),
			slog.String("err", oErr.Error()))
		updateStatus(bConst.RepoWikiStatusFailed, "", oErr.Error())
		return oErr
	}

	// Context 取消检查
	if ctx.Err() != nil {
		updateStatus(bConst.RepoWikiStatusCancelled, "", ctx.Err().Error())
		return xError.NewError(ctx, xError.ServerInternalError,
			xError.ErrMessage("分析任务已取消"), false, ctx.Err())
	}

	// ════════════════════════════════════════════════════════════════
	// Step 3: 完成处理（登记产出路径 + 自动指向新版本 + 清理旧路径）
	// ════════════════════════════════════════════════════════════════
	version.StoragePath = p.logic.svc.storage.GetVersionPath(versionID)
	version.WikiPath = p.logic.svc.storage.GetWikiPath(versionID)
	version.ExploreOutputsPath = p.logic.svc.storage.GetExploreOutputsPath(versionID)
	version.ArchitecturePath = p.logic.svc.storage.GetArchitecturePath(versionID)
	version.ManifestPath = p.logic.svc.storage.GetManifestPath(versionID)

	// 自动指向新版本（SelectedVersionID）
	config.SelectedVersionID = &version.ID

	updateStatus(bConst.RepoWikiStatusCompleted, "", "")

	// 清理旧版 config 级 wiki 目录（best-effort，失败仅记录日志）
	legacyWikiPath := p.logic.svc.storage.GetLegacyWikiPath(config.ID.Int64())
	if info, statErr := os.Stat(legacyWikiPath); statErr == nil && info.IsDir() {
		if rmErr := os.RemoveAll(legacyWikiPath); rmErr != nil {
			p.log.Warn(ctx, "清理旧版 wiki 目录失败（不影响新版本）",
				slog.Int64("versionID", versionID),
				slog.String("legacyPath", legacyWikiPath),
				slog.String("err", rmErr.Error()))
		} else {
			p.log.Info(ctx, "已清理旧版 config 级 wiki 目录",
				slog.Int64("versionID", versionID),
				slog.String("legacyPath", legacyWikiPath))
		}
	}

	p.log.Info(ctx, "分析管道执行完成",
		slog.Int64("versionID", versionID),
		slog.String("commitHash", version.CommitHash),
		slog.Int("durationMs", version.DurationMs),
		slog.String("wikiPath", version.WikiPath))

	return nil
}
