package middleware

import (
	"mime"
	"strings"

	"github.com/gin-gonic/gin"
)

// mcpAcceptHeader 是 MCP Streamable HTTP 协议要求客户端发送的 Accept 头部。
// go-sdk 的 StreamableHTTPHandler 会严格校验 Accept 头必须同时包含
// application/json 和 text/event-stream，否则返回 400 Bad Request。
//
// 部分客户端（如旧版 Claude Code）不会自动发送完整的 Accept 头，
// 导致服务端拒绝请求并陷入重试循环。
// 此中间件在 Accept 头缺失或不完整时自动补全，提升兼容性。
const mcpAcceptHeader = "application/json, text/event-stream"

// MCPCompat 为 MCP Streamable HTTP 端点提供客户端兼容性处理。
//
// 当前处理：
//   - 补全缺失或不完整的 Accept 头，确保包含 application/json 和 text/event-stream。
//     若请求已携带完整 Accept（包含两者），则不修改，保留客户端原始偏好。
func MCPCompat(ctx *gin.Context) {
	accept := ctx.GetHeader("Accept")

	// Accept 为空时，直接设置完整值
	if accept == "" {
		ctx.Request.Header.Set("Accept", mcpAcceptHeader)
		ctx.Next()
		return
	}

	// 检查是否已包含 application/json 和 text/event-stream
	hasJSON, hasSSE := checkAccept(accept)

	// 已经完整，无需修改
	if hasJSON && hasSSE {
		ctx.Next()
		return
	}

	// 缺失部分，追加到已有 Accept 头
	ctx.Request.Header.Set("Accept", accept+", "+mcpAcceptHeader)
	ctx.Next()
}

// checkAccept 检查 Accept 头是否包含 application/json 和 text/event-stream。
// 逻辑与 go-sdk 的 streamableAccepts 保持一致。
func checkAccept(accept string) (hasJSON, hasSSE bool) {
	for _, raw := range strings.Split(accept, ",") {
		// 去除参数部分（如 ;charset=utf-8），使用 mime 解析 media type
		base, _, _ := mime.ParseMediaType(strings.TrimSpace(raw))
		switch strings.ToLower(strings.TrimSpace(base)) {
		case "application/json", "application/*":
			hasJSON = true
		case "text/event-stream", "text/*":
			hasSSE = true
		case "*/*":
			hasJSON = true
			hasSSE = true
		}
	}
	return
}

// compile-time check
var _ gin.HandlerFunc = MCPCompat
