package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	xCtxUtil "github.com/bamboo-services/bamboo-base-go/major/utility/context"
	apiPreview "github.com/xiaolfeng/Lumina/api/preview"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"github.com/xiaolfeng/Lumina/internal/repository"
)

// previewRepo Preview 模块依赖的仓储集合
type previewRepo struct {
	session *repository.PreviewSessionRepo
	file    *repository.PreviewFileRepo
	page    *repository.PageRepo
	info    *repository.InfoRepo
}

// PreviewLogic Preview 业务逻辑层，负责预览会话管理、文件上传与渲染内容编排
type PreviewLogic struct {
	logic
	repo previewRepo
}

// OnPreviewChanged 预览内容变更后的回调钩子，由 WebSocket 层设置以广播 preview_sync 到在线设备
//
// eventType 取值：upload（文件上传/覆写）、delete（文件删除）、delete_session（会话删除）
var OnPreviewChanged func(sessionID string, eventType string)

// NewPreviewLogic 创建 PreviewLogic 实例
//
// 通过上下文获取 db，构造 PreviewSessionRepo 与 PreviewFileRepo 注入到 previewRepo 聚合结构。
func NewPreviewLogic(ctx context.Context) *PreviewLogic {
	db := xCtxUtil.MustGetDB(ctx)

	return &PreviewLogic{
		logic: logic{
			log: xLog.WithName(xLog.NamedLOGC, "PreviewLogic"),
		},
		repo: previewRepo{
			session: repository.NewPreviewSessionRepo(db),
			file:    repository.NewPreviewFileRepo(db),
			page:    repository.NewPageRepo(db),
			info:    repository.NewInfoRepo(db),
		},
	}
}

// CreateSession 创建预览会话（活动会话，1:N 多会话）
//
// 生成雪花 ID 与访问哈希后持久化，title 为空时回退为「未命名预览」。
func (l *PreviewLogic) CreateSession(ctx context.Context, projectID xSnowflake.SnowflakeID, title string) (*apiPreview.PreviewSessionResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("CreateSession - 创建预览会话 [projectID=%d, title=%s]", projectID.Int64(), title))

	// 生成雪花 ID 与访问哈希
	id := xSnowflake.GenerateID(bConst.GenePreviewSession)
	if title == "" {
		title = "未命名预览"
	}

	expiresAt := time.Now().Add(l.sessionTTL(ctx))
	session := &entity.PreviewSession{
		BaseEntity: xModels.BaseEntity{ID: id},
		ProjectID:  projectID,
		Title:      title,
		Hash:       generateSessionHash(id),
		Status:     bConst.PreviewSessionStatusActive,
		ExpiresAt:  &expiresAt,
	}

	if xErr := l.repo.session.Create(ctx, session); xErr != nil {
		return nil, xErr
	}

	return l.toPreviewSessionResponse(ctx, session), nil
}

// ListSessions 分页获取预览会话列表（projectID 为零值时不过滤），并批量填充各会话文件数。
func (l *PreviewLogic) ListSessions(ctx context.Context, projectID, workspaceID xSnowflake.SnowflakeID, page, size int) (*apiPreview.PreviewSessionListResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("ListSessions - 分页获取预览会话列表 [projectID=%d, workspace=%d, page=%d, size=%d]", projectID.Int64(), workspaceID.Int64(), page, size))

	sessions, total, xErr := l.repo.session.List(ctx, projectID, workspaceID, page, size)
	if xErr != nil {
		return nil, xErr
	}

	// 批量统计文件数，避免 N+1 查询
	sessionIDs := make([]xSnowflake.SnowflakeID, 0, len(sessions))
	for _, s := range sessions {
		sessionIDs = append(sessionIDs, s.ID)
	}
	fileCounts, xErr := l.repo.file.CountBySessions(ctx, sessionIDs)
	if xErr != nil {
		return nil, xErr
	}

	slugs := l.sourcePageSlugs(ctx, sessions)
	items := make([]apiPreview.PreviewSessionResponse, 0, len(sessions))
	for _, s := range sessions {
		resp := toPreviewSessionResponse(s, slugs[sourcePageIDKey(s)])
		resp.FileCount = fileCounts[s.ID.Int64()]
		items = append(items, *resp)
	}

	return &apiPreview.PreviewSessionListResponse{
		Items: items,
		Total: total,
	}, nil
}

