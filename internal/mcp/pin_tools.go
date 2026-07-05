package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	apiPin "github.com/xiaolfeng/Lumina/api/pin"
	"github.com/xiaolfeng/Lumina/internal/logic"
)

// pinLogic 保存 PinLogic 实例，供 MCP 工具处理器使用。
var pinLogic *logic.PinLogic

// SetPinLogic 设置 PinLogic 实例，供 MCP 工具处理器使用。
func SetPinLogic(l *logic.PinLogic) {
	pinLogic = l
}

// pinToolDefs 定义 Pin 模块的全部 MCP 工具。
var pinToolDefs = []struct {
	name        string
	description string
	inputSchema map[string]any
}{
	{
		name: "pin_push",
		description: `推送跨项目约束到目标项目。Agent 用于向另一个项目传递依赖约束、注意事项或接口变更通知。

触发场景：当你在一个项目中发现了会影响另一个项目的约定、接口变更、依赖升级、兼容性约束等信息时，使用本工具将约束推送到目标项目队列，目标项目的消费方会按 FIFO 顺序处理。

to_project_name 支持两种输入：项目名称/别名（如 "lumina"、"Lumina"）或雪花 ID 字符串（如 "1234567890123456789"）。Logic 层的 ResolveProject 会自动识别并解析。

from_project_id 为可选，表示约束来源项目；不传时该约束将被标记为无来源。`,
		inputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"title": map[string]any{
					"type":        "string",
					"description": "约束标题（简洁概述）",
				},
				"content": map[string]any{
					"type":        "string",
					"description": "约束详细内容（Markdown）",
				},
				"priority": map[string]any{
					"type":        "string",
					"description": "优先级：high（高）/ medium（中）/ low（低）",
					"enum":        []string{"high", "medium", "low"},
				},
				"to_project_name": map[string]any{
					"type":        "string",
					"description": "目标项目名称/别名或雪花 ID 字符串",
				},
				"category": map[string]any{
					"type":        "string",
					"description": "约束分类，默认 notice",
					"enum":        []string{"notice", "dependency", "api_change", "other"},
				},
				"from_project_id": map[string]any{
					"type":        "string",
					"description": "来源项目 ID（雪花 ID 字符串或别名，可选）",
				},
			},
			"required": []string{"title", "content", "priority", "to_project_name"},
		},
	},
	{
		name: "pin_consume",
		description: `消费目标项目队列中的约束（FIFO 先进先出）。Agent 用于按序处理待处理约束。

支持两种消费模式：
  - 不传 id 时：消费队首约束（最旧的 pending，FIFO）
  - 传 id 时：精确消费指定 ID 的约束（仅当该约束归属此项目且状态为 pending 时成功）

project_name 支持名称/别名或雪花 ID 字符串，由后端自动解析。

队列已空时返回「暂无待处理约束」提示，不视为错误。`,
		inputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_name": map[string]any{
					"type":        "string",
					"description": "目标项目名称/别名或雪花 ID 字符串",
				},
				"id": map[string]any{
					"type":        "string",
					"description": "精确消费的 Pin ID（雪花 ID 字符串）。不传则消费队首（FIFO）",
				},
				"category": map[string]any{
					"type":        "string",
					"description": "分类筛选（可选，仅 FIFO 模式下生效）",
					"enum":        []string{"notice", "dependency", "api_change", "other"},
				},
				"priority": map[string]any{
					"type":        "string",
					"description": "优先级筛选（可选，仅 FIFO 模式下生效）",
					"enum":        []string{"high", "medium", "low"},
				},
			},
			"required": []string{"project_name"},
		},
	},
	{
		name: "pin_list",
		description: `列出目标项目的约束列表，支持状态/分类/优先级筛选和分页。

project_name 支持名称/别名或雪花 ID 字符串。

默认返回 pending 状态的约束；可指定 status 查看 consumed（已消费）等历史约束。
排序为 FIFO（创建时间升序），便于消费场景查看队列顺序。`,
		inputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_name": map[string]any{
					"type":        "string",
					"description": "目标项目名称/别名或雪花 ID 字符串",
				},
				"status": map[string]any{
					"type":        "string",
					"description": "状态筛选（可选）：pending / consumed",
					"enum":        []string{"pending", "consumed"},
				},
				"category": map[string]any{
					"type":        "string",
					"description": "分类筛选（可选）",
					"enum":        []string{"notice", "dependency", "api_change", "other"},
				},
				"priority": map[string]any{
					"type":        "string",
					"description": "优先级筛选（可选）",
					"enum":        []string{"high", "medium", "low"},
				},
				"from_project_id": map[string]any{
					"type":        "string",
					"description": "来源项目 ID 筛选（可选）",
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
			"required": []string{"project_name"},
		},
	},
	{
		name: "pin_update",
		description: `更新约束的元数据（优先级、分类）。

状态不可通过此工具修改——约束状态只能通过 pin_consume 流转（pending → consumed）。
这保证了消费动作的原子性和唯一入口，避免误操作。

至少提供 priority 或 category 中的一个，否则等同于只读查询。`,
		inputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id": map[string]any{
					"type":        "string",
					"description": "待更新的 Pin ID（雪花 ID 字符串）",
				},
				"priority": map[string]any{
					"type":        "string",
					"description": "新优先级（可选）",
					"enum":        []string{"high", "medium", "low"},
				},
				"category": map[string]any{
					"type":        "string",
					"description": "新分类（可选）",
					"enum":        []string{"notice", "dependency", "api_change", "other"},
				},
			},
			"required": []string{"id"},
		},
	},
	{
		name: "pin_peek",
		description: `查看指定约束的详情（只读，不改变状态）。已消费的约束也可查看。

触发场景：需要在消费前预览约束完整内容、或在消费后回查历史约束时使用。
返回完整字段：ID、标题、内容、分类、状态、优先级、来源项目、目标项目、消费时间（如已消费）、创建/更新时间。`,
		inputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id": map[string]any{
					"type":        "string",
					"description": "要查看的 Pin ID（雪花 ID 字符串）",
				},
			},
			"required": []string{"id"},
		},
	},
}

