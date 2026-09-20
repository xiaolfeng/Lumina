package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	apiPreview "github.com/xiaolfeng/Lumina/api/preview"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/logic"
)

// previewLogic 保存 PreviewLogic 实例，供 MCP 工具处理器使用。
var previewLogic *logic.PreviewLogic

// SetPreviewLogic 设置 PreviewLogic 实例，供 MCP 工具处理器使用。
func SetPreviewLogic(l *logic.PreviewLogic) {
	previewLogic = l
}

// handlePreviewSessionCreate 创建预览会话
func handlePreviewSessionCreate(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if previewLogic == nil {
		return previewErrorResult("PreviewLogic 未初始化，请联系管理员"), nil
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
		return previewErrorResult(fmt.Sprintf("无效的 project_id: %s", projectIDStr)), nil
	}

	title, _ := args["title"].(string)
	title = strings.TrimSpace(title)
	if len([]rune(title)) > 255 {
		return previewErrorResult("title 长度不能超过 255 个字符"), nil
	}

	resp, xErr := previewLogic.CreateSession(ctx, projectID, title)
	if xErr != nil {
		return previewErrorResult(fmt.Sprintf("创建预览会话失败: %s", xErr.Error())), nil
	}

	previewURL := previewLogic.BuildSessionURL(ctx, resp.Hash, "")
	result := map[string]any{
		"status":      "success",
		"message":     "预览会话已创建；当前为空会话，尚不可交付评审。",
		"session":     previewSessionData(resp, previewURL),
		"preview_url": previewURL,
		"workflow": map[string]any{
			"state":     "awaiting_files",
			"next_tool": "preview_file_upload",
			"instructions": []string{
				"先上传 HTML 入口及其引用的 CSS/JavaScript 文件。",
				"不要打开或向用户交付空的 preview_url。",
				"全部文件上传后调用 preview_file_list 做最终核对。",
			},
		},
	}
	return previewStructuredResult(result), nil
}

// handlePreviewSessionList 分页获取预览会话列表
func handlePreviewSessionList(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if previewLogic == nil {
		return previewErrorResult("PreviewLogic 未初始化，请联系管理员"), nil
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
		return previewErrorResult(fmt.Sprintf("无效的 project_id: %s", projectIDStr)), nil
	}

	page := 1
	size := 10
	if p, ok := args["page"].(float64); ok && p > 0 {
		page = int(p)
	}
	if s, ok := args["size"].(float64); ok && s > 0 && s <= 100 {
		size = int(s)
	}

	resp, xErr := previewLogic.ListSessions(ctx, projectID, 0, page, size)
	if xErr != nil {
		return previewErrorResult(fmt.Sprintf("获取预览会话列表失败: %s", xErr.Error())), nil
	}

	totalPages := (resp.Total + int64(size) - 1) / int64(size)
	items := make([]map[string]any, 0, len(resp.Items))
	for i := range resp.Items {
		session := &resp.Items[i]
		previewURL := previewLogic.BuildSessionURL(ctx, session.Hash, "")
		items = append(items, previewSessionData(session, previewURL))
	}

	state := "session_available"
	nextTool := "preview_file_list"
	instructions := []string{
		"按标题和任务语义选择匹配的 active 会话，不要覆盖无关任务。",
		"选定后调用 preview_file_list 核对现有文件，再决定复用或新建。",
	}
	message := fmt.Sprintf("找到 %d 个预览会话。", len(items))
	if len(items) == 0 {
		state = "no_session"
		nextTool = "preview_session_create"
		message = "当前项目没有可复用的预览会话。"
		instructions = []string{"需要可视化预览时调用 preview_session_create 创建新会话。"}
	}

	return previewStructuredResult(map[string]any{
		"status":      "success",
		"message":     message,
		"items":       items,
		"total":       resp.Total,
		"page":        page,
		"size":        size,
		"total_pages": totalPages,
		"workflow": map[string]any{
			"state":        state,
			"next_tool":    nextTool,
			"instructions": instructions,
		},
	}), nil
}