// GetSessionByID 根据会话 ID 获取预览会话（填充文件数）。
func (l *PreviewLogic) GetSessionByID(ctx context.Context, sessionID xSnowflake.SnowflakeID) (*apiPreview.PreviewSessionResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("GetSessionByID - 获取预览会话 [sessionID=%d]", sessionID.Int64()))

	session, xErr := l.repo.session.GetByID(ctx, sessionID)
	if xErr != nil {
		return nil, xErr
	}

	// 填充文件数，保证 MCP 上传/列表工具返回的 session.file_count 反映真实数量
	counts, xErr := l.repo.file.CountBySessions(ctx, []xSnowflake.SnowflakeID{sessionID})
	if xErr != nil {
		return nil, xErr
	}

	resp := l.toPreviewSessionResponse(ctx, session)
	resp.FileCount = counts[sessionID.Int64()]
	return resp, nil
}

// BuildSessionURL 构造面向用户的预览会话访问地址。
//
// filename 为空时返回会话首页；非空时返回指定文件的深链。
func (l *PreviewLogic) BuildSessionURL(ctx context.Context, hash, filename string) string {
	return buildPreviewURL(resolveRuntimeDomain(ctx, l.repo.info), hash, filename)
}

// UploadFile 上传或覆写预览文件（扁平单层，同 Session 同文件名覆盖）
//
// 校验文件名合法性（禁止路径分隔符与目录穿越）与文件大小上限，按扩展名推断 MIME 类型。
func (l *PreviewLogic) UploadFile(ctx context.Context, sessionID xSnowflake.SnowflakeID, filename, content string) (*apiPreview.PreviewFileResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("UploadFile - 上传预览文件 [sessionID=%d, filename=%s]", sessionID.Int64(), filename))

	// 校验文件名（扁平单层）
	if err := validateFilename(filename); err != nil {
		return nil, xError.NewError(ctx, xError.ParameterError, xError.ErrMessage(err.Error()), false, nil)
	}

	// 校验文件大小（MySQL TEXT 列上限低于通用上限，按驱动收窄有效上限）
	maxBytes := l.repo.file.MaxContentBytes()
	if len(content) > maxBytes {
		return nil, xError.NewError(ctx, xError.ParameterError, xError.ErrMessage(fmt.Sprintf("文件大小超出上限(%dKB)", maxBytes/1024)), false, nil)
	}

	// 校验 LPW 语法合规性（上传失败不得覆盖已有文件）
	if err := validateLpwContent(filename, content); err != nil {
		return nil, xError.NewError(ctx, xError.ParameterError, xError.ErrMessage(err.Error()), false, nil)
	}

	session, xErr := l.repo.session.GetByID(ctx, sessionID)
	if xErr != nil {
		return nil, xErr
	}
	if xErr := l.rejectIfUnusable(ctx, session); xErr != nil {
		return nil, xErr
	}

	// 生成雪花 ID 并构建实体
	id := xSnowflake.GenerateID(bConst.GenePreviewFile)
	file := &entity.PreviewFile{
		BaseEntity: xModels.BaseEntity{ID: id},
		SessionID:  sessionID,
		Filename:   filename,
		MimeType:   inferMimeType(filename),
		Content:    content,
		Size:       len(content),
	}

	// 创建或覆写
	result, xErr := l.repo.file.CreateOrUpdate(ctx, file)
	if xErr != nil {
		return nil, xErr
	}

	// 触摸会话更新时间，使会话 updated_at 与内容变更保持一致（失败降级为警告，文件已落库不应标失败）
	if xErr := l.repo.session.TouchUpdatedAt(ctx, sessionID); xErr != nil {
		l.log.Warn(ctx, xErr.Error())
	}

	// 广播预览同步（上传/覆写均为内容变更）
	if OnPreviewChanged != nil {
		OnPreviewChanged(sessionID.String(), "upload")
	}

	return toPreviewFileResponse(result), nil
}

