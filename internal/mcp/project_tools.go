package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	apiProject "github.com/xiaolfeng/Lumina/api/project"
	"github.com/xiaolfeng/Lumina/internal/logic"
)

// projectLogic 保存 ProjectLogic 实例，供 MCP 工具处理器使用。
var projectLogic *logic.ProjectLogic

// SetProjectLogic 设置 ProjectLogic 实例，供 MCP 工具处理器使用。
func SetProjectLogic(l *logic.ProjectLogic) {
	projectLogic = l
}

// projectToolDefs 定义 Project 模块的全部 MCP 工具。
var projectToolDefs = []struct {
	name        string
	description string
	inputSchema map[string]any
}{
	{
		name: "project_create",
		description: `创建一个新项目。项目用于标识和组织代码库，AI Agent 通过该项目关联代码上下文。

必填参数：
- name: 项目名称（全局唯一）
- match_path: 路径匹配列表（用于 AI 根据文件路径自动识别项目）

可选参数：
- alias_name: 项目别名（人类可读的中文/英文名称）
- description: 项目描述

创建成功后返回 project_id 和项目基本信息。`,
		inputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name":        map[string]any{"type": "string", "description": "项目名称（全局唯一）"},
				"match_path":  map[string]any{"type": "array", "description": "路径匹配列表，AI 可根据文件路径自动匹配到该项目", "items": map[string]any{"type": "string"}},
				"alias_name":  map[string]any{"type": "string", "description": "项目别名（人类可读，可选）"},
				"description": map[string]any{"type": "string", "description": "项目描述（可选）"},
			},
			"required": []string{"name", "match_path"},
		},
	},
	{
		name: "project_get",
		description: `查询项目详情。支持三种查询方式（按优先级从高到低）：

1. project_id: 通过项目 ID 精确查询
2. name: 通过项目名称精确查询
3. match_path: 通过路径前缀匹配查询

至少指定一个查询条件。多个条件同时传入时按优先级取第一个非空条件。

match_path 查询示例：
- MatchPath=["/home/user/Lumina"] 可匹配 "/home/user/Lumina/src/main.go"
- 匹配逻辑：match_path 中的任一元素是查询路径的前缀

返回项目完整信息（ID、名称、别名、路径匹配列表、描述等）。`,
		inputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id": map[string]any{"type": "string", "description": "项目 ID（雪花 ID 字符串）"},
				"name":       map[string]any{"type": "string", "description": "项目名称"},
				"match_path": map[string]any{"type": "string", "description": "路径前缀，匹配 MatchPath 中任一元素是该路径前缀的项目"},
			},
		},
	},
	{
		name: "project_list",
		description: `获取项目列表。不传 match_path 时返回全部项目。

支持分页：page（页码，从 1 开始）、size（每页数量，默认 20，最大 100）。

可选 match_path 参数用于按路径前缀过滤项目。`,
		inputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"match_path": map[string]any{"type": "string", "description": "路径前缀过滤（可选，不传则返回全部）"},
				"page":       map[string]any{"type": "integer", "description": "页码（从 1 开始，默认 1）"},
				"size":       map[string]any{"type": "integer", "description": "每页数量（默认 20，最大 100）"},
			},
		},
	},
}

// handleProjectCreate 创建项目
func handleProjectCreate(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if projectLogic == nil {
		return textResult("ProjectLogic 未初始化，请联系管理员"), nil
	}
	args := parseArgs(req.Params.Arguments)
	name, _ := args["name"].(string)
	if name == "" {
		return textResult("缺少必填参数: name"), nil
	}
	var matchPath []string
	if mp, ok := args["match_path"].([]interface{}); ok {
		for _, item := range mp {
			if s, ok := item.(string); ok {
				matchPath = append(matchPath, s)
			}
		}
	}
	if len(matchPath) == 0 {
		return textResult("缺少必填参数: match_path"), nil
	}
	aliasName, _ := args["alias_name"].(string)
	description, _ := args["description"].(string)
	apiReq := &apiProject.CreateProjectRequest{
		Name: name, AliasName: aliasName, MatchPath: matchPath, Description: description,
	}
	resp, xErr := projectLogic.Create(context.Background(), apiReq)
	if xErr != nil {
		return textResult(fmt.Sprintf("创建项目失败: %s", xErr.Error())), nil
	}
	return textResult(fmt.Sprintf(`项目创建成功！

项目 ID: %s
名称: %s
别名: %s
路径匹配: %v
描述: %s`,
		resp.ID, resp.Name, resp.AliasName, resp.MatchPath, resp.Description)), nil
}

