// Package logic RepoWiki Webhook 业务编排层。
//
// 本文件实现 RepoWiki 模块的 Webhook 自动更新能力，职责包括：
//   - 接收 Git Provider 推送事件并全链路审计记录（WebhookEvent）
//   - 提供商检测、签名校验、payload 解析、仓库 URL 校验、分支匹配、去重
//   - 命中后调用 AnalyzeRepo 触发增量分析（非阻塞）
//   - Webhook 凭据生成 / 重置 / 配置查询 / 事件列表 / 已完成 Wiki 列表
//
// 设计约束：
//   - 每个请求（含失败/忽略）必须落库一条 WebhookEvent 审计记录
//   - AnalyzeRepo 不被修改，webhook 仅作为调用方
//   - 分支匹配采用精确字符串比较，不支持通配符
package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xUtil "github.com/bamboo-services/bamboo-base-go/common/utility"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"

	apiRepowiki "github.com/xiaolfeng/Lumina/api/repowiki"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"github.com/xiaolfeng/Lumina/internal/service"
	"gorm.io/datatypes"
)

// ──────────────────────────────────────────────────────────────────────
// Webhook 相关 DTO（Logic 层内部使用）
// ──────────────────────────────────────────────────────────────────────

// WikiVersionSummary 已完成 Wiki 版本摘要（ListCompletedWikis 返回值）
type WikiVersionSummary struct {
	VersionID   xSnowflake.SnowflakeID
	ConfigID    xSnowflake.SnowflakeID
	ConfigName  string
	Branch      string
	Language    string
	CommitHash  string
	CompletedAt *time.Time
}

// WebhookConfigInfo Webhook 配置信息（GetWebhookConfig 返回值）
type WebhookConfigInfo struct {
	URL       string   // 完整 Webhook URL
	Token     string   // 脱敏 Token（前 8 字符 + ****）
	HasSecret bool     // 是否设置 Secret
	Branches  []string // 监控分支列表
}

// ──────────────────────────────────────────────────────────────────────
// Webhook 核心处理
// ──────────────────────────────────────────────────────────────────────

