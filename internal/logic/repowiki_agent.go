// Package logic RepoWiki Agent 分析 Pass 运行器。
//
// AgentPassRunner 管理 4 个串行执行的 LLM 分析 Pass（概览/模块/架构/指南）。
// 每个 Pass 通过 bamboo-agent 框架执行：构建 Agent → Run → 提取输出 → 解析 JSON → 失败重试。
// Pass 之间存在数据依赖（后序 Pass 依赖前序输出），因此必须串行执行。
package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
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
// PassResult
// ──────────────────────────────────────────────────────────────────────

// PassResult 单个 Pass 的执行结果
//
// 无论成功还是失败都会产出 PassResult，失败时 Success 为 false 且 Error 包含原因。
// JSON 字段保存解析后的原始 JSON 数据（成功时）或 nil（失败时）。
type PassResult struct {
	Name       string          `json:"name"`                 // Pass 名称（pass1/pass2/pass3/pass4）
	Success    bool            `json:"success"`              // 是否成功
	JSON       json.RawMessage `json:"json,omitempty"`       // 解析后的 JSON 原始数据（成功时）
	RawOutput  string          `json:"raw_output,omitempty"` // Agent 原始输出文本（调试用，截断至 2000 字符）
	Error      string          `json:"error,omitempty"`      // 错误信息（失败时）
	DurationMs int64           `json:"duration_ms"`          // 执行耗时（毫秒）
	TokenCount int64           `json:"token_count"`          // Token 消耗（输入 + 输出）
	Attempts   int             `json:"attempts"`             // 实际尝试次数（含首次）
}

// ──────────────────────────────────────────────────────────────────────
// AgentPassRunner
// ──────────────────────────────────────────────────────────────────────

// defaultMaxRetries JSON 解析失败时的默认最大重试次数
const defaultMaxRetries = 3

// rawOutputTruncateLimit RawOutput 字段的截断字符数上限
const rawOutputTruncateLimit = 2000

// AgentPassRunner 管理 4 个 Agent 分析 Pass 的串行执行
//
// 职责：
//   - 为每个 Pass 构建 Agent（使用 bamboo-agent 框架）
//   - 执行 Agent 分析并提取最终输出文本
//   - 从输出中解析 JSON（处理 markdown 代码块包裹等情况）
//   - JSON 解析失败时自动重试（在 prompt 中追加格式提醒）
//   - 将每个 Pass 的结果持久化到文件系统
//   - 通过 progressCallback 回调通知上层更新 current_stage
//
// 非职责：
//   - 不负责 Git 克隆（由 GitCloneService 处理）
//   - 不负责文件扫描（由 FileScannerService 处理）
//   - 不负责文档组装（由 Document Assembler 处理）
//   - 不负责状态机管理（由 Pipeline Orchestrator 处理）
type AgentPassRunner struct {
	client     bamboo.BambooClient // LLM 客户端（真实 Provider 或 Stub）
	storage    *service.WikiStorageService
	log        *xLog.LogNamedLogger
	tools      []tool.Tool // Agent 可用工具集（file_read + file_search），仅只读工具
	maxRetries int         // JSON 解析失败重试次数
}

// NewAgentPassRunner 创建 AgentPassRunner 实例
//
// 参数说明:
//   - client: 已配置好的 LLM 客户端（由 service.NewLLMProvider 或 bamboo.NewClient(stub) 生成）
//   - storage: Wiki 存储服务（用于持久化 Pass 结果）
//   - log: 日志记录器（建议使用 xLog.WithName(xLog.NamedLOGC, "AgentPassRunner")）
//   - tools: Agent 可用工具集（仅 file_read + file_search 等只读工具，禁止 shell 等可写工具）
//
// maxRetries 默认为 defaultMaxRetries（3 次）。
func NewAgentPassRunner(
	client bamboo.BambooClient,
	storage *service.WikiStorageService,
	log *xLog.LogNamedLogger,
	tools []tool.Tool,
) *AgentPassRunner {
	return &AgentPassRunner{
		client:     client,
		storage:    storage,
		log:        log,
		tools:      tools,
		maxRetries: defaultMaxRetries,
	}
}