// ListFiles 获取指定会话的全部预览文件列表（按文件名升序）
func (l *PreviewLogic) ListFiles(ctx context.Context, sessionID xSnowflake.SnowflakeID) ([]apiPreview.PreviewFileResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("ListFiles - 获取预览文件列表 [sessionID=%d]", sessionID.Int64()))

	files, xErr := l.repo.file.ListBySession(ctx, sessionID)
	if xErr != nil {
		return nil, xErr
	}

	items := make([]apiPreview.PreviewFileResponse, 0, len(files))
	for _, f := range files {
		items = append(items, *toPreviewFileResponse(f))
	}

	return items, nil
}

// EditFile 行级编辑既有预览文件（insert/replace/delete，1 起始闭区间，无需全量重传）
//
// 编辑基于服务端存储内容按行变换后整体写回：换行统一归一为 LF，原文件结尾换行符原样保留；
// startLine 传 0 时仅 insert 允许（表示追加到文件末尾）。文件必须已存在，创建新文件请走 UploadFile。
// 返回编辑后的文件元数据、总行数与编辑落点区域（变更主体 ±3 行上下文、上限 40 行，带行号）。
func (l *PreviewLogic) EditFile(ctx context.Context, sessionID xSnowflake.SnowflakeID, filename, operation string, startLine, endLine int, content string) (*apiPreview.PreviewFileEditResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("EditFile - 行级编辑预览文件 [sessionID=%d, filename=%s, operation=%s, start=%d, end=%d]", sessionID.Int64(), filename, operation, startLine, endLine))

	// 校验文件名（扁平单层）
	if err := validateFilename(filename); err != nil {
		return nil, xError.NewError(ctx, xError.ParameterError, xError.ErrMessage(err.Error()), false, nil)
	}
	// delete 操作不接受 content（防误传导致意图歧义）
	if operation == bConst.PreviewEditOperationDelete && content != "" {
		return nil, xError.NewError(ctx, xError.ParameterError, xError.ErrMessage("delete 操作不接受 content 参数"), false, nil)
	}

	session, xErr := l.repo.session.GetByID(ctx, sessionID)
	if xErr != nil {
		return nil, xErr
	}
	if xErr := l.rejectIfUnusable(ctx, session); xErr != nil {
		return nil, xErr
	}

	file, xErr := l.repo.file.GetBySessionAndFilename(ctx, sessionID, filename)
	if xErr != nil {
		return nil, xErr
	}

	// 行变换（CRLF 归一 + 边界校验 + 应用操作）
	parsed := splitFileLines(file.Content)
	newLines, changedStart, changedEnd, err := applyLineEdit(parsed.lines, previewLineEdit{
		operation: operation,
		startLine: startLine,
		endLine:   endLine,
		content:   content,
	})
	if err != nil {
		return nil, xError.NewError(ctx, xError.ParameterError, xError.ErrMessage(err.Error()), false, nil)
	}
	newContent := joinFileLines(newLines, parsed.trailingNewline)

	// 校验编辑后大小上限（MySQL TEXT 列上限低于通用上限，按驱动收窄有效上限）
	maxBytes := l.repo.file.MaxContentBytes()
	if len(newContent) > maxBytes {
		return nil, xError.NewError(ctx, xError.ParameterError, xError.ErrMessage(fmt.Sprintf("编辑后文件大小超出上限(%dKB)", maxBytes/1024)), false, nil)
	}

	// 原位写回（保留原文件 ID 与创建时间）
	file.Content = newContent
	file.Size = len(newContent)
	result, xErr := l.repo.file.CreateOrUpdate(ctx, file)
	if xErr != nil {
		return nil, xErr
	}

	// 触摸会话更新时间，使会话 updated_at 与内容变更保持一致（失败降级为警告，文件已落库不应标失败）
	if xErr := l.repo.session.TouchUpdatedAt(ctx, sessionID); xErr != nil {
		l.log.Warn(ctx, xErr.Error())
	}

	// 广播预览同步（编辑属内容变更，复用 upload 事件类型，前端按全量详情刷新）
	if OnPreviewChanged != nil {
		OnPreviewChanged(sessionID.String(), "upload")
	}

	regionStart, regionEnd := editRegionWindow(len(newLines), changedStart, changedEnd)
	return &apiPreview.PreviewFileEditResponse{
		PreviewFileResponse: *toPreviewFileResponse(result),
		TotalLines:          len(newLines),
		RegionStart:         regionStart,
		RegionEnd:           regionEnd,
		Region:              numberedRegion(newLines, regionStart, regionEnd),
	}, nil
}

