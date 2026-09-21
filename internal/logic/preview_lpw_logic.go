package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	apiPreview "github.com/xiaolfeng/Lumina/api/preview"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/service"
)

// lpwStorage 抽象预览文件读写接口，支持单测纯内存注入
type lpwStorage interface {
	GetFile(ctx context.Context, sessionID xSnowflake.SnowflakeID, filename string) (content string, updatedAt time.Time, err *xError.Error)
	SaveFile(ctx context.Context, sessionID xSnowflake.SnowflakeID, filename, content string) (*apiPreview.PreviewFileResponse, *xError.Error)
}

type defaultLpwStorage struct {
	previewLogic *PreviewLogic
}

func (s *defaultLpwStorage) GetFile(ctx context.Context, sessionID xSnowflake.SnowflakeID, filename string) (string, time.Time, *xError.Error) {
	file, err := s.previewLogic.GetFileBySessionAndFilename(ctx, sessionID, filename)
	if err != nil {
		return "", time.Time{}, err
	}
	return file.Content, file.UpdatedAt, nil
}

func (s *defaultLpwStorage) SaveFile(ctx context.Context, sessionID xSnowflake.SnowflakeID, filename, content string) (*apiPreview.PreviewFileResponse, *xError.Error) {
	return s.previewLogic.UploadFile(ctx, sessionID, filename, content)
}

// PreviewLpwLogic LPW 文档分块增量操作业务逻辑层
type PreviewLpwLogic struct {
	logic
	previewLogic *PreviewLogic
	schema       *service.LpwSchemaLoader
	storage      lpwStorage
}

// LpwWriteResult 节点级操作成功响应结构
type LpwWriteResult struct {
	NodeID     string                          `json:"node_id,omitempty"`
	NodeKind   string                          `json:"node_kind,omitempty"`
	TotalNodes int                             `json:"total_nodes"`
	FileSize   int                             `json:"file_size"`
	Revision   time.Time                       `json:"revision"`
	File       *apiPreview.PreviewFileResponse `json:"file"`
}

// LpwOutlineItem 大纲条目
type LpwOutlineItem struct {
	ID               string `json:"id"`
	Kind             string `json:"kind"`
	Type             string `json:"type"`
	PatternOrVariant string `json:"pattern_or_variant,omitempty"` // layout→pattern；container→variant
	JSONPath         string `json:"json_path"`
	Depth            int    `json:"depth"`
	Children         int    `json:"children"`
	PropsBytes       int    `json:"props_bytes"`
}

// LpwOutline 大纲响应结构
type LpwOutline struct {
	Version              string           `json:"version"`
	NodeCount            int              `json:"node_count"`
	TotalSize            int              `json:"total_size"`
	Revision             string           `json:"revision"`
	Items                []LpwOutlineItem `json:"items"`
	CompletenessWarnings []string         `json:"completeness_warnings"`
}

// NewPreviewLpwLogic 构造 PreviewLpwLogic 实例
func NewPreviewLpwLogic(previewLogic *PreviewLogic, schema *service.LpwSchemaLoader) *PreviewLpwLogic {
	return &PreviewLpwLogic{
		logic:        logic{log: xLog.WithName(xLog.NamedLOGC, "PreviewLpwLogic")},
		previewLogic: previewLogic,
		schema:       schema,
		storage:      &defaultLpwStorage{previewLogic: previewLogic},
	}
}

// lpwLockCtxKey 用于在调用链上下文中标记「已持有某 LPW 文件锁」，实现同 key 可重入
type lpwLockCtxKey struct{}

// lpwFileLocks 跨写通道共享的 LPW 文件互斥锁（key: "<sessionID>:<filename>"）
var lpwFileLocks sync.Map

func withLpwFileLock(
	ctx context.Context,
	sessionID xSnowflake.SnowflakeID,
	filename string,
	fn func(ctx context.Context),
) {
	key := fmt.Sprintf("%d:%s", sessionID.Int64(), filename)
	if held, _ := ctx.Value(lpwLockCtxKey{}).(string); held == key {
		fn(ctx)
		return
	}
	actual, _ := lpwFileLocks.LoadOrStore(key, &sync.Mutex{})
	mtx := actual.(*sync.Mutex)
	mtx.Lock()
	defer mtx.Unlock()
	fn(context.WithValue(ctx, lpwLockCtxKey{}, key))
}

