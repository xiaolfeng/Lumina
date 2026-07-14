// Package logic RepoWiki 子 Agent 编排引擎（SubAgentOrchestrator）。
//
// SubAgentOrchestrator 实现 RepoWiki Wiki 生成的 5 阶段流水线：
//  1. 概要分析（Overview）—— Coordinator Agent 产出项目整体概览 Markdown
//  2. 代码探索（Explore）—— 多个 Explore Agent 并发分析各模块/目录，产出 XML 骨架
//  3. 架构规划（Architect）—— Architect Agent 综合概要与探索产出，构建 Wiki 目录大纲 JSON
//  4. 文档撰写（Writer）—— 多个 Writer Agent 分批并发撰写 Wiki 页面 Markdown
//  5. 文档校验（Validator）—— Validator Agent 校验完整性，失败时由 Execute 重驱动 Writer（最多 2 次）
//
// Orchestrator 只负责 LLM 编排与中间文件读写，不持有 db/rdb 引用；
// 数据持久化（WikiVersion 状态机）由上层 Pipeline 处理。
package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/bamboo-services/bamboo-agent/agent"
	"github.com/bamboo-services/bamboo-agent/tool"
	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	"github.com/bamboo-services/bamboo-messages/bamboo"

	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/service"
)

// ──────────────────────────────────────────────────────────────────────
// 内部类型
// ──────────────────────────────────────────────────────────────────────

// exploreOutcome 单路 Explore 分析的执行结果
//
// success 时 filePath 非空；failure 时 err 包含失败原因。
type exploreOutcome struct {
	scope    string // 分析范围
	filePath string // 产出文件路径（成功时）
	err      error  // 失败原因（失败时）
}

// ──────────────────────────────────────────────────────────────────────
// 常量
// ──────────────────────────────────────────────────────────────────────

const (
	// exploreMaxScopes 单次 Explore 阶段的最大分析范围数（超过则截断）
	exploreMaxScopes = 8
	// exploreFailureThresholdPct Explore 阶段失败比例阈值（超过则整体失败），50 表示 50%
	exploreFailureThresholdPct = 50
	// architectMaxParseRetries Architect JSON 解析失败的最大重试次数
	architectMaxParseRetries = 2
	// writerFileMinSize 判定 Writer 产出文件非空的最小字节数
	writerFileMinSize = 100
	// exploreOutputTruncateLen 截断 Explore 产出供 Architect 参考的最大字符数
	exploreOutputTruncateLen = 8000
)

// ──────────────────────────────────────────────────────────────────────
// SubAgentOrchestrator 结构体
// ──────────────────────────────────────────────────────────────────────

// SubAgentOrchestrator RepoWiki 子 Agent 编排引擎
//
// 持有 5 个角色的 LLM client 与模型配置，串联 5 个阶段生成 Wiki 文档。
// 每个阶段独立超时，中间产出持久化到文件系统，超时或失败时保留已产出文件（不回滚）。
//
// 字段说明:
//   - roleClients: 5 角色（coordinator/explore/architect/write/validator）的 LLM client
//   - roleModels:  5 角色的模型运行配置（复用 ModelRunConfig）
//   - storage:     Wiki 存储服务（路径管理与文件 I/O）
//   - log:         命名日志器
//   - versionID:   当前 Wiki 版本 ID（定位 versions/{vid}/ 下各子目录）
//   - repoPath:    克隆的仓库根目录（Agent 工具的作用域根）
type SubAgentOrchestrator struct {
	roleClients map[string]bamboo.BambooClient // 5 角色的 LLM client
	roleModels  map[string]*ModelRunConfig     // 5 角色的模型配置
	storage     *service.WikiStorageService    // Wiki 存储服务
	log         *xLog.LogNamedLogger           // 命名日志器
	versionID   int64                          // Wiki 版本 ID
	repoPath    string                         // 仓库根目录绝对路径
}

// NewSubAgentOrchestrator 创建 SubAgentOrchestrator 实例
//
// 参数说明:
//   - roleClients: 5 角色的 LLM client map（key 为 bConst.AgentRoleRepoWiki* 常量）
//   - roleModels:  5 角色的模型配置 map（key 与 roleClients 对齐）
//   - storage:     Wiki 存储服务
//   - log:         命名日志器（建议 xLog.WithName(xLog.NamedLOGC, "SubAgentOrchestrator")）
//   - versionID:   Wiki 版本 ID
//   - repoPath:    仓库根目录绝对路径
func NewSubAgentOrchestrator(
	roleClients map[string]bamboo.BambooClient,
	roleModels map[string]*ModelRunConfig,
	storage *service.WikiStorageService,
	log *xLog.LogNamedLogger,
	versionID int64,
	repoPath string,
) *SubAgentOrchestrator {
	return &SubAgentOrchestrator{
		roleClients: roleClients,
		roleModels:  roleModels,
		storage:     storage,
		log:         log,
		versionID:   versionID,
		repoPath:    repoPath,
	}
}

// ──────────────────────────────────────────────────────────────────────
// buildRoleAgent / buildAgentSession（私有 Agent 构造）
// ──────────────────────────────────────────────────────────────────────

// buildRoleAgent 按角色构建 Agent，使用角色名作为 session 子目录
//
// 适用于单实例阶段（Overview / Architect / Validator）。
// 并发阶段（Explore / Writer）请使用 buildAgentSession 传入唯一 sessionID。
func (o *SubAgentOrchestrator) buildRoleAgent(role string, systemPrompt string, tools []tool.Tool) (agent.Agent, error) {
	return o.buildAgentSession(role, role, systemPrompt, tools)
}