// GetSessionByHash 根据访问哈希获取预览会话（公开访问鉴权用）
func (l *PreviewLogic) GetSessionByHash(ctx context.Context, hash string) (*apiPreview.PreviewSessionResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("GetSessionByHash - 根据哈希获取预览会话 [%s]", hash))

	session, xErr := l.repo.session.GetByHash(ctx, hash)
	if xErr != nil {
		return nil, xErr
	}
	if xErr := l.rejectIfUnusable(ctx, session); xErr != nil {
		return nil, xErr
	}

	return l.toPreviewSessionResponse(ctx, session), nil
}

// GetSessionDetailByID 根据会话 ID 获取预览会话详情（含文件列表）
//
// 组合 GetSessionByID 与 ListFiles 逻辑，并填充 FileCount，
// 供 WebSocket 连接快照与内容变更广播一次性返回完整会话状态。
func (l *PreviewLogic) GetSessionDetailByID(ctx context.Context, sessionID xSnowflake.SnowflakeID) (*apiPreview.PreviewSessionDetailResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("GetSessionDetailByID - 获取预览会话详情 [sessionID=%d]", sessionID.Int64()))

	session, xErr := l.GetSessionByID(ctx, sessionID)
	if xErr != nil {
		return nil, xErr
	}

	files, xErr := l.ListFiles(ctx, sessionID)
	if xErr != nil {
		return nil, xErr
	}

	session.FileCount = int64(len(files))

	return &apiPreview.PreviewSessionDetailResponse{
		Session: *session,
		Files:   files,
	}, nil
}

// GetFileContent 根据访问哈希与文件名获取预览文件完整内容（serve 接口专用）
func (l *PreviewLogic) GetFileContent(ctx context.Context, hash, filename string) (*apiPreview.PreviewFileContentResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("GetFileContent - 获取预览文件内容 [hash=%s, filename=%s]", hash, filename))

	session, xErr := l.repo.session.GetByHash(ctx, hash)
	if xErr != nil {
		return nil, xErr
	}
	if xErr := l.rejectIfUnusable(ctx, session); xErr != nil {
		return nil, xErr
	}

	file, xErr := l.repo.file.GetBySessionAndFilename(ctx, session.ID, filename)
	if xErr != nil {
		return nil, xErr
	}

	return &apiPreview.PreviewFileContentResponse{
		Filename: file.Filename,
		MimeType: file.MimeType,
		Content:  file.Content,
	}, nil
}