// withDocument 原子读改验写流水线；strict=true 走完备态校验（一次性提交语义），否则渐进式
func (l *PreviewLpwLogic) withDocument(
	ctx context.Context,
	sessionID xSnowflake.SnowflakeID,
	filename string,
	revision string,
	allowCreate bool,
	strict bool,
	mutate func(doc *lpwDocument) (affectedNodeID string, affectedNodeKind string, err error),
) (*LpwWriteResult, *xError.Error) {
	if strings.ToLower(filepath.Ext(filename)) != ".lpw" {
		return nil, xError.NewError(ctx, xError.ParameterError, "文件扩展名必须为 .lpw", false, nil)
	}

	var result *LpwWriteResult
	var xErr *xError.Error

	withLpwFileLock(ctx, sessionID, filename, func(lockedCtx context.Context) {
		result, xErr = l.withDocumentLocked(lockedCtx, sessionID, filename, revision, allowCreate, strict, mutate)
	})
	return result, xErr
}

// withDocumentLocked 在已持有文件锁的前提下执行读改验写流水线；strict=true 走完备态校验
func (l *PreviewLpwLogic) withDocumentLocked(
	ctx context.Context,
	sessionID xSnowflake.SnowflakeID,
	filename string,
	revision string,
	allowCreate bool,
	strict bool,
	mutate func(doc *lpwDocument) (affectedNodeID string, affectedNodeKind string, err error),
) (*LpwWriteResult, *xError.Error) {
	// 2. 读当前文件
	content, updatedAt, getErr := l.storage.GetFile(ctx, sessionID, filename)
	if getErr != nil {
		if !allowCreate || getErr.GetErrorCode() != xError.NotFound {
			if getErr.GetErrorCode() == xError.NotFound {
				return nil, xError.NewError(ctx, xError.NotFound, "预览文件不存在，请先调用 preview_lpw_init 初始化文档", false, nil)
			}
			return nil, getErr
		}
	}

	// 3. revision 预检
	if getErr == nil && revision != "" {
		currentRev := updatedAt.UTC().Format(time.RFC3339)
		if revision != currentRev {
			return nil, xError.NewError(
				ctx,
				xError.BusinessError,
				xError.ErrMessage(fmt.Sprintf("文件已被修改（期望版本: %s，当前版本: %s），请先调用 preview_lpw_outline 重读后再重试", revision, currentRev)),
				false,
				nil,
			)
		}
	}

	// 4. 反序列化
	var doc lpwDocument
	if getErr == nil && content != "" {
		if err := json.Unmarshal([]byte(content), &doc); err != nil {
			return nil, xError.NewError(ctx, xError.ParameterError, "文档已损坏或不是合法的 LPW JSON，请先调用 preview_lpw_init 重置或使用 preview_file_upload 修复", false, err)
		}
		if doc.Version != "1.1" {
			return nil, xError.NewError(ctx, xError.ParameterError, xError.ErrMessage(fmt.Sprintf(
				"不支持的 LPW 版本 %q：仅支持 1.1；旧版文档请用 preview_lpw_init 重建", doc.Version)), false, nil)
		}
	} else {
		doc = lpwDocument{
			Version: "1.1",
			Content: []lpwNode{},
		}
	}

	// 5. mutate 变更操作
	affectedID, affectedKind, mutateErr := mutate(&doc)
	if mutateErr != nil {
		return nil, xError.NewError(ctx, xError.ParameterError, xError.ErrMessage(mutateErr.Error()), false, nil)
	}

	// 6. 稳定序列化（2 空格缩进）
	serialized, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, xError.NewError(ctx, xError.ParameterError, "序列化 LPW 文档失败", false, err)
	}

	// 7. Schema 规范校验
	if l.schema != nil {
		if failPath, reason, ok := l.schema.Validate(serialized, doc.Version); !ok {
			return nil, xError.NewError(ctx, xError.ParameterError, xError.ErrMessage(fmt.Sprintf("LPW Schema 校验失败 [%s]: %s", failPath, reason)), false, nil)
		}
	}

	// 8. 结构规则走查（strict=完备态下限全查；否则渐进式放宽下限，允许分步挂载 Layout 与 Container）
	structureErr := validateDocument11Progressive(&doc)
	if strict {
		structureErr = validateDocument11(&doc)
	}
	if structureErr != nil {
		return nil, xError.NewError(ctx, xError.ParameterError, xError.ErrMessage(fmt.Sprintf("LPW 结构规则校验失败: %s", structureErr.Error())), false, nil)
	}

	// 9. 大小上限检查
	if len(serialized) > bConst.PreviewFileMaxSize {
		return nil, xError.NewError(ctx, xError.ParameterError, xError.ErrMessage(fmt.Sprintf("LPW 文件大小超出上限（当前 %d 字节，最大上限 %d 字节）", len(serialized), bConst.PreviewFileMaxSize)), false, nil)
	}

	// 10. 落库并回读最新状态
	fileResp, saveErr := l.storage.SaveFile(ctx, sessionID, filename, string(serialized))
	if saveErr != nil {
		return nil, saveErr
	}

	_, totalCount := collectNodeIDs(&doc)

	var parsedRev time.Time
	if fileResp.UpdatedAt != "" {
		if t, parseErr := time.Parse(time.RFC3339, fileResp.UpdatedAt); parseErr == nil {
			parsedRev = t
		}
	}

	return &LpwWriteResult{
		NodeID:     affectedID,
		NodeKind:   affectedKind,
		TotalNodes: totalCount,
		FileSize:   len(serialized),
		Revision:   parsedRev,
		File:       fileResp,
	}, nil
}

