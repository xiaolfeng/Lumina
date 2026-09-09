package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	apiWorkspace "github.com/xiaolfeng/Lumina/api/workspace"
	"github.com/xiaolfeng/Lumina/internal/logic"
)

var workspaceLogic *logic.WorkspaceLogic

// SetWorkspaceLogic 设置 WorkspaceLogic 实例，供 MCP 工具处理器使用。
func SetWorkspaceLogic(l *logic.WorkspaceLogic) {
	workspaceLogic = l
}

var workspaceToolDefs = []struct {
	name        string
	description string
	inputSchema map[string]any
}{
	{
		name:        "workspace_list",
		description: `列出全部工作空间。用 match_path / name 解析项目之前先列出并选定空间；不要假设只有一个空间。默认空间 slug 为 default。`,
		inputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"page": map[string]any{"type": "integer", "description": "页码，从 1 开始，默认 1"},
				"size": map[string]any{"type": "integer", "description": "每页数量，默认 50，最大 100"},
			},
		},
	},
	{
		name:        "workspace_get",
		description: `按空间雪花 ID 或 slug 查看工作空间详情。优先级：workspace_id > slug。默认空间 slug 为 default。`,
		inputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"workspace_id": map[string]any{"type": "string", "description": "空间雪花 ID"},
				"slug":         map[string]any{"type": "string", "description": "空间标识，默认空间为 default"},
			},
		},
	},
}

func handleWorkspaceList(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if workspaceLogic == nil {
		return textResult("WorkspaceLogic 未初始化，请联系管理员"), nil
	}
	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return textResult(errMsg), nil
	}
	page := 1
	size := 50
	if p, ok := args["page"].(float64); ok && p > 0 {
		page = int(p)
	}
	if s, ok := args["size"].(float64); ok && s > 0 && s <= 100 {
		size = int(s)
	}
	resp, xErr := workspaceLogic.List(context.Background(), page, size)
	if xErr != nil {
		return textResult(fmt.Sprintf("获取空间列表失败: %s", xErr.Error())), nil
	}
	totalPages := (resp.Total + int64(size) - 1) / int64(size)
	result := fmt.Sprintf("空间列表（共 %d 个，第 %d/%d 页）：\n\n", resp.Total, page, totalPages)
	for i, item := range resp.Items {
		line := fmt.Sprintf("%d. [%s] %s（slug: %s）", i+1, item.ID, item.Name, item.Slug)
		if item.IsDefault {
			line += " [默认]"
		}
		result += line + "\n"
	}
	if len(resp.Items) == 0 {
		result += "（暂无空间）\n"
	}
	return textResult(result), nil
}

func handleWorkspaceGet(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if workspaceLogic == nil {
		return textResult("WorkspaceLogic 未初始化，请联系管理员"), nil
	}
	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return textResult(errMsg), nil
	}
	workspaceID, _ := args["workspace_id"].(string)
	slug, _ := args["slug"].(string)
	if workspaceID == "" && slug == "" {
		return textResult("请指定 workspace_id 或 slug"), nil
	}

	var resp *apiWorkspace.WorkspaceResponse
	if workspaceID != "" {
		got, xErr := workspaceLogic.GetByID(context.Background(), workspaceID)
		if xErr != nil {
			return textResult(fmt.Sprintf("查询空间失败: %s", xErr.Error())), nil
		}
		resp = got
	} else {
		got, xErr := workspaceLogic.GetBySlug(context.Background(), slug)
		if xErr != nil {
			return textResult(fmt.Sprintf("查询空间失败: %s", xErr.Error())), nil
		}
		resp = got
	}
	return textResult(formatWorkspaceDetail(resp)), nil
}

func formatWorkspaceDetail(resp *apiWorkspace.WorkspaceResponse) string {
	defaultLabel := "否"
	if resp.IsDefault {
		defaultLabel = "是"
	}
	icon := resp.Icon
	if icon == "" {
		icon = "（无）"
	}
	desc := resp.Description
	if desc == "" {
		desc = "（无）"
	}
	return fmt.Sprintf(`空间详情：

ID: %s
名称: %s
标识: %s
默认空间: %s
图标: %s
描述: %s
创建时间: %s
更新时间: %s`,
		resp.ID, resp.Name, resp.Slug, defaultLabel, icon, desc, resp.CreatedAt, resp.UpdatedAt)
}

// RegisterWorkspaceTools 将 Workspace 模块的 2 个 MCP 工具注册到 Server。
func RegisterWorkspaceTools(server *mcp.Server) {
	for _, def := range workspaceToolDefs {
		schemaBytes, _ := json.Marshal(def.inputSchema)
		tool := &mcp.Tool{
			Name: def.name, Description: def.description, InputSchema: json.RawMessage(schemaBytes),
		}
		var handler mcp.ToolHandler
		switch def.name {
		case "workspace_list":
			handler = handleWorkspaceList
		case "workspace_get":
			handler = handleWorkspaceGet
		default:
			handler = stubToolHandler(def.name)
		}
		server.AddTool(tool, handler)
	}
}