// buildAgentSession 按角色构建 Agent，使用指定 sessionID 作为 session 子目录
//
// 参数说明:
//   - role:       角色常量（用于查 roleClients/roleModels）
//   - sessionID:  session 子目录标识（并发阶段需唯一，如 "explore-internal"、"writer-batch0-0"）
//   - systemPrompt: 系统提示词
//   - tools:      Agent 可用工具集
func (o *SubAgentOrchestrator) buildAgentSession(role, sessionID, systemPrompt string, tools []tool.Tool) (agent.Agent, error) {
	client, ok := o.roleClients[role]
	if !ok || client == nil {
		return nil, fmt.Errorf("未配置角色 %s 的 LLM client", role)
	}
	mc, ok := o.roleModels[role]
	if !ok || mc == nil {
		return nil, fmt.Errorf("未配置角色 %s 的模型运行配置", role)
	}
	sessionDir := filepath.Join(o.storage.GetSessionPath(o.versionID), sessionID)
	return service.NewRepoWikiSubAgent(
		client, role,
		mc.ModelName, mc.MaxTokens, mc.ContextWindow, mc.Temperature,
		systemPrompt, tools, sessionDir,
	)
}

// ──────────────────────────────────────────────────────────────────────
// 阶段 1: runOverview（Coordinator 概要分析）
// ──────────────────────────────────────────────────────────────────────

// runOverview 执行概要分析阶段
//
// 构建 Coordinator Agent（工具：file_read + file_search + list_dir，作用域 o.repoPath），
// 执行 BuildOverviewUserPrompt，将产出写入 versions/{vid}/overview.md，返回概要文本。
func (o *SubAgentOrchestrator) runOverview(ctx context.Context) (string, *xError.Error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(bConst.RepoWikiOverviewTimeoutMin)*time.Minute)
	defer cancel()

	tools := []tool.Tool{
		service.NewFileReadTool(o.repoPath),
		service.NewFileSearchTool(o.repoPath),
		service.NewListDirTool(o.repoPath),
	}
	ag, err := o.buildRoleAgent(bConst.AgentRoleRepoWikiCoordinator, CoordinatorSystemPrompt, tools)
	if err != nil {
		return "", xError.NewError(ctx, xError.ServerInternalError,
			xError.ErrMessage("构建 Coordinator Agent 失败: "+err.Error()), false, err)
	}

	o.log.Info(ctx, "概要分析阶段开始",
		slog.Int64("version_id", o.versionID))
	start := time.Now()

	userInput := BuildOverviewUserPrompt(o.repoPath)
	result, runErr := ag.Run(ctx, userInput)
	o.saveSessionArtifact(ctx, bConst.AgentRoleRepoWikiCoordinator, bConst.AgentRoleRepoWikiCoordinator, userInput, result, runErr, start, 0)
	if runErr != nil {
		if ctx.Err() != nil {
			return "", xError.NewError(ctx, xError.ServerInternalError,
				xError.ErrMessage("概要分析阶段超时或被取消"), false, ctx.Err())
		}
		return "", xError.NewError(ctx, xError.ServerInternalError,
			xError.ErrMessage("Coordinator Agent 执行失败: "+runErr.Error()), false, runErr)
	}

	overview := strings.TrimSpace(result.Content)
	if overview == "" {
		return "", xError.NewError(ctx, xError.ServerInternalError,
			xError.ErrMessage("Coordinator Agent 输出为空"), false, nil)
	}

	// 持久化概要产出
	overviewPath := filepath.Join(o.storage.GetVersionPath(o.versionID), "overview.md")
	if writeErr := o.storage.WriteMarkdown(overviewPath, overview); writeErr != nil {
		o.log.Warn(ctx, "写入 overview.md 失败（继续流程）",
			slog.String("err", writeErr.Error()))
	}

	o.log.Info(ctx, "概要分析阶段完成",
		slog.Int64("version_id", o.versionID),
		slog.Int64("duration_ms", time.Since(start).Milliseconds()),
		slog.Int64("tokens", result.Usage.InputTokens+result.Usage.OutputTokens),
		slog.Int("content_len", len(overview)))

	return overview, nil
}

// ──────────────────────────────────────────────────────────────────────
// 阶段 2: runExploreConcurrent（Explore 并发探索）
// ──────────────────────────────────────────────────────────────────────

// runExploreConcurrent 执行代码探索阶段
//
// 从 overviewSummary 文本中正则提取顶层目录作为分析 scope（不足时回退到仓库文件系统），
// 以 RepoWikiExploreMaxConcurrent 并发驱动 Explore Agent，每路独立超时。
// 单 scope 失败跳过继续；超 50% scope 失败则整体返回错误。
//
// 返回 map[scope]filePath（scope → versions/{vid}/explore/{sanitized}.xml）。
func (o *SubAgentOrchestrator) runExploreConcurrent(ctx context.Context, overviewSummary string) (map[string]string, *xError.Error) {
	// 解析分析 scope（Go 代码正则解析，非 LLM 决策）
	scopes := extractExploreScopes(overviewSummary)
	if len(scopes) == 0 {
		scopes = listTopLevelDirs(o.repoPath)
	}
	if len(scopes) == 0 {
		return nil, xError.NewError(ctx, xError.ServerInternalError,
			xError.ErrMessage("未能从概要或仓库中提取任何分析 scope"), false, nil)
	}
	if len(scopes) > exploreMaxScopes {
		scopes = scopes[:exploreMaxScopes]
	}

	o.log.Info(ctx, "代码探索阶段开始",
		slog.Int64("version_id", o.versionID),
		slog.Int("scope_count", len(scopes)),
		slog.Any("scopes", scopes))

	// 并发探索（信号量限流）
	outcomes := make([]exploreOutcome, len(scopes))
	sem := make(chan struct{}, bConst.RepoWikiExploreMaxConcurrent)
	var wg sync.WaitGroup
	for i, scope := range scopes {
		wg.Add(1)
		go func(idx int, s string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			outcomes[idx] = o.runSingleExplore(ctx, s)
		}(i, scope)
	}
	wg.Wait()

	// 汇总结果
	outputs := make(map[string]string, len(scopes))
	failures := 0
	for _, oc := range outcomes {
		if oc.err != nil {
			failures++
			o.log.Warn(ctx, "单个 Explore scope 失败（跳过）",
				slog.String("scope", oc.scope),
				slog.String("err", oc.err.Error()))
			continue
		}
		outputs[oc.scope] = oc.filePath
	}

	// 失败比例检查
	if len(scopes) > 0 {
		failurePct := failures * 100 / len(scopes)
		if failurePct > exploreFailureThresholdPct {
			return nil, xError.NewError(ctx, xError.ServerInternalError,
				xError.ErrMessage(fmt.Sprintf("Explore 阶段失败比例过高: %d/%d (%d%%)",
					failures, len(scopes), failurePct)), false, nil)
		}
	}

	o.log.Info(ctx, "代码探索阶段完成",
		slog.Int64("version_id", o.versionID),
		slog.Int("success", len(outputs)),
		slog.Int("failure", failures))
	return outputs, nil
}

