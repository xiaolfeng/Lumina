// Package logic RepoWiki 增量更新引擎。
//
// IncrementalEngine 在已有 Wiki 版本的基础上，对比 Git commit 变更范围，
// 仅重跑受影响的 Agent Pass，避免每次代码更新都进行全量 LLM 分析。
//
// 三阶段流程：
//
//	CheckUpdateNeeded  → 判断是否需要更新（commit hash 对比）
//	AnalyzeImpactScope → 变更文件 → 受影响模块 → 受影响 Pass
//	ExecuteIncremental → 拉取最新 + 部分重跑 + 文档组装
//
// 增量策略：
//   - 入口文件变更（main.go / go.mod / package.json）→ 全部 4 个 Pass
//   - 架构文件变更（config / middleware / route）→ Pass 3
//   - 文档文件变更（.md / docs/）→ Pass 4
//   - 依赖清单变更（go.mod / package.json / Cargo.toml）→ Pass 1+2
//   - 普通源码文件变更 → Pass 2+3（默认）
//
// 未受影响的 Pass 直接复用上一版本结果，写入新版本 passes 目录保持完整。
// 文档组装阶段始终执行——即使所有 Pass 都复用，组装产物也需刷新元数据。
package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"path/filepath"
	"sort"
	"strings"
	"time"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"

	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"github.com/xiaolfeng/Lumina/internal/service"
)

// ──────────────────────────────────────────────────────────────────────
// 类型定义
// ──────────────────────────────────────────────────────────────────────

// UpdateNeedType 增量更新判断结果类型
type UpdateNeedType int

const (
	UpdateNotNeeded    UpdateNeedType = iota // commit 未变更，无需更新
	UpdateNeeded                             // 检测到变更，需要增量更新
	UpdateFullRequired                       // 无历史版本或异常，需要全量重新分析
)

// String 返回 UpdateNeedType 的可读字符串（用于日志和调试）
func (t UpdateNeedType) String() string {
	switch t {
	case UpdateNotNeeded:
		return "not_needed"
	case UpdateNeeded:
		return "needed"
	case UpdateFullRequired:
		return "full_required"
	default:
		return fmt.Sprintf("unknown(%d)", int(t))
	}
}

// UpdateCheckResult 增量判断结果
//
// 由 CheckUpdateNeeded 返回，描述是否需要更新以及（如果需要）变更文件列表。
type UpdateCheckResult struct {
	Type         UpdateNeedType // 判断结果
	OldHash      string         // 旧 commit hash（无历史版本时为空）
	NewHash      string         // 新 commit hash（无历史版本时为空）
	ChangedFiles []string       // 变更文件列表（仅 Type==UpdateNeeded 时填充）
}

// ImpactScope 影响范围分析结果
//
// 由 AnalyzeImpactScope 返回，描述变更波及的模块和需要重跑的 Pass 编号。
type ImpactScope struct {
	AffectedModules []string // 受影响的模块列表（目录路径，去重）
	AffectedPasses  []int    // 受影响的 Pass 编号（1-4，升序去重）
	ChangeSummary   string   // 变更摘要（用于日志和调试）
}

// ──────────────────────────────────────────────────────────────────────
// IncrementalEngine
// ──────────────────────────────────────────────────────────────────────

// IncrementalEngine 增量更新引擎
//
// 职责：
//   - 对比 commit hash 判断是否需要增量更新
//   - 分析变更文件的影响范围（模块 + Pass）
//   - 执行增量更新：拉取最新代码 → 重新扫描 → 按需重跑 Pass → 合并 → 文档组装
//
// 非职责：
//   - 不负责版本记录创建（由 RepoWikiLogic 调用方完成）
//   - 不负责并发控制（由 RepoWikiLogic.semaphore 管理）
//   - 不负责 LLM Provider 初始化（复用 AgentPassRunner 已有配置）
//
// 线程安全：
//   - IncrementalEngine 本身无可变状态，可并发使用
//   - 但同一次增量更新应串行调用 CheckUpdateNeeded → AnalyzeImpactScope → ExecuteIncremental
type IncrementalEngine struct {
	gitService *service.GitCloneService    // Git 操作服务（commit hash / diff / pull）
	storage    *service.WikiStorageService // 文件系统路径管理 + JSON 读写
	passRunner *AgentPassRunner            // Agent Pass 运行器（用于重跑受影响的 Pass）
	assembler  *DocumentAssembler          // 文档组装器（始终重新组装）
	logic      *RepoWikiLogic              // RepoWikiLogic 引用（访问 scanner / extractor / versionRepo）
	log        *xLog.LogNamedLogger        // 专用日志记录器
}

