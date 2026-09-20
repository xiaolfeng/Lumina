package startup

import (
	"context"
	"fmt"

	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xCtx "github.com/bamboo-services/bamboo-base-go/defined/context"
	"github.com/xiaolfeng/Lumina/internal/logic"
	"github.com/xiaolfeng/Lumina/internal/mcp"
	"github.com/xiaolfeng/Lumina/internal/service"
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
	previewLogic := logic.NewPreviewLogic(ctx)
	pagesLogic := logic.NewPagesLogic(ctx)
	workspaceLogic := logic.NewWorkspaceLogic(ctx)
	repoWikiLogic := logic.GetRepoWikiLogicFromContext(ctx)
	if repoWikiLogic == nil {
		log.Warn(ctx, "context 中未找到 RepoWikiLogic，MCP 的 RepoWiki 工具将不可用")
	}

	schemaLoader, err := service.NewLpwSchemaLoader()
	if err != nil {
		log.Error(ctx, fmt.Sprintf("初始化 LPW Schema Loader 失败: %s", err.Error()))
		return nil, err
	}
	previewLpwLogic := logic.NewPreviewLpwLogic(previewLogic, schemaLoader)

	mcp.SetQaLogic(qaLogic)
	mcp.SetProjectLogic(projectLogic)
	mcp.SetPinLogic(pinLogic)
	mcp.SetPreviewLogic(previewLogic)
	mcp.SetPreviewLpwLogic(previewLpwLogic)
	mcp.SetPagesLogic(pagesLogic)
	mcp.SetWorkspaceLogic(workspaceLogic)
	mcp.SetRepoWikiLogic(repoWikiLogic)

	handler := mcp.InitMCPServer(ctx)

	log.Info(ctx, "MCP Server 初始化完成")
	return handler, nil
}
