package logic

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	xCtxUtil "github.com/bamboo-services/bamboo-base-go/major/utility/context"
	apiPages "github.com/xiaolfeng/Lumina/api/pages"
	apiPreview "github.com/xiaolfeng/Lumina/api/preview"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"github.com/xiaolfeng/Lumina/internal/repository"
	"github.com/xiaolfeng/Lumina/internal/service"
)

type pagesRepo struct {
	page           *repository.PageRepo
	version        *repository.PageVersionRepo
	file           *repository.PageFileRepo
	previewSession *repository.PreviewSessionRepo
	previewFile    *repository.PreviewFileRepo
	project        *repository.ProjectRepo
	info           *repository.InfoRepo
}

// PagesLogic Pages 持久交付编排：晋升、Fork、版本指针、密码门与文件直出
type PagesLogic struct {
	logic
	repo      pagesRepo
	authToken *service.PageAuthTokenService
	authLimit *pagesAuthLimiter
}

// NewPagesLogic 创建 PagesLogic 实例
func NewPagesLogic(ctx context.Context) *PagesLogic {
	db := xCtxUtil.MustGetDB(ctx)
	rdb := xCtxUtil.MustGetRDB(ctx)
	return &PagesLogic{
		logic: logic{log: xLog.WithName(xLog.NamedLOGC, "PagesLogic")},
		repo: pagesRepo{
			page:           repository.NewPageRepo(db),
			version:        repository.NewPageVersionRepo(db),
			file:           repository.NewPageFileRepo(db),
			previewSession: repository.NewPreviewSessionRepo(db),
			previewFile:    repository.NewPreviewFileRepo(db),
			project:        repository.NewProjectRepo(db, rdb),
			info:           repository.NewInfoRepo(db),
		},
		authToken: service.NewPageAuthTokenService(),
		authLimit: pagesAuthLimit,
	}
}

// AuthToken 暴露密码门签名服务给中间件 / Handler
func (l *PagesLogic) AuthToken() *service.PageAuthTokenService {
	return l.authToken
}

// List 分页列出页面
func (l *PagesLogic) List(ctx context.Context, projectID, workspaceID xSnowflake.SnowflakeID, page, size int) (*apiPages.PageListResponse, *xError.Error) {
	pages, total, xErr := l.repo.page.List(ctx, projectID, workspaceID, page, size)
	if xErr != nil {
		return nil, xErr
	}
	items := make([]apiPages.PageResponse, 0, len(pages))
	for _, item := range pages {
		resp, mapErr := l.toPageResponse(ctx, item)
		if mapErr != nil {
			return nil, mapErr
		}
		items = append(items, *resp)
	}
	return &apiPages.PageListResponse{Items: items, Total: total}, nil
}

// GetByID 管理端详情
func (l *PagesLogic) GetByID(ctx context.Context, id xSnowflake.SnowflakeID) (*apiPages.PageResponse, *xError.Error) {
	page, xErr := l.repo.page.GetByID(ctx, id)
	if xErr != nil {
		return nil, xErr
	}
	return l.toPageResponse(ctx, page)
}

// GetByProjectNameAndSlug 按项目名与 slug 解析页面
func (l *PagesLogic) GetByProjectNameAndSlug(ctx context.Context, projectName, slug string) (*entity.Page, *entity.Project, *xError.Error) {
	project, xErr := l.repo.project.GetByName(ctx, projectName)
	if xErr != nil {
		return nil, nil, xErr
	}
	page, xErr := l.repo.page.GetByProjectAndSlug(ctx, project.ID, slug)
	if xErr != nil {
		return nil, nil, xErr
	}
	return page, project, nil
}

// GetPasswordHash 供 PagesAuth 中间件查询；空串表示公开
func (l *PagesLogic) GetPasswordHash(ctx context.Context, pageID int64) (string, error) {
	page, xErr := l.repo.page.GetByID(ctx, xSnowflake.SnowflakeID(pageID))
	if xErr != nil {
		return "", xErr
	}
	if page.AccessMode != bConst.PageAccessModePassword {
		return "", nil
	}
	return page.PasswordHash, nil
}