// NewIncrementalEngine 创建 IncrementalEngine 实例
//
// 参数说明:
//   - gitService: Git 操作服务（用于 GetCommitHash / GetChangedFiles / PullLatest）
//   - storage:    Wiki 存储服务（用于路径计算和 Pass JSON 读写）
//   - passRunner: Agent Pass 运行器（用于重跑受影响的 Pass，nil 时增量更新不可用）
//   - assembler:  文档组装器（nil 时跳过文档组装并记录警告）
//   - logic:      RepoWikiLogic 引用（访问 fileScanner / depExtractor / versionRepo）
//   - log:        日志记录器（建议使用 xLog.WithName(xLog.NamedLOGC, "IncrementalEngine")）
func NewIncrementalEngine(
	gitService *service.GitCloneService,
	storage *service.WikiStorageService,
	passRunner *AgentPassRunner,
	assembler *DocumentAssembler,
	logic *RepoWikiLogic,
	log *xLog.LogNamedLogger,
) *IncrementalEngine {
	if log == nil {
		log = xLog.WithName(xLog.NamedLOGC, "IncrementalEngine")
	}
	return &IncrementalEngine{
		gitService: gitService,
		storage:    storage,
		passRunner: passRunner,
		assembler:  assembler,
		logic:      logic,
		log:        log,
	}
}

// ──────────────────────────────────────────────────────────────────────
// CheckUpdateNeeded
// ──────────────────────────────────────────────────────────────────────

// CheckUpdateNeeded 判断是否需要增量更新
//
// 判断逻辑：
//  1. 无历史版本（latestVersion == nil）→ UpdateFullRequired
//  2. 历史版本的 commit hash 为空或 "unknown" → UpdateFullRequired（无法做 diff）
//  3. 获取仓库当前 HEAD commit hash
//  4. hash 相同 → UpdateNotNeeded
//  5. hash 不同 → UpdateNeeded，并获取两个 commit 之间的变更文件列表
//
// 参数说明:
//   - ctx:           上下文
//   - config:        RepoWiki 配置实体（使用 ID 定位仓库克隆路径）
//   - latestVersion: 最近一次成功的版本记录（nil 表示无历史版本）
//
// 返回值:
//   - *UpdateCheckResult: 判断结果
//   - *xError.Error:       仓库打开失败或 commit hash 获取失败
func (e *IncrementalEngine) CheckUpdateNeeded(
	ctx context.Context,
	config *entity.RepoWikiConfig,
	latestVersion *entity.WikiVersion,
) (*UpdateCheckResult, *xError.Error) {
	// 1. 无历史版本 → 全量
	if latestVersion == nil {
		e.log.Info(ctx, "无历史版本，需要全量分析",
			slog.Int64("configID", config.ID.Int64()))
		return &UpdateCheckResult{Type: UpdateFullRequired}, nil
	}

	// 2. 历史 commit hash 缺失 → 全量（无法 diff）
	oldHash := latestVersion.CommitHash
	if oldHash == "" || oldHash == "unknown" {
		e.log.Info(ctx, "历史版本 commit hash 缺失，需要全量分析",
			slog.Int64("configID", config.ID.Int64()),
			slog.Int64("versionID", latestVersion.ID.Int64()),
			slog.String("oldHash", oldHash))
		return &UpdateCheckResult{Type: UpdateFullRequired}, nil
	}

	// 3. 获取仓库当前 HEAD
	repoPath := e.storage.GetRepoPath(config.ID.Int64())
	newHash, err := e.gitService.GetCommitHash(repoPath)
	if err != nil {
		e.log.Error(ctx, "获取仓库 commit hash 失败",
			slog.Int64("configID", config.ID.Int64()),
			slog.String("repoPath", repoPath),
			slog.String("err", err.Error()))
		return nil, xError.NewError(ctx, xError.ServerInternalError,
			xError.ErrMessage("获取仓库 commit hash 失败: "+err.Error()), false, err)
	}

	// 4. hash 相同 → 无需更新
	if newHash == oldHash {
		e.log.Info(ctx, "commit hash 未变更，无需更新",
			slog.Int64("configID", config.ID.Int64()),
			slog.String("commitHash", newHash))
		return &UpdateCheckResult{
			Type:    UpdateNotNeeded,
			OldHash: oldHash,
			NewHash: newHash,
		}, nil
	}

	// 5. hash 不同 → 获取变更文件列表
	changedFiles, err := e.gitService.GetChangedFiles(repoPath, oldHash, newHash)
	if err != nil {
		e.log.Error(ctx, "获取变更文件列表失败",
			slog.Int64("configID", config.ID.Int64()),
			slog.String("oldHash", oldHash),
			slog.String("newHash", newHash),
			slog.String("err", err.Error()))
		return nil, xError.NewError(ctx, xError.ServerInternalError,
			xError.ErrMessage("获取变更文件列表失败: "+err.Error()), false, err)
	}

	e.log.Info(ctx, "检测到 commit 变更，需要增量更新",
		slog.Int64("configID", config.ID.Int64()),
		slog.String("oldHash", oldHash),
		slog.String("newHash", newHash),
		slog.Int("changedFiles", len(changedFiles)))

	return &UpdateCheckResult{
		Type:         UpdateNeeded,
		OldHash:      oldHash,
		NewHash:      newHash,
		ChangedFiles: changedFiles,
	}, nil
}

