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

// LpwWriteResult 块级操作成功响应结构
type LpwWriteResult struct {
	BlockID     string                          `json:"block_id,omitempty"`
	TotalBlocks int                             `json:"total_blocks"`
	FileSize    int                             `json:"file_size"`
	Revision    time.Time                       `json:"revision"`
	File        *apiPreview.PreviewFileResponse `json:"file"`
}

// LpwOutlineItem 大纲条目
type LpwOutlineItem struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Depth      int    `json:"depth"`
	Children   int    `json:"children"`
	PropsBytes int    `json:"props_bytes"`
}

// LpwOutline 大纲响应结构
type LpwOutline struct {
	Version    string           `json:"version"`
	BlockCount int              `json:"block_count"`
	TotalSize  int              `json:"total_size"`
	Revision   string           `json:"revision"`
	Items      []LpwOutlineItem `json:"items"`
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

// lpwFileLocks 跨写通道共享的 LPW 文件互斥锁（key: "<sessionID>:<filename>"）。
// Q-04 修复：preview_lpw_* 的读改验写流水线与 preview_file_upload 整体覆写
// 必须串行化，否则并发时后写者会覆盖前者的更新（丢写）。
var lpwFileLocks sync.Map

// withLpwFileLock 以跨通道共享的互斥锁执行 fn；若当前调用链已持有同 key 锁
// （经 ctx 传递，如 withDocument → UploadFile），则直接执行避免自死锁。
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

// withDocument 原子读改验写流水线
func (l *PreviewLpwLogic) withDocument(
	ctx context.Context,
	sessionID xSnowflake.SnowflakeID,
	filename string,
	revision string,
	allowCreate bool,
	mutate func(doc *lpwDocument) (affectedBlockID string, err error),
) (*LpwWriteResult, *xError.Error) {
	if strings.ToLower(filepath.Ext(filename)) != ".lpw" {
		return nil, xError.NewError(ctx, xError.ParameterError, "文件扩展名必须为 .lpw", false, nil)
	}

	var result *LpwWriteResult
	var xErr *xError.Error

	withLpwFileLock(ctx, sessionID, filename, func(lockedCtx context.Context) {
		result, xErr = l.withDocumentLocked(lockedCtx, sessionID, filename, revision, allowCreate, mutate)
	})
	return result, xErr
}

// withDocumentLocked 在已持有文件锁的前提下执行读改验写流水线
func (l *PreviewLpwLogic) withDocumentLocked(
	ctx context.Context,
	sessionID xSnowflake.SnowflakeID,
	filename string,
	revision string,
	allowCreate bool,
	mutate func(doc *lpwDocument) (affectedBlockID string, err error),
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
	} else {
		doc = lpwDocument{
			Version: "1.0",
			Blocks:  []lpwBlock{},
		}
	}

	// 5. mutate 变更操作
	affectedID, mutateErr := mutate(&doc)
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

	// 8. 结构规则走查
	if err := validateContainerRules(&doc); err != nil {
		return nil, xError.NewError(ctx, xError.ParameterError, xError.ErrMessage(fmt.Sprintf("LPW 结构规则校验失败: %s", err.Error())), false, nil)
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

	_, totalCount := collectBlockIDs(&doc)

	var parsedRev time.Time
	if fileResp.UpdatedAt != "" {
		if t, parseErr := time.Parse(time.RFC3339, fileResp.UpdatedAt); parseErr == nil {
			parsedRev = t
		}
	}

	return &LpwWriteResult{
		BlockID:     affectedID,
		TotalBlocks: totalCount,
		FileSize:    len(serialized),
		Revision:    parsedRev,
		File:        fileResp,
	}, nil
}

// InitDocument 初始化空白 LPW 文档并写入 Meta 元数据；initialBlocks 非空时作为初始块列表
func (l *PreviewLpwLogic) InitDocument(
	ctx context.Context,
	sessionID xSnowflake.SnowflakeID,
	filename string,
	title string,
	desc string,
	author string,
	tags []string,
	initialBlocks []lpwBlock,
	revision string,
) (*LpwWriteResult, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("InitDocument - 初始化 LPW 文档 [session=%d, file=%s, title=%s, initialBlocks=%d]", sessionID.Int64(), filename, title, len(initialBlocks)))

	return l.withDocument(ctx, sessionID, filename, revision, true, func(doc *lpwDocument) (string, error) {
		doc.Version = "1.0"
		doc.Meta = &lpwMeta{
			Title:       title,
			Description: desc,
			Author:      author,
			Tags:        tags,
		}
		if initialBlocks != nil {
			doc.Blocks = initialBlocks // Q-05：支持携带初始块，仍走统一校验管线
		} else {
			doc.Blocks = []lpwBlock{} // 重置清空
		}
		return "", nil
	})
}

