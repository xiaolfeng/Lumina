package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/xiaolfeng/Lumina/internal/logic"
)

// repoWikiLogic 保存 RepoWikiLogic 实例，供 MCP 工具处理器使用。
var repoWikiLogic *logic.RepoWikiLogic

// SetRepoWikiLogic 设置 RepoWikiLogic 实例，供 MCP 工具处理器使用。
func SetRepoWikiLogic(l *logic.RepoWikiLogic) {
	repoWikiLogic = l
}

// repoWikiToolDefs 定义 RepoWiki 模块的全部 MCP 工具。
var repoWikiToolDefs = []struct {
	name        string
	description string
	inputSchema map[string]any
}{
	{
		name: "repoWiki_query",
		description: `查询已生成的 Wiki 文档内容。

触发场景：Agent 需要获取某个仓库的 Wiki 文档以理解代码库结构、架构设计、模块说明等内容时调用。

wiki_id 为 Wiki 版本 ID（由 repoWiki_list 返回的 version_id）。
query 参数为可选的关键词搜索（当前版本返回 Wiki 首页摘要，关键词搜索为预留扩展）。

Wiki 必须处于 completed 状态才可查询；分析中或失败的版本会返回错误提示。`,
		inputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"wiki_id": map[string]any{
					"type":        "integer",
					"description": "Wiki 版本 ID（由 repoWiki_list 返回的 version_id）",
				},
				"query": map[string]any{
					"type":        "string",
					"description": "查询关键词（可选，不传返回 Wiki 概览）",
				},
			},
			"required": []string{"wiki_id"},
		},
	},
	{
		name: "repoWiki_list",
		description: `列出所有已完成的 Wiki 版本（分页）。返回 version_id、项目名称、分支、语言、commit hash、完成时间。Agent 可通过此工具查看有哪些 Wiki 可读，然后通过 repoWiki_query 读取具体内容。`,
		inputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"page": map[string]any{
					"type":        "integer",
					"description": "页码（从 1 开始，默认 1）",
				},
				"size": map[string]any{
					"type":        "integer",
					"description": "每页数量（默认 20，最大 100）",
				},
			},
		},
	},
}

// ─── Tool Handlers ──────────────────────────────────────────────────────

// handleRepoWikiQuery 查询 Wiki 内容
func handleRepoWikiQuery(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if repoWikiLogic == nil {
		return textResult("RepoWikiLogic 未初始化，请联系管理员"), nil
	}
	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return textResult(errMsg), nil
	}
	wikiID, ok := parseSnowflakeInt(args["wiki_id"])
	if !ok {
		return textResult("缺少必填参数: wiki_id（整数）"), nil
	}
	query, _ := args["query"].(string)

	content, xErr := repoWikiLogic.QueryWiki(context.Background(), wikiID, query)
	if xErr != nil {
		return textResult(fmt.Sprintf("查询 Wiki 失败: %s", xErr.Error())), nil
	}

	return textResult(content), nil
}

// handleRepoWikiList 列出所有已完成的 Wiki 版本
func handleRepoWikiList(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if repoWikiLogic == nil {
		return textResult("RepoWikiLogic 未初始化，请联系管理员"), nil
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

	wikis, total, xErr := repoWikiLogic.ListCompletedWikis(context.Background(), page, size)
	if xErr != nil {
		return textResult(fmt.Sprintf("获取 Wiki 版本列表失败: %s", xErr.Error())), nil
	}

	totalPages := (total + int64(size) - 1) / int64(size)
	result := fmt.Sprintf("Wiki 版本列表（共 %d 个，第 %d/%d 页）：\n\n", total, page, totalPages)
	for i, w := range wikis {
		result += fmt.Sprintf("%d. [version_id: %d] %s\n", i+1, w.VersionID.Int64(), w.ConfigName)
		result += fmt.Sprintf("   分支: %s | 语言: %s | commit: %s\n", w.Branch, w.Language, w.CommitHash)
		completedAt := "-"
		if w.CompletedAt != nil {
			completedAt = w.CompletedAt.Format(time.RFC3339)
		}
		result += fmt.Sprintf("   完成时间: %s\n", completedAt)
	}
	if len(wikis) == 0 {
		result += "（暂无已完成的 Wiki 版本）\n"
	}
	return textResult(result), nil
}

// ─── Helpers ────────────────────────────────────────────────────────────

// parseSnowflakeInt 从 MCP 参数中解析整数型雪花 ID
//
// MCP JSON 参数以 float64 传递，此函数统一处理类型转换。
// 返回 (int64值, 是否成功)。
func parseSnowflakeInt(val any) (int64, bool) {
	if f, ok := val.(float64); ok {
		return int64(f), true
	}
	return 0, false
}

// RegisterRepoWikiTools 将 RepoWiki 模块的 MCP 工具注册到 Server。
func RegisterRepoWikiTools(server *mcp.Server) {
	for _, def := range repoWikiToolDefs {
		schemaBytes, _ := json.Marshal(def.inputSchema)
		tool := &mcp.Tool{
			Name: def.name, Description: def.description, InputSchema: json.RawMessage(schemaBytes),
		}
		var handler mcp.ToolHandler
		switch def.name {
		case "repoWiki_query":
			handler = handleRepoWikiQuery
		case "repoWiki_list":
			handler = handleRepoWikiList
		default:
			handler = stubToolHandler(def.name)
		}
		server.AddTool(tool, handler)
	}
}