// LookupAuth 按项目名与 slug 解析页面 ID 与密码哈希（空串=公开）
//
// 与 PublicMeta/ServeFile 对齐：非 published 页面对外整体 404，归档页不可探测。
func (l *PagesLogic) LookupAuth(ctx context.Context, projectName, slug string) (int64, string, error) {
	page, _, xErr := l.GetByProjectNameAndSlug(ctx, projectName, slug)
	if xErr != nil {
		return 0, "", xErr
	}
	if page.Status != bConst.PageStatusPublished {
		return 0, "", xError.NewError(ctx, xError.NotFound, "页面不存在", false, nil)
	}
	if page.AccessMode != bConst.PageAccessModePassword {
		return page.ID.Int64(), "", nil
	}
	return page.ID.Int64(), page.PasswordHash, nil
}

// ListVersions 列出页面全部版本
func (l *PagesLogic) ListVersions(ctx context.Context, pageID xSnowflake.SnowflakeID) (*apiPages.PageVersionListResponse, *xError.Error) {
	page, xErr := l.repo.page.GetByID(ctx, pageID)
	if xErr != nil {
		return nil, xErr
	}
	versions, xErr := l.repo.version.ListByPage(ctx, pageID)
	if xErr != nil {
		return nil, xErr
	}
	items := make([]apiPages.PageVersionResponse, 0, len(versions))
	for _, version := range versions {
		items = append(items, *toPageVersionResponse(version, page.LatestVersionID))
	}
	return &apiPages.PageVersionListResponse{Items: items}, nil
}

// ListVersionsPublic 展示态版本列表（密码门校验，供公开端点调用）
//
// 与 PublicMeta 对齐：非 published 页面对外整体 404，归档页不可探测；
// 密码页需携带有效解锁 Cookie，否则 401。
func (l *PagesLogic) ListVersionsPublic(ctx context.Context, projectName, slug, cookieValue string) (*apiPages.PageVersionListResponse, *xError.Error) {
	page, _, xErr := l.GetByProjectNameAndSlug(ctx, projectName, slug)
	if xErr != nil {
		return nil, xErr
	}
	if page.Status != bConst.PageStatusPublished {
		return nil, xError.NewError(ctx, xError.NotFound, "页面不存在", false, nil)
	}
	required := page.AccessMode == bConst.PageAccessModePassword && page.PasswordHash != ""
	if required && !l.authToken.ValidateToken(cookieValue, page.ID.Int64()) {
		return nil, xError.NewError(ctx, xError.Unauthorized, "page authentication required", false, nil)
	}
	versions, xErr := l.repo.version.ListByPage(ctx, page.ID)
	if xErr != nil {
		return nil, xErr
	}
	items := make([]apiPages.PageVersionResponse, 0, len(versions))
	for _, version := range versions {
		items = append(items, *toPageVersionResponse(version, page.LatestVersionID))
	}
	return &apiPages.PageVersionListResponse{Items: items}, nil
}

// PublicMeta 展示态元信息（不含密码哈希）
func (l *PagesLogic) PublicMeta(ctx context.Context, projectName, slug, versionLabel string) (*apiPages.PagePublicMetaResponse, *xError.Error) {
	page, project, xErr := l.GetByProjectNameAndSlug(ctx, projectName, slug)
	if xErr != nil {
		return nil, xErr
	}
	if page.Status != bConst.PageStatusPublished {
		return nil, xError.NewError(ctx, xError.NotFound, "页面不存在", false, nil)
	}
	version, xErr := l.resolveVersion(ctx, page, versionLabel)
	if xErr != nil {
		return nil, xErr
	}
	files, xErr := l.repo.file.ListByVersion(ctx, version.ID)
	if xErr != nil {
		return nil, xErr
	}
	pageResp, xErr := l.toPageResponse(ctx, page)
	if xErr != nil {
		return nil, xErr
	}
	if pageResp.ProjectName == "" {
		pageResp.ProjectName = project.Name
	}
	fileItems := make([]apiPages.PageFileResponse, 0, len(files))
	for _, file := range files {
		fileItems = append(fileItems, *toPageFileResponse(file))
	}
	return &apiPages.PagePublicMetaResponse{
		Page:    *pageResp,
		Version: *toPageVersionResponse(version, page.LatestVersionID),
		Files:   fileItems,
	}, nil
}