// handlePreviewFileUpload 上传预览文件
func handlePreviewFileUpload(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if previewLogic == nil {
		return previewErrorResult("PreviewLogic 未初始化，请联系管理员"), nil
	}
	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return previewErrorResult(errMsg), nil
	}

	sessionIDStr, _ := args["session_id"].(string)
	if sessionIDStr == "" {
		return previewErrorResult("缺少必填参数: session_id"), nil
	}
	sessionID, err := xSnowflake.ParseSnowflakeID(sessionIDStr)
	if err != nil {
		return previewErrorResult(fmt.Sprintf("无效的 session_id: %s", sessionIDStr)), nil
	}

	filename, _ := args["filename"].(string)
	if filename == "" {
		return previewErrorResult("缺少必填参数: filename"), nil
	}
	content, contentOK := args["content"].(string)
	if !contentOK {
		return previewErrorResult("缺少必填参数或参数类型错误: content"), nil
	}

	resp, xErr := previewLogic.UploadFile(ctx, sessionID, filename, content)
	if xErr != nil {
		return previewErrorResult(fmt.Sprintf("上传预览文件失败: %s", xErr.Error())), nil
	}

	snapshot, errMsg := loadPreviewSessionSnapshot(ctx, sessionID, resp.Filename)
	if errMsg != "" {
		return previewErrorResult(errMsg), nil
	}
	state, nextTool, message, instructions := previewWriteWorkflow(snapshot, "文件已写入")

	return previewStructuredResult(map[string]any{
		"status":        "success",
		"message":       message,
		"session":       previewSessionData(snapshot.session, snapshot.previewURL),
		"file":          previewFileData(resp),
		"entry_file":    snapshot.entryFilename,
		"preview_url":   snapshot.previewURL,
		"qa_supplement": previewSupplementData(sessionIDStr, snapshot.entryID),
		"workflow": map[string]any{
			"state":        state,
			"next_tool":    nextTool,
			"instructions": instructions,
		},
	}), nil
}

// handlePreviewFileEdit 行级编辑预览文件
func handlePreviewFileEdit(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if previewLogic == nil {
		return previewErrorResult("PreviewLogic 未初始化，请联系管理员"), nil
	}
	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return previewErrorResult(errMsg), nil
	}

	sessionIDStr, _ := args["session_id"].(string)
	if sessionIDStr == "" {
		return previewErrorResult("缺少必填参数: session_id"), nil
	}
	sessionID, err := xSnowflake.ParseSnowflakeID(sessionIDStr)
	if err != nil {
		return previewErrorResult(fmt.Sprintf("无效的 session_id: %s", sessionIDStr)), nil
	}

	filename, _ := args["filename"].(string)
	if filename == "" {
		return previewErrorResult("缺少必填参数: filename"), nil
	}

	operation, _ := args["operation"].(string)
	switch operation {
	case bConst.PreviewEditOperationInsert, bConst.PreviewEditOperationReplace, bConst.PreviewEditOperationDelete:
	default:
		return previewErrorResult("缺少必填参数或参数类型错误: operation（可选 insert/replace/delete）"), nil
	}

	startLine, startPresent, startErr := parsePreviewLineArg(args, "start_line")
	if startErr != "" {
		return previewErrorResult(startErr), nil
	}
	endLine, endPresent, endErr := parsePreviewLineArg(args, "end_line")
	if endErr != "" {
		return previewErrorResult(endErr), nil
	}
	if operation != bConst.PreviewEditOperationInsert {
		if !startPresent {
			return previewErrorResult(fmt.Sprintf("%s 操作缺少必填参数: start_line", operation)), nil
		}
		if !endPresent {
			return previewErrorResult(fmt.Sprintf("%s 操作缺少必填参数: end_line", operation)), nil
		}
	}

	content, contentOK := args["content"].(string)
	if operation != bConst.PreviewEditOperationDelete && !contentOK {
		return previewErrorResult(fmt.Sprintf("%s 操作缺少必填参数或参数类型错误: content", operation)), nil
	}

	resp, xErr := previewLogic.EditFile(ctx, sessionID, filename, operation, startLine, endLine, content)
	if xErr != nil {
		return previewErrorResult(fmt.Sprintf("编辑预览文件失败: %s", xErr.Error())), nil
	}

	snapshot, errMsg := loadPreviewSessionSnapshot(ctx, sessionID, resp.Filename)
	if errMsg != "" {
		return previewErrorResult(errMsg), nil
	}
	state, nextTool, message, instructions := previewWriteWorkflow(snapshot, "文件已编辑")

	return previewStructuredResult(map[string]any{
		"status":        "success",
		"message":       message,
		"session":       previewSessionData(snapshot.session, snapshot.previewURL),
		"file":          previewFileDataWithLines(&resp.PreviewFileResponse, resp.TotalLines),
		"edited_region": map[string]any{"start_line": resp.RegionStart, "end_line": resp.RegionEnd, "content": resp.Region},
		"entry_file":    snapshot.entryFilename,
		"preview_url":   snapshot.previewURL,
		"qa_supplement": previewSupplementData(sessionIDStr, snapshot.entryID),
		"workflow": map[string]any{
			"state":        state,
			"next_tool":    nextTool,
			"instructions": instructions,
		},
	}), nil
}