// ─── Tool Handlers ──────────────────────────────────────────────────────

// handlePinPush 推送跨项目约束
func handlePinPush(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if pinLogic == nil {
		return textResult("PinLogic 未初始化，请联系管理员"), nil
	}
	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return textResult(errMsg), nil
	}
	title, _ := args["title"].(string)
	if title == "" {
		return textResult("缺少必填参数: title"), nil
	}
	content, _ := args["content"].(string)
	if content == "" {
		return textResult("缺少必填参数: content"), nil
	}
	priority, _ := args["priority"].(string)
	if priority == "" {
		return textResult("缺少必填参数: priority"), nil
	}
	toProjectName, _ := args["to_project_name"].(string)
	if toProjectName == "" {
		return textResult("缺少必填参数: to_project_name"), nil
	}
	category, _ := args["category"].(string)
	fromProjectID, _ := args["from_project_id"].(string)

	apiReq := &apiPin.CreatePinRequest{
		Title:         title,
		Content:       content,
		Category:      category,
		Priority:      priority,
		FromProjectID: fromProjectID,
		ToProjectID:   toProjectName,
	}
	resp, xErr := pinLogic.Push(context.Background(), apiReq)
	if xErr != nil {
		return textResult(fmt.Sprintf("Pin 推送失败: %s", xErr.Error())), nil
	}
	return textResult(fmt.Sprintf(`Pin 推送成功！

ID: %s
标题: %s
目标项目: %s → %s`,
		resp.ID, resp.Title, toProjectName, resp.ToProjectID)), nil
}

