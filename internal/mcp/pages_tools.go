package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	apiPreview "github.com/xiaolfeng/Lumina/api/preview"
	"github.com/xiaolfeng/Lumina/internal/logic"
)

var pagesLogic *logic.PagesLogic

// SetPagesLogic 注入 PagesLogic，供 MCP 工具使用。
func SetPagesLogic(l *logic.PagesLogic) {
	pagesLogic = l
}

type pagesToolDef struct {
	name         string
	title        string
	description  string
	inputSchema  map[string]any
	outputSchema map[string]any
	annotations  *mcp.ToolAnnotations
}

var pagesToolDefs = []pagesToolDef{
	{
		name:  "pages_list",
		title: "列出项目已发布页面",
		description: `用途：分页列出指定项目的 Pages 持久页面，返回 slug、生效版本与绝对 page_url；不返回文件正文。

何时调用：需要查看某项目已发布的即时页面、或准备晋升前确认 slug 是否已被占用时调用。

该工具只读且可安全重试。page_url 为路径式地址，例如 /pages/<project_name>/<slug>/<filename>。`,
		inputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"project_id": map[string]any{
					"type":        "string",
					"pattern":     "^[0-9]+$",
					"description": "关联的 Lumina 项目雪花 ID。",
				},
				"page": map[string]any{"type": "integer", "minimum": 1, "default": 1, "description": "页码，从 1 开始。"},
				"size": map[string]any{"type": "integer", "minimum": 1, "maximum": 100, "default": 10, "description": "每页数量。"},
			},
			"required": []string{"project_id"},
		},
		outputSchema: pagesListOutputSchema(),
		annotations:  previewToolAnnotations(true, false, true),
	},
	{
		name:  "pages_promote",
		title: "将预览会话晋升为 Pages",
		description: `用途：把维护完成的 Preview 会话深拷贝为项目级不可变 Pages 快照。不设置密码；访问策略由控制台 /console/pages 管理。

何时调用：Preview 文件齐全、用户确认将该草稿发布为持久页面时调用。无 HTML 入口时会被拒绝。

副作用：创建 Page 或追加 PageVersion；勾选 set_as_active 时切换线上生效指针。线上指针已前进且未 confirm_conflict 时返回冲突，不会改指针。`,
		inputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"session_id":       map[string]any{"type": "string", "pattern": "^[0-9]+$", "description": "Preview 会话雪花 ID。"},
				"slug":             map[string]any{"type": "string", "maxLength": 64, "description": "项目内访问标识，仅小写字母、数字与短横线。"},
				"title":            map[string]any{"type": "string", "maxLength": 255, "description": "页面标题。"},
				"description":      map[string]any{"type": "string", "description": "页面描述。"},
				"version":          map[string]any{"type": "string", "maxLength": 32, "description": "语义化版本号；空则自动递增。"},
				"changelog":        map[string]any{"type": "string", "description": "版本更新说明。"},
				"set_as_active":    map[string]any{"type": "boolean", "default": true, "description": "是否立即设为线上生效版本。"},
				"confirm_conflict": map[string]any{"type": "boolean", "description": "OCC 冲突时确认覆盖指针。"},
			},
			"required": []string{"session_id", "title"},
		},
		outputSchema: pagesPromoteOutputSchema(),
		annotations:  previewToolAnnotations(false, false, false),
	},
	{
		name:  "pages_fork",
		title: "从已发布页面派生预览会话",
		description: `用途：把指定 Pages 版本深拷贝为新的 Preview 草稿，并写入 source_page_id / source_page_slug / source_version_id。用于在已发布页面上继续改，再 pages_promote 发新版本。

何时调用：用户要基于线上页面迭代，而不是从空白 Preview 重做时调用。默认拷贝当前生效版本；可传 version_id 指定历史快照。

副作用：新建 Preview 会话与文件副本，不改线上指针。新会话仍须登录访问。密码策略不会带到草稿里。

下一步：用 preview_file_get / preview_file_upload 修改草稿，核对后调用 pages_promote（slug 可省略，锁定来源页面）。`,
		inputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"page_id":    map[string]any{"type": "string", "pattern": "^[0-9]+$", "description": "Pages 雪花 ID。"},
				"version_id": map[string]any{"type": "string", "pattern": "^[0-9]+$", "description": "可选。指定历史版本；省略则拷贝当前生效版本。"},
			},
			"required": []string{"page_id"},
		},
		outputSchema: pagesForkOutputSchema(),
		annotations:  previewToolAnnotations(false, false, false),
	},
}