// runSingleExplore 执行单路 Explore 分析
//
// 每路独立超时（RepoWikiExploreTimeoutMin），构建 Explore Agent（工具：file_read + file_search），
// 产出写入 versions/{vid}/explore/{sanitized_scope}.xml。
func (o *SubAgentOrchestrator) runSingleExplore(ctx context.Context, scope string) exploreOutcome {
	start := time.Now()
	exploreCtx, cancel := context.WithTimeout(ctx, time.Duration(bConst.RepoWikiExploreTimeoutMin)*time.Minute)
	defer cancel()

	tools := []tool.Tool{
		service.NewFileReadTool(o.repoPath),
		service.NewFileSearchTool(o.repoPath),
	}
	sessionID := "explore-" + sanitizeScopeForFilename(scope)
	ag, err := o.buildAgentSession(bConst.AgentRoleRepoWikiExplore, sessionID, ExploreSystemPrompt, tools)
	if err != nil {
		return exploreOutcome{scope: scope, err: fmt.Errorf("构建 Explore Agent 失败: %w", err)}
	}

	userInput := BuildExploreUserPrompt(scope)
	result, runErr := ag.Run(exploreCtx, userInput)
	o.saveSessionArtifact(exploreCtx, bConst.AgentRoleRepoWikiExplore, sessionID, userInput, result, runErr, start, 0)
	if runErr != nil {
		if exploreCtx.Err() != nil {
			return exploreOutcome{scope: scope, err: fmt.Errorf("scope %s 探索超时或被取消: %w", scope, exploreCtx.Err())}
		}
		return exploreOutcome{scope: scope, err: fmt.Errorf("scope %s Explore Agent 执行失败: %w", scope, runErr)}
	}

	content := strings.TrimSpace(result.Content)
	if content == "" {
		return exploreOutcome{scope: scope, err: fmt.Errorf("scope %s Explore Agent 输出为空", scope)}
	}

	// 持久化 Explore 产出（XML 骨架文本）
	exploreDir := o.storage.GetExploreOutputsPath(o.versionID)
	fileName := sanitizeScopeForFilename(scope) + ".xml"
	filePath := filepath.Join(exploreDir, fileName)
	if writeErr := o.storage.WriteMarkdown(filePath, content); writeErr != nil {
		return exploreOutcome{scope: scope, err: fmt.Errorf("scope %s 写入 Explore 产出失败: %w", scope, writeErr)}
	}

	o.log.Debug(ctx, "Explore scope 完成",
		slog.String("scope", scope),
		slog.String("file", filePath),
		slog.Int64("duration_ms", time.Since(start).Milliseconds()),
		slog.Int64("tokens", result.Usage.InputTokens+result.Usage.OutputTokens))

	return exploreOutcome{scope: scope, filePath: filePath}
}

// ──────────────────────────────────────────────────────────────────────
// 阶段 3: runArchitect（架构规划）
// ──────────────────────────────────────────────────────────────────────

// runArchitect 执行架构规划阶段
//
// 构建 Architect Agent（工具：file_read），汇总概要与全部 Explore 产出构建 Wiki 目录大纲。
// JSON 解析失败时在 prompt 追加格式提醒重试（最多 architectMaxParseRetries 次）。
// 产出写入 versions/{vid}/architecture.json，返回 outline 与原始 JSON 文本。
func (o *SubAgentOrchestrator) runArchitect(ctx context.Context, overviewSummary string, exploreOutputs map[string]string) ([]WikiEntry, *xError.Error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(bConst.RepoWikiArchitectTimeoutMin)*time.Minute)
	defer cancel()

	// 读取 Explore 产出内容，构建 []ExploreOutput
	exploreList := o.loadExploreOutputs(exploreOutputs)

	tools := []tool.Tool{
		service.NewFileReadTool(o.repoPath),
	}
	ag, err := o.buildRoleAgent(bConst.AgentRoleRepoWikiArchitect, ArchitectSystemPrompt, tools)
	if err != nil {
		return nil, xError.NewError(ctx, xError.ServerInternalError,
			xError.ErrMessage("构建 Architect Agent 失败: "+err.Error()), false, err)
	}

	o.log.Info(ctx, "架构规划阶段开始",
		slog.Int64("version_id", o.versionID),
		slog.Int("explore_count", len(exploreList)))

	start := time.Now()
	userInput := BuildArchitectUserPrompt(overviewSummary, exploreList)
	var outline []WikiEntry

	for attempt := 0; attempt <= architectMaxParseRetries; attempt++ {
		if ctx.Err() != nil {
			return nil, xError.NewError(ctx, xError.ServerInternalError,
				xError.ErrMessage("架构规划阶段超时或被取消"), false, ctx.Err())
		}

		currentInput := userInput
		if attempt > 0 {
			currentInput = userInput + buildArchitectRetryHint(attempt)
		}

		result, runErr := ag.Run(ctx, currentInput)
		o.saveSessionArtifact(ctx, bConst.AgentRoleRepoWikiArchitect, bConst.AgentRoleRepoWikiArchitect, currentInput, result, runErr, start, attempt+1)
		if runErr != nil {
			if ctx.Err() != nil {
				return nil, xError.NewError(ctx, xError.ServerInternalError,
					xError.ErrMessage("架构规划阶段超时或被取消"), false, ctx.Err())
			}
			return nil, xError.NewError(ctx, xError.ServerInternalError,
				xError.ErrMessage(fmt.Sprintf("Architect Agent 执行失败 (第 %d 次): %s", attempt+1, runErr.Error())), false, runErr)
		}

		rawOutput := strings.TrimSpace(result.Content)
		parsed, parseErr := parseAgentJSON(rawOutput)
		if parseErr != nil {
			o.log.Warn(ctx, "Architect JSON 解析失败，准备重试",
				slog.Int("attempt", attempt+1),
				slog.String("err", parseErr.Error()))
			continue
		}

		if uErr := json.Unmarshal(parsed, &outline); uErr != nil {
			o.log.Warn(ctx, "Architect JSON 反序列化 WikiEntry 失败，准备重试",
				slog.Int("attempt", attempt+1),
				slog.String("err", uErr.Error()))
			continue
		}

		break
	}

	if outline == nil {
		return nil, xError.NewError(ctx, xError.ServerInternalError,
			xError.ErrMessage(fmt.Sprintf("Architect 输出解析失败，已重试 %d 次", architectMaxParseRetries)), false, nil)
	}

	// 持久化架构大纲
	archPath := o.storage.GetArchitecturePath(o.versionID)
	if writeErr := o.storage.WriteJSON(archPath, outline); writeErr != nil {
		o.log.Warn(ctx, "写入 architecture.json 失败（继续流程）",
			slog.String("err", writeErr.Error()))
	}

	o.log.Info(ctx, "架构规划阶段完成",
		slog.Int64("version_id", o.versionID),
		slog.Int("entry_count", len(outline)))
	return outline, nil
}

