package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/xiaolfeng/Lumina/internal/logic"
)

var previewLpwLogic *logic.PreviewLpwLogic

// SetPreviewLpwLogic 设置 PreviewLpwLogic 实例
func SetPreviewLpwLogic(l *logic.PreviewLpwLogic) {
	previewLpwLogic = l
}

var previewLpwToolHandlers = map[string]mcp.ToolHandler{
	"preview_lpw_init":         handlePreviewLpwInit,
	"preview_lpw_block_add":    handlePreviewLpwBlockAdd,
	"preview_lpw_block_edit":   handlePreviewLpwBlockEdit,
	"preview_lpw_block_remove": handlePreviewLpwBlockRemove,
	"preview_lpw_block_sort":   handlePreviewLpwBlockSort,
	"preview_lpw_meta_set":     handlePreviewLpwMetaSet,
	"preview_lpw_outline":      handlePreviewLpwOutline,
}

func parseLpwSessionAndFilename(args map[string]any) (sessionID xSnowflake.SnowflakeID, filename string, revision string, errResult *mcp.CallToolResult) {
	sIDStr, _ := args["session_id"].(string)
	if sIDStr == "" {
		return 0, "", "", previewErrorResult("缺少必填参数 session_id")
	}
	sIDInt, err := strconv.ParseInt(sIDStr, 10, 64)
	if err != nil {
		return 0, "", "", previewErrorResult(fmt.Sprintf("参数 session_id 非法: %s", err.Error()))
	}

	fn, _ := args["filename"].(string)
	if fn == "" {
		fn = "index.lpw"
	}

	rev, _ := args["revision"].(string)
	return xSnowflake.SnowflakeID(sIDInt), fn, rev, nil
}

func buildLpwWriteResponse(
	ctx context.Context,
	sessionID xSnowflake.SnowflakeID,
	writeResult *logic.LpwWriteResult,
	message string,
) *mcp.CallToolResult {
	snapshot, errMsg := loadPreviewSessionSnapshot(ctx, sessionID, writeResult.File.Filename)
	if errMsg != "" {
		return previewErrorResult(errMsg)
	}

	state, nextTool, workflowMsg, instructions := previewWriteWorkflow(snapshot, "分块写入成功")

	resultData := map[string]any{
		"status":        "success",
		"message":       message,
		"total_blocks":  writeResult.TotalBlocks,
		"file_size":     writeResult.FileSize,
		"revision":      writeResult.Revision.UTC().Format(time.RFC3339),
		"session":       previewSessionData(snapshot.session, snapshot.previewURL),
		"file":          previewFileData(writeResult.File),
		"entry_file":    snapshot.entryFilename,
		"preview_url":   snapshot.previewURL,
		"qa_supplement": previewSupplementData(snapshot.session.ID.String(), snapshot.entryID),
		"workflow": map[string]any{
			"state":        state,
			"next_tool":    nextTool,
			"message":      workflowMsg,
			"instructions": instructions,
		},
	}
	if writeResult.BlockID != "" {
		resultData["block_id"] = writeResult.BlockID
	}

	return previewStructuredResult(resultData)
}

