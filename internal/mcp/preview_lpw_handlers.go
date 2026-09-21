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
	"preview_lpw_init":        handlePreviewLpwInit,
	"preview_lpw_node_add":    handlePreviewLpwNodeAdd,
	"preview_lpw_node_edit":   handlePreviewLpwNodeEdit,
	"preview_lpw_node_remove": handlePreviewLpwNodeRemove,
	"preview_lpw_node_sort":   handlePreviewLpwNodeSort,
	"preview_lpw_meta_set":    handlePreviewLpwMetaSet,
	"preview_lpw_outline":     handlePreviewLpwOutline,
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

	state, nextTool, workflowMsg, instructions := previewWriteWorkflow(snapshot, "节点写入成功")

	resultData := map[string]any{
		"status":        "success",
		"message":       message,
		"total_nodes":   writeResult.TotalNodes,
		"node_kind":     writeResult.NodeKind,
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
	if writeResult.NodeID != "" {
		resultData["node_id"] = writeResult.NodeID
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
	if m.Title == "" {
		return previewErrorResult("meta.title 不能为空"), nil
	}

	var initialNodes []logic.LpwNodeExport
	if rawContent, ok := args["content"].([]any); ok && len(rawContent) > 0 {
		initialNodes = make([]logic.LpwNodeExport, 0, len(rawContent))
		for i, rn := range rawContent {
			rnMap, isMap := rn.(map[string]any)
			if !isMap || rnMap == nil {
				return previewErrorResult(fmt.Sprintf("content[%d] 必须是对象", i)), nil
			}
			nodeBytes, err := json.Marshal(rnMap)
			if err != nil {
				return previewErrorResult(fmt.Sprintf("序列化 content[%d] 失败: %s", i, err.Error())), nil
			}
			var n logic.LpwNodeRaw
			if err := json.Unmarshal(nodeBytes, &n); err != nil {
				return previewErrorResult(fmt.Sprintf("解析 content[%d] 结构失败: %s", i, err.Error())), nil
			}
			initialNodes = append(initialNodes, n.ToInternal())
		}
	}

	res, xErr := previewLpwLogic.InitDocument(ctx, sessionID, filename, m, initialNodes, revision)
	if xErr != nil {
		return previewErrorResult(xErr.Error()), nil
	}

	return buildLpwWriteResponse(ctx, sessionID, res, fmt.Sprintf("已成功初始化 LPW 文档 %s", filename)), nil
}

func handlePreviewLpwNodeAdd(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	rawNode, ok := args["node"].(map[string]any)
	if !ok || rawNode == nil {
		return previewErrorResult("缺少必填参数 node"), nil
	}

	nodeBytes, err := json.Marshal(rawNode)
	if err != nil {
		return previewErrorResult("序列化 node 参数失败: " + err.Error()), nil
	}

	var n logic.LpwNodeRaw
	if err := json.Unmarshal(nodeBytes, &n); err != nil {
		return previewErrorResult("解析 node 结构失败: " + err.Error()), nil
	}

	parentID, _ := args["parent_id"].(string)
	var posPtr *int
	if posRaw, ok := args["position"].(float64); ok {
		p := int(posRaw)
		posPtr = &p
	}

	node := n.ToInternal()
	res, xErr := previewLpwLogic.AddNode(ctx, sessionID, filename, node, parentID, posPtr, revision)
	if xErr != nil {
		return previewErrorResult(xErr.Error()), nil
	}

	return buildLpwWriteResponse(ctx, sessionID, res, fmt.Sprintf("成功添加节点 [%s/%s#%s]", node.Kind, node.Type, node.ID)), nil
}

func handlePreviewLpwNodeEdit(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	nodeID, _ := args["node_id"].(string)
	if nodeID == "" {
		return previewErrorResult("缺少必填参数 node_id"), nil
	}

	var propsPatch map[string]any
	if rawPatch, ok := args["props"].(map[string]any); ok {
		propsPatch = rawPatch
	}

	var annPatch *logic.LpwNodeAnnotationPatch
	if rawAnn, exists := args["annotation"]; exists {
		if rawAnn == nil {
			annPatch = &logic.LpwNodeAnnotationPatch{Clear: true}
		} else {
			m, ok := rawAnn.(map[string]any)
			if !ok {
				return previewErrorResult("annotation 参数必须为对象或 null"), nil
			}
			b, err := json.Marshal(m)
			if err != nil {
				return previewErrorResult("序列化 annotation 参数失败: " + err.Error()), nil
			}
			var a logic.LpwAnnotationExport
			if err := json.Unmarshal(b, &a); err != nil {
				return previewErrorResult("解析 annotation 结构失败: " + err.Error()), nil
			}
			annPatch = &logic.LpwNodeAnnotationPatch{Value: &a}
		}
	}

	var replacement *logic.LpwNodeExport
	if rawNode, exists := args["node"]; exists && rawNode != nil {
		m, ok := rawNode.(map[string]any)
		if !ok {
			return previewErrorResult("node 参数必须为对象"), nil
		}
		nodeBytes, err := json.Marshal(m)
		if err != nil {
			return previewErrorResult("序列化 node 参数失败: " + err.Error()), nil
		}
		var n logic.LpwNodeRaw
		if err := json.Unmarshal(nodeBytes, &n); err != nil {
			return previewErrorResult("解析 node 结构失败: " + err.Error()), nil
		}
		internalNode := n.ToInternal()
		replacement = &internalNode
	}

	if replacement != nil && (propsPatch != nil || annPatch != nil) {
		return previewErrorResult("整节点替换 node 与 props/annotation 互斥，不能同时提供"), nil
	}

	if propsPatch == nil && annPatch == nil && replacement == nil {
		return previewErrorResult("必须提供 props、annotation 或 node 之一进行编辑"), nil
	}

	res, xErr := previewLpwLogic.EditNode(ctx, sessionID, filename, nodeID, propsPatch, annPatch, replacement, revision)
	if xErr != nil {
		return previewErrorResult(xErr.Error()), nil
	}

	return buildLpwWriteResponse(ctx, sessionID, res, fmt.Sprintf("成功修改节点 [%s]", nodeID)), nil
}

func handlePreviewLpwNodeRemove(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	rawIDs, ok := args["node_ids"].([]any)
	if !ok || len(rawIDs) == 0 {
		return previewErrorResult("缺少必填参数 node_ids"), nil
	}
	var nodeIDs []string
	for _, it := range rawIDs {
		if s, ok := it.(string); ok && s != "" {
			nodeIDs = append(nodeIDs, s)
		}
	}

	res, xErr := previewLpwLogic.RemoveNodes(ctx, sessionID, filename, nodeIDs, revision)
	if xErr != nil {
		return previewErrorResult(xErr.Error()), nil
	}

	return buildLpwWriteResponse(ctx, sessionID, res, fmt.Sprintf("成功删除 %d 个节点", len(nodeIDs))), nil
}

func handlePreviewLpwNodeSort(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	res, xErr := previewLpwLogic.SortNodes(ctx, sessionID, filename, parentID, order, revision)
	if xErr != nil {
		return previewErrorResult(xErr.Error()), nil
	}

	return buildLpwWriteResponse(ctx, sessionID, res, fmt.Sprintf("成功重排 %d 个节点", len(order))), nil
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

	message := fmt.Sprintf("获取大纲成功，包含 %d 个节点", outline.NodeCount)
	if len(outline.CompletenessWarnings) > 0 {
		message = fmt.Sprintf(
			"获取大纲成功，包含 %d 个节点；注意：文档尚未满足完成态契约（%d 条完备性告警），收工前需补齐",
			outline.NodeCount, len(outline.CompletenessWarnings),
		)
	}

	resultData := map[string]any{
		"status":  "success",
		"message": message,
		"outline": outline,
	}

	return previewStructuredResult(resultData), nil
}