// loadExploreOutputs 读取所有 Explore 产出文件内容，构建 []ExploreOutput 供 Architect 使用
//
// 单个文件读取失败时跳过该 scope（不影响整体流程）。
// 内容超长时截断至 exploreOutputTruncateLen 字符。
func (o *SubAgentOrchestrator) loadExploreOutputs(exploreFiles map[string]string) []ExploreOutput {
	list := make([]ExploreOutput, 0, len(exploreFiles))
	for scope, file := range exploreFiles {
		data, err := os.ReadFile(file)
		if err != nil {
			o.log.Warn(nil, "读取 Explore 产出失败（跳过）",
				slog.String("scope", scope),
				slog.String("file", file),
				slog.String("err", err.Error()))
			continue
		}
		content := string(data)
		if len(content) > exploreOutputTruncateLen {
			content = truncate(content, exploreOutputTruncateLen)
		}
		list = append(list, ExploreOutput{
			Scope:    scope,
			FilePath: file,
			Content:  content,
		})
	}
	// 按 scope 排序，保证 prompt 顺序稳定
	sort.Slice(list, func(i, j int) bool {
		return list[i].Scope < list[j].Scope
	})
	return list
}

// ──────────────────────────────────────────────────────────────────────
// 阶段 4: runWritersConcurrent（Writer 并发撰写）
// ──────────────────────────────────────────────────────────────────────

// runWritersConcurrent 执行文档撰写阶段
//
// 按 Complexity 分组（high → 1 板块/Writer；medium/low → 2 板块/Writer），
// 分批以 RepoWikiWriterMaxConcurrent 并发，每批独立超时（RepoWikiWriterTimeoutMin）。
// 等待一批完成再发下一批。Writer 通过 save_wiki_page 工具写入 Wiki 目录。
//
// 缺失/空文件记录为警告（不硬性中断），由 Validator 阶段触发重试。
func (o *SubAgentOrchestrator) runWritersConcurrent(ctx context.Context, outline []WikiEntry, exploreOutputs map[string]string) *xError.Error {
	if len(outline) == 0 {
		return xError.NewError(ctx, xError.ServerInternalError,
			xError.ErrMessage("Writer 阶段 outline 为空，无法撰写"), false, nil)
	}

	// 按 Complexity 拆分 Writer 分配组
	groups := splitWriterGroups(outline)
	wikiDir := o.storage.GetWikiPath(o.versionID)

	o.log.Info(ctx, "文档撰写阶段开始",
		slog.Int64("version_id", o.versionID),
		slog.Int("entry_count", len(outline)),
		slog.Int("writer_group_count", len(groups)))

	// 分批并发，每批最多 RepoWikiWriterMaxConcurrent 个 Writer
	maxConcurrent := bConst.RepoWikiWriterMaxConcurrent
	if maxConcurrent <= 0 {
		maxConcurrent = 1
	}

	for batchStart := 0; batchStart < len(groups); batchStart += maxConcurrent {
		batchEnd := batchStart + maxConcurrent
		if batchEnd > len(groups) {
			batchEnd = len(groups)
		}
		batchIdx := batchStart / maxConcurrent
		batch := groups[batchStart:batchEnd]

		// 每批独立超时
		batchCtx, cancel := context.WithTimeout(ctx, time.Duration(bConst.RepoWikiWriterTimeoutMin)*time.Minute)

		var wg sync.WaitGroup
		for j, group := range batch {
			wg.Add(1)
			go func(workerIdx int, entries []WikiEntry) {
				defer wg.Done()
				o.runSingleWriter(batchCtx, batchIdx, workerIdx, entries, exploreOutputs, wikiDir)
			}(j, group)
		}
		wg.Wait()
		cancel()
	}

	// 校验所有 outline 条目对应文件是否存在且非空（警告级别，由 Validator 驱动重试）
	o.checkWriterOutputs(ctx, outline, wikiDir)

	o.log.Info(ctx, "文档撰写阶段完成",
		slog.Int64("version_id", o.versionID))
	return nil
}