// InitDocument 创建或重置新格式 LPW 文档，根字段为 content
func (l *PreviewLpwLogic) InitDocument(
	ctx context.Context,
	sessionID xSnowflake.SnowflakeID,
	filename string,
	meta lpwMeta,
	content []LpwNodeExport,
	revision string,
) (*LpwWriteResult, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("InitDocument - 初始化 LPW 文档 [session=%d, file=%s, title=%s, initialNodes=%d]", sessionID.Int64(), filename, meta.Title, len(content)))

	// init 是一次性提交语义：content 视为完整文档，走完备态校验
	return l.withDocument(ctx, sessionID, filename, revision, true, true, func(doc *lpwDocument) (string, string, error) {
		doc.Version = "1.1"
		doc.Meta = &meta
		if content != nil {
			doc.Content = content
			// Q-12 修复：与 insertNode 对齐，初始化全文档节点总数上限 500 检查
			_, initTotal := collectNodeIDs(doc)
			if initTotal > 500 {
				return "", "", fmt.Errorf("初始化总节点数（%d）超过上限 500", initTotal)
			}
		} else {
			doc.Content = []lpwNode{}
		}
		return "", "", nil
	})
}

// AddNode 插入单个 Layout / Container / Block 节点或子树
func (l *PreviewLpwLogic) AddNode(
	ctx context.Context,
	sessionID xSnowflake.SnowflakeID,
	filename string,
	node LpwNodeExport,
	parentID string,
	position *int,
	revision string,
) (*LpwWriteResult, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("AddNode - 添加 LPW 节点 [session=%d, file=%s, kind=%s, type=%s, id=%s]", sessionID.Int64(), filename, node.Kind, node.Type, node.ID))

	return l.withDocument(ctx, sessionID, filename, revision, false, false, func(doc *lpwDocument) (string, string, error) {
		if err := insertNode(doc, parentID, position, node); err != nil {
			return "", "", err
		}
		return node.ID, node.Kind, nil
	})
}

// EditNode 编辑节点（支持 props patch、annotation patch 或整节点 replace）
func (l *PreviewLpwLogic) EditNode(
	ctx context.Context,
	sessionID xSnowflake.SnowflakeID,
	filename string,
	nodeID string,
	propsPatch map[string]any,
	annotationPatch *LpwNodeAnnotationPatch,
	replacement *LpwNodeExport,
	revision string,
) (*LpwWriteResult, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("EditNode - 编辑 LPW 节点 [session=%d, file=%s, id=%s]", sessionID.Int64(), filename, nodeID))

	return l.withDocument(ctx, sessionID, filename, revision, false, false, func(doc *lpwDocument) (string, string, error) {
		if replacement != nil {
			if err := replaceNode(doc, nodeID, *replacement); err != nil {
				return "", "", err
			}
			return replacement.ID, replacement.Kind, nil
		}

		node, _, found := findNodeDirect(doc, nodeID)
		if !found {
			return "", "", fmt.Errorf("待编辑的目标节点 id %q 不存在", nodeID)
		}
		nodeKind := node.Kind

		if propsPatch != nil {
			if err := patchNodeProps(doc, nodeID, propsPatch); err != nil {
				return "", "", err
			}
		}
		if annotationPatch != nil {
			if err := patchNodeAnnotation(doc, nodeID, annotationPatch); err != nil {
				return "", "", err
			}
		}
		if propsPatch == nil && annotationPatch == nil {
			return "", "", fmt.Errorf("必须提供 props_patch、annotation 或 replace 之一")
		}
		return nodeID, nodeKind, nil
	})
}

// RemoveNodes 批量删除指定节点及其合法 children
func (l *PreviewLpwLogic) RemoveNodes(
	ctx context.Context,
	sessionID xSnowflake.SnowflakeID,
	filename string,
	nodeIDs []string,
	revision string,
) (*LpwWriteResult, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("RemoveNodes - 删除 LPW 节点 [session=%d, file=%s, ids=%v]", sessionID.Int64(), filename, nodeIDs))

	return l.withDocument(ctx, sessionID, filename, revision, false, false, func(doc *lpwDocument) (string, string, error) {
		if err := removeNodes(doc, nodeIDs); err != nil {
			return "", "", err
		}
		return "", "", nil
	})
}

