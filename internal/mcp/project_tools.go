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
		description: `创建一个新项目，将代码库注册到 Lumina 系统。后续 Q&A 等模块的会话需要关联到具体项目。

触发场景：用户提到"新项目"、"初始化项目"，或 Agent 发现当前工作目录尚未注册为项目时主动建议创建。

match_path 填写项目根目录的绝对路径。Agent 应先执行 pwd（Unix/macOS）或 cd（Windows）获取当前工作目录绝对路径，将其作为 match_path 传入。AI 会根据该路径前缀自动将文件匹配到对应项目。

创建成功后返回 project_id（雪花 ID）和项目基本信息，后续通过 project_id 或 name 查询详情。`,
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
		description: `查询项目详情。支持三种查询方式，按优先级依次取第一个非空条件：project_id > name > match_path。

触发场景：需要获取 project_id、确认项目是否存在、查看项目的完整配置信息时调用。

match_path 匹配逻辑：项目的 MatchPath 字段中任一元素是查询路径的前缀时命中。例如项目 MatchPath=["/home/user/Lumina"] 可匹配查询路径 "/home/user/Lumina/src/main.go"。

返回项目完整信息（ID、名称、别名、路径匹配列表、描述等），若不存在会返回错误提示。`,
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
		description: `获取项目列表。支持按路径前缀过滤，不传过滤条件时返回全部项目（分页）。

触发场景：Agent 不知道 project_id 但需要找到对应项目时，可用当前工作目录的绝对路径作为 match_path 过滤匹配。

match_path 过滤模式：在应用层执行，匹配逻辑为双向前缀匹配——项目的 MatchPath 任一元素是查询路径的前缀，或查询路径是 MatchPath 元素的前缀时命中。使用过滤时不分页，返回全部匹配结果。

每项返回项目 ID、名称、别名和路径匹配列表。需要查看完整详情时使用 project_get。`,
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
	if errMsg := checkParseError(args); errMsg != "" {
		return textResult(errMsg), nil
	}
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
路径匹配: %s
描述: %s`,
		resp.ID, resp.Name, resp.AliasName, formatPathList(resp.MatchPath), resp.Description)), nil
}

// handleProjectGet 查询项目详情
func handleProjectGet(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if projectLogic == nil {
		return textResult("ProjectLogic 未初始化，请联系管理员"), nil
	}
	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return textResult(errMsg), nil
	}

	if projectID, _ := args["project_id"].(string); projectID != "" {
		resp, xErr := projectLogic.GetByID(context.Background(), projectID)
		if xErr != nil {
			return textResult(fmt.Sprintf("查询项目失败: %s", xErr.Error())), nil
		}
		return textResult(formatProjectDetail(resp)), nil
	}
	if name, _ := args["name"].(string); name != "" {
		resp, xErr := projectLogic.GetByName(context.Background(), name)
		if xErr != nil {
			return textResult(fmt.Sprintf("查询项目失败: %s", xErr.Error())), nil
		}
		return textResult(formatProjectDetail(resp)), nil
	}
	if mp, _ := args["match_path"].(string); mp != "" {
		resp, xErr := projectLogic.GetByMatchPath(context.Background(), mp)
		if xErr != nil {
			return textResult(fmt.Sprintf("查询项目失败: %s", xErr.Error())), nil
		}
		return textResult(formatProjectDetail(resp)), nil
	}
	return textResult("请至少指定一个查询条件: project_id、name 或 match_path"), nil
}

// handleProjectList 获取项目列表
func handleProjectList(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if projectLogic == nil {
		return textResult("ProjectLogic 未初始化，请联系管理员"), nil
	}
	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return textResult(errMsg), nil
	}
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

	var items []apiProject.ProjectResponse
	if matchPathFilter != "" {
		for _, p := range resp.Items {
			for _, mp := range p.MatchPath {
				if strings.HasPrefix(matchPathFilter, mp) || strings.HasPrefix(mp, matchPathFilter) {
					items = append(items, p)
					break
				}
			}
		}
	} else {
		items = resp.Items
	}

	var result string
	if matchPathFilter != "" {
		result = fmt.Sprintf("项目列表（路径匹配「%s」，共 %d 个）：\n\n", matchPathFilter, len(items))
	} else {
		totalPages := (resp.Total + int64(size) - 1) / int64(size)
		result = fmt.Sprintf("项目列表（共 %d 个，第 %d/%d 页）：\n\n", resp.Total, page, totalPages)
	}
	for i, p := range items {
		result += fmt.Sprintf("%d. [%s] %s", i+1, p.ID, p.Name)
		if p.AliasName != "" {
			result += fmt.Sprintf("（%s）", p.AliasName)
		}
		result += fmt.Sprintf(" | 路径: %s\n", formatPathList(p.MatchPath))
	}
	if len(items) == 0 {
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

// formatPathList 将路径列表格式化为可读字符串
func formatPathList(paths []string) string {
	if len(paths) == 0 {
		return "（无）"
	}
	return strings.Join(paths, ", ")
}

// formatProjectDetail 格式化项目详情为可读字符串
func formatProjectDetail(resp *apiProject.ProjectResponse) string {
	return fmt.Sprintf(`项目详情：

ID: %s
名称: %s
别名: %s
路径匹配: %s
描述: %s
创建时间: %s
更新时间: %s`,
		resp.ID, resp.Name, resp.AliasName, formatPathList(resp.MatchPath), resp.Description, resp.CreatedAt, resp.UpdatedAt)
}
