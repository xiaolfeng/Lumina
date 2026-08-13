package logic

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
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
}

// PreviewLogic Preview 业务逻辑层，负责预览会话管理、文件上传与渲染内容编排
type PreviewLogic struct {
	logic
	repo previewRepo
}

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
		},
	}
}

// CreateSession 创建预览会话（活动工作区，1:N 多工作区）
//
// 生成雪花 ID 与 16 位访问哈希后持久化，title 为空时回退为「未命名预览」。
func (l *PreviewLogic) CreateSession(ctx context.Context, projectID xSnowflake.SnowflakeID, title string) (*apiPreview.PreviewSessionResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("CreateSession - 创建预览会话 [projectID=%d, title=%s]", projectID.Int64(), title))

	// 生成雪花 ID 与访问哈希
	id := xSnowflake.GenerateID(bConst.GenePreviewSession)
	if title == "" {
		title = "未命名预览"
	}

	session := &entity.PreviewSession{
		BaseEntity: xModels.BaseEntity{ID: id},
		ProjectID:  projectID,
		Title:      title,
		Hash:       generateSessionHash(id),
		Status:     bConst.PreviewSessionStatusActive,
	}

	if xErr := l.repo.session.Create(ctx, session); xErr != nil {
		return nil, xErr
	}

	return toPreviewSessionResponse(session), nil
}

// ListSessions 分页获取预览会话列表（projectID 为零值时不过滤）
func (l *PreviewLogic) ListSessions(ctx context.Context, projectID xSnowflake.SnowflakeID, page, size int) (*apiPreview.PreviewSessionListResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("ListSessions - 分页获取预览会话列表 [projectID=%d, page=%d, size=%d]", projectID.Int64(), page, size))

	sessions, total, xErr := l.repo.session.List(ctx, projectID, page, size)
	if xErr != nil {
		return nil, xErr
	}

	items := make([]apiPreview.PreviewSessionResponse, 0, len(sessions))
	for _, s := range sessions {
		items = append(items, *toPreviewSessionResponse(s))
	}

	return &apiPreview.PreviewSessionListResponse{
		Items: items,
		Total: total,
	}, nil
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

	// 校验文件大小
	if len(content) > bConst.PreviewFileMaxSize {
		return nil, xError.NewError(ctx, xError.ParameterError, "文件大小超出上限(256KB)", false, nil)
	}

	// 校验会话存在
	if _, xErr := l.repo.session.GetByID(ctx, sessionID); xErr != nil {
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

// GetSessionByHash 根据访问哈希获取预览会话（公开访问鉴权用）
func (l *PreviewLogic) GetSessionByHash(ctx context.Context, hash string) (*apiPreview.PreviewSessionResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("GetSessionByHash - 根据哈希获取预览会话 [%s]", hash))

	session, xErr := l.repo.session.GetByHash(ctx, hash)
	if xErr != nil {
		return nil, xErr
	}

	return toPreviewSessionResponse(session), nil
}

// GetFileContent 根据访问哈希与文件名获取预览文件完整内容（serve 接口专用）
func (l *PreviewLogic) GetFileContent(ctx context.Context, hash, filename string) (*apiPreview.PreviewFileContentResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("GetFileContent - 获取预览文件内容 [hash=%s, filename=%s]", hash, filename))

	session, xErr := l.repo.session.GetByHash(ctx, hash)
	if xErr != nil {
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

// GetFileContentBySession 根据会话 ID 与文件名获取预览文件完整内容（MCP 提取代码用）
func (l *PreviewLogic) GetFileContentBySession(ctx context.Context, sessionID xSnowflake.SnowflakeID, filename string) (*apiPreview.PreviewFileContentResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("GetFileContentBySession - 获取预览文件内容 [sessionID=%d, filename=%s]", sessionID.Int64(), filename))

	// 校验会话存在
	if _, xErr := l.repo.session.GetByID(ctx, sessionID); xErr != nil {
		return nil, xErr
	}

	file, xErr := l.repo.file.GetBySessionAndFilename(ctx, sessionID, filename)
	if xErr != nil {
		return nil, xErr
	}

	return &apiPreview.PreviewFileContentResponse{
		Filename: file.Filename,
		MimeType: file.MimeType,
		Content:  file.Content,
	}, nil
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

	return &apiPreview.PreviewFileDetailResponse{
		PreviewFileResponse: *toPreviewFileResponse(file),
		SessionHash:         session.Hash,
	}, nil
}

// DeleteSession 删除预览会话（级联删除其下全部预览文件）
func (l *PreviewLogic) DeleteSession(ctx context.Context, sessionID xSnowflake.SnowflakeID) *xError.Error {
	l.log.Info(ctx, fmt.Sprintf("DeleteSession - 删除预览会话 [%d]", sessionID.Int64()))

	// 校验会话存在
	if _, xErr := l.repo.session.GetByID(ctx, sessionID); xErr != nil {
		return xErr
	}

	// 级联删除文件
	if xErr := l.repo.file.DeleteBySession(ctx, sessionID); xErr != nil {
		return xErr
	}

	// 删除会话
	return l.repo.session.Delete(ctx, sessionID)
}

// DeleteFile 删除单个预览文件
func (l *PreviewLogic) DeleteFile(ctx context.Context, fileID xSnowflake.SnowflakeID) *xError.Error {
	l.log.Info(ctx, fmt.Sprintf("DeleteFile - 删除预览文件 [%d]", fileID.Int64()))
	return l.repo.file.Delete(ctx, fileID)
}

// ─── Helpers ────────────────────────────────────────────────────────────

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
	case ".js", ".mjs":
		return bConst.PreviewMimeJS
	case ".json":
		return bConst.PreviewMimeJSON
	case ".svg":
		return bConst.PreviewMimeSVG
	default:
		return bConst.PreviewMimePlain
	}
}

// toPreviewSessionResponse 将预览会话实体映射为响应 DTO
func toPreviewSessionResponse(session *entity.PreviewSession) *apiPreview.PreviewSessionResponse {
	return &apiPreview.PreviewSessionResponse{
		ID:        session.ID,
		ProjectID: session.ProjectID,
		Title:     session.Title,
		Hash:      session.Hash,
		Status:    session.Status,
		CreatedAt: session.CreatedAt.Format(time.RFC3339),
		UpdatedAt: session.UpdatedAt.Format(time.RFC3339),
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