// SetMaxRetries 设置 JSON 解析失败时的最大重试次数（测试用）
func (r *AgentPassRunner) SetMaxRetries(n int) {
	if n > 0 {
		r.maxRetries = n
	}
}

// ──────────────────────────────────────────────────────────────────────
// RunAllPasses
// ──────────────────────────────────────────────────────────────────────

// RunAllPasses 执行全部 4 个 Pass（串行，存在数据依赖）
//
// 执行流程:
//  1. Pass 1（概览）：输入 = 仓库路径 + 文件扫描摘要
//  2. Pass 2（模块）：输入 = Pass 1 输出 + 依赖摘要 JSON
//  3. Pass 3（架构）：输入 = Pass 1 + Pass 2 输出
//  4. Pass 4（指南）：输入 = Pass 1 + Pass 2 + Pass 3 输出
//
// 每个 Pass 完成后调用 progressCallback 通知上层更新 current_stage。
// 如果某个 Pass 失败（所有重试均失败），立即停止后续 Pass 并返回错误。
//
// 参数说明:
//   - ctx: 上下文（支持取消）
//   - versionID: 版本 ID（用于定位 passes/ 和 sessions/ 目录）
//   - repoPath: 克隆的仓库路径（Agent 工作目录的根）
//   - fileScan: 文件扫描结果（可为 nil，跳过该上下文）
//   - depSummary: 依赖摘要（可为 nil，跳过该上下文）
//   - progressCallback: 每个 Pass 开始前回调（参数为 stage 字符串，如 "pass1"）
//
// 返回值:
//   - map[string]*PassResult: 所有已执行 Pass 的结果（key 为 "pass1"/"pass2"/...）
//   - *xError.Error: 某个 Pass 彻底失败时返回错误（results 中仍包含已完成的部分结果）
func (r *AgentPassRunner) RunAllPasses(
	ctx context.Context,
	versionID int64,
	repoPath string,
	fileScan *service.FileScanResult,
	depSummary *service.DependencySummary,
	progressCallback func(stage string),
) (map[string]*PassResult, *xError.Error) {
	results := make(map[string]*PassResult, 4)

	// 预构建上下文 JSON（如果可用）
	var fileScanJSON, depSummaryJSON []byte
	if fileScan != nil {
		fileScanJSON, _ = json.MarshalIndent(fileScan, "", "  ")
	}
	if depSummary != nil {
		depSummaryJSON, _ = json.MarshalIndent(depSummary, "", "  ")
	}

	// ── Pass 1: 项目概览 ──
	r.notifyProgress(progressCallback, bConst.RepoWikiStagePass1)
	pass1Input := buildPass1UserInput(repoPath, fileScanJSON)
	pass1Result := r.runSinglePass(ctx, 1, pass1Input, versionID, repoPath)
	results["pass1"] = pass1Result
	if !pass1Result.Success {
		return results, xError.NewError(ctx, xError.ServerInternalError,
			xError.ErrMessage("Pass 1（项目概览）分析失败: "+pass1Result.Error), false, nil)
	}

	// ── Pass 2: 模块分析（依赖 Pass 1 输出 + depSummary）──
	r.notifyProgress(progressCallback, bConst.RepoWikiStagePass2)
	pass2Input := buildPass2UserInput(pass1Result, depSummaryJSON)
	pass2Result := r.runSinglePass(ctx, 2, pass2Input, versionID, repoPath)
	results["pass2"] = pass2Result
	if !pass2Result.Success {
		return results, xError.NewError(ctx, xError.ServerInternalError,
			xError.ErrMessage("Pass 2（模块分析）分析失败: "+pass2Result.Error), false, nil)
	}

	// ── Pass 3: 架构分析（依赖 Pass 1+2 输出）──
	r.notifyProgress(progressCallback, bConst.RepoWikiStagePass3)
	pass3Input := buildPass3UserInput(pass1Result, pass2Result)
	pass3Result := r.runSinglePass(ctx, 3, pass3Input, versionID, repoPath)
	results["pass3"] = pass3Result
	if !pass3Result.Success {
		return results, xError.NewError(ctx, xError.ServerInternalError,
			xError.ErrMessage("Pass 3（架构分析）分析失败: "+pass3Result.Error), false, nil)
	}

	// ── Pass 4: 阅读指南（依赖 Pass 1+2+3 输出）──
	r.notifyProgress(progressCallback, bConst.RepoWikiStagePass4)
	pass4Input := buildPass4UserInput(pass1Result, pass2Result, pass3Result)
	pass4Result := r.runSinglePass(ctx, 4, pass4Input, versionID, repoPath)
	results["pass4"] = pass4Result
	if !pass4Result.Success {
		return results, xError.NewError(ctx, xError.ServerInternalError,
			xError.ErrMessage("Pass 4（阅读指南）分析失败: "+pass4Result.Error), false, nil)
	}

	r.log.Info(ctx, "全部 4 个 Agent 分析 Pass 执行完成",
		slog.Int("versionID", int(versionID)),
		slog.Int("total_tokens", int(totalTokens(results))))
	return results, nil
}

