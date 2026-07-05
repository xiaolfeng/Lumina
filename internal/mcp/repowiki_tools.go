package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	apiRepowiki "github.com/xiaolfeng/Lumina/api/repowiki"
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
		name: "repoWiki_analyze",
		description: `克隆 Git 仓库并通过 LLM 分析生成结构化 Wiki 文档。

触发场景：用户需要对一个代码仓库进行深度分析并生成 Wiki 文档时调用。工具会先创建 RepoWiki 配置，再触发异步分析管道（克隆 → 文件扫描 → LLM 多 Pass 分析 → 文档组装）。

分析是异步执行的：工具立即返回 version_id（状态 pending），Agent 可通过 repoWiki_query 工具查询 Wiki 内容（分析完成后才可读）。

ssh_key 用于私有仓库克隆（PEM 格式私钥），公开仓库不需要传。`,
		inputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"repo_url": map[string]any{
					"type":        "string",
					"description": "Git 仓库地址（HTTPS 或 SSH）",
				},
				"name": map[string]any{
					"type":        "string",
					"description": "配置名称（便于识别）",
				},
				"branch": map[string]any{
					"type":        "string",
					"description": "分析分支（可选，默认 main）",
				},
				"ssh_key": map[string]any{
					"type":        "string",
					"description": "SSH 私钥（PEM 格式，私有仓库用，可选）",
				},
			},
			"required": []string{"repo_url", "name"},
		},
	},
	{
		name: "repoWiki_query",
		description: `查询已生成的 Wiki 文档内容。

触发场景：Agent 需要获取某个仓库的 Wiki 文档以理解代码库结构、架构设计、模块说明等内容时调用。

wiki_id 为分析版本 ID（由 repoWiki_analyze 返回的 version_id）。
query 参数为可选的关键词搜索（当前版本返回 Wiki 首页摘要，关键词搜索为预留扩展）。

Wiki 必须处于 completed 状态才可查询；分析中或失败的版本会返回错误提示。`,
		inputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"wiki_id": map[string]any{
					"type":        "integer",
					"description": "Wiki 版本 ID（由 repoWiki_analyze 返回的 version_id）",
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
		name: "repoWiki_update",
		description: `增量更新 Wiki（对已分析过的仓库重新触发分析）。

触发场景：仓库代码有更新后，Agent 需要刷新 Wiki 文档时调用。会基于现有配置创建新的分析版本，执行完整的分析管道。

config_id 为 RepoWiki 配置 ID（由 repoWiki_analyze 或 repoWiki_list 返回）。

更新是异步执行的，立即返回新的 version_id（状态 pending）。`,
		inputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"config_id": map[string]any{
					"type":        "integer",
					"description": "RepoWiki 配置 ID",
				},
				"branch": map[string]any{
					"type":        "string",
					"description": "分析分支（可选，默认使用配置的 default_branch）",
				},
			},
			"required": []string{"config_id"},
		},
	},
	{
		name: "repoWiki_list",
		description: `列出所有 RepoWiki 配置（分页）。

触发场景：Agent 需要查看已分析过的仓库列表、获取 config_id 或查看分析状态时调用。

每项返回配置 ID、仓库地址、分支、语言、状态等基本信息。`,
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
	{
		name: "repoWiki_delete",
		description: `删除 RepoWiki 配置及其相关数据。

触发场景：用户不再需要某个仓库的 Wiki 文档，或需要清理配置时调用。

config_id 为 RepoWiki 配置 ID。
删除后关联的版本记录和 Wiki 文档将不可恢复。

注意：当前版本仅删除数据库记录，文件系统中的克隆仓库和 Wiki 文件可能需要手动清理。`,
		inputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"config_id": map[string]any{
					"type":        "integer",
					"description": "要删除的 RepoWiki 配置 ID",
				},
			},
			"required": []string{"config_id"},
		},
	},
}

// ─── Tool Handlers ──────────────────────────────────────────────────────

// handleRepoWikiAnalyze 克隆并分析仓库（创建配置 + 触发分析）
func handleRepoWikiAnalyze(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if repoWikiLogic == nil {
		return textResult("RepoWikiLogic 未初始化，请联系管理员"), nil
	}
	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return textResult(errMsg), nil
	}
	repoURL, _ := args["repo_url"].(string)
	if repoURL == "" {
		return textResult("缺少必填参数: repo_url"), nil
	}
	name, _ := args["name"].(string)
	if name == "" {
		return textResult("缺少必填参数: name"), nil
	}
	branch, _ := args["branch"].(string)
	sshKey, _ := args["ssh_key"].(string)

	// Step 1: 创建配置
	createReq := &apiRepowiki.CreateConfigRequest{
		RepoURL:       repoURL,
		Name:          name,
		DefaultBranch: branch,
		SSHKey:        sshKey,
	}
	config, xErr := repoWikiLogic.CreateConfig(context.Background(), createReq)
	if xErr != nil {
		return textResult(fmt.Sprintf("创建 RepoWiki 配置失败: %s", xErr.Error())), nil
	}

	// Step 2: 触发分析
	analyzeReq := &apiRepowiki.AnalyzeRequest{
		Branch: branch,
	}
	version, xErr := repoWikiLogic.AnalyzeRepo(context.Background(), config.ID, analyzeReq)
	if xErr != nil {
		return textResult(fmt.Sprintf("触发分析失败: %s\n\n配置已创建（config_id: %d），可稍后通过 repoWiki_update 重试分析。",
			xErr.Error(), config.ID.Int64())), nil
	}

	return textResult(fmt.Sprintf(`RepoWiki 分析已触发！

配置 ID: %d
版本 ID: %d
仓库地址: %s
分析分支: %s
当前状态: %s

分析是异步执行的，请稍后通过 repoWiki_query（wiki_id: %d）查询 Wiki 内容。`,
		config.ID.Int64(), version.ID.Int64(), repoURL, version.Branch, version.Status, version.ID.Int64())), nil
}

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