// ServeFile 读取指定版本（或生效版本）的文件正文
func (l *PagesLogic) ServeFile(ctx context.Context, projectName, slug, filename, versionLabel string) (*apiPages.PageFileContentResponse, *xError.Error) {
	if err := validateFilename(filename); err != nil {
		return nil, xError.NewError(ctx, xError.ParameterError, xError.ErrMessage(err.Error()), false, nil)
	}
	page, _, xErr := l.GetByProjectNameAndSlug(ctx, projectName, slug)
	if xErr != nil {
		return nil, xErr
	}
	if page.Status != bConst.PageStatusPublished {
		return nil, xError.NewError(ctx, xError.NotFound, "页面不存在", false, nil)
	}
	version, xErr := l.resolveVersion(ctx, page, versionLabel)
	if xErr != nil {
		return nil, xErr
	}
	file, xErr := l.repo.file.GetByVersionAndFilename(ctx, version.ID, filename)
	if xErr != nil {
		return nil, xErr
	}
	return &apiPages.PageFileContentResponse{
		Filename: file.Filename,
		MimeType: file.MimeType,
		Content:  file.Content,
	}, nil
}

// EntryFilename 返回指定版本入口；versionLabel 空则走生效版本
func (l *PagesLogic) EntryFilename(ctx context.Context, projectName, slug, versionLabel string) (string, *xError.Error) {
	page, _, xErr := l.GetByProjectNameAndSlug(ctx, projectName, slug)
	if xErr != nil {
		return "", xErr
	}
	version, xErr := l.resolveVersion(ctx, page, versionLabel)
	if xErr != nil {
		return "", xErr
	}
	if version.EntryFilename != "" {
		return version.EntryFilename, nil
	}
	return bConst.PagesDefaultEntryFilename, nil
}

// Promote 将预览会话深拷贝为不可变 Pages 快照
func (l *PagesLogic) Promote(ctx context.Context, sessionID xSnowflake.SnowflakeID, req *apiPreview.PromoteSessionRequest) (*apiPages.PromoteSessionResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("Promote - 晋升预览会话 [%d] slug=%s", sessionID.Int64(), req.Slug))

	title := strings.TrimSpace(req.Title)
	if title == "" || utf8.RuneCountInString(title) > 255 {
		return nil, xError.NewError(ctx, xError.ParameterError, "页面标题不能为空且不超过 255 字符", false, nil)
	}

	session, xErr := l.repo.previewSession.GetByID(ctx, sessionID)
	if xErr != nil {
		return nil, xErr
	}
	if session.Status != bConst.PreviewSessionStatusActive {
		return nil, xError.NewError(ctx, xError.NotFound, "预览会话不存在", false, nil)
	}

	slug := strings.TrimSpace(req.Slug)
	if session.SourcePageID == nil || session.SourcePageID.IsZero() {
		if xErr := l.validateSlug(ctx, slug); xErr != nil {
			return nil, xErr
		}
	}

	previewFiles, xErr := l.repo.previewFile.ListBySession(ctx, sessionID)
	if xErr != nil {
		return nil, xErr
	}
	entry := FindPreviewEntryFromPreviewFiles(previewFiles)
	if entry == "" {
		return nil, xError.NewError(ctx, xError.ParameterError, "会话内没有可评审入口（HTML、LPW 或 TSX），无法晋升", false, nil)
	}

	project, xErr := l.repo.project.GetByID(ctx, session.ProjectID)
	if xErr != nil {
		return nil, xErr
	}

	page, createPage, xErr := l.resolvePromoteTarget(ctx, session, project.ID, slug, title, req.Description)
	if xErr != nil {
		return nil, xErr
	}

	if !createPage && session.SourceVersionID != nil && !session.SourceVersionID.IsZero() &&
		page.LatestVersionID != *session.SourceVersionID && !req.ConfirmConflict {
		return l.conflictResponse(ctx, page, *session.SourceVersionID)
	}

	versionLabel := strings.TrimSpace(req.Version)
	if versionLabel == "" {
		versionLabel = bConst.PagesInitialVersion
		if !createPage {
			if latest, latestErr := l.repo.version.GetLatestByPage(ctx, page.ID); latestErr == nil {
				versionLabel = incrementPatchVersion(latest.Version)
			}
		}
	}
	if !createPage {
		if existing, existErr := l.repo.version.GetByPageAndVersionLabel(ctx, page.ID, versionLabel); existErr == nil && existing != nil {
			return nil, xError.NewError(ctx, xError.BusinessError, "版本号已存在", false, nil)
		} else if existErr != nil && existErr.GetErrorCode() != xError.NotFound {
			return nil, existErr
		}
	}

	setAsActive := req.SetAsActive || createPage
	createdBy, _ := l.repo.info.GetByKey(ctx, bConst.InfoKeyOwnerUsername)

	versionID := xSnowflake.GenerateID(bConst.GenePageVersion)
	var sourceSessionID *xSnowflake.SnowflakeID
	sid := session.ID
	sourceSessionID = &sid
	version := &entity.PageVersion{
		BaseEntity:      xModels.BaseEntity{ID: versionID},
		PageID:          page.ID,
		Version:         versionLabel,
		Changelog:       req.Changelog,
		SourceSessionID: sourceSessionID,
		BaseVersionID:   session.SourceVersionID,
		EntryFilename:   entry,
		FileCount:       len(previewFiles),
		CreatedBy:       createdBy,
	}

	// 复制文件无需重复大小校验：来源均为已通过 UploadFile 驱动感知上限校验的预览文件
	pageFiles := make([]*entity.PageFile, 0, len(previewFiles))
	var totalSize int64
	for _, previewFile := range previewFiles {
		totalSize += int64(previewFile.Size)
		pageFiles = append(pageFiles, &entity.PageFile{
			BaseEntity: xModels.BaseEntity{ID: xSnowflake.GenerateID(bConst.GenePageFile)},
			VersionID:  versionID,
			Filename:   previewFile.Filename,
			MimeType:   previewFile.MimeType,
			Size:       previewFile.Size,
			Content:    previewFile.Content,
		})
	}
	version.TotalSize = totalSize

	// 晋升即消耗草稿：PersistPromotion 在同一事务内写入快照并将源会话软删除，
	// 保证「快照落库」与「草稿下线」原子提交，杜绝半提交态（快照已出、草稿仍 active）；
	// Pages 快照为深拷贝，消耗草稿不影响已发布文件
	if xErr := l.repo.page.PersistPromotion(ctx, page, createPage, version, pageFiles, setAsActive, session.ID); xErr != nil {
		return nil, xErr
	}

	// 草稿已被消耗，广播会话结束让工作台 WS 及时下线（与 DeleteSession / ExpireStaleSessions 一致）
	if OnPreviewChanged != nil {
		OnPreviewChanged(sessionID.String(), "delete_session")
	}

	pageResp, xErr := l.toPageResponse(ctx, page)
	if xErr != nil {
		return nil, xErr
	}
	return &apiPages.PromoteSessionResponse{
		Page:    pageResp,
		Version: toPageVersionResponse(version, page.LatestVersionID),
	}, nil
}