// ──────────────────────────────────────────────────────────────────────
// runSinglePass（内部方法）
// ──────────────────────────────────────────────────────────────────────

// runSinglePass 执行单个 Pass 的完整流程
//
// 流程:
//  1. 从 allPasses 获取 Pass 元信息（systemPrompt）
//  2. 构建 Agent（使用 service.NewRepoWikiAgent）
//  3. 调用 executeWithRetry 执行分析（含重试逻辑）
//  4. 将结果持久化到 {versionPath}/passes/pass-{N}.json
//
// 返回的 PassResult 总是非 nil（即使失败也包含错误信息）。
func (r *AgentPassRunner) runSinglePass(
	ctx context.Context,
	passNum int,
	userInput string,
	versionID int64,
	repoPath string,
) *PassResult {
	info := allPasses[passNum-1]
	start := time.Now()

	// 构建工具：每个 Pass 都创建作用域限定在 repoPath 的 file_read + file_search 工具
	tools := []tool.Tool{
		service.NewFileReadTool(repoPath),
		service.NewFileSearchTool(repoPath),
	}
	if len(r.tools) > 0 {
		// 保留运行器级别注入的扩展工具，只读工具放在前面避免被覆盖
		tools = append(tools, r.tools...)
	}

	// 构建 Agent（每个 Pass 使用独立的 session 目录）
	sessionDir := filepath.Join(r.storage.GetSessionPath(versionID), info.Name)
	ag, err := service.NewRepoWikiAgent(r.client, info.SystemPrompt, tools, sessionDir)
	if err != nil {
		return &PassResult{
			Name:       info.Name,
			Success:    false,
			Error:      fmt.Sprintf("创建 Agent 失败: %v", err),
			DurationMs: time.Since(start).Milliseconds(),
		}
	}

	// 执行分析（含重试）
	result := r.executeWithRetry(ctx, ag, userInput, info.Name)
	result.DurationMs = time.Since(start).Milliseconds()

	// 持久化结果
	passPath := filepath.Join(r.storage.GetPassesPath(versionID), fmt.Sprintf("pass-%d.json", passNum))
	if writeErr := r.storage.WriteJSON(passPath, result); writeErr != nil {
		r.log.Warn(ctx, "写入 Pass 结果失败",
			slog.String("pass", info.Name),
			slog.String("path", passPath),
			slog.String("err", writeErr.Error()))
	} else {
		r.log.Debug(ctx, "Pass 结果已写入",
			slog.String("pass", info.Name),
			slog.String("path", passPath))
	}

	// 日志记录
	if result.Success {
		r.log.Info(ctx, "Pass 执行成功",
			slog.String("pass", info.Name),
			slog.Int64("duration_ms", result.DurationMs),
			slog.Int64("tokens", result.TokenCount),
			slog.Int("attempts", result.Attempts))
	} else {
		r.log.Error(ctx, "Pass 执行失败",
			slog.String("pass", info.Name),
			slog.String("error", result.Error),
			slog.Int("attempts", result.Attempts))
	}

	return result
}