// ──────────────────────────────────────────────────────────────────────
// AnalyzeImpactScope
// ──────────────────────────────────────────────────────────────────────

// AnalyzeImpactScope 分析变更文件的影响范围
//
// 将变更文件映射到所属模块，再依据文件类型决定受影响的 Agent Pass。
//
// 模块映射规则：
//   - 文件所在目录即为模块（与 DependencyExtractor 的模块定义一致）
//   - 若某模块依赖被变更的模块，则该模块也视为受影响
//
// Pass 影响规则（见 determineAffectedPasses）：
//   - 入口文件 → 全部 Pass 1+2+3+4
//   - 架构文件（config/middleware/route）→ Pass 3
//   - 文档文件（.md / docs/）→ Pass 4
//   - 依赖清单（go.mod / package.json / Cargo.toml）→ Pass 1+2
//   - 普通源码文件 → Pass 2+3
//
// 参数说明:
//   - changedFiles: 变更文件相对路径列表（来自 CheckUpdateResult.ChangedFiles）
//   - depSummary:   当前依赖摘要（用于模块依赖传播分析，nil 时仅按目录映射）
func (e *IncrementalEngine) AnalyzeImpactScope(
	changedFiles []string,
	depSummary *service.DependencySummary,
) *ImpactScope {
	scope := &ImpactScope{
		AffectedModules: []string{},
		AffectedPasses:  []int{},
	}

	if len(changedFiles) == 0 {
		scope.ChangeSummary = "无变更文件"
		return scope
	}

	// 1. 将变更文件映射到所属模块（按目录）
	affectedModules := make(map[string]bool)
	for _, file := range changedFiles {
		dir := filepath.ToSlash(filepath.Dir(file))
		if dir == "" {
			dir = "."
		}
		affectedModules[dir] = true
	}

	// 2. 依赖传播：依赖被变更模块的父模块也视为受影响
	if depSummary != nil {
		for _, mod := range depSummary.Modules {
			for _, dep := range mod.Dependencies {
				if affectedModules[dep] {
					affectedModules[mod.Name] = true
				}
			}
		}
	}

	// 收集排序后的模块列表
	for mod := range affectedModules {
		scope.AffectedModules = append(scope.AffectedModules, mod)
	}
	sort.Strings(scope.AffectedModules)

	// 3. 根据变更文件类型决定受影响的 Pass
	scope.AffectedPasses = e.determineAffectedPasses(changedFiles)

	// 4. 生成变更摘要
	scope.ChangeSummary = fmt.Sprintf("变更文件 %d 个，受影响模块 %d 个，需重跑 Pass %v",
		len(changedFiles), len(affectedModules), scope.AffectedPasses)

	e.log.Debug(context.Background(), "影响范围分析完成",
		slog.Int("changedFiles", len(changedFiles)),
		slog.Int("affectedModules", len(scope.AffectedModules)),
		slog.Any("affectedPasses", scope.AffectedPasses))

	return scope
}

// determineAffectedPasses 根据变更文件类型决定受影响的 Pass 编号
//
// 判断顺序（短路优先）：
//  1. 任一入口文件 → 直接返回 [1,2,3,4]
//  2. 按文件类型累加 Pass 到 set：
//     - 架构文件 → Pass 3
//     - 文档文件 → Pass 4
//     - 依赖清单 → Pass 1+2
//     - 普通源码 → Pass 2+3
//  3. set 为空时返回默认值 [2,3]
//
// 返回的切片始终升序且去重。
func (e *IncrementalEngine) determineAffectedPasses(changedFiles []string) []int {
	passes := make(map[int]bool)

	for _, file := range changedFiles {
		base := filepath.Base(file)

		// 入口文件 → 全部 Pass（立即返回）
		if isEntryPointFile(base) {
			return []int{1, 2, 3, 4}
		}

		// 架构文件 → Pass 3
		if isArchitectureFile(file) {
			passes[3] = true
		}

		// 文档文件 → Pass 4
		if isDocumentationFile(file) {
			passes[4] = true
		}

		// 依赖清单 → Pass 1+2
		if isDependencyFile(base) {
			passes[1] = true
			passes[2] = true
		}

		// 普通源码 → Pass 2+3
		if isSourceFile(file) {
			passes[2] = true
			passes[3] = true
		}
	}

	if len(passes) == 0 {
		return []int{2, 3} // 默认：模块+架构分析
	}
	return sortedIntKeys(passes)
}