func (l *PagesLogic) resolvePromoteTarget(ctx context.Context, session *entity.PreviewSession, projectID xSnowflake.SnowflakeID, slug, title, description string) (*entity.Page, bool, *xError.Error) {
	if session.SourcePageID != nil && !session.SourcePageID.IsZero() {
		page, xErr := l.repo.page.GetByID(ctx, *session.SourcePageID)
		if xErr != nil {
			return nil, false, xErr
		}
		page.Title = title
		page.Description = description
		return page, false, nil
	}

	existing, xErr := l.repo.page.GetByProjectAndSlug(ctx, projectID, slug)
	if xErr != nil {
		if xErr.GetErrorCode() != xError.NotFound {
			return nil, false, xErr
		}
		pageID := xSnowflake.GenerateID(bConst.GenePage)
		return &entity.Page{
			BaseEntity:  xModels.BaseEntity{ID: pageID},
			ProjectID:   projectID,
			Slug:        slug,
			Title:       title,
			Description: description,
			Status:      bConst.PageStatusPublished,
			AccessMode:  bConst.PageAccessModePublic,
		}, true, nil
	}
	existing.Title = title
	existing.Description = description
	return existing, false, nil
}

func (l *PagesLogic) conflictResponse(ctx context.Context, page *entity.Page, sourceVersionID xSnowflake.SnowflakeID) (*apiPages.PromoteSessionResponse, *xError.Error) {
	current, xErr := l.repo.version.GetByID(ctx, page.LatestVersionID)
	if xErr != nil {
		return nil, xErr
	}
	return &apiPages.PromoteSessionResponse{
		Conflict: &apiPreview.PromoteConflictResponse{
			CurrentVersion:   current.Version,
			CurrentVersionID: current.ID,
			CurrentChangelog: current.Changelog,
			CurrentCreatedBy: current.CreatedBy,
			CurrentCreatedAt: current.CreatedAt.Format(time.RFC3339),
			SourceVersionID:  sourceVersionID,
			Message:          fmt.Sprintf("线上已有更新版本发布：当前生效版本已推进至 %s。确认后可继续发布并覆盖生效指针。", current.Version),
		},
	}, nil
}