// GetFileLinesBySession 根据会话 ID 与文件名读取预览文件内容，支持行区间（MCP 行级读取入口）
//
// startLine/endLine 均为 0 时返回原始全量内容（与既有行为兼容）；指定区间时内容按「行号| 文本」
// 格式返回（1 起始闭区间），endLine 传 0 或超出总行数时钳制到末行，便于「从第 N 行读到末尾」。
func (l *PreviewLogic) GetFileLinesBySession(ctx context.Context, sessionID xSnowflake.SnowflakeID, filename string, startLine, endLine int) (*apiPreview.PreviewFileLinesResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("GetFileLinesBySession - 行级读取预览文件 [sessionID=%d, filename=%s, start=%d, end=%d]", sessionID.Int64(), filename, startLine, endLine))

	session, xErr := l.repo.session.GetByID(ctx, sessionID)
	if xErr != nil {
		return nil, xErr
	}
	if xErr := l.rejectIfUnusable(ctx, session); xErr != nil {
		return nil, xErr
	}

	file, xErr := l.repo.file.GetBySessionAndFilename(ctx, sessionID, filename)
	if xErr != nil {
		return nil, xErr
	}

	parsed := splitFileLines(file.Content)
	resp := &apiPreview.PreviewFileLinesResponse{
		Filename:   file.Filename,
		MimeType:   file.MimeType,
		Size:       file.Size,
		TotalLines: len(parsed.lines),
	}

	// 全量模式：原始内容原样返回（兼容既有消费方对源码做字符串匹配）
	if startLine == 0 && endLine == 0 {
		if len(parsed.lines) > 0 {
			resp.StartLine, resp.EndLine = 1, len(parsed.lines)
		}
		resp.Content = file.Content
		return resp, nil
	}

	// 区间模式：带行号返回（空文件无行可读，明确报错）
	if len(parsed.lines) == 0 {
		return nil, xError.NewError(ctx, xError.ParameterError, xError.ErrMessage("文件为空，没有可读取的行"), false, nil)
	}
	if startLine < 1 || startLine > len(parsed.lines) {
		return nil, xError.NewError(ctx, xError.ParameterError, xError.ErrMessage(fmt.Sprintf("start_line 无效：需要 1 ≤ start_line ≤ %d（当前 %d）", len(parsed.lines), startLine)), false, nil)
	}
	effectiveEnd := endLine
	if effectiveEnd == 0 || effectiveEnd > len(parsed.lines) {
		effectiveEnd = len(parsed.lines)
	}
	if effectiveEnd < startLine {
		return nil, xError.NewError(ctx, xError.ParameterError, xError.ErrMessage(fmt.Sprintf("end_line 不能小于 start_line（start=%d, end=%d）", startLine, endLine)), false, nil)
	}

	resp.StartLine = startLine
	resp.EndLine = effectiveEnd
	resp.Content = formatNumberedLines(startLine, parsed.lines[startLine-1:effectiveEnd])
	return resp, nil
}

// GetFileByID 根据文件 ID 获取预览文件详情（含关联会话哈希）
//
// 供 Q&A supplement preview 类型渲染时，由 file_id 解析出 (session_hash, filename) 以构造 serve 地址。
func (l *PreviewLogic) GetFileByID(ctx context.Context, fileID xSnowflake.SnowflakeID) (*apiPreview.PreviewFileDetailResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("GetFileByID - 根据 ID 获取预览文件详情 [%d]", fileID.Int64()))

	file, xErr := l.repo.file.GetByID(ctx, fileID)
	if xErr != nil {
		return nil, xErr
	}

	session, xErr := l.repo.session.GetByID(ctx, file.SessionID)
	if xErr != nil {
		return nil, xErr
	}
	if xErr := l.rejectIfUnusable(ctx, session); xErr != nil {
		return nil, xErr
	}

	return &apiPreview.PreviewFileDetailResponse{
		PreviewFileResponse: *toPreviewFileResponse(file),
		SessionHash:         session.Hash,
	}, nil
}