// ──────────────────────────────────────────────────────────────────────
// ExecuteIncremental
// ──────────────────────────────────────────────────────────────────────

// ExecuteIncremental 执行增量更新
//
// 执行流程：
//  1. 拉取最新代码（PullLatest）
//  2. 重新扫描文件 + 依赖提取（基于最新代码）
//  3. 加载上一版本全部 4 个 Pass 结果作为基线
//  4. 仅执行受影响的 Pass（按 1→4 顺序，前序结果可来自重跑或复用）
//  5. 将未变化的 Pass 结果复制到新版本 passes 目录（保持完整）
//  6. 文档组装（始终执行，刷新元数据）
//  7. 更新版本状态为 completed
//
// 错误处理：
//   - 任意步骤失败 → 标记版本为 failed 并返回错误
//   - 重跑 Pass 失败 → 同上
//
// 非阻塞约定：
//   - 本方法应在后台 goroutine 中调用（与 AnalysisPipeline.Execute 一致）
//   - 调用方负责信号量获取/释放和 context 生命周期
//
// 参数说明:
//   - ctx:         上下文（建议使用 context.Background()）
//   - config:      RepoWiki 配置实体
//   - newVersion:  新创建的版本记录（status=pending，commit_hash 待填充）
//   - oldVersion:  上一版本记录（用于加载 Pass 基线和 commit hash）
//   - impactScope: 影响范围分析结果（来自 AnalyzeImpactScope）
func (e *IncrementalEngine) ExecuteIncremental(
	ctx context.Context,
	config *entity.RepoWikiConfig,
	newVersion *entity.WikiVersion,
	oldVersion *entity.WikiVersion,
	impactScope *ImpactScope,
) *xError.Error {
	startedAt := time.Now()
	newVersionID := newVersion.ID.Int64()

	e.log.Info(ctx, "增量更新启动",
		slog.Int64("newVersionID", newVersionID),
		slog.Int64("oldVersionID", oldVersion.ID.Int64()),
		slog.Any("affectedPasses", impactScope.AffectedPasses))

	// 辅助函数：更新版本和配置状态（独立 context 防止取消影响 DB 写入）
	updateStatus := func(status, stage, errMsg string) {
		bgCtx := context.Background()

		newVersion.Status = status
		if stage != "" {
			newVersion.CurrentStage = stage
		}
		if errMsg != "" {
			newVersion.ErrorMsg = errMsg
		}
		if status == bConst.RepoWikiStatusCompleted {
			now := time.Now()
			newVersion.CompletedAt = &now
			newVersion.DurationMs = int(now.Sub(startedAt).Milliseconds())
		}
		if status == bConst.RepoWikiStatusScanning && newVersion.StartedAt == nil {
			now := time.Now()
			newVersion.StartedAt = &now
		}

		if uErr := e.logic.repo.version.Update(bgCtx, newVersion); uErr != nil {
			e.log.Error(bgCtx, "更新版本状态失败",
				slog.Int64("versionID", newVersionID),
				slog.String("targetStatus", status),
				slog.String("err", uErr.Error()))
		}

		// 同步配置的冗余 status（列表页快速展示）
		config.Status = status
		if uErr := e.logic.repo.config.Update(bgCtx, config); uErr != nil {
			e.log.Error(bgCtx, "更新配置状态失败",
				slog.Int64("configID", config.ID.Int64()),
				slog.String("err", uErr.Error()))
		}
	}

	// ════════════════════════════════════════════════════════════════
	// Step 1: 拉取最新代码
	// ════════════════════════════════════════════════════════════════
	updateStatus(bConst.RepoWikiStatusScanning, "", "")

	repoPath := e.storage.GetRepoPath(config.ID.Int64())
	branch := config.DefaultBranch
	if branch == "" {
		branch = newVersion.Branch
	}

	e.log.Info(ctx, "Step 1/6: 拉取最新代码",
		slog.Int64("versionID", newVersionID),
		slog.String("repoPath", repoPath),
		slog.String("branch", branch))

	if err := e.gitService.PullLatest(repoPath, branch); err != nil {
		e.log.Error(ctx, "增量拉取失败",
			slog.Int64("versionID", newVersionID),
			slog.String("err", err.Error()))
		updateStatus(bConst.RepoWikiStatusFailed, "", "增量拉取失败: "+err.Error())
		return xError.NewError(ctx, xError.ServerInternalError,
			xError.ErrMessage("增量拉取失败: "+err.Error()), false, err)
	}

	// 获取最新 commit hash（覆盖 newVersion.CommitHash）
	if commitHash, hashErr := e.gitService.GetCommitHash(repoPath); hashErr != nil {
		e.log.Warn(ctx, "获取最新 commit hash 失败，使用 'unknown'",
			slog.Int64("versionID", newVersionID),
			slog.String("err", hashErr.Error()))
		newVersion.CommitHash = "unknown"
	} else {
		newVersion.CommitHash = commitHash
	}

	// Context 取消检查
	if ctx.Err() != nil {
		updateStatus(bConst.RepoWikiStatusCancelled, "", ctx.Err().Error())
		return xError.NewError(ctx, xError.ServerInternalError,
			xError.ErrMessage("增量更新任务已取消"), false, ctx.Err())
	}

	// ════════════════════════════════════════════════════════════════
	// Step 2: 重新扫描文件 + 依赖提取
	// ════════════════════════════════════════════════════════════════
	updateStatus(bConst.RepoWikiStatusScanning, bConst.RepoWikiStageScan, "")

	e.log.Info(ctx, "Step 2/6: 重新扫描文件",
		slog.Int64("versionID", newVersionID))

	fileScan, xErr := e.logic.svc.scanner.Scan(ctx, repoPath)
	if xErr != nil {
		e.log.Error(ctx, "文件扫描失败",
			slog.Int64("versionID", newVersionID),
			slog.String("err", xErr.Error()))
		updateStatus(bConst.RepoWikiStatusFailed, "", xErr.Error())
		return xErr
	}
	newVersion.FileCount = fileScan.TotalFiles

	// 持久化 file_scan.json
	fileScanPath := e.storage.GetFileScanPath(newVersionID)
	if writeErr := e.storage.WriteJSON(fileScanPath, fileScan); writeErr != nil {
		e.log.Warn(ctx, "写入 file_scan.json 失败（不阻塞流程）",
			slog.Int64("versionID", newVersionID),
			slog.String("err", writeErr.Error()))
	} else {
		newVersion.FileScanPath = fileScanPath
	}

	// 依赖提取
	updateStatus(bConst.RepoWikiStatusScanning, bConst.RepoWikiStageDepExtract, "")

	depSummary, err := e.logic.svc.extractor.Extract(fileScan, repoPath)
	if err != nil {
		e.log.Error(ctx, "依赖提取失败",
			slog.Int64("versionID", newVersionID),
			slog.String("err", err.Error()))
		updateStatus(bConst.RepoWikiStatusFailed, "", "依赖提取失败: "+err.Error())
		return xError.NewError(ctx, xError.ServerInternalError,
			xError.ErrMessage("依赖提取失败: "+err.Error()), false, err)
	}

	// 持久化 dep_summary.json
	depSummaryPath := e.storage.GetDepSummaryPath(newVersionID)
	if writeErr := e.storage.WriteJSON(depSummaryPath, depSummary); writeErr != nil {
		e.log.Warn(ctx, "写入 dep_summary.json 失败（不阻塞流程）",
			slog.Int64("versionID", newVersionID),
			slog.String("err", writeErr.Error()))
	} else {
		newVersion.DepSummaryPath = depSummaryPath
	}

	if ctx.Err() != nil {
		updateStatus(bConst.RepoWikiStatusCancelled, "", ctx.Err().Error())
		return xError.NewError(ctx, xError.ServerInternalError,
			xError.ErrMessage("增量更新任务已取消"), false, ctx.Err())
	}

	// ════════════════════════════════════════════════════════════════
	// Step 3: 加载上一版本 Pass 结果作为基线
	// ════════════════════════════════════════════════════════════════
	e.log.Info(ctx, "Step 3/6: 加载上一版本 Pass 基线",
		slog.Int64("versionID", newVersionID),
		slog.Int64("oldVersionID", oldVersion.ID.Int64()))

	oldPassResults := e.loadOldPassResults(oldVersion.ID.Int64())
	e.log.Info(ctx, "上一版本 Pass 基线加载完成",
		slog.Int64("versionID", newVersionID),
		slog.Int("loadedPasses", len(oldPassResults)))

	// ════════════════════════════════════════════════════════════════
	// Step 4: 仅执行受影响的 Pass（按 1→4 顺序）
	// ════════════════════════════════════════════════════════════════
	updateStatus(bConst.RepoWikiStatusAnalyzing, "", "")

	if e.passRunner == nil {
		e.log.Error(ctx, "AgentPassRunner 未注入，无法执行增量 Pass 重跑",
			slog.Int64("versionID", newVersionID))
		updateStatus(bConst.RepoWikiStatusFailed, "", "AgentPassRunner 未注入")
		return xError.NewError(ctx, xError.ServerInternalError,
			xError.ErrMessage("AgentPassRunner 未注入，无法执行增量 Pass 重跑"), false, nil)
	}

	// 预构建上下文 JSON
	var fileScanJSON, depSummaryJSON []byte
	if fileScan != nil {
		fileScanJSON, _ = json.MarshalIndent(fileScan, "", "  ")
	}
	if depSummary != nil {
		depSummaryJSON, _ = json.MarshalIndent(depSummary, "", "  ")
	}

	// passResults 初始化为上一版本结果的副本（未受影响的 Pass 保留）
	passResults := make(map[string]*PassResult, 4)
	maps.Copy(passResults, oldPassResults)

	affectedSet := make(map[int]bool, len(impactScope.AffectedPasses))
	for _, p := range impactScope.AffectedPasses {
		affectedSet[p] = true
	}

	e.log.Info(ctx, "Step 4/6: 按需重跑受影响的 Pass",
		slog.Int64("versionID", newVersionID),
		slog.Any("affectedPasses", impactScope.AffectedPasses))

	for passNum := 1; passNum <= 4; passNum++ {
		if !affectedSet[passNum] {
			continue
		}

		// 更新 current_stage（status 保持 analyzing）
		updateStatus(bConst.RepoWikiStatusAnalyzing, allPasses[passNum-1].Stage, "")

		// 构建 userInput（使用 passResults 中的最新前序结果）
		userInput, bErr := buildUserInputForPass(passNum, passResults, repoPath, fileScanJSON, depSummaryJSON)
		if bErr != nil {
			e.log.Error(ctx, "构建 Pass userInput 失败",
				slog.Int64("versionID", newVersionID),
				slog.Int("passNum", passNum),
				slog.String("err", bErr.Error()))
			updateStatus(bConst.RepoWikiStatusFailed, "", fmt.Sprintf("Pass %d userInput 构建失败: %s", passNum, bErr.Error()))
			return xError.NewError(ctx, xError.ServerInternalError,
				xError.ErrMessage(fmt.Sprintf("Pass %d userInput 构建失败: %s", passNum, bErr.Error())), false, nil)
		}

		// 执行单个 Pass（结果自动写入新版本 passes 目录）
		result := e.passRunner.runSinglePass(ctx, passNum, userInput, newVersionID, repoPath)
		passResults[fmt.Sprintf("pass%d", passNum)] = result

		if !result.Success {
			e.log.Error(ctx, "增量 Pass 重跑失败",
				slog.Int64("versionID", newVersionID),
				slog.Int("passNum", passNum),
				slog.String("error", result.Error))
			updateStatus(bConst.RepoWikiStatusFailed, "", fmt.Sprintf("Pass %d 重跑失败: %s", passNum, result.Error))
			return xError.NewError(ctx, xError.ServerInternalError,
				xError.ErrMessage(fmt.Sprintf("Pass %d 增量重跑失败: %s", passNum, result.Error)), false, nil)
		}

		e.log.Info(ctx, "增量 Pass 重跑成功",
			slog.Int64("versionID", newVersionID),
			slog.Int("passNum", passNum),
			slog.Int64("durationMs", result.DurationMs),
			slog.Int64("tokens", result.TokenCount))
	}

	// 累计 Token 消耗
	newVersion.TokenCount = totalTokens(passResults)

	if ctx.Err() != nil {
		updateStatus(bConst.RepoWikiStatusCancelled, "", ctx.Err().Error())
		return xError.NewError(ctx, xError.ServerInternalError,
			xError.ErrMessage("增量更新任务已取消"), false, ctx.Err())
	}

	// ════════════════════════════════════════════════════════════════
	// Step 5: 复制未变化的 Pass 结果到新版本目录
	// ════════════════════════════════════════════════════════════════
	e.log.Info(ctx, "Step 5/6: 复制未变化的 Pass 结果",
		slog.Int64("versionID", newVersionID))

	newPassesPath := e.storage.GetPassesPath(newVersionID)
	for passNum := 1; passNum <= 4; passNum++ {
		// 受影响的 Pass 已由 runSinglePass 写入
		if affectedSet[passNum] {
			continue
		}
		key := fmt.Sprintf("pass%d", passNum)
		result, ok := passResults[key]
		if !ok {
			e.log.Warn(ctx, "上一版本 Pass 结果缺失，跳过复制",
				slog.Int64("versionID", newVersionID),
				slog.Int("passNum", passNum))
			continue
		}
		passPath := filepath.Join(newPassesPath, fmt.Sprintf("pass-%d.json", passNum))
		if writeErr := e.storage.WriteJSON(passPath, result); writeErr != nil {
			e.log.Warn(ctx, "复制 Pass 结果失败（不阻塞流程）",
				slog.Int64("versionID", newVersionID),
				slog.Int("passNum", passNum),
				slog.String("err", writeErr.Error()))
		}
	}

	// ════════════════════════════════════════════════════════════════
	// Step 6: 文档组装（始终执行）
	// ════════════════════════════════════════════════════════════════
	updateStatus(bConst.RepoWikiStatusAssembling, bConst.RepoWikiStageAssemble, "")

	e.log.Info(ctx, "Step 6/6: 文档组装",
		slog.Int64("versionID", newVersionID))

	if e.assembler != nil {
		if aErr := e.assembler.Assemble(
			ctx,
			passResults,
			fileScan,
			depSummary,
			config.ID.Int64(),
			newVersion.Language,
		); aErr != nil {
			e.log.Error(ctx, "文档组装失败",
				slog.Int64("versionID", newVersionID),
				slog.String("err", aErr.Error()))
			updateStatus(bConst.RepoWikiStatusFailed, "", aErr.Error())
			return aErr
		}
		e.log.Info(ctx, "文档组装完成",
			slog.Int64("versionID", newVersionID),
			slog.Int64("configID", config.ID.Int64()))
	} else {
		e.log.Error(ctx, "DocumentAssembler 未注入，无法生成 Wiki 文档",
			slog.Int64("versionID", newVersionID))
		err := xError.NewError(ctx, xError.ServerInternalError,
			xError.ErrMessage("DocumentAssembler 未注入，Wiki 文档生成被跳过"), false, nil)
		updateStatus(bConst.RepoWikiStatusFailed, "", err.Error())
		return err
	}

	// 记录版本数据存储路径
	newVersion.StoragePath = e.storage.GetVersionPath(newVersionID)

	// ════════════════════════════════════════════════════════════════
	// 完成
	// ════════════════════════════════════════════════════════════════
	updateStatus(bConst.RepoWikiStatusCompleted, "", "")

	e.log.Info(ctx, "增量更新执行完成",
		slog.Int64("versionID", newVersionID),
		slog.Int("fileCount", newVersion.FileCount),
		slog.Int64("tokenCount", newVersion.TokenCount),
		slog.Int("durationMs", newVersion.DurationMs),
		slog.String("commitHash", newVersion.CommitHash),
		slog.Any("rerunPasses", impactScope.AffectedPasses))

	return nil
}