// Fork 从指定版本深拷贝为新 Preview 会话
func (l *PagesLogic) Fork(ctx context.Context, pageID xSnowflake.SnowflakeID, versionID xSnowflake.SnowflakeID) (*apiPages.ForkPageResponse, *xError.Error) {
	page, xErr := l.repo.page.GetByID(ctx, pageID)
	if xErr != nil {
		return nil, xErr
	}
	if versionID.IsZero() {
		versionID = page.LatestVersionID
	}
	version, xErr := l.repo.version.GetByID(ctx, versionID)
	if xErr != nil {
		return nil, xErr
	}
	if version.PageID != page.ID {
		return nil, xError.NewError(ctx, xError.ParameterError, "版本不属于该页面", false, nil)
	}
	files, xErr := l.repo.file.ListByVersion(ctx, version.ID)
	if xErr != nil {
		return nil, xErr
	}

	sessionID := xSnowflake.GenerateID(bConst.GenePreviewSession)
	expiresAt := time.Now().Add(previewSessionTTL(ctx, l.repo.info))
	sourcePageID := page.ID
	sourceVersionID := version.ID
	session := &entity.PreviewSession{
		BaseEntity:      xModels.BaseEntity{ID: sessionID},
		ProjectID:       page.ProjectID,
		Title:           fmt.Sprintf("%s · 基于 %s", page.Title, version.Version),
		Hash:            generateSessionHash(sessionID),
		Status:          bConst.PreviewSessionStatusActive,
		ExpiresAt:       &expiresAt,
		SourcePageID:    &sourcePageID,
		SourceVersionID: &sourceVersionID,
	}

	// 复制文件无需重复大小校验：来源均为已通过 UploadFile 驱动感知上限校验的页面快照文件
	previewFiles := make([]*entity.PreviewFile, 0, len(files))
	for _, file := range files {
		previewFiles = append(previewFiles, &entity.PreviewFile{
			BaseEntity: xModels.BaseEntity{ID: xSnowflake.GenerateID(bConst.GenePreviewFile)},
			SessionID:  sessionID,
			Filename:   file.Filename,
			MimeType:   file.MimeType,
			Size:       file.Size,
			Content:    file.Content,
		})
	}
	if xErr := l.repo.previewSession.CreateWithFiles(ctx, session, previewFiles); xErr != nil {
		return nil, xErr
	}

	resp := toPreviewSessionResponse(session, page.Slug)
	resp.FileCount = int64(len(previewFiles))
	entry := version.EntryFilename
	if entry == "" {
		entry = FindPreviewEntryFromPageFiles(files)
	}
	return &apiPages.ForkPageResponse{
		Session:    *resp,
		PreviewURL: l.BuildPreviewURL(ctx, session.Hash, entry),
	}, nil
}

// SwitchActive 原子切换生效版本指针
func (l *PagesLogic) SwitchActive(ctx context.Context, pageID, versionID xSnowflake.SnowflakeID) (*apiPages.SwitchActiveResponse, *xError.Error) {
	page, xErr := l.repo.page.GetByID(ctx, pageID)
	if xErr != nil {
		return nil, xErr
	}
	version, xErr := l.repo.version.GetByID(ctx, versionID)
	if xErr != nil {
		return nil, xErr
	}
	if version.PageID != page.ID {
		return nil, xError.NewError(ctx, xError.ParameterError, "版本不属于该页面", false, nil)
	}
	if xErr := l.repo.page.UpdateLatestVersionID(ctx, pageID, versionID); xErr != nil {
		return nil, xErr
	}
	page.LatestVersionID = versionID
	pageResp, xErr := l.toPageResponse(ctx, page)
	if xErr != nil {
		return nil, xErr
	}
	return &apiPages.SwitchActiveResponse{
		Page:    *pageResp,
		Version: *toPageVersionResponse(version, versionID),
	}, nil
}