// handlePreviewFileDelete 按文件名删除预览文件
func handlePreviewFileDelete(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if previewLogic == nil {
		return previewErrorResult("PreviewLogic 未初始化，请联系管理员"), nil
	}
	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return previewErrorResult(errMsg), nil
	}

	sessionIDStr, _ := args["session_id"].(string)
	if sessionIDStr == "" {
		return previewErrorResult("缺少必填参数: session_id"), nil
	}
	sessionID, err := xSnowflake.ParseSnowflakeID(sessionIDStr)
	if err != nil {
		return previewErrorResult(fmt.Sprintf("无效的 session_id: %s", sessionIDStr)), nil
	}

	filename, _ := args["filename"].(string)
	if filename == "" {
		return previewErrorResult("缺少必填参数: filename"), nil
	}

	deleted, xErr := previewLogic.DeleteFileByName(ctx, sessionID, filename)
	if xErr != nil {
		return previewErrorResult(fmt.Sprintf("删除预览文件失败: %s", xErr.Error())), nil
	}

	snapshot, errMsg := loadPreviewSessionSnapshot(ctx, sessionID, "")
	if errMsg != "" {
		return previewErrorResult(errMsg), nil
	}

	state := "reviewable_unverified"
	nextTool := "preview_file_list"
	message := fmt.Sprintf("文件 %s 已删除，会话仍有 HTML 入口。", deleted.Filename)
	instructions := []string{
		"调用 preview_file_list 核对剩余文件与入口依赖的完整性。",
		"若被删文件仍被 HTML 引用，删除引用或重新补齐对应文件。",
	}
	if len(snapshot.files) == 0 {
		state = "empty"
		nextTool = "preview_file_upload"
		message = fmt.Sprintf("文件 %s 已删除，会话已没有任何文件。", deleted.Filename)
		instructions = []string{
			"需要继续预览时调用 preview_file_upload 重新上传 HTML 入口及其依赖。",
			"会话已无内容，不要向用户交付当前 preview_url。",
		}
	} else if snapshot.entry == nil {
		state = "awaiting_html_entry"
		nextTool = "preview_file_upload"
		message = fmt.Sprintf("文件 %s 已删除，剩余文件中没有 HTML 入口，暂不可作为前端页面交付评审。", deleted.Filename)
		instructions = []string{
			"调用 preview_file_upload 重新上传 HTML 入口，并以相对路径引用剩余文件。",
			"补齐入口后调用 preview_file_list 做最终核对。",
		}
	}

	fileItems := make([]map[string]any, 0, len(snapshot.files))
	for i := range snapshot.files {
		fileItems = append(fileItems, previewFileData(&snapshot.files[i]))
	}

	return previewStructuredResult(map[string]any{
		"status":        "success",
		"message":       message,
		"deleted_file":  map[string]any{"id": deleted.ID.String(), "filename": deleted.Filename},
		"session":       previewSessionData(snapshot.session, snapshot.previewURL),
		"files":         fileItems,
		"entry_file":    snapshot.entryFilename,
		"preview_url":   snapshot.previewURL,
		"qa_supplement": previewSupplementData(sessionIDStr, snapshot.entryID),
		"workflow": map[string]any{
			"state":        state,
			"next_tool":    nextTool,
			"instructions": instructions,
		},
	}), nil
}