func pagesListOutputSchema() map[string]any {
	return objectSchema(map[string]any{
		"status":      map[string]any{"type": "string", "const": "success"},
		"message":     map[string]any{"type": "string"},
		"items":       map[string]any{"type": "array", "items": pageItemSchema()},
		"total":       map[string]any{"type": "integer", "minimum": 0},
		"page":        map[string]any{"type": "integer", "minimum": 1},
		"size":        map[string]any{"type": "integer", "minimum": 1},
		"total_pages": map[string]any{"type": "integer", "minimum": 0},
	}, "status", "message", "items", "total", "page", "size", "total_pages")
}

func pagesForkOutputSchema() map[string]any {
	return objectSchema(map[string]any{
		"status":      map[string]any{"type": "string", "const": "success"},
		"message":     map[string]any{"type": "string"},
		"session":     previewSessionSchema(),
		"preview_url": map[string]any{"type": "string", "format": "uri"},
	}, "status", "message", "session", "preview_url")
}

func pagesPromoteOutputSchema() map[string]any {
	return objectSchema(map[string]any{
		"status":  map[string]any{"type": "string", "const": "success"},
		"message": map[string]any{"type": "string"},
		"conflict": objectSchema(map[string]any{
			"current_version":    map[string]any{"type": "string"},
			"current_version_id": map[string]any{"type": "string"},
			"current_changelog":  map[string]any{"type": "string"},
			"current_created_by": map[string]any{"type": "string"},
			"current_created_at": map[string]any{"type": "string"},
			"source_version_id":  map[string]any{"type": "string"},
			"message":            map[string]any{"type": "string"},
		}, "current_version", "current_version_id", "current_changelog", "current_created_by", "current_created_at", "source_version_id", "message"),
		"page":    pageItemSchema(),
		"version": pageVersionSchema(),
	}, "status", "message")
}

func pageItemSchema() map[string]any {
	return objectSchema(map[string]any{
		"id":                map[string]any{"type": "string"},
		"project_id":        map[string]any{"type": "string"},
		"project_name":      map[string]any{"type": "string"},
		"slug":              map[string]any{"type": "string"},
		"title":             map[string]any{"type": "string"},
		"description":       map[string]any{"type": "string"},
		"status":            map[string]any{"type": "string"},
		"access_mode":       map[string]any{"type": "string"},
		"latest_version_id": map[string]any{"type": "string"},
		"latest_version":    map[string]any{"type": "string"},
		"page_url":          map[string]any{"type": "string", "format": "uri"},
		"created_at":        map[string]any{"type": "string"},
		"updated_at":        map[string]any{"type": "string"},
	}, "id", "project_id", "project_name", "slug", "title", "status", "access_mode", "latest_version_id", "page_url", "created_at", "updated_at")
}

func pageVersionSchema() map[string]any {
	return objectSchema(map[string]any{
		"id":                map[string]any{"type": "string"},
		"page_id":           map[string]any{"type": "string"},
		"version":           map[string]any{"type": "string"},
		"changelog":         map[string]any{"type": "string"},
		"source_session_id": map[string]any{"type": []string{"string", "null"}},
		"base_version_id":   map[string]any{"type": []string{"string", "null"}},
		"entry_filename":    map[string]any{"type": "string"},
		"file_count":        map[string]any{"type": "integer"},
		"total_size":        map[string]any{"type": "integer"},
		"created_by":        map[string]any{"type": "string"},
		"is_active":         map[string]any{"type": "boolean"},
		"created_at":        map[string]any{"type": "string"},
	}, "id", "page_id", "version", "entry_filename", "file_count", "total_size", "created_by", "is_active", "created_at")
}

func handlePagesList(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if pagesLogic == nil {
		return previewErrorResult("PagesLogic 未初始化，请联系管理员"), nil
	}
	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return previewErrorResult(errMsg), nil
	}
	projectIDStr, _ := args["project_id"].(string)
	if projectIDStr == "" {
		return previewErrorResult("缺少必填参数: project_id"), nil
	}
	projectID, err := xSnowflake.ParseSnowflakeID(projectIDStr)
	if err != nil {
		return previewErrorResult("无效的 project_id"), nil
	}
	page := 1
	size := 10
	if p, ok := args["page"].(float64); ok && p > 0 {
		page = int(p)
	}
	if s, ok := args["size"].(float64); ok && s > 0 && s <= 100 {
		size = int(s)
	}
	resp, xErr := pagesLogic.List(ctx, projectID, 0, page, size)
	if xErr != nil {
		return previewErrorResult(fmt.Sprintf("获取页面列表失败: %s", xErr.Error())), nil
	}
	items := make([]map[string]any, 0, len(resp.Items))
	for i := range resp.Items {
		items = append(items, pageItemData(&resp.Items[i]))
	}
	totalPages := (resp.Total + int64(size) - 1) / int64(size)
	return previewStructuredResult(map[string]any{
		"status":      "success",
		"message":     fmt.Sprintf("找到 %d 个页面。", len(items)),
		"items":       items,
		"total":       resp.Total,
		"page":        page,
		"size":        size,
		"total_pages": totalPages,
	}), nil
}