// HandleWebhookPush 处理 Git Provider 推送事件
//
// 全链路审计：每个请求（含失败/忽略）均落库 WebhookEvent 记录。
// Provider 检测在 Logic 内部完成，Handler 仅传递 token / headers / body。
//
// 处理流程：
//  1. Token → Config 查找；失败记录 404 事件
//  2. 立即创建 status=received 事件记录
//  3. DetectProvider；未知 → failed(401)
//  4. VerifyWebhookSignature；失败 → failed(401)
//  5. ParseWebhookPayload；非 push → ignored(200)
//  6. 仓库 URL 归一化比对；不匹配 → ignored(200)
//  7. WebhookBranches 为空 → ignored(200)
//  8. 分支不在监控列表 → ignored(200)
//  9. 同一 CommitHash 已完成分析 → ignored(200)
//  10. 调用 AnalyzeRepo；成功 → accepted(200)
//  11. AnalyzeRepo 失败 → failed(500)
//  12. 所有路径均设置 processed_at
func (l *RepoWikiLogic) HandleWebhookPush(ctx context.Context, token string, headers http.Header, body []byte) (*entity.WebhookEvent, *xError.Error) {
	l.log.Info(ctx, "HandleWebhookPush - 接收 Webhook 推送事件")

	now := time.Now()

	// Step 1: Token → Config 查找
	config, xErr := l.repo.config.GetByWebhookToken(ctx, token)
	if xErr != nil {
		// Token 无效：记录 config_id=nil 的失败事件
		event := &entity.WebhookEvent{
			BaseEntity:   xModels.BaseEntity{ID: xSnowflake.GenerateID(bConst.GeneWebhookEvent)},
			ConfigID:     nil,
			Provider:     "",
			EventType:    "unknown",
			Status:       bConst.WebhookEventStatusFailed,
			Reason:       "token_not_found",
			ResponseCode: http.StatusNotFound,
			ReceivedAt:   now,
		}
		created, createErr := l.repo.webhookEvent.Create(ctx, event)
		if createErr != nil {
			l.log.Error(ctx, "HandleWebhookPush - 审计事件落库失败(token_not_found)", slog.Any("err", createErr))
			return nil, xErr
		}
		processedAt := time.Now()
		_ = l.repo.webhookEvent.UpdateStatus(ctx, created.ID, bConst.WebhookEventStatusFailed, "token_not_found", 0, http.StatusNotFound, &processedAt)
		return created, xErr
	}

	// Step 2: 立即创建 status=received 事件记录
	event := &entity.WebhookEvent{
		BaseEntity: xModels.BaseEntity{ID: xSnowflake.GenerateID(bConst.GeneWebhookEvent)},
		ConfigID:   &config.ID,
		Provider:   "",
		EventType:  "unknown",
		Status:     bConst.WebhookEventStatusReceived,
		ReceivedAt: now,
	}
	created, xErr := l.repo.webhookEvent.Create(ctx, event)
	if xErr != nil {
		l.log.Error(ctx, "HandleWebhookPush - 审计事件落库失败(received)", slog.Any("err", xErr))
		return nil, xErr
	}

	// finishEvent 是所有路径的统一收尾：更新状态 + 设置 processed_at
	finishEvent := func(status, reason string, versionID xSnowflake.SnowflakeID, responseCode int) {
		processedAt := time.Now()
		if uErr := l.repo.webhookEvent.UpdateStatus(ctx, created.ID, status, reason, versionID, responseCode, &processedAt); uErr != nil {
			l.log.Warn(ctx, "HandleWebhookPush - 更新审计事件状态失败", slog.Any("err", uErr))
		}
		created.Status = status
		created.Reason = reason
		created.VersionID = versionID
		created.ResponseCode = responseCode
		created.ProcessedAt = &processedAt
	}

	// Step 3: DetectProvider
	provider, headerValue := service.DetectProvider(headers)
	if provider == "" {
		finishEvent(bConst.WebhookEventStatusFailed, "unknown_provider", 0, http.StatusUnauthorized)
		return created, xError.NewError(ctx, xError.BusinessError, "无法识别 Webhook 提供商", false, nil)
	}

	// 更新事件 Provider 字段（best-effort，不阻塞流程）
	created.Provider = provider

	// Step 4: VerifyWebhookSignature
	if !service.VerifyWebhookSignature(provider, body, headerValue, config.WebhookSecret) {
		finishEvent(bConst.WebhookEventStatusFailed, "signature_failed", 0, http.StatusUnauthorized)
		return created, xError.NewError(ctx, xError.BusinessError, "Webhook 签名校验失败", false, nil)
	}

	// Step 5: ParseWebhookPayload
	payload, isPush, err := service.ParseWebhookPayload(provider, headers, body)
	if err != nil {
		finishEvent(bConst.WebhookEventStatusFailed, "parse_error: "+err.Error(), 0, http.StatusBadRequest)
		return created, xError.NewError(ctx, xError.BusinessError, "Webhook payload 解析失败", false, err)
	}
	if !isPush || payload == nil {
		// 非 push 事件：记录 detected type 后忽略
		eventType := headers.Get("X-GitHub-Event")
		if eventType == "" {
			eventType = headers.Get("X-Gitee-Event")
		}
		if eventType == "" {
			eventType = headers.Get("X-Gitea-Event")
		}
		if eventType == "" {
			eventType = "non_push"
		}
		created.EventType = eventType
		finishEvent(bConst.WebhookEventStatusIgnored, "non_push_event", 0, http.StatusOK)
		return created, nil
	}

	// 填充事件详情字段
	created.EventType = "push"
	created.Branch = payload.Branch
	created.CommitBefore = payload.BeforeHash
	created.CommitAfter = payload.AfterHash
	created.ChangedCount = len(payload.ChangedFiles)

	// Step 6: 仓库 URL 归一化比对
	if !normalizeRepoURL(payload.RepoURL, config.GitURL) {
		finishEvent(bConst.WebhookEventStatusIgnored, "repo_url_mismatch", 0, http.StatusOK)
		return created, nil
	}

	// Step 7: WebhookBranches 为空检查
	var branches []string
	if err := json.Unmarshal(config.WebhookBranches, &branches); err != nil {
		// JSON 解析失败视为空
		branches = nil
	}
	if len(branches) == 0 {
		finishEvent(bConst.WebhookEventStatusIgnored, "no_branches_configured", 0, http.StatusOK)
		return created, nil
	}

	// Step 8: 分支精确匹配
	branchMatched := false
	for _, b := range branches {
		if b == payload.Branch {
			branchMatched = true
			break
		}
	}
	if !branchMatched {
		finishEvent(bConst.WebhookEventStatusIgnored, "branch_not_monitored", 0, http.StatusOK)
		return created, nil
	}

	// Step 9: 去重检查 — 同一 CommitHash 已完成分析
	latest, xErr := l.repo.version.GetLatestByConfigID(ctx, config.ID)
	if xErr == nil && latest != nil {
		if latest.CommitHash == payload.AfterHash && latest.Status == bConst.RepoWikiStatusCompleted {
			finishEvent(bConst.WebhookEventStatusIgnored, "already_analyzed", 0, http.StatusOK)
			return created, nil
		}
	}
	// GetLatestByConfigID 返回 NotFound 时不视为错误，继续流程

	// Step 10: 调用 AnalyzRepo 触发增量分析
	version, xErr := l.AnalyzeRepo(ctx, config.ID, &apiRepowiki.AnalyzeRequest{
		Branch:   payload.Branch,
		Language: config.DefaultLanguage,
	})
	if xErr != nil {
		// Step 11: AnalyzeRepo 失败
		finishEvent(bConst.WebhookEventStatusFailed, "analyze_failed: "+xErr.Error(), 0, http.StatusInternalServerError)
		return created, xErr
	}

	// Step 10 成功
	finishEvent(bConst.WebhookEventStatusAccepted, "", version.ID, http.StatusOK)
	return created, nil
}