// ──────────────────────────────────────────────────────────────────────
// 内部辅助方法
// ──────────────────────────────────────────────────────────────────────

// loadOldPassResults 加载上一版本全部 4 个 Pass 的结果
//
// 从 {oldVersionPath}/passes/pass-{N}.json 读取，缺失或解析失败的 Pass 跳过（记录警告）。
// 返回的 map key 为 "pass1"/"pass2"/"pass3"/"pass4"。
func (e *IncrementalEngine) loadOldPassResults(oldVersionID int64) map[string]*PassResult {
	results := make(map[string]*PassResult, 4)
	passesPath := e.storage.GetPassesPath(oldVersionID)

	for passNum := 1; passNum <= 4; passNum++ {
		passPath := filepath.Join(passesPath, fmt.Sprintf("pass-%d.json", passNum))
		var result PassResult
		if xErr := e.storage.ReadJSON(passPath, &result); xErr != nil {
			e.log.Warn(context.Background(), "读取上一版本 Pass 结果失败，跳过",
				slog.Int64("oldVersionID", oldVersionID),
				slog.Int("passNum", passNum),
				slog.String("path", passPath),
				slog.String("err", xErr.Error()))
			continue
		}
		results[fmt.Sprintf("pass%d", passNum)] = &result
	}

	return results
}

