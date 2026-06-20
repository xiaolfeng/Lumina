package startup

import (
	"context"

	xCtx "github.com/bamboo-services/bamboo-base-go/defined/context"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	"github.com/xiaolfeng/Lumina/internal/logic"
	"github.com/xiaolfeng/Lumina/internal/mcp"
)

// MCPHandlerKey MCP Server HTTP Handler 在 context 中的存储键。
const MCPHandlerKey xCtx.ContextKey = "mcp_handler"

// mcpInit 初始化 MCP Server 并将 HTTP Handler 注册到 context。
func (r *reg) mcpInit(ctx context.Context) (any, error) {
	log := xLog.WithName(xLog.NamedINIT)
	log.Debug(ctx, "正在初始化 MCP Server...")

	qaLogic := logic.NewQaLogic(ctx)
	projectLogic := logic.NewProjectLogic(ctx)
	pinLogic := logic.NewPinLogic(ctx)
	mcp.SetQaLogic(qaLogic)
	mcp.SetProjectLogic(projectLogic)
	mcp.SetPinLogic(pinLogic)

	handler := mcp.InitMCPServer(ctx)

	log.Info(ctx, "MCP Server 初始化完成")
	return handler, nil
}