// ──────────────────────────────────────────────────────────────────────
// executeWithRetry（内部方法）
// ──────────────────────────────────────────────────────────────────────

// executeWithRetry 执行 Agent 并在 JSON 解析失败时重试
//
// 重试策略:
//  1. 调用 agent.Run(ctx, input) 获取输出
//  2. 尝试 parseAgentJSON 解析 JSON
//  3. 解析成功 → 返回成功结果
//  4. 解析失败 → 在 input 末尾追加格式提醒，进入下一次重试
//  5. 所有重试均失败 → 返回失败结果（包含最后一次错误信息）
//
// 每次 Agent.Run 失败（非 JSON 格式错误）也会触发重试。
// 重试次数由 r.maxRetries 控制（默认 3 次）。
func (r *AgentPassRunner) executeWithRetry(
	ctx context.Context,
	ag agent.Agent,
	userInput string,
	passName string,
) *PassResult {
	currentInput := userInput
	var lastErrMsg string
	var lastRawOutput string
	var lastTokens int64

	for attempt := 1; attempt <= r.maxRetries; attempt++ {
		// 检查上下文取消
		if ctx.Err() != nil {
			return &PassResult{
				Name:     passName,
				Success:  false,
				Error:    fmt.Sprintf("上下文已取消: %v", ctx.Err()),
				Attempts: attempt - 1,
			}
		}

		// 执行 Agent
		result, runErr := ag.Run(ctx, currentInput)
		if runErr != nil {
			lastErrMsg = fmt.Sprintf("Agent 执行错误 (第 %d 次): %v", attempt, runErr)
			lastRawOutput = ""
			r.log.Warn(ctx, "Agent 执行出错，准备重试",
				slog.String("pass", passName),
				slog.Int("attempt", attempt),
				slog.String("err", runErr.Error()))
			continue
		}

		// 提取输出
		rawOutput := result.Content
		lastRawOutput = rawOutput
		lastTokens = result.Usage.InputTokens + result.Usage.OutputTokens

		// 尝试解析 JSON
		parsed, parseErr := parseAgentJSON(rawOutput)
		if parseErr == nil {
			// 成功！
			return &PassResult{
				Name:       passName,
				Success:    true,
				JSON:       parsed,
				RawOutput:  truncate(rawOutput, rawOutputTruncateLimit),
				TokenCount: lastTokens,
				Attempts:   attempt,
			}
		}

		// JSON 解析失败，准备重试
		lastErrMsg = fmt.Sprintf("JSON 解析失败 (第 %d 次): %v", attempt, parseErr)
		r.log.Warn(ctx, "JSON 解析失败，准备重试",
			slog.String("pass", passName),
			slog.Int("attempt", attempt),
			slog.String("err", parseErr.Error()))

		// 在 input 末尾追加格式提醒（仅追加一次基础信息，后续重试递增提醒强度）
		currentInput = userInput + buildRetryHint(attempt, parseErr)
	}

	// 所有重试均失败
	return &PassResult{
		Name:       passName,
		Success:    false,
		RawOutput:  truncate(lastRawOutput, rawOutputTruncateLimit),
		Error:      lastErrMsg,
		TokenCount: lastTokens,
		Attempts:   r.maxRetries,
	}
}

// ──────────────────────────────────────────────────────────────────────
// parseAgentJSON（包级函数）
// ──────────────────────────────────────────────────────────────────────