// buildUserInputForPass 根据 Pass 编号构建对应的 userInput
//
// 前序 Pass 结果从 results map 中获取（可能来自上一版本复用或本次重跑）。
// 任一必需的前序结果缺失或失败 → 返回错误。
func buildUserInputForPass(
	passNum int,
	results map[string]*PassResult,
	repoPath string,
	fileScanJSON []byte,
	depSummaryJSON []byte,
) (string, error) {
	switch passNum {
	case 1:
		return buildPass1UserInput(repoPath, fileScanJSON), nil

	case 2:
		pass1 := results["pass1"]
		if pass1 == nil || !pass1.Success {
			return "", fmt.Errorf("Pass 1 结果不可用（缺失或失败），无法构建 Pass 2 输入")
		}
		return buildPass2UserInput(pass1, depSummaryJSON), nil

	case 3:
		pass1 := results["pass1"]
		pass2 := results["pass2"]
		if pass1 == nil || !pass1.Success {
			return "", fmt.Errorf("Pass 1 结果不可用（缺失或失败），无法构建 Pass 3 输入")
		}
		if pass2 == nil || !pass2.Success {
			return "", fmt.Errorf("Pass 2 结果不可用（缺失或失败），无法构建 Pass 3 输入")
		}
		return buildPass3UserInput(pass1, pass2), nil

	case 4:
		pass1 := results["pass1"]
		pass2 := results["pass2"]
		pass3 := results["pass3"]
		if pass1 == nil || !pass1.Success {
			return "", fmt.Errorf("Pass 1 结果不可用（缺失或失败），无法构建 Pass 4 输入")
		}
		if pass2 == nil || !pass2.Success {
			return "", fmt.Errorf("Pass 2 结果不可用（缺失或失败），无法构建 Pass 4 输入")
		}
		if pass3 == nil || !pass3.Success {
			return "", fmt.Errorf("Pass 3 结果不可用（缺失或失败），无法构建 Pass 4 输入")
		}
		return buildPass4UserInput(pass1, pass2, pass3), nil

	default:
		return "", fmt.Errorf("无效的 Pass 编号: %d", passNum)
	}
}

