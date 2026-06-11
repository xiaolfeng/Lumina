package mcp

import (
	"context"
	"net/http"

	xEnv "github.com/bamboo-services/bamboo-base-go/defined/env"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
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
		nil, // ServerOptions
	)

	// 注册 Q&A 模块工具
	RegisterQATools(server)

	// 注册 Project 模块工具
	RegisterProjectTools(server)

	log := xLog.WithName(xLog.NamedINIT)
	log.Info(ctx, "MCP Server initialized with QA and Project tools registered")

	// 创建 Streamable HTTP Handler，每个请求使用同一个 Server 实例
	return mcp.NewStreamableHTTPHandler(
		func(_ *http.Request) *mcp.Server { return server },
		nil, // StreamableHTTPOptions
	)
}