func handlePreviewLpwInit(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if previewLpwLogic == nil {
		return previewErrorResult("PreviewLpwLogic 未初始化"), nil
	}
	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return previewErrorResult(errMsg), nil
	}

	sessionID, filename, revision, errRes := parseLpwSessionAndFilename(args)
	if errRes != nil {
		return errRes, nil
	}

	var title, desc, author string
	var tags []string
	if rawMeta, ok := args["meta"].(map[string]any); ok && rawMeta != nil {
		title, _ = rawMeta["title"].(string)
		desc, _ = rawMeta["description"].(string)
		author, _ = rawMeta["author"].(string)
		if rawTags, ok := rawMeta["tags"].([]any); ok {
			for _, t := range rawTags {
				if tStr, ok := t.(string); ok {
					tags = append(tags, tStr)
				}
			}
		}
	}
	if title == "" {
		title = "LPW 预览文档"
	}

	// Q-05：解析可选初始块列表（design 0003 init 契约）
	var initialBlocks []logic.LpwBlockExport
	if rawBlocks, ok := args["blocks"].([]any); ok && len(rawBlocks) > 0 {
		initialBlocks = make([]logic.LpwBlockExport, 0, len(rawBlocks))
		for i, rb := range rawBlocks {
			rbMap, isMap := rb.(map[string]any)
			if !isMap || rbMap == nil {
				return previewErrorResult(fmt.Sprintf("blocks[%d] 必须是对象", i)), nil
			}
			blockBytes, err := json.Marshal(rbMap)
			if err != nil {
				return previewErrorResult(fmt.Sprintf("序列化 blocks[%d] 失败: %s", i, err.Error())), nil
			}
			var b logic.LpwBlockRaw
			if err := json.Unmarshal(blockBytes, &b); err != nil {
				return previewErrorResult(fmt.Sprintf("解析 blocks[%d] 结构失败: %s", i, err.Error())), nil
			}
			initialBlocks = append(initialBlocks, b.ToInternal())
		}
	}

	res, xErr := previewLpwLogic.InitDocument(ctx, sessionID, filename, title, desc, author, tags, initialBlocks, revision)
	if xErr != nil {
		return previewErrorResult(xErr.Error()), nil
	}

	return buildLpwWriteResponse(ctx, sessionID, res, fmt.Sprintf("已成功初始化 LPW 文档 %s", filename)), nil
}

func handlePreviewLpwBlockAdd(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if previewLpwLogic == nil {
		return previewErrorResult("PreviewLpwLogic 未初始化"), nil
	}
	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return previewErrorResult(errMsg), nil
	}

	sessionID, filename, revision, errRes := parseLpwSessionAndFilename(args)
	if errRes != nil {
		return errRes, nil
	}

	rawBlock, ok := args["block"].(map[string]any)
	if !ok || rawBlock == nil {
		return previewErrorResult("缺少必填参数 block"), nil
	}

	blockBytes, err := json.Marshal(rawBlock)
	if err != nil {
		return previewErrorResult("序列化 block 参数失败: " + err.Error()), nil
	}

	var b logic.LpwBlockRaw
	if err := json.Unmarshal(blockBytes, &b); err != nil {
		return previewErrorResult("解析 block 结构失败: " + err.Error()), nil
	}

	parentID, _ := args["parent_id"].(string)
	var posPtr *int
	if posRaw, ok := args["position"].(float64); ok {
		p := int(posRaw)
		posPtr = &p
	}

	res, xErr := previewLpwLogic.AddBlock(ctx, sessionID, filename, parentID, posPtr, b.ToInternal(), revision)
	if xErr != nil {
		return previewErrorResult(xErr.Error()), nil
	}

	return buildLpwWriteResponse(ctx, sessionID, res, fmt.Sprintf("成功添加块 [%s#%s]", b.Type, b.ID)), nil
}

func handlePreviewLpwBlockEdit(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if previewLpwLogic == nil {
		return previewErrorResult("PreviewLpwLogic 未初始化"), nil
	}
	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return previewErrorResult(errMsg), nil
	}

	sessionID, filename, revision, errRes := parseLpwSessionAndFilename(args)
	if errRes != nil {
		return errRes, nil
	}

	blockID, _ := args["block_id"].(string)
	if blockID == "" {
		return previewErrorResult("缺少必填参数 block_id"), nil
	}

	var propsPatch map[string]any
	if rawPatch, ok := args["props"].(map[string]any); ok {
		propsPatch = rawPatch
	}

	var replaceBlock *logic.LpwBlockExport
	if rawBlock, ok := args["block"].(map[string]any); ok && rawBlock != nil {
		blockBytes, _ := json.Marshal(rawBlock)
		var b logic.LpwBlockRaw
		if err := json.Unmarshal(blockBytes, &b); err == nil {
			internalB := b.ToInternal()
			replaceBlock = &internalB
		}
	}

	if propsPatch == nil && replaceBlock == nil {
		return previewErrorResult("必须提供 props 或 block 之一进行编辑"), nil
	}

	res, xErr := previewLpwLogic.EditBlock(ctx, sessionID, filename, blockID, propsPatch, replaceBlock, revision)
	if xErr != nil {
		return previewErrorResult(xErr.Error()), nil
	}

	return buildLpwWriteResponse(ctx, sessionID, res, fmt.Sprintf("成功修改块 [%s]", blockID)), nil
}