// sortedIntKeys 返回 map[int]bool 中 true 值对应的 key 升序切片
func sortedIntKeys(m map[int]bool) []int {
	keys := make([]int, 0, len(m))
	for k, v := range m {
		if v {
			keys = append(keys, k)
		}
	}
	sort.Ints(keys)
	return keys
}

// ──────────────────────────────────────────────────────────────────────
// 文件类型判断辅助函数
// ──────────────────────────────────────────────────────────────────────

// entryPointFileNames 入口文件名集合（与 file_scanner.go 的 entryPointFiles 保持一致）
//
// 入口文件变更意味着项目结构可能发生根本性变化，需重跑全部 Pass。
var entryPointFileNames = map[string]bool{
	"main.go": true, "main.py": true, "main.rs": true, "main.java": true,
	"index.ts": true, "index.js": true, "index.tsx": true, "index.jsx": true,
	"app.py": true, "app.js": true,
	"Main.java": true,
	"go.mod":    true, "package.json": true, "Cargo.toml": true,
	"requirements.txt": true, "Gemfile": true, "pom.xml": true,
	"build.gradle": true, "Makefile": true, "CMakeLists.txt": true,
}

// architecturePathKeywords 架构相关路径关键词（小写匹配）
//
// 命中任一关键词的文件视为架构文件，影响 Pass 3（架构分析）。
var architecturePathKeywords = []string{
	"config", "middleware", "route", "router", "startup", "bootstrap",
	"app/", "/app", "server", "handler", "interceptor", "filter",
}