func handlePagesPromote(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if pagesLogic == nil {
		return previewErrorResult("PagesLogic 未初始化，请联系管理员"), nil
	}
	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return previewErrorResult(errMsg), nil
	}
	sessionIDStr, _ := args["session_id"].(string)
	sessionID, err := xSnowflake.ParseSnowflakeID(sessionIDStr)
	if err != nil || sessionIDStr == "" {
		return previewErrorResult("缺少或无效的 session_id"), nil
	}
	slug, _ := args["slug"].(string)
	title, _ := args["title"].(string)
	description, _ := args["description"].(string)
	version, _ := args["version"].(string)
	changelog, _ := args["changelog"].(string)
	setAsActive := true
	if v, ok := args["set_as_active"].(bool); ok {
		setAsActive = v
	}
	confirmConflict, _ := args["confirm_conflict"].(bool)
	resp, xErr := pagesLogic.Promote(ctx, sessionID, &apiPreview.PromoteSessionRequest{
		Slug:            slug,
		Title:           title,
		Description:     description,
		Version:         version,
		Changelog:       changelog,
		SetAsActive:     setAsActive,
		ConfirmConflict: confirmConflict,
	})
	if xErr != nil {
		return previewErrorResult(fmt.Sprintf("晋升失败: %s", xErr.Error())), nil
	}
	result := map[string]any{
		"status":  "success",
		"message": "晋升成功。",
	}
	if resp.Conflict != nil {
		result["message"] = resp.Conflict.Message
		result["conflict"] = map[string]any{
			"current_version":    resp.Conflict.CurrentVersion,
			"current_version_id": resp.Conflict.CurrentVersionID.String(),
			"current_changelog":  resp.Conflict.CurrentChangelog,
			"current_created_by": resp.Conflict.CurrentCreatedBy,
			"current_created_at": resp.Conflict.CurrentCreatedAt,
			"source_version_id":  resp.Conflict.SourceVersionID.String(),
			"message":            resp.Conflict.Message,
		}
	}
	if resp.Page != nil {
		result["page"] = pageItemData(resp.Page)
	}
	if resp.Version != nil {
		result["version"] = pageVersionData(resp.Version)
	}
	return previewStructuredResult(result), nil
}

func handlePagesFork(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if pagesLogic == nil {
		return previewErrorResult("PagesLogic 未初始化，请联系管理员"), nil
	}
	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return previewErrorResult(errMsg), nil
	}
	pageIDStr, _ := args["page_id"].(string)
	pageID, err := xSnowflake.ParseSnowflakeID(pageIDStr)
	if err != nil || pageIDStr == "" {
		return previewErrorResult("缺少或无效的 page_id"), nil
	}
	var versionID xSnowflake.SnowflakeID
	if versionIDStr, ok := args["version_id"].(string); ok && versionIDStr != "" {
		parsed, parseErr := xSnowflake.ParseSnowflakeID(versionIDStr)
		if parseErr != nil {
			return previewErrorResult("无效的 version_id"), nil
		}
		versionID = parsed
	}
	resp, xErr := pagesLogic.Fork(ctx, pageID, versionID)
	if xErr != nil {
		return previewErrorResult(fmt.Sprintf("Fork 失败: %s", xErr.Error())), nil
	}
	return previewStructuredResult(map[string]any{
		"status":      "success",
		"message":     "已从页面派生新的 Preview 草稿。",
		"session":     previewSessionData(&resp.Session, resp.PreviewURL),
		"preview_url": resp.PreviewURL,
	}), nil
}

func pageItemData(page any) map[string]any {
	payload, _ := json.Marshal(page)
	data := map[string]any{}
	_ = json.Unmarshal(payload, &data)
	return data
}

func pageVersionData(version any) map[string]any {
	payload, _ := json.Marshal(version)
	data := map[string]any{}
	_ = json.Unmarshal(payload, &data)
	return data
}

// RegisterPagesTools 注册 Pages MCP 工具。
func RegisterPagesTools(server *mcp.Server) {
	for _, def := range pagesToolDefs {
		inputSchemaBytes, _ := json.Marshal(def.inputSchema)
		outputSchemaBytes, _ := json.Marshal(def.outputSchema)
		tool := &mcp.Tool{
			Name:         def.name,
			Title:        def.title,
			Description:  def.description,
			InputSchema:  json.RawMessage(inputSchemaBytes),
			OutputSchema: json.RawMessage(outputSchemaBytes),
			Annotations:  def.annotations,
		}
		var handler mcp.ToolHandler
		switch def.name {
		case "pages_list":
			handler = handlePagesList
		case "pages_promote":
			handler = handlePagesPromote
		case "pages_fork":
			handler = handlePagesFork
		default:
			handler = stubToolHandler(def.name)
		}
		server.AddTool(tool, handler)
	}
}