// handlePreviewFileList 获取预览文件列表
func handlePreviewFileList(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if previewLogic == nil {
		return previewErrorResult("PreviewLogic 未初始化，请联系管理员"), nil
	}
	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return previewErrorResult(errMsg), nil
	}

	sessionIDStr, _ := args["session_id"].(string)
	if sessionIDStr == "" {
		return previewErrorResult("缺少必填参数: session_id"), nil
	}
	sessionID, err := xSnowflake.ParseSnowflakeID(sessionIDStr)
	if err != nil {
		return previewErrorResult(fmt.Sprintf("无效的 session_id: %s", sessionIDStr)), nil
	}

	snapshot, errMsg := loadPreviewSessionSnapshot(ctx, sessionID, "")
	if errMsg != "" {
		return previewErrorResult(errMsg), nil
	}

	fileItems := make([]map[string]any, 0, len(snapshot.files))
	for i := range snapshot.files {
		fileItems = append(fileItems, previewFileData(&snapshot.files[i]))
	}

	state := "empty"
	nextTool := "preview_file_upload"
	message := "会话中没有文件。"
	instructions := []string{"调用 preview_file_upload 上传 HTML 入口及其依赖。"}

	if len(snapshot.files) > 0 && snapshot.entry == nil {
		state = "awaiting_html_entry"
		message = "文件清单已核对，但没有 HTML 入口，暂不可作为前端页面交付评审。"
		instructions = []string{
			"调用 preview_file_upload 添加 HTML 入口，并以相对路径引用已有文件。",
			"上传完成后再次调用 preview_file_list。",
		}
	}
	if snapshot.entry != nil {
		state = "ready_for_review"
		nextTool = ""
		message = fmt.Sprintf("文件清单已核对，共 %d 个文件，HTML 入口为 %s。", len(snapshot.files), snapshot.entry.Filename)
		instructions = []string{
			"若用户请求视觉评审，优先使用客户端原生浏览器/打开链接能力访问 preview_url；能力不可用时向用户提供可点击 URL。",
			"若这是 Q&A 的问题或选项补充，调用 qa_push_supplement，并将 content_type 设为 preview、content 原样设为 qa_supplement.content。",
			"挂载 Q&A 补充后调用 qa_get_answer；不要把 Preview 当作真实项目实现已完成的证据。",
		}
	}

	return previewStructuredResult(map[string]any{
		"status":        "success",
		"message":       message,
		"session":       previewSessionData(snapshot.session, snapshot.previewURL),
		"files":         fileItems,
		"entry_file":    snapshot.entryFilename,
		"preview_url":   snapshot.previewURL,
		"qa_supplement": previewSupplementData(sessionIDStr, snapshot.entryID),
		"workflow": map[string]any{
			"state":        state,
			"next_tool":    nextTool,
			"instructions": instructions,
		},
	}), nil
}

// handlePreviewFileGet 获取预览文件内容（支持行区间带行号读取）
func handlePreviewFileGet(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if previewLogic == nil {
		return previewErrorResult("PreviewLogic 未初始化，请联系管理员"), nil
	}
	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return previewErrorResult(errMsg), nil
	}

	sessionIDStr, _ := args["session_id"].(string)
	if sessionIDStr == "" {
		return previewErrorResult("缺少必填参数: session_id"), nil
	}
	sessionID, err := xSnowflake.ParseSnowflakeID(sessionIDStr)
	if err != nil {
		return previewErrorResult(fmt.Sprintf("无效的 session_id: %s", sessionIDStr)), nil
	}

	filename, _ := args["filename"].(string)
	if filename == "" {
		return previewErrorResult("缺少必填参数: filename"), nil
	}

	startLine, startPresent, startErr := parsePreviewLineArg(args, "start_line")
	if startErr != "" {
		return previewErrorResult(startErr), nil
	}
	endLine, endPresent, endErr := parsePreviewLineArg(args, "end_line")
	if endErr != "" {
		return previewErrorResult(endErr), nil
	}
	if endPresent && !startPresent {
		return previewErrorResult("指定 end_line 时必须同时指定 start_line"), nil
	}
	startArg, endArg := 0, 0
	if startPresent {
		startArg = startLine
	}
	if endPresent {
		endArg = endLine
	}

	resp, xErr := previewLogic.GetFileLinesBySession(ctx, sessionID, filename, startArg, endArg)
	if xErr != nil {
		return previewErrorResult(fmt.Sprintf("获取预览文件内容失败: %s", xErr.Error())), nil
	}

	metadata := map[string]any{
		"status":  "success",
		"message": "已读取预览文件源码；该内容仍属于 Preview 会话，不等同于真实项目文件。",
		"file": map[string]any{
			"session_id":  sessionIDStr,
			"filename":    resp.Filename,
			"mime_type":   resp.MimeType,
			"size":        resp.Size,
			"total_lines": resp.TotalLines,
			"start_line":  resp.StartLine,
			"end_line":    resp.EndLine,
			"content":     resp.Content,
		},
		"workflow": map[string]any{
			"state":     "source_loaded",
			"next_tool": "",
			"instructions": []string{
				"只在用户要求迭代或提取已确认规范时使用这份源码。",
				"局部修改优先调用 preview_file_edit 行级编辑（行号与本工具区间模式的行号一致）。",
				"整文件重写才使用 preview_file_upload 覆写；改动后调用 preview_file_list 核对并重新打开预览。",
			},
		},
	}
	return previewFileContentResult(metadata, resp.Filename, resp.MimeType, resp.Content), nil
}