// dependencyManifestFiles 依赖清单文件名集合
//
// 依赖清单变更意味着技术栈或模块边界可能变化，影响 Pass 1+2。
var dependencyManifestFiles = map[string]bool{
	"go.mod": true, "go.sum": true,
	"package.json": true, "package-lock.json": true, "yarn.lock": true, "pnpm-lock.yaml": true,
	"Cargo.toml": true, "Cargo.lock": true,
	"requirements.txt": true, "Pipfile": true, "Pipfile.lock": true, "pyproject.toml": true,
	"Gemfile": true, "Gemfile.lock": true,
	"pom.xml": true, "build.gradle": true, "build.gradle.kts": true,
	"composer.json": true, "composer.lock": true,
}

// sourceFileExtensions 源码文件扩展名集合（小写，含点号）
var sourceFileExtensions = map[string]bool{
	".go": true, ".py": true, ".ts": true, ".tsx": true, ".js": true, ".jsx": true,
	".java": true, ".kt": true, ".scala": true, ".rs": true,
	".c": true, ".h": true, ".cpp": true, ".cc": true, ".cxx": true, ".hpp": true,
	".rb": true, ".php": true, ".swift": true, ".cs": true,
	".sh": true, ".sql": true,
}

// isEntryPointFile 判断文件名是否为入口文件
func isEntryPointFile(name string) bool {
	return entryPointFileNames[name]
}

// isArchitectureFile 判断文件路径是否为架构相关文件
//
// 路径中包含 config / middleware / route / startup 等关键词的文件视为架构文件。
// 匹配不区分大小写。
func isArchitectureFile(path string) bool {
	lower := strings.ToLower(filepath.ToSlash(path))
	for _, kw := range architecturePathKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// isDocumentationFile 判断文件是否为文档文件
//
// .md 扩展名或路径包含 docs/ 目录的文件视为文档文件。
func isDocumentationFile(path string) bool {
	slashPath := filepath.ToSlash(path)
	if strings.HasPrefix(filepath.Ext(slashPath), ".md") {
		return true
	}
	lower := strings.ToLower(slashPath)
	return strings.Contains(lower, "docs/") || strings.Contains(lower, "doc/") ||
		strings.Contains(lower, "documentation/")
}

// isDependencyFile 判断文件名是否为依赖清单文件
func isDependencyFile(name string) bool {
	return dependencyManifestFiles[name]
}

// isSourceFile 判断文件是否为源码文件
func isSourceFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return sourceFileExtensions[ext]
}