// Archive 将页面归档
func (l *PagesLogic) Archive(ctx context.Context, pageID xSnowflake.SnowflakeID) (*apiPages.PageResponse, *xError.Error) {
	page, xErr := l.repo.page.GetByID(ctx, pageID)
	if xErr != nil {
		return nil, xErr
	}
	if xErr := l.repo.page.UpdateStatus(ctx, pageID, bConst.PageStatusArchived); xErr != nil {
		return nil, xErr
	}
	page.Status = bConst.PageStatusArchived
	return l.toPageResponse(ctx, page)
}

// UpdateAccessPolicy 仅管理员配置公开/密码
func (l *PagesLogic) UpdateAccessPolicy(ctx context.Context, pageID xSnowflake.SnowflakeID, accessMode, password string) (*apiPages.PageResponse, *xError.Error) {
	page, xErr := l.repo.page.GetByID(ctx, pageID)
	if xErr != nil {
		return nil, xErr
	}
	accessMode = strings.TrimSpace(accessMode)
	var hash string
	switch accessMode {
	case bConst.PageAccessModePublic:
		hash = ""
	case bConst.PageAccessModePassword:
		if strings.TrimSpace(password) == "" {
			return nil, xError.NewError(ctx, xError.ParameterError, "密码保护模式必须设置访问密码", false, nil)
		}
		hashed, err := service.HashPassword(password)
		if err != nil {
			return nil, xError.NewError(ctx, xError.ServerInternalError, "密码哈希失败", false, err)
		}
		hash = hashed
	default:
		return nil, xError.NewError(ctx, xError.ParameterError, "访问策略仅支持 public 或 password", false, nil)
	}
	if xErr := l.repo.page.UpdateAccessPolicy(ctx, pageID, accessMode, hash); xErr != nil {
		return nil, xErr
	}
	page.AccessMode = accessMode
	page.PasswordHash = hash
	return l.toPageResponse(ctx, page)
}

// CheckAuth 返回密码门状态
func (l *PagesLogic) CheckAuth(ctx context.Context, projectName, slug, cookieValue string) (*apiPages.PageAuthCheckResponse, *xError.Error) {
	page, _, xErr := l.GetByProjectNameAndSlug(ctx, projectName, slug)
	if xErr != nil {
		return nil, xErr
	}
	// 与 PublicMeta/ServeFile 对齐：归档页不可探测，公开端点一律 404
	if page.Status != bConst.PageStatusPublished {
		return nil, xError.NewError(ctx, xError.NotFound, "页面不存在", false, nil)
	}
	required := page.AccessMode == bConst.PageAccessModePassword && page.PasswordHash != ""
	authenticated := !required
	if required && cookieValue != "" {
		authenticated = l.authToken.ValidateToken(cookieValue, page.ID.Int64())
	}
	return &apiPages.PageAuthCheckResponse{
		Authenticated:    authenticated,
		PasswordRequired: required,
	}, nil
}

// ── Pages 密码门暴力破解限流 ──
//
// 内存级失败计数限流（Lumina 为单实例单用户部署，内存态即可），
// 对每个页面的解锁失败次数计数，超过阈值后锁定一段时间。
// 与 Wiki 限流器（handler 层）不同：Pages 的密码校验与 pageID 解析都在 logic 层
// 的 Unlock 内完成，handler 无法先于校验拿到限流维度，故限流器内聚在 logic 层。
const (
	pagesAuthMaxFailures  = 10               // 最大连续失败次数（与 Wiki 密码门阈值对齐）
	pagesAuthLockDuration = 15 * time.Minute // 失败锁定窗口
)

type pagesAuthLimiter struct {
	mu       sync.Mutex
	failures map[int64]int       // pageID → 连续失败次数
	lockedAt map[int64]time.Time // pageID → 锁定截止时间
}

// pagesAuthLimit 进程级共享限流器（PagesLogic 存在多处构造，计数必须全进程归一）
var pagesAuthLimit = newPagesAuthLimiter()

func newPagesAuthLimiter() *pagesAuthLimiter {
	return &pagesAuthLimiter{
		failures: make(map[int64]int),
		lockedAt: make(map[int64]time.Time),
	}
}

// allow 返回该页面是否允许继续尝试解锁
func (l *pagesAuthLimiter) allow(pageID int64) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if t, ok := l.lockedAt[pageID]; ok && time.Now().Before(t) {
		return false
	}
	return true
}