// DeleteSession 删除预览会话（事务级联删除其下全部预览文件）
func (l *PreviewLogic) DeleteSession(ctx context.Context, sessionID xSnowflake.SnowflakeID) *xError.Error {
	l.log.Info(ctx, fmt.Sprintf("DeleteSession - 删除预览会话 [%d]", sessionID.Int64()))

	// 事务内级联删除文件与会话本体，避免孤儿文件
	if xErr := l.repo.session.DeleteCascade(ctx, sessionID); xErr != nil {
		return xErr
	}

	// 广播预览同步（会话删除后前端跳转关闭）
	if OnPreviewChanged != nil {
		OnPreviewChanged(sessionID.String(), "delete_session")
	}

	return nil
}

// DeleteFile 删除单个预览文件
func (l *PreviewLogic) DeleteFile(ctx context.Context, fileID xSnowflake.SnowflakeID) *xError.Error {
	l.log.Info(ctx, fmt.Sprintf("DeleteFile - 删除预览文件 [%d]", fileID.Int64()))

	// 先查询文件获取关联会话 ID（删除成功后需广播同步）
	file, xErr := l.repo.file.GetByID(ctx, fileID)
	if xErr != nil {
		return xErr
	}

	if xErr := l.repo.file.Delete(ctx, fileID); xErr != nil {
		return xErr
	}

	// 触摸会话更新时间，使会话 updated_at 与内容变更保持一致（失败降级为警告，文件已删除不应标失败）
	if xErr := l.repo.session.TouchUpdatedAt(ctx, file.SessionID); xErr != nil {
		l.log.Warn(ctx, xErr.Error())
	}

	// 广播预览同步（文件删除为内容变更）
	if OnPreviewChanged != nil {
		OnPreviewChanged(file.SessionID.String(), "delete")
	}

	return nil
}

// DeleteFileByName 根据会话 ID 与文件名删除单个预览文件（MCP 按名删除入口）
//
// 与 DeleteFile(fileID) 语义等价：先解析实体再做会话可用性校验，返回被删文件快照供调用方回显。
func (l *PreviewLogic) DeleteFileByName(ctx context.Context, sessionID xSnowflake.SnowflakeID, filename string) (*apiPreview.PreviewFileResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("DeleteFileByName - 按名删除预览文件 [sessionID=%d, filename=%s]", sessionID.Int64(), filename))

	// 校验文件名（扁平单层）
	if err := validateFilename(filename); err != nil {
		return nil, xError.NewError(ctx, xError.ParameterError, xError.ErrMessage(err.Error()), false, nil)
	}

	session, xErr := l.repo.session.GetByID(ctx, sessionID)
	if xErr != nil {
		return nil, xErr
	}
	if xErr := l.rejectIfUnusable(ctx, session); xErr != nil {
		return nil, xErr
	}

	file, xErr := l.repo.file.GetBySessionAndFilename(ctx, sessionID, filename)
	if xErr != nil {
		return nil, xErr
	}
	snapshot := toPreviewFileResponse(file)

	if xErr := l.repo.file.Delete(ctx, file.ID); xErr != nil {
		return nil, xErr
	}

	// 触摸会话更新时间，使会话 updated_at 与内容变更保持一致（失败降级为警告，文件已删除不应标失败）
	if xErr := l.repo.session.TouchUpdatedAt(ctx, sessionID); xErr != nil {
		l.log.Warn(ctx, xErr.Error())
	}

	// 广播预览同步（文件删除为内容变更）
	if OnPreviewChanged != nil {
		OnPreviewChanged(sessionID.String(), "delete")
	}

	return snapshot, nil
}