func handlePreviewLpwBlockRemove(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if previewLpwLogic == nil {
		return previewErrorResult("PreviewLpwLogic 未初始化"), nil
	}
	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return previewErrorResult(errMsg), nil
	}

	sessionID, filename, revision, errRes := parseLpwSessionAndFilename(args)
	if errRes != nil {
		return errRes, nil
	}

	rawIDs, ok := args["block_ids"].([]any)
	if !ok || len(rawIDs) == 0 {
		return previewErrorResult("缺少必填参数 block_ids"), nil
	}
	var blockIDs []string
	for _, it := range rawIDs {
		if s, ok := it.(string); ok && s != "" {
			blockIDs = append(blockIDs, s)
		}
	}

	res, xErr := previewLpwLogic.RemoveBlocks(ctx, sessionID, filename, blockIDs, revision)
	if xErr != nil {
		return previewErrorResult(xErr.Error()), nil
	}

	return buildLpwWriteResponse(ctx, sessionID, res, fmt.Sprintf("成功删除 %d 个块", len(blockIDs))), nil
}

func handlePreviewLpwBlockSort(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if previewLpwLogic == nil {
		return previewErrorResult("PreviewLpwLogic 未初始化"), nil
	}
	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return previewErrorResult(errMsg), nil
	}

	sessionID, filename, revision, errRes := parseLpwSessionAndFilename(args)
	if errRes != nil {
		return errRes, nil
	}

	rawOrder, ok := args["order"].([]any)
	if !ok || len(rawOrder) == 0 {
		return previewErrorResult("缺少必填参数 order"), nil
	}
	var order []string
	for _, it := range rawOrder {
		if s, ok := it.(string); ok {
			order = append(order, s)
		}
	}

	parentID, _ := args["parent_id"].(string)

	res, xErr := previewLpwLogic.SortBlocks(ctx, sessionID, filename, parentID, order, revision)
	if xErr != nil {
		return previewErrorResult(xErr.Error()), nil
	}

	return buildLpwWriteResponse(ctx, sessionID, res, fmt.Sprintf("成功重排 %d 个块", len(order))), nil
}

func handlePreviewLpwMetaSet(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if previewLpwLogic == nil {
		return previewErrorResult("PreviewLpwLogic 未初始化"), nil
	}
	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return previewErrorResult(errMsg), nil
	}

	sessionID, filename, revision, errRes := parseLpwSessionAndFilename(args)
	if errRes != nil {
		return errRes, nil
	}

	rawMeta, ok := args["meta"].(map[string]any)
	if !ok || rawMeta == nil {
		return previewErrorResult("缺少必填参数 meta"), nil
	}

	metaBytes, err := json.Marshal(rawMeta)
	if err != nil {
		return previewErrorResult("序列化 meta 参数失败: " + err.Error()), nil
	}

	var m logic.LpwMetaExport
	if err := json.Unmarshal(metaBytes, &m); err != nil {
		return previewErrorResult("解析 meta 结构失败: " + err.Error()), nil
	}

	res, xErr := previewLpwLogic.SetMeta(ctx, sessionID, filename, m, revision)
	if xErr != nil {
		return previewErrorResult(xErr.Error()), nil
	}

	return buildLpwWriteResponse(ctx, sessionID, res, "成功更新文档元数据"), nil
}

func handlePreviewLpwOutline(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if previewLpwLogic == nil {
		return previewErrorResult("PreviewLpwLogic 未初始化"), nil
	}
	args := parseArgs(req.Params.Arguments)
	if errMsg := checkParseError(args); errMsg != "" {
		return previewErrorResult(errMsg), nil
	}

	sessionID, filename, _, errRes := parseLpwSessionAndFilename(args)
	if errRes != nil {
		return errRes, nil
	}

	outline, xErr := previewLpwLogic.Outline(ctx, sessionID, filename)
	if xErr != nil {
		return previewErrorResult(xErr.Error()), nil
	}

	resultData := map[string]any{
		"status":  "success",
		"message": fmt.Sprintf("获取大纲成功，包含 %d 个块", outline.BlockCount),
		"outline": outline,
	}

	return previewStructuredResult(resultData), nil
}
