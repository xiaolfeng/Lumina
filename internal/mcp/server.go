package mcp

import (
	"context"
	"net/http"

	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xEnv "github.com/bamboo-services/bamboo-base-go/defined/env"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// InitMCPServer 创建并配置 MCP Server 实例，注册所有业务工具。
// 返回可用于路由挂载的 StreamableHTTPHandler。
func InitMCPServer(ctx context.Context) http.Handler {
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "lumina",
			Title:   "Lumina · 微明",
			Version: xEnv.GetEnvString("APP_VERSION", "v0.1.0"),
		},
		&mcp.ServerOptions{
			Instructions: "Lumina 是 AI 代码认知与长期记忆的知识中枢。提供项目管理（project_create/get/list）、交互式 Q&A（qa_session_create/push_question/get_answer 等）、跨项目约束推送（pin_push/consume/list/update/peek）以及代码仓库 Wiki 读取（repoWiki_query/repoWiki_list）能力。Wiki 更新由 Git Webhook 自动触发，Agent 无需也无法通过 MCP 触发更新。当用户需要管理项目注册、与用户进行富交互式问答、向其他项目传递依赖约束、或读取代码仓库的 Wiki 文档时，使用这些工具。",
		},
	)

	// 注册 Q&A 模块工具
	RegisterQATools(server)

	// 注册 Project 模块工具
	RegisterProjectTools(server)

	// 注册 Pin 模块工具
	RegisterPinTools(server)

	// 注册 RepoWiki 模块工具
	RegisterRepoWikiTools(server)

	log := xLog.WithName(xLog.NamedINIT)
	log.Info(ctx, "MCP Server initialized with QA, Project, Pin and RepoWiki tools registered")

	// 创建 Streamable HTTP Handler，每个请求使用同一个 Server 实例
	return mcp.NewStreamableHTTPHandler(
		func(_ *http.Request) *mcp.Server { return server },
		nil, // StreamableHTTPOptions
	)
}