// handleRepoWikiUpdate 增量更新 Wiki（重新分析）
func handleRepoWikiUpdate(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if repoWikiLogic == nil {
		return textResult("RepoWikiLogic 未初始化，请联系管理员"), nil
	}
	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return textResult(errMsg), nil
	}
	configID, ok := parseSnowflakeInt(args["config_id"])
	if !ok {
		return textResult("缺少必填参数: config_id（整数）"), nil
	}
	branch, _ := args["branch"].(string)

	analyzeReq := &apiRepowiki.AnalyzeRequest{
		Branch: branch,
	}
	version, xErr := repoWikiLogic.AnalyzeRepo(context.Background(), xSnowflake.SnowflakeID(configID), analyzeReq)
	if xErr != nil {
		return textResult(fmt.Sprintf("触发更新分析失败: %s", xErr.Error())), nil
	}

	return textResult(fmt.Sprintf(`Wiki 更新分析已触发！

配置 ID: %d
新版本 ID: %d
分析分支: %s
当前状态: %s

分析是异步执行的，请稍后通过 repoWiki_query（wiki_id: %d）查询最新 Wiki 内容。`,
		configID, version.ID.Int64(), version.Branch, version.Status, version.ID.Int64())), nil
}

// handleRepoWikiList 列出所有 RepoWiki 配置
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

	configs, total, xErr := repoWikiLogic.ListConfigs(context.Background(), page, size)
	if xErr != nil {
		return textResult(fmt.Sprintf("获取配置列表失败: %s", xErr.Error())), nil
	}

	totalPages := (total + int64(size) - 1) / int64(size)
	result := fmt.Sprintf("RepoWiki 配置列表（共 %d 个，第 %d/%d 页）：\n\n", total, page, totalPages)
	for i, c := range configs {
		result += fmt.Sprintf("%d. [config_id: %d] %s\n", i+1, c.ID.Int64(), c.GitURL)
		result += fmt.Sprintf("   分支: %s | 语言: %s | 状态: %s\n", c.DefaultBranch, c.DefaultLanguage, c.Status)
	}
	if len(configs) == 0 {
		result += "（暂无 RepoWiki 配置）\n"
	}
	return textResult(result), nil
}

// handleRepoWikiDelete 删除 RepoWiki 配置
func handleRepoWikiDelete(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if repoWikiLogic == nil {
		return textResult("RepoWikiLogic 未初始化，请联系管理员"), nil
	}
	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return textResult(errMsg), nil
	}
	configID, ok := parseSnowflakeInt(args["config_id"])
	if !ok {
		return textResult("缺少必填参数: config_id（整数）"), nil
	}

	xErr := repoWikiLogic.DeleteConfig(context.Background(), xSnowflake.SnowflakeID(configID))
	if xErr != nil {
		return textResult(fmt.Sprintf("删除配置失败: %s", xErr.Error())), nil
	}

	return textResult(fmt.Sprintf(`RepoWiki 配置已删除！

配置 ID: %d

注意：数据库记录已删除，文件系统中的 Wiki 文档可能需要手动清理。`, configID)), nil
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

// RegisterRepoWikiTools 将 RepoWiki 模块的 5 个 MCP 工具注册到 Server。
func RegisterRepoWikiTools(server *mcp.Server) {
	for _, def := range repoWikiToolDefs {
		schemaBytes, _ := json.Marshal(def.inputSchema)
		tool := &mcp.Tool{
			Name: def.name, Description: def.description, InputSchema: json.RawMessage(schemaBytes),
		}
		var handler mcp.ToolHandler
		switch def.name {
		case "repoWiki_analyze":
			handler = handleRepoWikiAnalyze
		case "repoWiki_query":
			handler = handleRepoWikiQuery
		case "repoWiki_update":
			handler = handleRepoWikiUpdate
		case "repoWiki_list":
			handler = handleRepoWikiList
		case "repoWiki_delete":
			handler = handleRepoWikiDelete
		default:
			handler = stubToolHandler(def.name)
		}
		server.AddTool(tool, handler)
	}
}
