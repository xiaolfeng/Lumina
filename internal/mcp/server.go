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
			Instructions: `Lumina 是面向代码项目的知识与人机协作中枢。工具由模型调用，但必须服从用户当前任务；不要因为工具存在就主动创建项目、会话、预览或跨项目通知。

全局工作流：
1. 需要项目上下文时，先用 project_get（优先 match_path）或 project_list 解析已有 project_id；仅在确认尚未注册且用户任务确实需要时使用 project_create。
2. 需要用户作决定或补充信息时，使用 Q&A：qa_session_list/create → qa_what_question → qa_push_question →（如 supplement=true，先 qa_push_supplement）→ qa_get_answer。不要用猜测替代用户决策。
3. 需要可视化前端文件时，使用 Preview：preview_session_list/create → 逐文件 preview_file_upload → preview_file_list 最终核对。Preview 是评审和沟通媒介，不能替代真实项目文件的实现、测试与交付。
4. Preview 核对通过后分两种交付：直接视觉评审时，优先用 MCP 客户端原生浏览器/打开链接能力访问返回的绝对 preview_url；没有该能力时向用户提供可点击 URL。作为 Q&A 补充时，调用 qa_push_supplement，content_type=preview，content 原样使用 Preview 返回的 qa_supplement.content，然后再 qa_get_answer。
5. Preview 的 hash 仅用于网页 URL；Q&A preview 引用必须使用 session_id + file_id。不得自行拼错字段，也不得给 JSON 加 Markdown 围栏。
6. RepoWiki MCP 工具只读；Wiki 更新由 Git Webhook 驱动。Pin 仅用于明确的跨项目约束，消费前可先 pin_peek/list 核对。

所有工具都应按其 description 的「何时调用、不要调用、限制、返回与下一步」执行。工具返回 isError=true 时，先根据错误修正输入或停止流程，不得把失败结果当作成功。`,
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

	// 注册 Preview 模块工具
	RegisterPreviewTools(server)

	log := xLog.WithName(xLog.NamedINIT)
	log.Info(ctx, "MCP Server initialized with QA, Project, Pin, RepoWiki and Preview tools registered")

	// 创建 Streamable HTTP Handler，每个请求使用同一个 Server 实例
	return mcp.NewStreamableHTTPHandler(
		func(_ *http.Request) *mcp.Server { return server },
		nil, // StreamableHTTPOptions
	)
}