// ─── 共享响应骨架 ───────────────────────────────────────────────────────

// previewSessionSnapshot 会话写入/编辑/删除后的状态快照，供多个 Preview 工具组装响应骨架。
type previewSessionSnapshot struct {
	session       *apiPreview.PreviewSessionResponse
	files         []apiPreview.PreviewFileResponse
	entry         *apiPreview.PreviewFileResponse
	previewURL    string
	entryFilename string
	entryID       string
}

// loadPreviewSessionSnapshot 拉取会话详情与文件清单并完成 HTML 入口检测与预览地址构建。
//
// 无 HTML 入口时 preview_url 回退到 fallbackFilename（通常为刚写入的文件），便于用户直达该文件。
func loadPreviewSessionSnapshot(ctx context.Context, sessionID xSnowflake.SnowflakeID, fallbackFilename string) (*previewSessionSnapshot, string) {
	session, xErr := previewLogic.GetSessionByID(ctx, sessionID)
	if xErr != nil {
		return nil, fmt.Sprintf("读取预览会话失败: %s", xErr.Error())
	}
	files, xErr := previewLogic.ListFiles(ctx, sessionID)
	if xErr != nil {
		return nil, fmt.Sprintf("核对预览文件失败: %s", xErr.Error())
	}

	snapshot := &previewSessionSnapshot{session: session, files: files}
	if entry := logic.FindPreviewEntry(files); entry != nil {
		snapshot.entry = entry
		snapshot.entryFilename = entry.Filename
		snapshot.entryID = entry.ID.String()
		snapshot.previewURL = previewLogic.BuildSessionURL(ctx, session.Hash, entry.Filename)
	} else {
		snapshot.previewURL = previewLogic.BuildSessionURL(ctx, session.Hash, fallbackFilename)
	}
	return snapshot, ""
}

// previewWriteWorkflow 写入/编辑后的工作流状态机（有入口 → 待最终核对；无入口 → 待补入口）
func previewWriteWorkflow(snapshot *previewSessionSnapshot, writeNoun string) (state, nextTool, message string, instructions []string) {
	if snapshot.entry == nil {
		return "awaiting_html_entry", "preview_file_upload",
			fmt.Sprintf("%s，但会话还没有可评审入口（HTML 或 LPW），暂不可作为前端页面交付评审。", writeNoun),
			[]string{
				"继续上传可评审入口文件（HTML 或 LPW），并用相对路径引用同会话中的 CSS/JavaScript/静态资源。",
				"文件齐全后调用 preview_file_list 做最终核对。",
			}
	}
	return "reviewable_unverified", "preview_file_list",
		fmt.Sprintf("%s，会话已有可评审入口；完成其余依赖上传后仍需最终核对。", writeNoun),
		[]string{
			"若入口文件仍引用未上传的依赖，继续调用 preview_file_upload。",
			"全部文件写入后调用 preview_file_list；不要跳过最终核对。",
			"核对通过后再打开/分享 preview_url，或使用 qa_supplement 挂载到 Q&A。",
		}
}