// recordFailure 记录一次失败，累计达到阈值则锁定
func (l *pagesAuthLimiter) recordFailure(pageID int64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.failures[pageID]++
	if l.failures[pageID] >= pagesAuthMaxFailures {
		l.lockedAt[pageID] = time.Now().Add(pagesAuthLockDuration)
		l.failures[pageID] = 0
	}
}

// reset 解锁成功后清零计数
func (l *pagesAuthLimiter) reset(pageID int64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.failures, pageID)
	delete(l.lockedAt, pageID)
}

// Unlock 校验密码并签发 Cookie
func (l *PagesLogic) Unlock(ctx context.Context, projectName, slug, password string) (pageID int64, token string, maxAge int, xErr *xError.Error) {
	page, _, xErr := l.GetByProjectNameAndSlug(ctx, projectName, slug)
	if xErr != nil {
		return 0, "", 0, xErr
	}
	// 与 PublicMeta/ServeFile 对齐：归档页不可探测、不可解锁，一律 404
	if page.Status != bConst.PageStatusPublished {
		return 0, "", 0, xError.NewError(ctx, xError.NotFound, "页面不存在", false, nil)
	}
	if page.AccessMode != bConst.PageAccessModePassword || page.PasswordHash == "" {
		return page.ID.Int64(), "", 0, nil
	}
	// 锁定期内直接拒绝，防止匿名无限爆破 bcrypt 比对
	if !l.authLimit.allow(page.ID.Int64()) {
		return 0, "", 0, xError.NewError(ctx, xError.TooManyRequests, "尝试次数过多，请稍后再试", false, nil)
	}
	if !service.VerifyPassword(password, page.PasswordHash) {
		l.authLimit.recordFailure(page.ID.Int64())
		return 0, "", 0, xError.NewError(ctx, xError.Unauthorized, "页面密码错误", false, nil)
	}
	l.authLimit.reset(page.ID.Int64())
	maxAge = l.cookieMaxAge(ctx)
	token, err := l.authToken.GenerateToken(page.ID.Int64(), maxAge)
	if err != nil {
		return 0, "", 0, xError.NewError(ctx, xError.ServerInternalError, "Token 生成失败", false, err)
	}
	return page.ID.Int64(), token, maxAge, nil
}

// BuildPreviewURL 构造路径式预览地址
func (l *PagesLogic) BuildPreviewURL(ctx context.Context, hash, filename string) string {
	return buildPreviewURL(resolveRuntimeDomain(ctx, l.repo.info), hash, filename)
}

// BuildPageURL 构造路径式 Pages 地址
func (l *PagesLogic) BuildPageURL(ctx context.Context, projectName, slug, filename string) string {
	return buildPagesURL(resolveRuntimeDomain(ctx, l.repo.info), projectName, slug, filename)
}

func (l *PagesLogic) cookieMaxAge(ctx context.Context) int {
	value, xErr := l.repo.info.GetByKey(ctx, bConst.InfoKeyPagesAuthCookieMaxAge)
	if xErr != nil {
		return bConst.PagesCookieMaxAge
	}
	sec, err := strconv.Atoi(value)
	if err != nil || sec <= 0 {
		return bConst.PagesCookieMaxAge
	}
	return sec
}

func (l *PagesLogic) validateSlug(ctx context.Context, slug string) *xError.Error {
	slug = strings.TrimSpace(slug)
	if slug == "" || utf8.RuneCountInString(slug) > 64 || !bConst.PageSlugPattern.MatchString(slug) {
		return xError.NewError(ctx, xError.ParameterError, "页面标识仅允许小写字母、数字与短横线", false, nil)
	}
	return nil
}

func (l *PagesLogic) resolveVersion(ctx context.Context, page *entity.Page, versionLabel string) (*entity.PageVersion, *xError.Error) {
	label := strings.TrimSpace(versionLabel)
	// 合法语义化版本按标签精确查询；垃圾输入（如前端误传的 RFC3339 缓存戳）回退生效指针，
	// 避免展示页因错误缓存参数整页 404
	if normalized, ok := normalizeVersionLabel(label); ok {
		return l.repo.version.GetByPageAndVersionLabel(ctx, page.ID, normalized)
	}
	if page.LatestVersionID.IsZero() {
		return nil, xError.NewError(ctx, xError.NotFound, "页面尚未发布版本", false, nil)
	}
	return l.repo.version.GetByID(ctx, page.LatestVersionID)
}