// runSingleWriter 执行单个 Writer Agent 调用
//
// 参数说明:
//   - batchIdx:   批次索引（用于 session 目录命名）
//   - workerIdx:  批内 worker 索引
//   - entries:    本次 Writer 负责的 Wiki 条目（1-2 个）
//   - exploreOutputs: 全局 Explore 产出 map（scope → filePath）
//   - wikiDir:    Wiki 输出目录（save_wiki_page 工具的作用域）
func (o *SubAgentOrchestrator) runSingleWriter(
	ctx context.Context,
	batchIdx, workerIdx int,
	entries []WikiEntry,
	exploreOutputs map[string]string,
	wikiDir string,
) {
	start := time.Now()

	// 汇总本组条目引用的 Explore 产出内容
	relevantExplores := o.collectRelevantExplores(entries, exploreOutputs)

	tools := []tool.Tool{
		service.NewFileReadTool(o.repoPath),
		service.NewSaveWikiPageTool(wikiDir),
	}
	sessionID := fmt.Sprintf("writer-batch%d-%d", batchIdx, workerIdx)
	ag, err := o.buildAgentSession(bConst.AgentRoleRepoWikiWrite, sessionID, WriterSystemPrompt, tools)
	if err != nil {
		o.log.Error(ctx, "构建 Writer Agent 失败",
			slog.String("session", sessionID),
			slog.String("err", err.Error()))
		return
	}

	userInput := BuildWriterUserPrompt(entries, relevantExplores)
	result, runErr := ag.Run(ctx, userInput)
	o.saveSessionArtifact(ctx, bConst.AgentRoleRepoWikiWrite, sessionID, userInput, result, runErr, start, 0)
	if runErr != nil {
		if ctx.Err() != nil {
			o.log.Warn(ctx, "Writer 超时或被取消",
				slog.String("session", sessionID),
				slog.String("err", ctx.Err().Error()))
			return
		}
		o.log.Error(ctx, "Writer Agent 执行失败",
			slog.String("session", sessionID),
			slog.String("err", runErr.Error()))
		return
	}

	o.log.Debug(ctx, "Writer 完成",
		slog.String("session", sessionID),
		slog.Int("entry_count", len(entries)),
		slog.Int64("duration_ms", time.Since(start).Milliseconds()),
		slog.Int64("tokens", result.Usage.InputTokens+result.Usage.OutputTokens))
}

// checkWriterOutputs 检查 outline 条目对应的 Wiki 文件是否存在且非空
//
// 缺失或空文件记录为警告（不中断流程），由 Validator 阶段捕获并触发重试。
func (o *SubAgentOrchestrator) checkWriterOutputs(ctx context.Context, outline []WikiEntry, wikiDir string) {
	for _, entry := range outline {
		if entry.Path == "" {
			continue
		}
		fullPath := filepath.Join(wikiDir, entry.Path)
		info, err := os.Stat(fullPath)
		if err != nil {
			o.log.Warn(ctx, "Wiki 文件缺失，将由 Validator 捕获",
				slog.String("path", entry.Path),
				slog.String("title", entry.Title))
			continue
		}
		if info.Size() < writerFileMinSize {
			o.log.Warn(ctx, "Wiki 文件可能为空页面，将由 Validator 捕获",
				slog.String("path", entry.Path),
				slog.Int64("size", info.Size()))
		}
	}
}

// collectRelevantExplores 收集 entries 引用的 Explore 产出内容
//
// 匹配策略（多级 fallback，确保 Writer 不至于空手）:
//  1. 精确匹配：ref 与 scope 完全相等
//  2. 子串匹配：scope 包含 ref 或 ref 包含 scope
//  3. 去前缀匹配：剥离 ref 中的 "explore-" 前缀后做子串匹配（兼容 LLM 擅自加前缀的常见错误）
//  4. 最终兜底：若 entry 的所有 ref 都未命中，返回所有 Explore 产出（保证 Writer 有素材可用）
//
// 每个 entry 独立计算匹配；result 的 key 为 scope 原文。
func (o *SubAgentOrchestrator) collectRelevantExplores(entries []WikiEntry, exploreFiles map[string]string) map[string]string {
	result := make(map[string]string)
	// 预读所有 Explore 内容，避免重复 IO
	allContent := make(map[string]string, len(exploreFiles))
	for scope, file := range exploreFiles {
		if data, err := os.ReadFile(file); err == nil {
			allContent[scope] = string(data)
		}
	}

	for _, entry := range entries {
		matchedScopes := o.matchEntryExplores(entry.ExploreRefs, exploreFiles, allContent)
		if len(matchedScopes) == 0 && len(allContent) > 0 {
			// 最终兜底：所有 ref 未命中时给 Writer 全量素材
			o.log.Warn(nil, "Entry 的所有 ExploreRefs 均未命中，回退为全量 Explore 内容",
				slog.String("entry_title", entry.Title),
				slog.Any("explore_refs", entry.ExploreRefs))
			for scope, content := range allContent {
				result[scope] = content
			}
			continue
		}
		for _, scope := range matchedScopes {
			if content, ok := allContent[scope]; ok {
				result[scope] = content
			}
		}
	}
	return result
}

// matchEntryExplores 对单个 entry 的 ExploreRefs 执行多级匹配，返回命中的 scope 列表
func (o *SubAgentOrchestrator) matchEntryExplores(refs []string, exploreFiles map[string]string, allContent map[string]string) []string {
	matched := make(map[string]struct{})
	for _, rawRef := range refs {
		ref := strings.TrimSpace(rawRef)
		if ref == "" {
			continue
		}

		// 级别 1：精确匹配
		if _, ok := exploreFiles[ref]; ok {
			matched[ref] = struct{}{}
			continue
		}

		// 级别 2：子串匹配
		found := false
		for scope := range exploreFiles {
			if strings.Contains(scope, ref) || strings.Contains(ref, scope) {
				matched[scope] = struct{}{}
				found = true
				break
			}
		}
		if found {
			continue
		}

		// 级别 3：去 "explore-" 前缀后子串匹配（兼容 LLM 擅自加前缀）
		stripped := strings.TrimPrefix(ref, "explore-")
		if stripped != ref {
			for scope := range exploreFiles {
				if strings.Contains(scope, stripped) || strings.Contains(stripped, scope) {
					matched[scope] = struct{}{}
					found = true
					break
				}
			}
		}
		if found {
			continue
		}

		// 级别 4：单个 ref 全部未命中（不影响其他 ref 的匹配结果）
		o.log.Debug(nil, "Writer 引用的 Explore scope 未找到匹配",
			slog.String("ref", ref))
	}

	result := make([]string, 0, len(matched))
	for scope := range matched {
		result = append(result, scope)
	}
	sort.Strings(result)
	return result
}

