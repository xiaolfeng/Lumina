package logic

import (
	"context"
	"fmt"
	"log/slog"

	xAsync "github.com/bamboo-services/bamboo-base-go/plugins/async"
	xEnv "github.com/bamboo-services/bamboo-base-go/defined/env"

	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
)

// RetryStaleTask 扫描超时的非终态任务并通过 xAsync 重试
//
// 由 xCron 定时调用（每 5 分钟）。扫描逻辑：
//  1. 读取 REPOWIKI_TASK_TIMEOUT（默认 1800 秒）和 REPOWIKI_MAX_RETRY（默认 3）
//  2. 查询 GetStaleTasks 获取超时任务列表（查询用 <= maxRetry，含超限任务）
//  3. 遍历每个任务：
//     a. retry_count >= maxRetry → 标记 failed（error_msg="重试次数超限"）
//     b. retry_count < maxRetry → retry_count++ 持久化 → xAsync 触发重新分析
//  4. 重新分析时复用 resolveRunner（T3 抽取的方法）构建 LLM runner
//
// 幂等性：GetStaleTasks 的 updated_at 条件确保同一任务不会被重复扫描
// （retry_count++ 会更新 updated_at，使其在下一个扫描周期前不再次命中）
func (l *RepoWikiLogic) RetryStaleTask(ctx context.Context) {
	timeoutSec := xEnv.GetEnvInt("REPOWIKI_TASK_TIMEOUT", 1800)
	maxRetry := xEnv.GetEnvInt("REPOWIKI_MAX_RETRY", 3)

	staleTasks, xErr := l.repo.version.GetStaleTasks(ctx, timeoutSec, maxRetry)
	if xErr != nil {
		l.log.Error(ctx, "RetryStaleTask - 查询超时任务失败", slog.String("err", xErr.Error()))
		return
	}
	if len(staleTasks) == 0 {
		return // 无超时任务，正常退出
	}

	l.log.Info(ctx, "RetryStaleTask - 发现超时任务", slog.Int("count", len(staleTasks)))

	for _, task := range staleTasks {
		// 获取关联配置
		config, xErr := l.repo.config.GetByID(ctx, task.ConfigID)
		if xErr != nil {
			// 配置不存在（已删除），标记版本为 failed 避免无限扫描
			task.Status = bConst.RepoWikiStatusFailed
			task.ErrorMsg = "关联配置已不存在"
			_ = l.repo.version.Update(ctx, task)
			l.log.Warn(ctx, "RetryStaleTask - 关联配置不存在，标记失败",
				slog.Int64("versionID", task.ID.Int64()),
				slog.Int64("configID", task.ConfigID.Int64()))
			continue
		}

		if task.RetryCount >= maxRetry {
			// 超过重试上限，标记 failed（版本 + 配置冗余字段同步）
			task.Status = bConst.RepoWikiStatusFailed
			task.ErrorMsg = fmt.Sprintf("重试次数超限 (%d/%d)", task.RetryCount, maxRetry)
			config.Status = bConst.RepoWikiStatusFailed
			_ = l.repo.version.Update(ctx, task)
			_ = l.repo.config.Update(ctx, config)
			l.log.Warn(ctx, "RetryStaleTask - 任务重试超限，标记失败",
				slog.Int64("versionID", task.ID.Int64()),
				slog.Int("retryCount", task.RetryCount))
			continue
		}

		// 通过 xAsync 异步重试（retry_count 递增在 retryTaskAsync 内部，
		// 确认信号量获取成功后才递增，避免信号量满时白白消耗重试次数）
		l.retryTaskAsync(ctx, config, task)
	}
}

// retryTaskAsync 通过 xAsync 异步重试单个超时任务
//
// retry_count 递增在信号量获取成功后执行，避免信号量满或 resolveRunner 失败时白白消耗重试次数。
// 对于信号量满的任务，仅刷新 updated_at（不递增 retry_count），下个 cron 周期会重新扫描。
func (l *RepoWikiLogic) retryTaskAsync(ctx context.Context, config *entity.RepoWikiConfig, task *entity.WikiVersion) {
	// 先非阻塞获取信号量（满则跳过，不消耗 retry_count）
	select {
	case l.semaphore <- struct{}{}:
	default:
		// 信号量已满，仅刷新 updated_at 延迟下轮扫描（不递增 retry_count）
		_ = l.repo.version.Update(ctx, task)
		l.log.Warn(ctx, "retryTaskAsync - 信号量已满，延迟重试",
			slog.Int64("versionID", task.ID.Int64()))
		return
	}

	// 信号量获取成功后，递增 retry_count 并持久化（同时刷新 updated_at）
	task.RetryCount++
	if uErr := l.repo.version.Update(ctx, task); uErr != nil {
		<-l.semaphore // 持久化失败则释放信号量
		l.log.Warn(ctx, "retryTaskAsync - 更新 retry_count 失败",
			slog.Int64("versionID", task.ID.Int64()),
			slog.String("err", uErr.Error()))
		return
	}

	// 解析 LLM Runner（失败则释放信号量，不消耗 retry_count——因为递增已在上面完成，
	// 但 LLM 未配置属于环境问题，递增的 retry_count 会在下轮 cron 重新尝试）
	runner, proto, model, xErr := l.resolveRunner(ctx)
	if xErr != nil {
		<-l.semaphore
		l.log.Warn(ctx, "retryTaskAsync - LLM 解析失败",
			slog.Int64("versionID", task.ID.Int64()),
			slog.String("err", xErr.Error()))
		return
	}

	pipeline := NewAnalysisPipeline(l, l.log, runner, proto, model)

	xAsync.Async(ctx, func(asyncCtx context.Context) {
		defer func() { <-l.semaphore }()
		if pErr := pipeline.Execute(asyncCtx, config, task); pErr != nil {
			l.log.Error(asyncCtx, "retryTaskAsync - 重试执行失败",
				slog.Int64("versionID", task.ID.Int64()),
				slog.String("err", pErr.Error()))
		}
	},
		xAsync.WithName("RepoWiki-CronRetry"),
		xAsync.WithLogger(l.log),
	)
}
