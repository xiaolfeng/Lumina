package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/xiaolfeng/Lumina/internal/logic"
)

// previewLogic 保存 PreviewLogic 实例，供 MCP 工具处理器使用。
var previewLogic *logic.PreviewLogic

// SetPreviewLogic 设置 PreviewLogic 实例，供 MCP 工具处理器使用。
func SetPreviewLogic(l *logic.PreviewLogic) {
	previewLogic = l
}

// previewToolDefs 定义 Preview 模块的全部 MCP 工具。
var previewToolDefs = []struct {
	name        string
	description string
	inputSchema map[string]any
}{
	{
		name: "preview_session_create",
		description: `创建一个前端可视化预览会话（活动工作区），用于承载多个前端文件的实时渲染。

触发场景：Agent 需要在某个 Project 上下文中为前端文件创建预览工作区时使用。一个 Project 可开多个预览会话（多工作区并存）。

创建成功后返回 session_id 与 hash。hash 用于构造预览页访问链接（/preview?session={hash}）以及后续 Q&A supplement 的 preview 类型引用。

推荐流程：
1. 用 preview_session_create 创建会话拿到 session_id
2. 用 preview_file_upload 上传前端文件（html/js/css，扁平单层，多次调用）
3. 将 session_id 与 file_id 通过 Q&A supplement（content_type=preview）关联给用户预览`,
		inputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id": map[string]any{
					"type":        "string",
					"description": "关联的项目 ID（必填，雪花 ID 字符串）",
				},
				"title": map[string]any{
					"type":        "string",
					"description": "会话标题，可选（默认「未命名预览」）",
				},
			},
			"required": []string{"project_id"},
		},
	},
	{
		name: "preview_session_list",
		description: `分页获取指定项目下的预览会话列表。

触发场景：需要查看某个项目已有哪些预览工作区、或复用已有会话继续上传文件时使用。`,
		inputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id": map[string]any{
					"type":        "string",
					"description": "项目 ID（必填，雪花 ID 字符串）",
				},
				"page": map[string]any{
					"type":        "integer",
					"description": "页码（从 1 开始，默认 1）",
				},
				"size": map[string]any{
					"type":        "integer",
					"description": "每页数量（默认 10，最大 100）",
				},
			},
			"required": []string{"project_id"},
		},
	},
	{
		name: "preview_file_upload",
		description: `上传或覆写一个前端文件到指定预览会话（扁平单层）。

触发场景：Agent 生成前端代码后，将 html/js/css 等文件推送到预览会话供用户可视化查看。

约束：
  - 文件名为扁平单层（仅当前目录），禁止包含路径分隔符（/、\\）与目录穿越（..）
  - 同 Session 同文件名重复上传会覆盖旧内容
  - 单文件大小上限 256KB
  - 多文件上传需多次调用本工具（一次一个文件）

上传成功返回 file_id，可配合 Q&A supplement（content_type=preview）引用到具体渲染内容。`,
		inputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"session_id": map[string]any{
					"type":        "string",
					"description": "目标预览会话 ID（必填，雪花 ID 字符串）",
				},
				"filename": map[string]any{
					"type":        "string",
					"description": "文件名（必填，如 index.html、style.css、app.js）",
				},
				"content": map[string]any{
					"type":        "string",
					"description": "文件内容（必填）",
				},
			},
			"required": []string{"session_id", "filename", "content"},
		},
	},
	{
		name: "preview_file_list",
		description: `获取指定预览会话的全部文件列表（按文件名升序）。

触发场景：查看某个预览会话当前已上传哪些文件、或确认文件是否上传成功时使用。`,
		inputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"session_id": map[string]any{
					"type":        "string",
					"description": "目标预览会话 ID（必填，雪花 ID 字符串）",
				},
			},
			"required": []string{"session_id"},
		},
	},
	{
		name: "preview_file_get",
		description: `获取指定预览会话中某个文件的完整内容（提取代码）。

触发场景：需要读取预览工作区中已上传文件的源码内容时使用，例如基于已有代码继续迭代、将 preview 中已确立的代码作为规范基准提取回上下文。

推荐流程：
1. 用 preview_file_list 查看会话内有哪些文件（拿到文件名）
2. 用 preview_file_get 提取指定文件的完整代码`,
		inputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"session_id": map[string]any{
					"type":        "string",
					"description": "目标预览会话 ID（必填，雪花 ID 字符串）",
				},
				"filename": map[string]any{
					"type":        "string",
					"description": "文件名（必填，如 index.html、style.css、app.js）",
				},
			},
			"required": []string{"session_id", "filename"},
		},
	},
}