// ──────────────────────────────────────────────────────────────────────
// Webhook 凭据管理
// ──────────────────────────────────────────────────────────────────────

// GenerateWebhookCredentials 生成 Webhook Token 和 Secret
//
// 使用 xUtil.Security().GenerateLongKey() 生成 67 字符随机串（cs_ + 64 hex）。
// Token 用于路由定位配置，Secret 用于 HMAC 签名校验。
// 两者均以明文存储（Token 仅用于路由，Secret 需要 HMAC 原文参与计算）。
func (l *RepoWikiLogic) GenerateWebhookCredentials() (token string, secret string) {
	token = xUtil.Security().GenerateLongKey()
	secret = xUtil.Security().GenerateLongKey()
	return
}

// RegenerateWebhook 重新生成 Webhook 凭据
//
// 生成新的 Token 和 Secret，更新到配置（旧凭据立即失效）。
// 返回新的完整凭据（仅展示一次）。
func (l *RepoWikiLogic) RegenerateWebhook(ctx context.Context, configID xSnowflake.SnowflakeID) (token string, secret string, xErr *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("RegenerateWebhook - 重新生成 Webhook 凭据 [%d]", configID.Int64()))

	config, xErr := l.repo.config.GetByID(ctx, configID)
	if xErr != nil {
		return "", "", xErr
	}

	token, secret = l.GenerateWebhookCredentials()
	config.WebhookToken = token
	config.WebhookSecret = secret

	if xErr := l.repo.config.Update(ctx, config); xErr != nil {
		return "", "", xErr
	}

	return token, secret, nil
}

// UpdateWebhookBranches 更新 Webhook 监控分支列表
//
// 将 branches 序列化为 datatypes.JSON 存储到配置。
// 空切片将被存储为 []。
func (l *RepoWikiLogic) UpdateWebhookBranches(ctx context.Context, configID xSnowflake.SnowflakeID, branches []string) *xError.Error {
	l.log.Info(ctx, fmt.Sprintf("UpdateWebhookBranches - 更新 Webhook 监控分支 [%d, branches=%v]", configID.Int64(), branches))

	config, xErr := l.repo.config.GetByID(ctx, configID)
	if xErr != nil {
		return xErr
	}

	jsonData, err := json.Marshal(branches)
	if err != nil {
		return xError.NewError(ctx, xError.ServerInternalError, "序列化分支列表失败", false, err)
	}
	config.WebhookBranches = datatypes.JSON(jsonData)

	return l.repo.config.Update(ctx, config)
}