// AddBlock 插入单个块或子树
func (l *PreviewLpwLogic) AddBlock(
	ctx context.Context,
	sessionID xSnowflake.SnowflakeID,
	filename string,
	parentID string,
	position *int,
	block lpwBlock,
	revision string,
) (*LpwWriteResult, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("AddBlock - 添加 LPW 块 [session=%d, file=%s, type=%s, id=%s]", sessionID.Int64(), filename, block.Type, block.ID))

	return l.withDocument(ctx, sessionID, filename, revision, false, func(doc *lpwDocument) (string, error) {
		if err := insertBlock(doc, parentID, position, block); err != nil {
			return "", err
		}
		return block.ID, nil
	})
}

// EditBlock 编辑修改块（支持局部 patchProps 或整体 replaceBlock）
func (l *PreviewLpwLogic) EditBlock(
	ctx context.Context,
	sessionID xSnowflake.SnowflakeID,
	filename string,
	blockID string,
	propsPatch map[string]any,
	replace *lpwBlock,
	revision string,
) (*LpwWriteResult, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("EditBlock - 编辑 LPW 块 [session=%d, file=%s, id=%s]", sessionID.Int64(), filename, blockID))

	return l.withDocument(ctx, sessionID, filename, revision, false, func(doc *lpwDocument) (string, error) {
		if replace != nil {
			if err := replaceBlock(doc, blockID, *replace); err != nil {
				return "", err
			}
			return replace.ID, nil
		}
		if propsPatch != nil {
			if err := patchProps(doc, blockID, propsPatch); err != nil {
				return "", err
			}
			return blockID, nil
		}
		return "", fmt.Errorf("必须提供 props_patch 或 replace 之一")
	})
}

// RemoveBlocks 批量删除指定块
func (l *PreviewLpwLogic) RemoveBlocks(
	ctx context.Context,
	sessionID xSnowflake.SnowflakeID,
	filename string,
	ids []string,
	revision string,
) (*LpwWriteResult, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("RemoveBlocks - 删除 LPW 块 [session=%d, file=%s, ids=%v]", sessionID.Int64(), filename, ids))

	return l.withDocument(ctx, sessionID, filename, revision, false, func(doc *lpwDocument) (string, error) {
		if err := removeBlocks(doc, ids); err != nil {
			return "", err
		}
		return "", nil
	})
}

// SortBlocks 重排同一容器内子块顺序
func (l *PreviewLpwLogic) SortBlocks(
	ctx context.Context,
	sessionID xSnowflake.SnowflakeID,
	filename string,
	parentID string,
	order []string,
	revision string,
) (*LpwWriteResult, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("SortBlocks - 重新排序 LPW 块 [session=%d, file=%s, parent=%s, order=%v]", sessionID.Int64(), filename, parentID, order))

	return l.withDocument(ctx, sessionID, filename, revision, false, func(doc *lpwDocument) (string, error) {
		if err := reorderSiblings(doc, parentID, order); err != nil {
			return "", err
		}
		return "", nil
	})
}

// SetMeta 更新 Meta 元数据
func (l *PreviewLpwLogic) SetMeta(
	ctx context.Context,
	sessionID xSnowflake.SnowflakeID,
	filename string,
	meta lpwMeta,
	revision string,
) (*LpwWriteResult, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("SetMeta - 更新 LPW 文档 Meta [session=%d, file=%s, title=%s]", sessionID.Int64(), filename, meta.Title))

	return l.withDocument(ctx, sessionID, filename, revision, false, func(doc *lpwDocument) (string, error) {
		doc.Meta = &meta
		return "", nil
	})
}

// Outline 读取文档大纲结构（轻量，不返回 props 详情）
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

	var doc lpwDocument
	if err := json.Unmarshal([]byte(content), &doc); err != nil {
		return nil, xError.NewError(ctx, xError.ParameterError, "文档已损坏或不是合法的 LPW JSON", false, err)
	}

	var items []LpwOutlineItem
	var walk func(blocks []lpwBlock, depth int)
	walk = func(blocks []lpwBlock, depth int) {
		for _, b := range blocks {
			propsBytes := 0
			if b.Props != nil {
				if pb, mErr := json.Marshal(b.Props); mErr == nil {
					propsBytes = len(pb)
				}
			}
			items = append(items, LpwOutlineItem{
				ID:         b.ID,
				Type:       b.Type,
				Depth:      depth,
				Children:   len(b.Children),
				PropsBytes: propsBytes,
			})
			if len(b.Children) > 0 {
				walk(b.Children, depth+1)
			}
		}
	}
	walk(doc.Blocks, 1)

	return &LpwOutline{
		Version:    doc.Version,
		BlockCount: len(items),
		TotalSize:  len(content),
		Revision:   updatedAt.UTC().Format(time.RFC3339),
		Items:      items,
	}, nil
}