// handleProjectGet 查询项目详情
func handleProjectGet(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if projectLogic == nil {
		return textResult("ProjectLogic 未初始化，请联系管理员"), nil
	}
	args := parseArgs(req.Params.Arguments)
	if projectID, _ := args["project_id"].(string); projectID != "" {
		resp, xErr := projectLogic.GetByID(context.Background(), projectID)
		if xErr != nil {
			return textResult(fmt.Sprintf("查询项目失败: %s", xErr.Error())), nil
		}
		detail := fmt.Sprintf(`项目详情：

ID: %s
名称: %s
别名: %s
路径匹配: %v
描述: %s
创建时间: %s
更新时间: %s`,
			resp.ID, resp.Name, resp.AliasName, resp.MatchPath, resp.Description, resp.CreatedAt, resp.UpdatedAt)
		return textResult(detail), nil
	}
	if name, _ := args["name"].(string); name != "" {
		resp, xErr := projectLogic.GetByName(context.Background(), name)
		if xErr != nil {
			return textResult(fmt.Sprintf("查询项目失败: %s", xErr.Error())), nil
		}
		detail := fmt.Sprintf(`项目详情：

ID: %s
名称: %s
别名: %s
路径匹配: %v
描述: %s
创建时间: %s
更新时间: %s`,
			resp.ID, resp.Name, resp.AliasName, resp.MatchPath, resp.Description, resp.CreatedAt, resp.UpdatedAt)
		return textResult(detail), nil
	}
	if mp, _ := args["match_path"].(string); mp != "" {
		resp, xErr := projectLogic.GetByMatchPath(context.Background(), mp)
		if xErr != nil {
			return textResult(fmt.Sprintf("查询项目失败: %s", xErr.Error())), nil
		}
		detail := fmt.Sprintf(`项目详情：

ID: %s
名称: %s
别名: %s
路径匹配: %v
描述: %s
创建时间: %s
更新时间: %s`,
			resp.ID, resp.Name, resp.AliasName, resp.MatchPath, resp.Description, resp.CreatedAt, resp.UpdatedAt)
		return textResult(detail), nil
	}
	return textResult("请至少指定一个查询条件: project_id、name 或 match_path"), nil
}

// handleProjectList 获取项目列表
func handleProjectList(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if projectLogic == nil {
		return textResult("ProjectLogic 未初始化，请联系管理员"), nil
	}
	args := parseArgs(req.Params.Arguments)
	page := 1
	size := 20
	if p, ok := args["page"].(float64); ok && p > 0 {
		page = int(p)
	}
	if s, ok := args["size"].(float64); ok && s > 0 && s <= 100 {
		size = int(s)
	}
	matchPathFilter, _ := args["match_path"].(string)
	resp, xErr := projectLogic.List(context.Background(), page, size)
	if xErr != nil {
		return textResult(fmt.Sprintf("获取项目列表失败: %s", xErr.Error())), nil
	}
	if matchPathFilter != "" {
		var filtered []apiProject.ProjectResponse
		for _, p := range resp.Items {
			for _, mp := range p.MatchPath {
				if strings.HasPrefix(matchPathFilter, mp) || strings.HasPrefix(mp, matchPathFilter) {
					filtered = append(filtered, p)
					break
				}
			}
		}
		resp.Items = filtered
		resp.Total = int64(len(filtered))
	}
	totalPages := (resp.Total + int64(size) - 1) / int64(size)
	result := fmt.Sprintf(`项目列表（共 %d 个，第 %d/%d 页）：

`, resp.Total, page, totalPages)
	for i, p := range resp.Items {
		result += fmt.Sprintf("%d. [%s] %s", i+1, p.ID, p.Name)
		if p.AliasName != "" {
			result += fmt.Sprintf("（%s）", p.AliasName)
		}
		result += fmt.Sprintf(" | 路径: %v\n", p.MatchPath)
	}
	if len(resp.Items) == 0 {
		result += "（暂无项目）\n"
	}
	return textResult(result), nil
}

// RegisterProjectTools 将 Project 模块的 3 个 MCP 工具注册到 Server。
func RegisterProjectTools(server *mcp.Server) {
	for _, def := range projectToolDefs {
		schemaBytes, _ := json.Marshal(def.inputSchema)
		tool := &mcp.Tool{
			Name: def.name, Description: def.description, InputSchema: json.RawMessage(schemaBytes),
		}
		var handler mcp.ToolHandler
		switch def.name {
		case "project_create":
			handler = handleProjectCreate
		case "project_get":
			handler = handleProjectGet
		case "project_list":
			handler = handleProjectList
		default:
			handler = stubToolHandler(def.name)
		}
		server.AddTool(tool, handler)
	}
}