// parseAgentJSON 从 Agent 响应文本中提取有效 JSON
//
// 解析策略（按优先级依次尝试）:
//  1. 直接解析：TrimSpace 后直接 json.Valid
//  2. Markdown 代码块提取：提取 ```json ... ``` 或 ``` ... ``` 中的内容
//  3. 花括号范围提取：取第一个 '{' 到最后一个 '}' 之间的子串
//  4. 全部失败 → 返回错误
//
// 该函数只验证 JSON 语法正确性，不校验是否匹配特定 Schema。
func parseAgentJSON(rawOutput string) (json.RawMessage, error) {
	trimmed := strings.TrimSpace(rawOutput)
	if trimmed == "" {
		return nil, fmt.Errorf("Agent 输出为空")
	}

	// 策略 1: 直接解析
	if json.Valid([]byte(trimmed)) {
		return json.RawMessage(trimmed), nil
	}

	// 策略 2: Markdown 代码块提取
	if extracted := extractMarkdownCodeBlock(trimmed); extracted != "" {
		if json.Valid([]byte(extracted)) {
			return json.RawMessage(extracted), nil
		}
	}

	// 策略 3: 花括号范围提取
	if extracted := extractBraceRange(trimmed); extracted != "" {
		if json.Valid([]byte(extracted)) {
			return json.RawMessage(extracted), nil
		}
	}

	return nil, fmt.Errorf("无法从 Agent 输出中解析有效 JSON（前 200 字符: %s）", truncate(trimmed, 200))
}

// extractMarkdownCodeBlock 从文本中提取 ```json...``` 或 ```...``` 代码块内容
//
// 匹配格式：
//
//	```json
//	{...}
//	```
//
// 或不带语言标记的代码块。返回代码块内的原始内容（TrimSpace 后），未匹配返回空字符串。
func extractMarkdownCodeBlock(text string) string {
	// 尝试 ```json ... ```
	if content := extractFencedBlock(text, "json"); content != "" {
		return content
	}
	// 尝试不带语言标记的 ``` ... ```
	if content := extractFencedBlock(text, ""); content != "" {
		return content
	}
	return ""
}

// extractFencedBlock 提取指定语言标记的 markdown 围栏代码块内容
//
// lang 为空时匹配不带语言标记的围栏块。
func extractFencedBlock(text, lang string) string {
	openFence := "```" + lang
	closeFence := "```"

	startIdx := strings.Index(text, openFence)
	if startIdx == -1 {
		return ""
	}
	// 跳过开始围栏 + 换行
	contentStart := startIdx + len(openFence)
	// 跳过围栏后的换行符
	for contentStart < len(text) && (text[contentStart] == '\n' || text[contentStart] == '\r') {
		contentStart++
	}

	remaining := text[contentStart:]
	before, _, found := strings.Cut(remaining, closeFence)
	if !found {
		return ""
	}

	return strings.TrimSpace(before)
}

// extractBraceRange 提取文本中第一个 '{' 到最后一个 '}' 之间的内容
//
// 这是一种兜底策略：LLM 可能在 JSON 前后添加了解释性文字，
// 但 JSON 本身以 '{' 开头、'}' 结尾。
// 未找到花括号时返回空字符串。
func extractBraceRange(text string) string {
	first := strings.IndexByte(text, '{')
	last := strings.LastIndexByte(text, '}')
	if first == -1 || last == -1 || first >= last {
		return ""
	}
	return text[first : last+1]
}

// ──────────────────────────────────────────────────────────────────────
// User Input 构建函数
// ──────────────────────────────────────────────────────────────────────

// buildPass1UserInput 构建 Pass 1 的用户输入
//
// 包含仓库路径和文件扫描结果摘要（语言统计、入口文件）。
func buildPass1UserInput(repoPath string, fileScanJSON []byte) string {
	var sb strings.Builder
	sb.WriteString("请分析以下代码仓库并生成项目概览。\n\n")
	fmt.Fprintf(&sb, "仓库路径: %s\n\n", repoPath)

	if len(fileScanJSON) > 0 {
		sb.WriteString("## 文件扫描结果摘要\n")
		sb.WriteString("以下是文件扫描的完整结果（JSON 格式），包含文件列表、语言分布和入口文件：\n\n")
		sb.Write(fileScanJSON)
		sb.WriteString("\n\n")
	}

	sb.WriteString("请使用 file_read 和 file_search 工具阅读关键文件（README、清单文件、入口文件），然后输出分析结果 JSON。\n")
	return sb.String()
}

