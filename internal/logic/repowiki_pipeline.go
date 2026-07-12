// Package logic RepoWiki 分析管道编排器。
//
// AnalysisPipeline 串联 RepoWiki 的完整分析流程：
//
//	cloning → scanning（文件扫描 + 依赖提取）→ analyzing（4 Pass Agent 分析）→ assembling（文档组装）→ completed
//
// 每一步完成后立即更新 WikiVersion 和 RepoWikiConfig 的状态字段，
// 使 Agent 端轮询能实时感知分析进度。任意步骤失败将版本标记为 failed 并终止管道。
//
// Pipeline 由 RepoWikiLogic.AnalyzeRepo 在后台 goroutine 中启动，
// 通过指针引用 Logic 持有的全部 service 和 repository 依赖。
package logic

import (
	"context"
	"fmt"
	"log/slog"
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
//   - 串联 Git 克隆 → 文件扫描 → 依赖提取 → Agent 分析 → 文档组装完整流程
//   - 每一步完成后更新版本状态（status + current_stage + error_msg）
//   - 同步更新配置的冗余 status 字段（供列表页快速展示）
//   - 任意步骤失败时标记版本为 failed 并返回错误
//   - Context 取消时优雅停止并标记为 cancelled
//
// 非职责：
//   - 不负责版本记录创建（由 RepoWikiLogic.AnalyzeRepo 完成）
//   - 不负责并发控制（由 RepoWikiLogic.semaphore 管理）
//   - 不负责 LLM Provider 初始化（由 RepoWikiLogic 构造时完成）
type AnalysisPipeline struct {
	logic           *RepoWikiLogic    // 引用 Logic 以访问 service / repository / assembler
	log             *xLog.LogNamedLogger
	runner          *AgentPassRunner  // Agent 分析 Pass 运行器（由 AnalyzeRepo 热构建后注入）
	llmProviderName string            // LLM Provider 协议名称（写入版本记录）
	llmModelName    string            // LLM 模型名称（写入版本记录）
}

// NewAnalysisPipeline 创建 AnalysisPipeline 实例
//
// 参数说明:
//   - l:            RepoWikiLogic 引用（Pipeline 通过它访问全部依赖）
//   - log:          日志记录器（建议复用 RepoWikiLogic 的 log 实例）
//   - runner:       AgentPassRunner 实例（由 AnalyzeRepo 热构建后传入）
//   - providerName: LLM Provider 协议名称（写入 WikiVersion.LLMProvider）
//   - modelName:    LLM 模型名称（写入 WikiVersion.LLMModel）
func NewAnalysisPipeline(l *RepoWikiLogic, log *xLog.LogNamedLogger, runner *AgentPassRunner, providerName, modelName string) *AnalysisPipeline {
	return &AnalysisPipeline{
		logic:           l,
		log:             log,
		runner:          runner,
		llmProviderName: providerName,
		llmModelName:    modelName,
	}
}

// ──────────────────────────────────────────────────────────────────────
// Execute
// ──────────────────────────────────────────────────────────────────────

// Execute 执行完整分析流程
//
// 分析步骤（每步完成后更新版本状态）：
//  1. **Cloning** — 调用 GitCloneService.CloneRepo 克隆仓库到版本 raw 目录
//  2. **Scanning（文件扫描）** — FileScannerService.Scan 生成文件元数据 → 写 file_scan.json
//  3. **Scanning（依赖提取）** — DependencyExtractorService.Extract 构建模块依赖图 → 写 dep_summary.json
//  4. **Analyzing** — AgentPassRunner.RunAllPasses 执行 4 个串行 LLM 分析 Pass
//  5. **Assembling** — DocumentAssembler.Assemble 将 Pass JSON 组装为 Wiki Markdown
//  6. **Completed** — 标记完成，记录耗时和 Token 消耗
//
// 错误处理：
//   - 任意步骤失败 → updateStatus(failed) + 返回 *xError.Error
//   - Context 取消 → updateStatus(cancelled) + 返回取消错误
//
// 参数说明:
//   - ctx:    上下文（后台 goroutine 使用 context.Background()，支持取消）
//   - config: RepoWiki 配置实体（含 GitURL / SSHKeyID / 分支等）
//   - version: Wiki 版本实体（status 初始为 pending，管道执行中持续更新）
//
// 返回值:
//   - *xError.Error: 任意步骤失败时返回错误（版本状态已更新为 failed）
func (p *AnalysisPipeline) Execute(
	ctx context.Context,
	config *entity.RepoWikiConfig,
	version *entity.WikiVersion,
) *xError.Error {
	startedAt := time.Now()
	versionID := version.ID.Int64()

	p.log.Info(ctx, "分析管道启动",
		slog.Int64("versionID", versionID),
		slog.String("gitURL", config.GitURL),
		slog.String("branch", version.Branch))

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
	// Step 1: Git 克隆
	// ════════════════════════════════════════════════════════════════
	updateStatus(bConst.RepoWikiStatusCloning, "", "")

	repoPath := p.logic.svc.storage.GetRawPath(versionID)
	cloneBranch := config.DefaultBranch
	if cloneBranch == "" {
		cloneBranch = version.Branch
	}

	p.log.Info(ctx, "Step 1/5: 开始克隆仓库",
		slog.Int64("versionID", versionID),
		slog.String("gitURL", config.GitURL),
		slog.String("branch", cloneBranch),
		slog.String("destPath", repoPath))

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

	if err := p.logic.svc.git.CloneRepo(ctx, config.GitURL, cloneBranch, privateKey, repoPath); err != nil {
		p.log.Error(ctx, "仓库克隆失败",
			slog.Int64("versionID", versionID),
			slog.String("err", err.Error()))
		updateStatus(bConst.RepoWikiStatusFailed, "", err.Error())
		return xError.NewError(ctx, xError.ServerInternalError,
			xError.ErrMessage("仓库克隆失败: "+err.Error()), false, err)
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
	// Step 2: 文件扫描
	// ════════════════════════════════════════════════════════════════
	updateStatus(bConst.RepoWikiStatusScanning, bConst.RepoWikiStageScan, "")

	p.log.Info(ctx, "Step 2/5: 开始文件扫描",
		slog.Int64("versionID", versionID),
		slog.String("repoPath", repoPath))

	fileScan, xErr := p.logic.svc.scanner.Scan(ctx, repoPath)
	if xErr != nil {
		p.log.Error(ctx, "文件扫描失败",
			slog.Int64("versionID", versionID),
			slog.String("err", xErr.Error()))
		updateStatus(bConst.RepoWikiStatusFailed, "", xErr.Error())
		return xErr
	}

	version.FileCount = fileScan.TotalFiles

	// 持久化 file_scan.json
	fileScanPath := p.logic.svc.storage.GetFileScanPath(versionID)
	if writeErr := p.logic.svc.storage.WriteJSON(fileScanPath, fileScan); writeErr != nil {
		p.log.Warn(ctx, "写入 file_scan.json 失败（不阻塞流程）",
			slog.Int64("versionID", versionID),
			slog.String("err", writeErr.Error()))
	} else {
		version.FileScanPath = fileScanPath
		p.log.Debug(ctx, "file_scan.json 已写入",
			slog.String("path", fileScanPath),
			slog.Int("totalFiles", fileScan.TotalFiles))
	}

	// Context 取消检查（文件扫描可能耗时较长）
	if ctx.Err() != nil {
		updateStatus(bConst.RepoWikiStatusCancelled, "", ctx.Err().Error())
		return xError.NewError(ctx, xError.ServerInternalError,
			xError.ErrMessage("分析任务已取消"), false, ctx.Err())
	}

	// ════════════════════════════════════════════════════════════════
	// Step 3: 依赖提取
	// ════════════════════════════════════════════════════════════════
	updateStatus(bConst.RepoWikiStatusScanning, bConst.RepoWikiStageDepExtract, "")

	p.log.Info(ctx, "Step 3/5: 开始依赖提取",
		slog.Int64("versionID", versionID))

	depSummary, err := p.logic.svc.extractor.Extract(fileScan, repoPath)
	if err != nil {
		p.log.Error(ctx, "依赖提取失败",
			slog.Int64("versionID", versionID),
			slog.String("err", err.Error()))
		updateStatus(bConst.RepoWikiStatusFailed, "", err.Error())
		return xError.NewError(ctx, xError.ServerInternalError,
			xError.ErrMessage("依赖提取失败: "+err.Error()), false, err)
	}

	// 持久化 dep_summary.json
	depSummaryPath := p.logic.svc.storage.GetDepSummaryPath(versionID)
	if writeErr := p.logic.svc.storage.WriteJSON(depSummaryPath, depSummary); writeErr != nil {
		p.log.Warn(ctx, "写入 dep_summary.json 失败（不阻塞流程）",
			slog.Int64("versionID", versionID),
			slog.String("err", writeErr.Error()))
	} else {
		version.DepSummaryPath = depSummaryPath
		p.log.Debug(ctx, "dep_summary.json 已写入",
			slog.String("path", depSummaryPath),
			slog.Int("modules", len(depSummary.Modules)))
	}

	// Context 取消检查
	if ctx.Err() != nil {
		updateStatus(bConst.RepoWikiStatusCancelled, "", ctx.Err().Error())
		return xError.NewError(ctx, xError.ServerInternalError,
			xError.ErrMessage("分析任务已取消"), false, ctx.Err())
	}

	// ════════════════════════════════════════════════════════════════
	// Step 4: Agent 4 Pass 分析
	// ════════════════════════════════════════════════════════════════
	updateStatus(bConst.RepoWikiStatusAnalyzing, "", "")

	p.log.Info(ctx, "Step 4/5: 开始 Agent 4 Pass 分析",
		slog.Int64("versionID", versionID),
		slog.String("repoPath", repoPath))

	passResults, xErr := p.runner.RunAllPasses(
		ctx,
		versionID,
		repoPath,
		fileScan,
		depSummary,
		func(stage string) {
			// 每个 Pass 开始时更新 current_stage（status 保持 analyzing）
			updateStatus(bConst.RepoWikiStatusAnalyzing, stage, "")
		},
	)
	if xErr != nil {
		p.log.Error(ctx, "Agent 分析失败",
			slog.Int64("versionID", versionID),
			slog.String("err", xErr.Error()))
		updateStatus(bConst.RepoWikiStatusFailed, "", xErr.Error())
		return xErr
	}

	// 累计 Token 消耗
	version.TokenCount = totalTokens(passResults)

	p.log.Info(ctx, "Agent 分析完成",
		slog.Int64("versionID", versionID),
		slog.Int64("totalTokens", version.TokenCount),
		slog.Int("passCount", len(passResults)))

	// Context 取消检查
	if ctx.Err() != nil {
		updateStatus(bConst.RepoWikiStatusCancelled, "", ctx.Err().Error())
		return xError.NewError(ctx, xError.ServerInternalError,
			xError.ErrMessage("分析任务已取消"), false, ctx.Err())
	}

	// ════════════════════════════════════════════════════════════════
	// Step 5: 文档组装
	// ════════════════════════════════════════════════════════════════
	updateStatus(bConst.RepoWikiStatusAssembling, bConst.RepoWikiStageAssemble, "")

	p.log.Info(ctx, "Step 5/5: 开始文档组装",
		slog.Int64("versionID", versionID))

	if p.logic.assembler != nil {
		if aErr := p.logic.assembler.Assemble(
			ctx,
			passResults,
			fileScan,
			depSummary,
			config.ProjectID.Int64(),
			version.Language,
		); aErr != nil {
			p.log.Error(ctx, "文档组装失败",
				slog.Int64("versionID", versionID),
				slog.String("err", aErr.Error()))
			updateStatus(bConst.RepoWikiStatusFailed, "", aErr.Error())
			return aErr
		}
		p.log.Info(ctx, "文档组装完成",
			slog.Int64("versionID", versionID),
			slog.Int64("projectID", config.ProjectID.Int64()))
	} else {
		p.log.Warn(ctx, "DocumentAssembler 未注入，跳过文档组装步骤（Wiki 文档将不可用）",
			slog.Int64("versionID", versionID))
	}

	// 记录版本数据存储路径
	version.StoragePath = p.logic.svc.storage.GetVersionPath(versionID)

	// ════════════════════════════════════════════════════════════════
	// Step 6: 完成
	// ════════════════════════════════════════════════════════════════
	updateStatus(bConst.RepoWikiStatusCompleted, "", "")

	p.log.Info(ctx, "分析管道执行完成",
		slog.Int64("versionID", versionID),
		slog.Int("fileCount", version.FileCount),
		slog.Int64("tokenCount", version.TokenCount),
		slog.Int("durationMs", version.DurationMs),
		slog.String("commitHash", version.CommitHash))

	return nil
}