// ──────────────────────────────────────────────────────────────────────
// 阶段 5: runValidator（文档校验）
// ──────────────────────────────────────────────────────────────────────

// validatorResult Validator Agent 输出的校验结果 JSON 结构
type validatorResult struct {
	Valid             bool             `json:"valid"`                // 校验是否通过
	Errors            []ValidationError `json:"errors"`              // 校验错误项
	MetadataGenerated bool             `json:"metadata_generated"`   // metadata.json 是否已生成
}

// runValidator 执行文档校验阶段
//
// 构建 Validator Agent（工具：file_read + save_wiki_page，作用域 wikiDir），
// 执行校验并解析 JSON 结果 {valid, errors, metadata_generated}。
func (o *SubAgentOrchestrator) runValidator(ctx context.Context, outline []WikiEntry, architectRawJSON string) (bool, []ValidationError, *xError.Error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(bConst.RepoWikiValidatorTimeoutMin)*time.Minute)
	defer cancel()

	wikiDir := o.storage.GetWikiPath(o.versionID)
	tools := []tool.Tool{
		service.NewFileReadTool(wikiDir),
		service.NewSaveWikiPageTool(wikiDir),
	}
	ag, err := o.buildRoleAgent(bConst.AgentRoleRepoWikiValidator, ValidatorSystemPrompt, tools)
	if err != nil {
		return false, nil, xError.NewError(ctx, xError.ServerInternalError,
			xError.ErrMessage("构建 Validator Agent 失败: "+err.Error()), false, err)
	}

	o.log.Info(ctx, "文档校验阶段开始",
		slog.Int64("version_id", o.versionID),
		slog.Int("entry_count", len(outline)))

	start := time.Now()
	userInput := BuildValidatorUserPrompt(wikiDir, architectRawJSON)
	result, runErr := ag.Run(ctx, userInput)
	o.saveSessionArtifact(ctx, bConst.AgentRoleRepoWikiValidator, bConst.AgentRoleRepoWikiValidator, userInput, result, runErr, start, 0)
	if runErr != nil {
		if ctx.Err() != nil {
			return false, nil, xError.NewError(ctx, xError.ServerInternalError,
				xError.ErrMessage("文档校验阶段超时或被取消"), false, ctx.Err())
		}
		return false, nil, xError.NewError(ctx, xError.ServerInternalError,
			xError.ErrMessage("Validator Agent 执行失败: "+runErr.Error()), false, runErr)
	}

	// 解析校验结果 JSON
	rawOutput := strings.TrimSpace(result.Content)
	parsed, parseErr := parseAgentJSON(rawOutput)
	if parseErr != nil {
		return false, nil, xError.NewError(ctx, xError.ServerInternalError,
			xError.ErrMessage("Validator 输出 JSON 解析失败: "+parseErr.Error()), false, parseErr)
	}

	var vr validatorResult
	if uErr := json.Unmarshal(parsed, &vr); uErr != nil {
		return false, nil, xError.NewError(ctx, xError.ServerInternalError,
			xError.ErrMessage("Validator 结果反序列化失败: "+uErr.Error()), false, uErr)
	}

	o.log.Info(ctx, "文档校验阶段完成",
		slog.Int64("version_id", o.versionID),
		slog.Bool("valid", vr.Valid),
		slog.Int("error_count", len(vr.Errors)),
		slog.Bool("metadata_generated", vr.MetadataGenerated))

	return vr.Valid, vr.Errors, nil
}

// ──────────────────────────────────────────────────────────────────────
// Execute 主入口
// ──────────────────────────────────────────────────────────────────────

// Execute 执行完整的 5 阶段 Wiki 生成流水线
//
// 阶段顺序：overview → explore → architect → writers → validator
// progressCallback 依次回调：scanning → exploring → architecting → writing → validating → completed
//
// 重试策略:
//   - Validator 返回 valid=false 时，从 errors 提取缺失 path → 找对应 WikiEntry → 重驱动 Writer
//   - 最多重试 RepoWikiWriterMaxRetry 次（默认 2 次）
//   - 超限标记失败并返回错误
//   - 任意阶段超时 → 立即返回错误（保留已产出中间文件，不回滚）
func (o *SubAgentOrchestrator) Execute(ctx context.Context, progressCallback func(stage string)) *xError.Error {
	notify := func(stage string) {
		if progressCallback != nil {
			progressCallback(stage)
		}
	}

	// ── 阶段 1: 概要分析 ──
	notify(bConst.RepoWikiStatusScanning)
	overview, err := o.runOverview(ctx)
	if err != nil {
		return err
	}

	// ── 阶段 2: 代码探索 ──
	notify(bConst.RepoWikiStageExploring)
	exploreOutputs, err := o.runExploreConcurrent(ctx, overview)
	if err != nil {
		return err
	}

	// ── 阶段 3: 架构规划 ──
	notify(bConst.RepoWikiStageArchitecting)
	outline, err := o.runArchitect(ctx, overview, exploreOutputs)
	if err != nil {
		return err
	}
	// 读取原始 architecture JSON 文本供 Validator 参考
	archRawJSON := o.readArchitectureRawJSON()

	// ── 阶段 4: 文档撰写 ──
	notify(bConst.RepoWikiStageWriting)
	if wErr := o.runWritersConcurrent(ctx, outline, exploreOutputs); wErr != nil {
		return wErr
	}

	// ── 阶段 5: 文档校验（含重试循环） ──
	notify(bConst.RepoWikiStageValidating)
	valid, errors, vErr := o.runValidator(ctx, outline, archRawJSON)
	if vErr != nil {
		return vErr
	}

	// Validator 失败重试循环
	for retry := 0; retry < bConst.RepoWikiWriterMaxRetry && !valid; retry++ {
		missingEntries := findMissingEntries(errors, outline)
		if len(missingEntries) == 0 {
			// 无法定位缺失条目，无法重驱动，直接终止
			o.log.Warn(ctx, "Validator 失败但无法定位缺失条目，终止重试",
				slog.Int("retry", retry),
				slog.Int("error_count", len(errors)))
			break
		}

		o.log.Info(ctx, "Validator 失败，重驱动 Writer 补写缺失板块",
			slog.Int("retry", retry+1),
			slog.Int("missing_count", len(missingEntries)))

		// 重驱动缺失条目的 Writer
		if wErr := o.runWritersConcurrent(ctx, missingEntries, exploreOutputs); wErr != nil {
			o.log.Warn(ctx, "重驱动 Writer 失败（继续重新校验）",
				slog.String("err", wErr.Error()))
		}

		// 重新校验
		valid, errors, vErr = o.runValidator(ctx, outline, archRawJSON)
		if vErr != nil {
			return vErr
		}
	}

	if !valid {
		errMsg := fmt.Sprintf("Wiki 校验未通过（已重试 %d 次），剩余 %d 个错误",
			bConst.RepoWikiWriterMaxRetry, len(errors))
		return xError.NewError(ctx, xError.ServerInternalError,
			xError.ErrMessage(errMsg), false, nil)
	}

	notify(bConst.RepoWikiStatusCompleted)
	o.log.Info(ctx, "Wiki 生成流水线全部完成",
		slog.Int64("version_id", o.versionID),
		slog.Int("entry_count", len(outline)))
	return nil
}