// ExpireStaleSessions 将已到期仍为 active 的预览会话改为 deleted（保留文件）
func (l *PreviewLogic) ExpireStaleSessions(ctx context.Context) {
	l.log.Info(ctx, "ExpireStaleSessions - 扫描已到期的预览会话")
	ids, xErr := l.repo.session.ListActiveExpiredIDs(ctx, time.Now())
	if xErr != nil {
		l.log.Warn(ctx, xErr.Error())
		return
	}
	for _, id := range ids {
		if uErr := l.repo.session.UpdateStatus(ctx, id, bConst.PreviewSessionStatusDeleted); uErr != nil {
			l.log.Warn(ctx, uErr.Error())
			continue
		}
		if OnPreviewChanged != nil {
			OnPreviewChanged(id.String(), "delete_session")
		}
	}
}

// ─── Helpers ────────────────────────────────────────────────────────────

const defaultPreviewSessionTTL = 7 * 24 * time.Hour

func previewSessionTTL(ctx context.Context, infoRepo *repository.InfoRepo) time.Duration {
	if infoRepo == nil {
		return defaultPreviewSessionTTL
	}
	ttlStr, xErr := infoRepo.GetByKey(ctx, bConst.InfoKeyPreviewSessionTTL)
	if xErr != nil {
		return defaultPreviewSessionTTL
	}
	sec, err := strconv.Atoi(ttlStr)
	if err != nil || sec <= 0 {
		return defaultPreviewSessionTTL
	}
	return time.Duration(sec) * time.Second
}

func (l *PreviewLogic) sessionTTL(ctx context.Context) time.Duration {
	return previewSessionTTL(ctx, l.repo.info)
}

func sessionDue(session *entity.PreviewSession, now time.Time) bool {
	return session.Status == bConst.PreviewSessionStatusActive &&
		session.ExpiresAt != nil &&
		now.After(*session.ExpiresAt)
}

func (l *PreviewLogic) expireIfDue(ctx context.Context, session *entity.PreviewSession) {
	if !sessionDue(session, time.Now()) {
		return
	}
	if xErr := l.repo.session.UpdateStatus(ctx, session.ID, bConst.PreviewSessionStatusDeleted); xErr != nil {
		l.log.Warn(ctx, xErr.Error())
		return
	}
	session.Status = bConst.PreviewSessionStatusDeleted
	if OnPreviewChanged != nil {
		OnPreviewChanged(session.ID.String(), "delete_session")
	}
}

func (l *PreviewLogic) rejectIfUnusable(ctx context.Context, session *entity.PreviewSession) *xError.Error {
	l.expireIfDue(ctx, session)
	if session.Status != bConst.PreviewSessionStatusActive {
		return xError.NewError(ctx, xError.NotFound, "预览会话不存在", false, nil)
	}
	return nil
}

// validateFilename 校验文件名是否合法（扁平单层）
//
// 规则：非空、禁止路径分隔符（/ 与 \）、禁止目录穿越（. 与 ..）、长度不超过 255。
func validateFilename(filename string) error {
	if strings.TrimSpace(filename) == "" {
		return errors.New("文件名不能为空")
	}
	if strings.ContainsAny(filename, "/\\") {
		return errors.New("文件名禁止包含路径分隔符（仅支持扁平单层）")
	}
	if filename == "." || filename == ".." || strings.Contains(filename, "..") {
		return errors.New("文件名禁止目录穿越")
	}
	if len(filename) > 255 {
		return errors.New("文件名过长（上限 255 字符）")
	}
	return nil
}

// inferMimeType 根据文件扩展名推断 MIME 类型，未知扩展名回退 text/plain
func inferMimeType(filename string) string {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".html", ".htm":
		return bConst.PreviewMimeHTML
	case ".css":
		return bConst.PreviewMimeCSS
	case ".js", ".mjs", ".cjs":
		return bConst.PreviewMimeJS
	case ".json":
		return bConst.PreviewMimeJSON
	case ".lpw":
		return bConst.PreviewMimeLPW
	case ".md", ".markdown":
		return bConst.PreviewMimeMarkdown
	case ".ts", ".tsx", ".mts", ".cts":
		return bConst.PreviewMimePlain
	case ".svg":
		return bConst.PreviewMimeSVG
	default:
		return bConst.PreviewMimePlain
	}
}