// ─── Tool Handlers ──────────────────────────────────────────────────────

// handlePreviewSessionCreate 创建预览会话
func handlePreviewSessionCreate(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if previewLogic == nil {
		return textResult("PreviewLogic 未初始化，请联系管理员"), nil
	}
	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return textResult(errMsg), nil
	}

	projectIDStr, _ := args["project_id"].(string)
	if projectIDStr == "" {
		return textResult("缺少必填参数: project_id"), nil
	}
	projectID, err := xSnowflake.ParseSnowflakeID(projectIDStr)
	if err != nil {
		return textResult(fmt.Sprintf("无效的 project_id: %s", projectIDStr)), nil
	}

	title, _ := args["title"].(string)

	resp, xErr := previewLogic.CreateSession(context.Background(), projectID, title)
	if xErr != nil {
		return textResult(fmt.Sprintf("创建预览会话失败: %s", xErr.Error())), nil
	}

	return textResult(fmt.Sprintf(`预览会话创建成功！

session_id: %s
hash: %s
标题: %s
预览链接: /preview?session=%s`,
		resp.ID, resp.Hash, resp.Title, resp.Hash)), nil
}

// handlePreviewSessionList 分页获取预览会话列表
func handlePreviewSessionList(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if previewLogic == nil {
		return textResult("PreviewLogic 未初始化，请联系管理员"), nil
	}
	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return textResult(errMsg), nil
	}

	projectIDStr, _ := args["project_id"].(string)
	if projectIDStr == "" {
		return textResult("缺少必填参数: project_id"), nil
	}
	projectID, err := xSnowflake.ParseSnowflakeID(projectIDStr)
	if err != nil {
		return textResult(fmt.Sprintf("无效的 project_id: %s", projectIDStr)), nil
	}

	page := 1
	size := 10
	if p, ok := args["page"].(float64); ok && p > 0 {
		page = int(p)
	}
	if s, ok := args["size"].(float64); ok && s > 0 && s <= 100 {
		size = int(s)
	}

	resp, xErr := previewLogic.ListSessions(context.Background(), projectID, page, size)
	if xErr != nil {
		return textResult(fmt.Sprintf("获取预览会话列表失败: %s", xErr.Error())), nil
	}

	totalPages := (resp.Total + int64(size) - 1) / int64(size)
	result := fmt.Sprintf("预览会话列表（共 %d 个，第 %d/%d 页）：\n\n", resp.Total, page, totalPages)
	for i, s := range resp.Items {
		result += fmt.Sprintf("%d. [%s] %s | hash: %s | 状态: %s\n", i+1, s.ID, s.Title, s.Hash, s.Status)
	}
	if len(resp.Items) == 0 {
		result += "（暂无预览会话）\n"
	}
	return textResult(result), nil
}