// SortNodes 重排同一父节点下的子节点
func (l *PreviewLpwLogic) SortNodes(
	ctx context.Context,
	sessionID xSnowflake.SnowflakeID,
	filename string,
	parentID string,
	order []string,
	revision string,
) (*LpwWriteResult, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("SortNodes - 重新排序 LPW 节点 [session=%d, file=%s, parent=%s, order=%v]", sessionID.Int64(), filename, parentID, order))

	return l.withDocument(ctx, sessionID, filename, revision, false, false, func(doc *lpwDocument) (string, string, error) {
		if err := reorderSiblings(doc, parentID, order); err != nil {
			return "", "", err
		}
		return "", "", nil
	})
}

// SetMeta 浅合并或替换 Meta 元数据
func (l *PreviewLpwLogic) SetMeta(
	ctx context.Context,
	sessionID xSnowflake.SnowflakeID,
	filename string,
	meta lpwMeta,
	revision string,
) (*LpwWriteResult, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("SetMeta - 更新 LPW 文档 Meta [session=%d, file=%s, title=%s]", sessionID.Int64(), filename, meta.Title))

	return l.withDocument(ctx, sessionID, filename, revision, false, false, func(doc *lpwDocument) (string, string, error) {
		doc.Meta = &meta
		return "", "", nil
	})
}

// Outline 读取文档大纲结构（返回 kind、type、pattern/variant、json_path 与体积摘要）
func (l *PreviewLpwLogic) Outline(
	ctx context.Context,
	sessionID xSnowflake.SnowflakeID,
	filename string,
) (*LpwOutline, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("Outline - 获取 LPW 文档大纲 [session=%d, file=%s]", sessionID.Int64(), filename))

	content, updatedAt, err := l.storage.GetFile(ctx, sessionID, filename)
	if err != nil {
		return nil, err
	}

	var rawMap map[string]any
	if err := json.Unmarshal([]byte(content), &rawMap); err != nil {
		return nil, xError.NewError(ctx, xError.ParameterError, "文档已损坏或不是合法的 LPW JSON", false, err)
	}

	// Q-11 修复：版本不符与旧版 blocks 属于致命契约错误，直接报错返回，严禁混入 CompletenessWarnings
	if _, hasBlocks := rawMap["blocks"]; hasBlocks {
		return nil, xError.NewError(ctx, xError.ParameterError, "检测到旧版 LPW 文档（根字段 blocks）；仅支持 1.1 content 结构，请用 preview_lpw_init 重建", false, nil)
	}

	var doc lpwDocument
	if err := json.Unmarshal([]byte(content), &doc); err != nil {
		return nil, xError.NewError(ctx, xError.ParameterError, "文档已损坏或不是合法的 LPW JSON", false, err)
	}

	if doc.Version != "1.1" {
		return nil, xError.NewError(ctx, xError.ParameterError, xError.ErrMessage(fmt.Sprintf(
			"不支持的 LPW 版本 %q：仅支持 1.1；旧版文档请用 preview_lpw_init 重建", doc.Version)), false, nil)
	}

	var items []LpwOutlineItem
	var walk func(nodes []lpwNode, currentPath string, depth int)
	walk = func(nodes []lpwNode, currentPath string, depth int) {
		for i, n := range nodes {
			jsonPath := fmt.Sprintf("%s[%d]", currentPath, i)
			propsBytes := 0
			if n.Props != nil {
				if pb, mErr := json.Marshal(n.Props); mErr == nil {
					propsBytes = len(pb)
				}
			}

			pv := ""
			if n.Kind == "layout" {
				if pat, _ := n.Props["pattern"].(string); pat != "" {
					pv = pat
				}
			} else if n.Kind == "container" {
				if vr, _ := n.Props["variant"].(string); vr != "" {
					pv = vr
				}
			}

			items = append(items, LpwOutlineItem{
				ID:               n.ID,
				Kind:             n.Kind,
				Type:             n.Type,
				PatternOrVariant: pv,
				JSONPath:         jsonPath,
				Depth:            depth,
				Children:         len(n.Children),
				PropsBytes:       propsBytes,
			})
			if len(n.Children) > 0 {
				walk(n.Children, fmt.Sprintf("%s/children", jsonPath), depth+1)
			}
		}
	}
	walk(doc.Content, "/content", 1)

	// 完备态契约走查：渐进式构建的中间态在此暴露，供 Agent 收工前核对
	completenessWarnings := []string{}
	if err := validateDocument11(&doc); err != nil {
		completenessWarnings = append(completenessWarnings, err.Error())
	}

	return &LpwOutline{
		Version:              doc.Version,
		NodeCount:            len(items),
		TotalSize:            len(content),
		Revision:             updatedAt.UTC().Format(time.RFC3339),
		Items:                items,
		CompletenessWarnings: completenessWarnings,
	}, nil
}