// buildPass2UserInput 构建 Pass 2 的用户输入
//
// 包含 Pass 1 的输出 JSON 和依赖摘要。
func buildPass2UserInput(pass1Result *PassResult, depSummaryJSON []byte) string {
	var sb strings.Builder
	sb.WriteString("请分析以下代码仓库的模块划分。\n\n")

	sb.WriteString("## 前序分析结果\n\n")
	sb.WriteString("### Pass 1 - 项目概览\n")
	sb.Write(pass1Result.JSON)
	sb.WriteString("\n\n")

	if len(depSummaryJSON) > 0 {
		sb.WriteString("### 依赖摘要（dep_summary.json）\n")
		sb.Write(depSummaryJSON)
		sb.WriteString("\n\n")
	}

	sb.WriteString("请基于以上信息，使用 file_read 和 file_search 工具深入阅读各模块的入口文件，然后输出模块分析结果 JSON。\n")
	return sb.String()
}

// buildPass3UserInput 构建 Pass 3 的用户输入
//
// 包含 Pass 1 和 Pass 2 的输出 JSON。
func buildPass3UserInput(pass1Result, pass2Result *PassResult) string {
	var sb strings.Builder
	sb.WriteString("请分析以下代码仓库的整体架构设计。\n\n")

	sb.WriteString("## 前序分析结果\n\n")
	sb.WriteString("### Pass 1 - 项目概览\n")
	sb.Write(pass1Result.JSON)
	sb.WriteString("\n\n")

	sb.WriteString("### Pass 2 - 模块分析\n")
	sb.Write(pass2Result.JSON)
	sb.WriteString("\n\n")

	sb.WriteString("请基于以上信息，使用 file_read 工具阅读路由注册、中间件、启动流程等核心架构文件，然后输出架构分析结果 JSON。\n")
	return sb.String()
}

// buildPass4UserInput 构建 Pass 4 的用户输入
//
// 包含 Pass 1、2、3 的输出 JSON。
func buildPass4UserInput(pass1Result, pass2Result, pass3Result *PassResult) string {
	var sb strings.Builder
	sb.WriteString("请为以下代码仓库编写一份新人阅读指南。\n\n")

	sb.WriteString("## 前序分析结果\n\n")
	sb.WriteString("### Pass 1 - 项目概览\n")
	sb.Write(pass1Result.JSON)
	sb.WriteString("\n\n")

	sb.WriteString("### Pass 2 - 模块分析\n")
	sb.Write(pass2Result.JSON)
	sb.WriteString("\n\n")

	sb.WriteString("### Pass 3 - 架构分析\n")
	sb.Write(pass3Result.JSON)
	sb.WriteString("\n\n")

	sb.WriteString("请基于以上所有分析结果，站在新人视角编写阅读指南 JSON。\n")
	return sb.String()
}

// ──────────────────────────────────────────────────────────────────────
// 辅助函数
// ──────────────────────────────────────────────────────────────────────

// buildRetryHint 构建重试时追加到 user input 末尾的格式提醒
func buildRetryHint(attempt int, parseErr error) string {
	return fmt.Sprintf("\n\n---\n⚠️ 重要提醒（第 %d 次重试）：你上一次的输出无法解析为有效 JSON（错误: %v）。\n请确保你**仅**输出纯 JSON 对象，不要包含 markdown 代码块（```）、解释性文字或任何其他内容。\nJSON 必须以 '{' 开头、'}' 结尾。", attempt, parseErr)
}

// notifyProgress 安全调用 progressCallback（nil 安全）
func (r *AgentPassRunner) notifyProgress(cb func(string), stage string) {
	if cb != nil {
		cb(stage)
	}
}

// truncate 将字符串截断到指定长度，超长时追加 "..."
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// totalTokens 统算所有 Pass 结果的 Token 总量
func totalTokens(results map[string]*PassResult) int64 {
	var total int64
	for _, r := range results {
		total += r.TokenCount
	}
	return total
}