// handlePreviewFileUpload 上传预览文件
func handlePreviewFileUpload(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if previewLogic == nil {
		return textResult("PreviewLogic 未初始化，请联系管理员"), nil
	}
	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return textResult(errMsg), nil
	}

	sessionIDStr, _ := args["session_id"].(string)
	if sessionIDStr == "" {
		return textResult("缺少必填参数: session_id"), nil
	}
	sessionID, err := xSnowflake.ParseSnowflakeID(sessionIDStr)
	if err != nil {
		return textResult(fmt.Sprintf("无效的 session_id: %s", sessionIDStr)), nil
	}

	filename, _ := args["filename"].(string)
	if filename == "" {
		return textResult("缺少必填参数: filename"), nil
	}
	content, _ := args["content"].(string)

	resp, xErr := previewLogic.UploadFile(context.Background(), sessionID, filename, content)
	if xErr != nil {
		return textResult(fmt.Sprintf("上传预览文件失败: %s", xErr.Error())), nil
	}

	return textResult(fmt.Sprintf(`预览文件上传成功！

file_id: %s
文件名: %s
类型: %s
大小: %d 字节`,
		resp.ID, resp.Filename, resp.MimeType, resp.Size)), nil
}

// handlePreviewFileList 获取预览文件列表
func handlePreviewFileList(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if previewLogic == nil {
		return textResult("PreviewLogic 未初始化，请联系管理员"), nil
	}
	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return textResult(errMsg), nil
	}

	sessionIDStr, _ := args["session_id"].(string)
	if sessionIDStr == "" {
		return textResult("缺少必填参数: session_id"), nil
	}
	sessionID, err := xSnowflake.ParseSnowflakeID(sessionIDStr)
	if err != nil {
		return textResult(fmt.Sprintf("无效的 session_id: %s", sessionIDStr)), nil
	}

	files, xErr := previewLogic.ListFiles(context.Background(), sessionID)
	if xErr != nil {
		return textResult(fmt.Sprintf("获取预览文件列表失败: %s", xErr.Error())), nil
	}

	result := fmt.Sprintf("预览文件列表（共 %d 个）：\n\n", len(files))
	for i, f := range files {
		result += fmt.Sprintf("%d. %s | %s | %d 字节\n", i+1, f.Filename, f.MimeType, f.Size)
	}
	if len(files) == 0 {
		result += "（暂无文件）\n"
	}
	return textResult(result), nil
}

// handlePreviewFileGet 获取预览文件完整内容
func handlePreviewFileGet(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if previewLogic == nil {
		return textResult("PreviewLogic 未初始化，请联系管理员"), nil
	}
	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return textResult(errMsg), nil
	}

	sessionIDStr, _ := args["session_id"].(string)
	if sessionIDStr == "" {
		return textResult("缺少必填参数: session_id"), nil
	}
	sessionID, err := xSnowflake.ParseSnowflakeID(sessionIDStr)
	if err != nil {
		return textResult(fmt.Sprintf("无效的 session_id: %s", sessionIDStr)), nil
	}

	filename, _ := args["filename"].(string)
	if filename == "" {
		return textResult("缺少必填参数: filename"), nil
	}

	resp, xErr := previewLogic.GetFileContentBySession(context.Background(), sessionID, filename)
	if xErr != nil {
		return textResult(fmt.Sprintf("获取预览文件内容失败: %s", xErr.Error())), nil
	}

	return textResult(fmt.Sprintf("=== %s（%s）===\n\n%s", resp.Filename, resp.MimeType, resp.Content)), nil
}

// RegisterPreviewTools 将 Preview 模块的 5 个 MCP 工具注册到 Server。
func RegisterPreviewTools(server *mcp.Server) {
	for _, def := range previewToolDefs {
		schemaBytes, _ := json.Marshal(def.inputSchema)
		tool := &mcp.Tool{
			Name:        def.name,
			Description: def.description,
			InputSchema: json.RawMessage(schemaBytes),
		}

		var handler mcp.ToolHandler
		switch def.name {
		case "preview_session_create":
			handler = handlePreviewSessionCreate
		case "preview_session_list":
			handler = handlePreviewSessionList
		case "preview_file_upload":
			handler = handlePreviewFileUpload
		case "preview_file_list":
			handler = handlePreviewFileList
		case "preview_file_get":
			handler = handlePreviewFileGet
		default:
			handler = stubToolHandler(def.name)
		}

		server.AddTool(tool, handler)
	}
}
