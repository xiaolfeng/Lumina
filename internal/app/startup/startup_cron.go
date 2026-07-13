package startup

import (
	"context"
	"log/slog"
	"time"

	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xCron "github.com/bamboo-services/bamboo-base-go/plugins/cron"
	xCronRunner "github.com/bamboo-services/bamboo-base-go/plugins/cron/runner"

	"github.com/xiaolfeng/Lumina/internal/logic"
)

// NewCronRunner 返回 cron 附加协程启动函数，由 main.go 传给 xMain.Runner 的 goroutineFunc 参数。
// ctx 由 xMain.Runner 在调用时传入（为 reg.Init.Ctx 的派生 context，已注入所有 Logic）。
func NewCronRunner() func(ctx context.Context, option ...any) {
	return func(ctx context.Context, option ...any) {
		log := xLog.WithName(xLog.NamedCRON)
		runner := xCronRunner.New(
			xCronRunner.WithLogger(log),
			xCronRunner.WithRegister(
				xCron.NewJob("@every 5m", func(jobCtx context.Context) {
					defer func() {
						if rec := recover(); rec != nil {
							log.Error(jobCtx, "RepoWiki cron Job panic recovered", slog.Any("error", rec))
						}
					}()
					repoWikiLogic := logic.GetRepoWikiLogicFromContext(ctx)
					if repoWikiLogic == nil {
						log.Warn(jobCtx, "RepoWikiLogic 未注入 context，跳过兜底重试")
						return
					}
					repoWikiLogic.RetryStaleTask(jobCtx)
				}),
			),
			xCronRunner.WithGracefulStopTimeout(30*time.Second),
		)
		runner(ctx, option...)
	}
}