// validateLpwContent 校验 .lpw 文件内容是否为合法 JSON
func validateLpwContent(filename, content string) error {
	if strings.ToLower(filepath.Ext(filename)) == ".lpw" {
		if !json.Valid([]byte(content)) {
			return errors.New("LPW 文件必须是合法 JSON")
		}
	}
	return nil
}

// GetFileBySessionAndFilename 根据会话 ID 与文件名获取预览文件实体
func (l *PreviewLogic) GetFileBySessionAndFilename(ctx context.Context, sessionID xSnowflake.SnowflakeID, filename string) (*entity.PreviewFile, *xError.Error) {
	return l.repo.file.GetBySessionAndFilename(ctx, sessionID, filename)
}

func (l *PreviewLogic) toPreviewSessionResponse(ctx context.Context, session *entity.PreviewSession) *apiPreview.PreviewSessionResponse {
	return toPreviewSessionResponse(session, l.sourcePageSlug(ctx, session))
}

func sourcePageIDKey(session *entity.PreviewSession) int64 {
	if session == nil || session.SourcePageID == nil || session.SourcePageID.IsZero() {
		return 0
	}
	return session.SourcePageID.Int64()
}

func (l *PreviewLogic) sourcePageSlug(ctx context.Context, session *entity.PreviewSession) string {
	id := sourcePageIDKey(session)
	if id == 0 {
		return ""
	}
	page, xErr := l.repo.page.GetByID(ctx, *session.SourcePageID)
	if xErr != nil {
		l.log.Warn(ctx, xErr.Error())
		return ""
	}
	return page.Slug
}

func (l *PreviewLogic) sourcePageSlugs(ctx context.Context, sessions []*entity.PreviewSession) map[int64]string {
	ids := make([]xSnowflake.SnowflakeID, 0, len(sessions))
	seen := make(map[int64]struct{}, len(sessions))
	for _, session := range sessions {
		id := sourcePageIDKey(session)
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, *session.SourcePageID)
	}
	pages, xErr := l.repo.page.GetByIDs(ctx, ids)
	if xErr != nil {
		l.log.Warn(ctx, xErr.Error())
		return map[int64]string{}
	}
	slugs := make(map[int64]string, len(pages))
	for _, page := range pages {
		slugs[page.ID.Int64()] = page.Slug
	}
	return slugs
}

// toPreviewSessionResponse 将预览会话实体映射为响应 DTO
func toPreviewSessionResponse(session *entity.PreviewSession, sourcePageSlug string) *apiPreview.PreviewSessionResponse {
	expiresAt := ""
	if session.ExpiresAt != nil {
		expiresAt = session.ExpiresAt.Format(time.RFC3339)
	}
	return &apiPreview.PreviewSessionResponse{
		ID:              session.ID,
		ProjectID:       session.ProjectID,
		Title:           session.Title,
		Hash:            session.Hash,
		Status:          session.Status,
		ExpiresAt:       expiresAt,
		SourcePageID:    session.SourcePageID,
		SourcePageSlug:  sourcePageSlug,
		SourceVersionID: session.SourceVersionID,
		CreatedAt:       session.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       session.UpdatedAt.Format(time.RFC3339),
	}
}

// toPreviewFileResponse 将预览文件实体映射为响应 DTO（不含 Content）
func toPreviewFileResponse(file *entity.PreviewFile) *apiPreview.PreviewFileResponse {
	return &apiPreview.PreviewFileResponse{
		ID:        file.ID,
		SessionID: file.SessionID,
		Filename:  file.Filename,
		MimeType:  file.MimeType,
		Size:      file.Size,
		CreatedAt: file.CreatedAt.Format(time.RFC3339),
		UpdatedAt: file.UpdatedAt.Format(time.RFC3339),
	}
}