// readArchitectureRawJSON 读取 architecture.json 文件内容为字符串
//
// 用于 Execute 阶段向 Validator 传递原始目录大纲文本。
// 读取失败时返回空字符串（Validator 可在无大纲参考下进行基础校验）。
func (o *SubAgentOrchestrator) readArchitectureRawJSON() string {
	archPath := o.storage.GetArchitecturePath(o.versionID)
	data, err := os.ReadFile(archPath)
	if err != nil {
		o.log.Warn(nil, "读取 architecture.json 失败（Validator 将在无大纲参考下校验）",
			slog.String("err", err.Error()))
		return ""
	}
	return string(data)
}

// ──────────────────────────────────────────────────────────────────────
// 辅助函数（包级私有）
// ──────────────────────────────────────────────────────────────────────

// extractExploreScopes 从概要文本中正则提取分析 scope（顶层目录路径）
//
// 匹配策略（按优先级）:
//  1. 反引号包裹的路径（如 `internal/`、`web/src/`）
//  2. Markdown 列表项中的目录名（如 `- internal/`、`* cmd/`）
//
// 结果去重、过滤噪声（http/https/git 等协议前缀）、按字典序排序、截断至 exploreMaxScopes。
func extractExploreScopes(overviewSummary string) []string {
	set := make(map[string]bool)

	// 策略 1: 反引号包裹的路径 `xxx/yyy/`
	backtickRe := regexp.MustCompile("`([a-zA-Z0-9_.-]+(?:/[a-zA-Z0-9_.-]+)+)/?`")
	for _, m := range backtickRe.FindAllStringSubmatch(overviewSummary, -1) {
		scope := strings.Trim(m[1], "/")
		if scope != "" && !isNoiseScope(scope) {
			set[scope] = true
		}
	}

	// 策略 2: Markdown 列表项中的目录名
	listRe := regexp.MustCompile("(?m)^\\s*[-*]\\s+`?([a-zA-Z0-9_.-]+(?:/[a-zA-Z0-9_.-]+)?)/`?")
	for _, m := range listRe.FindAllStringSubmatch(overviewSummary, -1) {
		scope := strings.Trim(m[1], "/")
		if scope != "" && !isNoiseScope(scope) {
			set[scope] = true
		}
	}

	scopes := make([]string, 0, len(set))
	for s := range set {
		scopes = append(scopes, s)
	}
	sort.Strings(scopes)
	return scopes
}

// isNoiseScope 判断 scope 是否为噪声（协议前缀、常见非代码目录）
func isNoiseScope(scope string) bool {
	first := strings.SplitN(scope, "/", 2)[0]
	switch first {
	case "http", "https", "ftp", "git", "ssh":
		return true
	}
	return false
}

// listTopLevelDirs 列出仓库根目录下的顶层子目录作为 Explore scope 的回退方案
//
// 排除 .git、node_modules、vendor、dist、build 等非源码目录。
func listTopLevelDirs(repoPath string) []string {
	entries, err := os.ReadDir(repoPath)
	if err != nil {
		return nil
	}
	excluded := map[string]bool{
		".git": true, "node_modules": true, "vendor": true,
		"dist": true, "build": true, "target": true,
		".idea": true, ".vscode": true, "bin": true,
	}
	var dirs []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if excluded[e.Name()] {
			continue
		}
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		dirs = append(dirs, e.Name())
	}
	sort.Strings(dirs)
	return dirs
}

// sanitizeScopeForFilename 将 scope 转换为安全的文件名片段
//
// 用于 Explore 产出文件命名（如 "internal/log" → "internal_log"）。
func sanitizeScopeForFilename(scope string) string {
	s := strings.ReplaceAll(scope, "/", "_")
	reg := regexp.MustCompile("[^a-zA-Z0-9_.-]")
	s = reg.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-_.")
	if s == "" {
		s = "scope"
	}
	return s
}

// splitWriterGroups 按 Complexity 将 outline 条目拆分为 Writer 分配组
//
// 分配策略:
//   - complexity 为 "high" 的条目：每组 1 个（复杂模块需 Writer 专注）
//   - complexity 为 "medium" 或 "low"（或其他值）：每组最多 2 个
//
// 返回 [][]WikiEntry，每个子切片是单个 Writer Agent 的负责条目。
func splitWriterGroups(outline []WikiEntry) [][]WikiEntry {
	var groups [][]WikiEntry
	var pendingMedium []WikiEntry
	for _, entry := range outline {
		if strings.ToLower(entry.Complexity) == "high" {
			// high 复杂度：单独一组
			if len(pendingMedium) > 0 {
				groups = append(groups, pendingMedium)
				pendingMedium = nil
			}
			groups = append(groups, []WikiEntry{entry})
			continue
		}
		// medium/low：累积 2 个一组
		pendingMedium = append(pendingMedium, entry)
		if len(pendingMedium) == 2 {
			groups = append(groups, pendingMedium)
			pendingMedium = nil
		}
	}
	if len(pendingMedium) > 0 {
		groups = append(groups, pendingMedium)
	}
	return groups
}