// normalizeVersionLabel 将外部传入的版本标签规范化为 vX.Y.Z 形态（兼容无 v 前缀写法）。
//
// 仅接受三段非负整数的语义化版本（如 v1.0.0 / 1.2.3，对齐 incrementPatchVersion 的解析口径）；
// 其余输入（缓存时间戳、随机串等）返回 false，由 resolveVersion 回退生效版本指针。
func normalizeVersionLabel(label string) (string, bool) {
	body := strings.TrimPrefix(strings.TrimSpace(label), "v")
	parts := strings.Split(body, ".")
	if len(parts) != 3 {
		return "", false
	}
	for _, part := range parts {
		if !isDigits(part) {
			return "", false
		}
	}
	return "v" + body, true
}

// isDigits 判断字符串是否为非空纯数字（拒绝符号、空白与字母，比 strconv.Atoi 更严）
func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func (l *PagesLogic) toPageResponse(ctx context.Context, page *entity.Page) (*apiPages.PageResponse, *xError.Error) {
	projectName := ""
	if project, xErr := l.repo.project.GetByID(ctx, page.ProjectID); xErr == nil {
		projectName = project.Name
	}
	latestVersion := ""
	if !page.LatestVersionID.IsZero() {
		if version, xErr := l.repo.version.GetByID(ctx, page.LatestVersionID); xErr == nil {
			latestVersion = version.Version
		}
	}
	filename := ""
	if latestVersion != "" {
		if version, xErr := l.repo.version.GetByID(ctx, page.LatestVersionID); xErr == nil {
			filename = version.EntryFilename
		}
	}
	return &apiPages.PageResponse{
		ID:              page.ID,
		ProjectID:       page.ProjectID,
		ProjectName:     projectName,
		Slug:            page.Slug,
		Title:           page.Title,
		Description:     page.Description,
		Status:          page.Status,
		AccessMode:      page.AccessMode,
		LatestVersionID: page.LatestVersionID,
		LatestVersion:   latestVersion,
		PageURL:         buildPagesURL(resolveRuntimeDomain(ctx, l.repo.info), projectName, page.Slug, filename),
		CreatedAt:       page.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       page.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func toPageVersionResponse(version *entity.PageVersion, latestID xSnowflake.SnowflakeID) *apiPages.PageVersionResponse {
	return &apiPages.PageVersionResponse{
		ID:              version.ID,
		PageID:          version.PageID,
		Version:         version.Version,
		Changelog:       version.Changelog,
		SourceSessionID: version.SourceSessionID,
		BaseVersionID:   version.BaseVersionID,
		EntryFilename:   version.EntryFilename,
		FileCount:       version.FileCount,
		TotalSize:       version.TotalSize,
		CreatedBy:       version.CreatedBy,
		IsActive:        version.ID == latestID,
		CreatedAt:       version.CreatedAt.Format(time.RFC3339),
	}
}

func toPageFileResponse(file *entity.PageFile) *apiPages.PageFileResponse {
	return &apiPages.PageFileResponse{
		ID:        file.ID,
		VersionID: file.VersionID,
		Filename:  file.Filename,
		MimeType:  file.MimeType,
		Size:      file.Size,
		CreatedAt: file.CreatedAt.Format(time.RFC3339),
		UpdatedAt: file.UpdatedAt.Format(time.RFC3339),
	}
}

func findHTMLEntry(files []*entity.PreviewFile) string {
	return FindPreviewEntryFromPreviewFiles(files)
}

func findHTMLEntryFromPageFiles(files []*entity.PageFile) string {
	return FindPreviewEntryFromPageFiles(files)
}

func isHTMLFilename(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".html" || ext == ".htm"
}

func isTSXFilename(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".tsx" || ext == ".jsx"
}

func incrementPatchVersion(current string) string {
	label := strings.TrimPrefix(strings.TrimSpace(current), "v")
	parts := strings.Split(label, ".")
	if len(parts) != 3 {
		return bConst.PagesInitialVersion
	}
	major, err1 := strconv.Atoi(parts[0])
	minor, err2 := strconv.Atoi(parts[1])
	patch, err3 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil || err3 != nil {
		return bConst.PagesInitialVersion
	}
	return fmt.Sprintf("v%d.%d.%d", major, minor, patch+1)
}