// GetWebhookConfig 获取 Webhook 配置信息
//
// 返回完整 URL、脱敏 Token、是否设置 Secret、监控分支列表。
// URL 格式：https://{requestHost}/api/v1/webhooks/repowiki/{token}
// Token 脱敏：前 8 字符 + ****
func (l *RepoWikiLogic) GetWebhookConfig(ctx context.Context, configID xSnowflake.SnowflakeID, requestHost string) (*WebhookConfigInfo, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("GetWebhookConfig - 获取 Webhook 配置 [%d]", configID.Int64()))

	config, xErr := l.repo.config.GetByID(ctx, configID)
	if xErr != nil {
		return nil, xErr
	}

	// 脱敏 Token：前 8 字符 + ****
	maskedToken := maskToken(config.WebhookToken)

	// 解析监控分支
	var branches []string
	if err := json.Unmarshal(config.WebhookBranches, &branches); err != nil {
		branches = nil
	}

	url := fmt.Sprintf("https://%s/api/v1/webhooks/repowiki/%s", requestHost, config.WebhookToken)

	return &WebhookConfigInfo{
		URL:       url,
		Token:     maskedToken,
		HasSecret: config.WebhookSecret != "",
		Branches:  branches,
	}, nil
}

// ──────────────────────────────────────────────────────────────────────
// Webhook 事件查询
// ──────────────────────────────────────────────────────────────────────

// ListWebhookEvents 按配置 ID 分页查询 Webhook 事件
func (l *RepoWikiLogic) ListWebhookEvents(ctx context.Context, configID xSnowflake.SnowflakeID, page, size int) ([]*entity.WebhookEvent, int64, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("ListWebhookEvents - 分页查询 Webhook 事件 [configID=%d, page=%d, size=%d]", configID.Int64(), page, size))

	pageReq := xModels.PageRequest{Page: int64(page), Size: int64(size)}.Normalize()
	return l.repo.webhookEvent.ListByConfigID(ctx, configID, int(pageReq.Page), int(pageReq.Size))
}

// ListCompletedWikis 分页查询所有已完成的 Wiki 版本（跨配置）
//
// 批量获取配置名称：先查询版本列表，收集唯一 ConfigID，
// 再逐个查询配置构建 ID→Name 映射，避免 N+1 查询。
func (l *RepoWikiLogic) ListCompletedWikis(ctx context.Context, page, size int) ([]*WikiVersionSummary, int64, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("ListCompletedWikis - 分页查询已完成 Wiki [page=%d, size=%d]", page, size))

	pageReq := xModels.PageRequest{Page: int64(page), Size: int64(size)}.Normalize()
	versions, total, xErr := l.repo.version.ListCompleted(ctx, int(pageReq.Page), int(pageReq.Size))
	if xErr != nil {
		return nil, 0, xErr
	}

	// 收集唯一 ConfigID
	configIDSet := make(map[xSnowflake.SnowflakeID]struct{})
	for _, v := range versions {
		configIDSet[v.ConfigID] = struct{}{}
	}

	// 批量查询配置名称
	configNameMap := make(map[xSnowflake.SnowflakeID]string)
	for configID := range configIDSet {
		if config, xErr := l.repo.config.GetByID(ctx, configID); xErr == nil {
			configNameMap[configID] = config.Name
		}
	}

	// 组装摘要
	summaries := make([]*WikiVersionSummary, 0, len(versions))
	for _, v := range versions {
		summaries = append(summaries, &WikiVersionSummary{
			VersionID:   v.ID,
			ConfigID:    v.ConfigID,
			ConfigName:  configNameMap[v.ConfigID],
			Branch:      v.Branch,
			Language:    v.Language,
			CommitHash:  v.CommitHash,
			CompletedAt: v.CompletedAt,
		})
	}

	return summaries, total, nil
}

// ──────────────────────────────────────────────────────────────────────
// 辅助函数
// ──────────────────────────────────────────────────────────────────────

// normalizeRepoURL 归一化仓库 URL 并比较是否匹配
//
// 归一化规则：
//   - 移除 .git 后缀
//   - 移除尾部 /
//   - 转小写
func normalizeRepoURL(a, b string) bool {
	return normalizeURL(a) == normalizeURL(b)
}

// normalizeURL 单个 URL 归一化
func normalizeURL(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, ".git")
	s = strings.TrimSuffix(s, "/")
	return strings.ToLower(s)
}

// maskToken 脱敏 Webhook Token
//
// 显示前 8 字符 + ****，不足 8 字符时全部显示 ****。
func maskToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:8] + "****"
}