// handlePinConsume 消费约束（FIFO 队首 / 精确 ID）
func handlePinConsume(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if pinLogic == nil {
		return textResult("PinLogic 未初始化，请联系管理员"), nil
	}
	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return textResult(errMsg), nil
	}
	projectName, _ := args["project_name"].(string)
	if projectName == "" {
		return textResult("缺少必填参数: project_name"), nil
	}

	// 解析目标项目（支持名称/别名或雪花 ID）
	project, xErr := pinLogic.ResolveProject(context.Background(), projectName)
	if xErr != nil {
		return textResult(fmt.Sprintf("解析目标项目失败: %s", xErr.Error())), nil
	}

	// 精确 ID 消费 vs FIFO 队首消费
	if idStr, _ := args["id"].(string); idStr != "" {
		parsedID, err := xSnowflake.ParseSnowflakeID(idStr)
		if err != nil {
			return textResult(fmt.Sprintf("无效的 Pin ID: %s", idStr)), nil
		}
		resp, xErr := pinLogic.Consume(context.Background(), project.ID, &parsedID)
		if xErr != nil {
			return textResult(fmt.Sprintf("消费失败: %s", xErr.Error())), nil
		}
		return textResult(formatPinConsumed(resp)), nil
	}

	resp, xErr := pinLogic.Consume(context.Background(), project.ID, nil)
	if xErr != nil {
		return textResult(fmt.Sprintf("消费失败: %s", xErr.Error())), nil
	}
	return textResult(formatPinConsumed(resp)), nil
}

// handlePinList 列出目标项目的约束列表
func handlePinList(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if pinLogic == nil {
		return textResult("PinLogic 未初始化，请联系管理员"), nil
	}
	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return textResult(errMsg), nil
	}
	projectName, _ := args["project_name"].(string)
	if projectName == "" {
		return textResult("缺少必填参数: project_name"), nil
	}

	// 解析目标项目
	project, xErr := pinLogic.ResolveProject(context.Background(), projectName)
	if xErr != nil {
		return textResult(fmt.Sprintf("解析目标项目失败: %s", xErr.Error())), nil
	}

	page := 1
	size := 10
	if p, ok := args["page"].(float64); ok && p > 0 {
		page = int(p)
	}
	if s, ok := args["size"].(float64); ok && s > 0 && s <= 100 {
		size = int(s)
	}
	status, _ := args["status"].(string)
	category, _ := args["category"].(string)
	priority, _ := args["priority"].(string)
	fromProjectID, _ := args["from_project_id"].(string)

	listReq := &apiPin.PinListRequest{
		ToProjectID:   project.ID.String(),
		FromProjectID: fromProjectID,
		Status:        status,
		Category:      category,
		Priority:      priority,
		Page:          page,
		Size:          size,
	}
	resp, xErr := pinLogic.List(context.Background(), listReq)
	if xErr != nil {
		return textResult(fmt.Sprintf("获取约束列表失败: %s", xErr.Error())), nil
	}

	totalPages := (resp.Total + int64(size) - 1) / int64(size)
	result := fmt.Sprintf("约束列表（共 %d 个，第 %d/%d 页）：\n\n", resp.Total, page, totalPages)
	for i, p := range resp.Items {
		result += fmt.Sprintf("%d. [%s] %s | 状态: %s | 优先级: %s\n", i+1, p.ID, p.Title, p.Status, p.Priority)
	}
	if len(resp.Items) == 0 {
		result += "（暂无约束）\n"
	}
	return textResult(result), nil
}