// parsePreviewLineArg 解析可选整型行号参数（present 表示是否传入；errMsg 表示类型或取值非法）
func parsePreviewLineArg(args map[string]any, key string) (value int, present bool, errMsg string) {
	raw, exists := args[key]
	if !exists || raw == nil {
		return 0, false, ""
	}
	number, ok := raw.(float64)
	if !ok {
		return 0, true, fmt.Sprintf("参数类型错误: %s 应为整数", key)
	}
	if number != float64(int(number)) || int(number) < 1 {
		return 0, true, fmt.Sprintf("参数无效: %s 应为不小于 1 的整数（当前 %v）", key, number)
	}
	return int(number), true, ""
}

// ─── 数据组装辅助 ───────────────────────────────────────────────────────

func previewSessionData(session *apiPreview.PreviewSessionResponse, previewURL string) map[string]any {
	data := map[string]any{
		"id":          session.ID.String(),
		"project_id":  session.ProjectID.String(),
		"title":       session.Title,
		"hash":        session.Hash,
		"status":      session.Status,
		"file_count":  session.FileCount,
		"expires_at":  nullableString(session.ExpiresAt),
		"created_at":  session.CreatedAt,
		"updated_at":  session.UpdatedAt,
		"preview_url": previewURL,
	}
	if session.SourcePageID != nil {
		data["source_page_id"] = session.SourcePageID.String()
	} else {
		data["source_page_id"] = nil
	}
	data["source_page_slug"] = nullableString(session.SourcePageSlug)
	if session.SourceVersionID != nil {
		data["source_version_id"] = session.SourceVersionID.String()
	} else {
		data["source_version_id"] = nil
	}
	return data
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func previewFileData(file *apiPreview.PreviewFileResponse) map[string]any {
	return map[string]any{
		"id":         file.ID.String(),
		"session_id": file.SessionID.String(),
		"filename":   file.Filename,
		"mime_type":  file.MimeType,
		"size":       file.Size,
		"created_at": file.CreatedAt,
		"updated_at": file.UpdatedAt,
	}
}

// previewFileDataWithLines 在文件元数据基础上附带总行数（编辑/行级读取场景）
func previewFileDataWithLines(file *apiPreview.PreviewFileResponse, totalLines int) map[string]any {
	data := previewFileData(file)
	data["total_lines"] = totalLines
	return data
}

// findPreviewEntry 从文件清单中选择入口文件（优先 index.lpw -> 首个 lpw -> 现有 HTML）
func findPreviewEntry(files []apiPreview.PreviewFileResponse) *apiPreview.PreviewFileResponse {
	return logic.FindPreviewEntry(files)
}

func previewSupplementData(sessionID, fileID string) map[string]any {
	content := ""
	if sessionID != "" && fileID != "" {
		reference := struct {
			SessionID string `json:"session_id"`
			FileID    string `json:"file_id"`
		}{
			SessionID: sessionID,
			FileID:    fileID,
		}
		if payload, err := json.Marshal(reference); err == nil {
			content = string(payload)
		}
	}
	return map[string]any{
		"content_type": "preview",
		"content":      content,
	}
}

// previewStructuredResult 同时返回结构化结果与其 JSON 文本，兼容尚未消费
// structuredContent 的 MCP 客户端。
func previewStructuredResult(structured map[string]any) *mcp.CallToolResult {
	payload, err := json.Marshal(structured)
	if err != nil {
		return previewErrorResult(fmt.Sprintf("序列化 Preview 工具结果失败: %s", err.Error()))
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(payload)},
		},
		StructuredContent: structured,
	}
}

// previewFileContentResult 将元数据作为结构化结果返回，并单独附加原始源码，避免
// structuredContent 与兼容 JSON 文本重复携带大段文件内容。
func previewFileContentResult(structured map[string]any, filename, mimeType, content string) *mcp.CallToolResult {
	result := previewStructuredResult(structured)
	if result.IsError {
		return result
	}
	result.Content = append(result.Content, &mcp.TextContent{
		Text: fmt.Sprintf("=== %s（%s）===\n\n%s", filename, mimeType, content),
	})
	return result
}

func previewErrorResult(message string) *mcp.CallToolResult {
	return errorTextResult(message)
}
