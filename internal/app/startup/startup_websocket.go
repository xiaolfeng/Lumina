package startup

import (
	"context"

	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	"github.com/xiaolfeng/Lumina/internal/websocket"
)

// NewWebSocketRunner 返回 WebSocket Hub 主循环启动函数，由 main.go 传给 xMain.Runner 的 goroutineFunc 参数。
//
// ctx 由 xMain.Runner 传入（为 reg.Init.Ctx 的 WithCancel 派生 context），
// Hub.Run 依赖 ctx.Done() 在收到 SIGINT/SIGTERM 时执行 shutdownAll() 优雅关闭全部连接。
//
// Hub 单例已在 route 包的 wsRouter（RouteRegistrar，Register 阶段）中通过
// websocket.GetHub 创建并完成回调装配；此处 GetGlobalHub 仅获取该单例，
// 避免在 RouteRegistrar 中启动主循环——其 ctx 未经 WithCancel，Done() 永不触发。
func NewWebSocketRunner() func(ctx context.Context, option ...any) {
	return func(ctx context.Context, option ...any) {
		log := xLog.WithName(xLog.NamedCONT)
		hub := websocket.GetGlobalHub()
		if hub == nil {
			log.Warn(ctx, "WebSocket Hub 未初始化，跳过主循环启动")
			return
		}
		hub.Run(ctx)
	}
}