// handlePinUpdate 更新约束元数据（优先级 / 分类）
func handlePinUpdate(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if pinLogic == nil {
		return textResult("PinLogic 未初始化，请联系管理员"), nil
	}
	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return textResult(errMsg), nil
	}
	idStr, _ := args["id"].(string)
	if idStr == "" {
		return textResult("缺少必填参数: id"), nil
	}
	parsedID, err := xSnowflake.ParseSnowflakeID(idStr)
	if err != nil {
		return textResult(fmt.Sprintf("无效的 Pin ID: %s", idStr)), nil
	}

	// 构造可选更新请求（指针字段为 nil 表示不更新）
	updateReq := &apiPin.UpdatePinRequest{}
	if p, ok := args["priority"].(string); ok && p != "" {
		updateReq.Priority = &p
	}
	if c, ok := args["category"].(string); ok && c != "" {
		updateReq.Category = &c
	}

	resp, xErr := pinLogic.Update(context.Background(), parsedID, updateReq)
	if xErr != nil {
		return textResult(fmt.Sprintf("更新失败: %s", xErr.Error())), nil
	}
	return textResult(fmt.Sprintf(`更新成功！

ID: %s
标题: %s
优先级: %s
分类: %s`,
		resp.ID, resp.Title, resp.Priority, resp.Category)), nil
}

// handlePinPeek 查看约束详情（只读）
func handlePinPeek(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if pinLogic == nil {
		return textResult("PinLogic 未初始化，请联系管理员"), nil
	}
	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return textResult(errMsg), nil
	}
	idStr, _ := args["id"].(string)
	if idStr == "" {
		return textResult("缺少必填参数: id"), nil
	}
	parsedID, err := xSnowflake.ParseSnowflakeID(idStr)
	if err != nil {
		return textResult(fmt.Sprintf("无效的 Pin ID: %s", idStr)), nil
	}

	resp, xErr := pinLogic.Peek(context.Background(), parsedID)
	if xErr != nil {
		return textResult(fmt.Sprintf("查看约束失败: %s", xErr.Error())), nil
	}
	return textResult(formatPinDetail(resp)), nil
}

// ─── Helpers ────────────────────────────────────────────────────────────

// formatPinConsumed 格式化消费成功的 Pin 响应为可读文本
func formatPinConsumed(resp *apiPin.PinResponse) string {
	return fmt.Sprintf(`消费成功！

ID: %s
标题: %s
分类: %s
优先级: %s

内容:
%s`,
		resp.ID, resp.Title, resp.Category, resp.Priority, resp.Content)
}

// formatPinDetail 格式化 Pin 详情为完整可读文本
func formatPinDetail(resp *apiPin.PinResponse) string {
	consumedAt := "（未消费）"
	if resp.ConsumedAt != "" {
		consumedAt = resp.ConsumedAt
	}
	fromProjectID := "（无来源）"
	if resp.FromProjectID != "" {
		fromProjectID = resp.FromProjectID
	}
	return fmt.Sprintf(`Pin 详情：

ID: %s
标题: %s
内容: %s
分类: %s
状态: %s
优先级: %s
来源项目: %s
目标项目: %s
消费时间: %s
创建时间: %s
更新时间: %s`,
		resp.ID, resp.Title, resp.Content, resp.Category, resp.Status, resp.Priority,
		fromProjectID, resp.ToProjectID, consumedAt, resp.CreatedAt, resp.UpdatedAt)
}

// RegisterPinTools 将 Pin 模块的 5 个 MCP 工具注册到 Server。
func RegisterPinTools(server *mcp.Server) {
	for _, def := range pinToolDefs {
		schemaBytes, _ := json.Marshal(def.inputSchema)
		tool := &mcp.Tool{
			Name: def.name, Description: def.description, InputSchema: json.RawMessage(schemaBytes),
		}
		var handler mcp.ToolHandler
		switch def.name {
		case "pin_push":
			handler = handlePinPush
		case "pin_consume":
			handler = handlePinConsume
		case "pin_list":
			handler = handlePinList
		case "pin_update":
			handler = handlePinUpdate
		case "pin_peek":
			handler = handlePinPeek
		default:
			handler = stubToolHandler(def.name)
		}
		server.AddTool(tool, handler)
	}
}