// findMissingEntries 从 ValidationError 列表中提取缺失/空页面对应的 WikiEntry
//
// 匹配 ValidationError.Path 与 WikiEntry.Path（精确匹配）。
// 错误类型为 missing_file / empty_page / orphan_file 时视为需要重写的条目。
func findMissingEntries(errors []ValidationError, outline []WikiEntry) []WikiEntry {
	pathSet := make(map[string]bool)
	for _, e := range errors {
		if e.Path == "" {
			continue
		}
		// 所有带 path 的错误都尝试匹配（missing_file / empty_page / orphan_file 等）
		pathSet[e.Path] = true
	}
	if len(pathSet) == 0 {
		return nil
	}
	var missing []WikiEntry
	for _, entry := range outline {
		if entry.Path == "" {
			continue
		}
		if pathSet[entry.Path] {
			missing = append(missing, entry)
		}
	}
	return missing
}

// buildArchitectRetryHint 构建 Architect 重试时追加到 user prompt 末尾的格式提醒
func buildArchitectRetryHint(attempt int) string {
	return fmt.Sprintf("\n\n---\n⚠️ 重要提醒（第 %d 次重试）：你上一次的输出无法解析为有效 JSON。\n请确保你**仅**输出纯 JSON 数组，不要包含 markdown 代码块（```）、解释性文字或任何其他内容。\nJSON 必须以 '[' 开头、']' 结尾，每个元素是包含 title/path/description/explore_refs/complexity 字段的对象。", attempt)
}

// ──────────────────────────────────────────────────────────────────────
// Session 留痕（审计 trail）
// ──────────────────────────────────────────────────────────────────────

// sessionArtifactMeta session 目录下 meta.json 的结构
type sessionArtifactMeta struct {
	Role         string            `json:"role"`
	Model        string            `json:"model"`
	SessionID    string            `json:"session_id"`
	StartedAt    string            `json:"started_at"`
	FinishedAt   string            `json:"finished_at"`
	DurationMS   int64             `json:"duration_ms"`
	Iterations   int               `json:"iterations,omitempty"`
	ExitReason   string            `json:"exit_reason,omitempty"`
	InputTokens  int64             `json:"input_tokens"`
	OutputTokens int64             `json:"output_tokens"`
	TotalTokens  int64             `json:"total_tokens"`
	ToolCalls    []toolCallSummary `json:"tool_calls,omitempty"`
	Attempt      int               `json:"attempt,omitempty"`
	Error        string            `json:"error,omitempty"`
}

type toolCallSummary struct {
	Name      string `json:"name"`
	IsError   bool   `json:"is_error,omitempty"`
	InputLen  int    `json:"input_len"`
	ResultLen int    `json:"result_len"`
}

// saveSessionArtifact 在 Agent 执行后持久化审计 trail 到 session 目录
//
// 写入文件:
//   - input.md       — user prompt 原文
//   - output.md      — Agent 最终输出（result.Content）
//   - meta.json      — 摘要元数据（角色/模型/耗时/token/工具调用/错误）
//   - messages.json  — 完整消息历史（仅当 result 非 nil 且有 Messages 时）
//
// 单次写入失败不影响主流程（仅 log.Warn）。
// attempt 用于 Architect 重试时区分每次尝试（其他阶段传 0）。
func (o *SubAgentOrchestrator) saveSessionArtifact(
	_ context.Context,
	role, sessionID, userInput string,
	result *agent.AgentResult,
	runErr error,
	start time.Time,
	attempt int,
) {
	sessionDir := filepath.Join(o.storage.GetSessionPath(o.versionID), sessionID)

	if mErr := o.storage.WriteMarkdown(filepath.Join(sessionDir, "input.md"), userInput); mErr != nil {
		o.log.Warn(nil, "写入 session input.md 失败",
			slog.String("session", sessionID),
			slog.String("err", mErr.Error()))
	}

	meta := sessionArtifactMeta{
		Role:      role,
		SessionID: sessionID,
		StartedAt: start.Format(time.RFC3339),
		FinishedAt: time.Now().Format(time.RFC3339),
		DurationMS: time.Since(start).Milliseconds(),
		Attempt:   attempt,
	}
	if mc, ok := o.roleModels[role]; ok && mc != nil {
		meta.Model = mc.ModelName
	}

	if runErr != nil {
		meta.Error = runErr.Error()
	}

	if result != nil {
		if mErr := o.storage.WriteMarkdown(filepath.Join(sessionDir, "output.md"), result.Content); mErr != nil {
			o.log.Warn(nil, "写入 session output.md 失败",
				slog.String("session", sessionID),
				slog.String("err", mErr.Error()))
		}

		meta.Iterations = result.Iterations
		meta.ExitReason = string(result.ExitReason)
		meta.InputTokens = result.Usage.InputTokens
		meta.OutputTokens = result.Usage.OutputTokens
		meta.TotalTokens = result.Usage.InputTokens + result.Usage.OutputTokens

		meta.ToolCalls = make([]toolCallSummary, 0, len(result.ToolCalls))
		for _, tc := range result.ToolCalls {
			meta.ToolCalls = append(meta.ToolCalls, toolCallSummary{
				Name:      tc.Name,
				IsError:   tc.IsError,
				InputLen:  len(tc.Input),
				ResultLen: len(tc.Result),
			})
		}

		if len(result.Messages) > 0 {
			if mErr := o.storage.WriteJSON(filepath.Join(sessionDir, "messages.json"), result.Messages); mErr != nil {
				o.log.Warn(nil, "写入 session messages.json 失败",
					slog.String("session", sessionID),
					slog.String("err", mErr.Error()))
			}
		}
	}

	if mErr := o.storage.WriteJSON(filepath.Join(sessionDir, "meta.json"), meta); mErr != nil {
		o.log.Warn(nil, "写入 session meta.json 失败",
			slog.String("session", sessionID),
			slog.String("err", mErr.Error()))
	}
}
