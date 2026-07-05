// Package logic RepoWiki 业务编排层。
//
// RepoWikiLogic 作为 RepoWiki 模块的核心编排者，负责：
//   - 配置管理（CRUD + SSH Key 加密 + 密码哈希）
//   - 版本管理（触发分析、状态查询、版本列表）
//   - 分析管道调度（信号量并发控制 + 后台 goroutine 执行）
//   - Wiki 内容查询（MCP 工具调用入口）
//
// 分析管道（AnalysisPipeline）在独立文件 repowiki_pipeline.go 中实现，
// Logic 通过持有 Pipeline 引用将业务方法与管道执行解耦。
//
// 严格遵循分层约定：Logic 仅持有 repo（+cache）和 service 引用，
// 禁止直接持有 db/rdb 或拼接 SQL/Redis 命令。
package logic

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xCtxUtil "github.com/bamboo-services/bamboo-base-go/common/utility/context"
	xEnv "github.com/bamboo-services/bamboo-base-go/defined/env"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"

	apiRepowiki "github.com/xiaolfeng/Lumina/api/repowiki"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"github.com/xiaolfeng/Lumina/internal/repository"
	"github.com/xiaolfeng/Lumina/internal/service"
)

// ──────────────────────────────────────────────────────────────────────
// RepoWikiLogic
//
// DocumentAssembler（文档组装器）定义在 repowiki_assembler.go 中，
// 通过 SetDocumentAssembler 注入。Pipeline 中通过 assembler.Assemble() 调用。
// ──────────────────────────────────────────────────────────────────────

// repowikiRepo RepoWiki 模块依赖的仓储集合
//
// 刻意聚合在独立结构体中，保持 RepoWikiLogic 主结构体字段简洁。
// config 和 version 分别对应 RepoWikiConfigRepo 和 WikiVersionRepo。
type repowikiRepo struct {
	config  *repository.RepoWikiConfigRepo // 配置 CRUD + Cache-Aside 缓存
	version *repository.WikiVersionRepo    // 版本 CRUD + 状态缓存
}

// repowikiSvc RepoWiki 模块依赖的共享服务集合
//
// 所有服务在 NewRepoWikiLogic 时一次性构造，通过聚合结构体注入，
// 避免主结构体字段过多且便于 Pipeline 统一引用。
type repowikiSvc struct {
	git       *service.GitCloneService            // Git 克隆 / commit hash / diff
	scanner   *service.FileScannerService         // 仓库文件扫描
	extractor *service.DependencyExtractorService // 模块级依赖提取
	storage   *service.WikiStorageService         // 文件系统路径管理 + JSON/MD 读写
	authToken *service.WikiAuthTokenService       // Wiki 访问 Cookie HMAC 签名
}

// RepoWikiLogic RepoWiki 业务编排层
//
// 职责边界：
//   - 配置 CRUD 编排（含 SSH Key AES-GCM 加密、密码 bcrypt 哈希）
//   - 版本记录生命周期管理（创建 → pending → 分析中 → completed/failed）
//   - 分析任务并发控制（semaphore 信号量，容量由 REPOWIKI_MAX_CONCURRENT 控制）
//   - 后台 goroutine 调度（AnalyzeRepo 不阻塞请求，管道在独立 goroutine 执行）
//   - Wiki 内容查询（为 MCP 工具提供入口）
//
// 非职责：
//   - 不直接执行 Git 克隆 / 文件扫描 / LLM 分析（由 service 层 + AgentPassRunner 完成）
//   - 不负责文档组装（由 DocumentAssembler — Task 16 — 完成）
//   - 不直接操作数据库（经由 repository 层）
type RepoWikiLogic struct {
	logic
	repo        repowikiRepo       // 仓储层依赖
	svc         repowikiSvc        // 服务层依赖
	runner      *AgentPassRunner   // Agent 分析 Pass 运行器（nil 表示 LLM Provider 未就绪）
	assembler   *DocumentAssembler // 文档组装器（nil 表示未注入，Pipeline 跳过组装并记录日志）
	semaphore   chan struct{}      // 并发控制信号量（容量 = REPOWIKI_MAX_CONCURRENT）
	llmProvider string             // LLM Provider 类型
	llmModel    string             // LLM 模型名称
}

// NewRepoWikiLogic 创建 RepoWikiLogic 实例
//
// 通过上下文获取 db/rdb，构造所有仓储和服务依赖。
// LLM Provider 初始化失败时不阻塞 Logic 创建（runner 为 nil），
// AnalyzeRepo 调用时会返回错误提示用户配置 LLM。
//
// 并发控制：
//   - 信号量容量由环境变量 REPOWIKI_MAX_CONCURRENT 控制（默认 1）
//   - 每次 AnalyzeRepo 启动后台 goroutine 前非阻塞获取信号量
//   - goroutine 结束后释放信号量
func NewRepoWikiLogic(ctx context.Context) *RepoWikiLogic {
	db := xCtxUtil.MustGetDB(ctx)
	rdb := xCtxUtil.MustGetRDB(ctx)

	// 读取并发上限（最小 1）
	maxConcurrent := max(xEnv.GetEnvInt("REPOWIKI_MAX_CONCURRENT", 1), 1)

	l := &RepoWikiLogic{
		logic: logic{
			log: xLog.WithName(xLog.NamedLOGC, "RepoWikiLogic"),
		},
		repo: repowikiRepo{
			config:  repository.NewRepoWikiConfigRepo(db, rdb),
			version: repository.NewWikiVersionRepo(db, rdb),
		},
		svc: repowikiSvc{
			git:       service.NewGitCloneService(),
			scanner:   service.NewFileScannerService(),
			extractor: service.NewDependencyExtractorService(),
			storage:   service.NewWikiStorageService(),
			authToken: service.NewWikiAuthTokenService(),
		},
		semaphore:   make(chan struct{}, maxConcurrent),
		llmProvider: xEnv.GetEnvString("LLM_PROVIDER", "openai"),
		llmModel:    xEnv.GetEnvString("LLM_MODEL", "gpt-4o"),
	}

	// 尝试初始化 LLM Provider 和 AgentPassRunner
	// 失败不阻塞 Logic 创建——非分析类方法（CRUD/查询）仍可正常工作
	if client, err := service.NewLLMProvider(); err != nil {
		l.log.Warn(ctx, "LLM Provider 初始化失败，AnalyzeRepo 将不可用",
			slog.String("err", err.Error()))
	} else {
		l.runner = NewAgentPassRunner(client, l.svc.storage, l.log, nil)
		l.log.Info(ctx, "AgentPassRunner 已就绪")
	}

	return l
}

// GetRepoWikiLogicFromContext 从 context 中获取 RepoWikiLogic 实例。
//
// 启动阶段将实例注入 context，handler 与 MCP 共享同一实例以保证
// semaphore 等状态全局唯一（否则每次新建 Logic 会导致并发控制失效）。
func GetRepoWikiLogicFromContext(ctx context.Context) *RepoWikiLogic {
	if l, ok := ctx.Value(bConst.RepoWikiLogicKey).(*RepoWikiLogic); ok {
		return l
	}
	return nil
}

// SetDocumentAssembler 注入文档组装器（由 startup 阶段调用）
//
// assembler 为 nil 时 Pipeline 跳过文档组装步骤并记录日志。
func (l *RepoWikiLogic) SetDocumentAssembler(assembler *DocumentAssembler) {
	l.assembler = assembler
}

// ──────────────────────────────────────────────────────────────────────
// 配置管理方法
// ──────────────────────────────────────────────────────────────────────

// CreateConfig 创建 RepoWiki 配置
//
// 业务流程：
//  1. SSH Key 非空时使用 AES-GCM 加密（EncryptSSHKey）
//  2. Wiki 密码非空时使用 bcrypt 哈希（HashPassword）
//  3. 构建实体并生成雪花 ID（GeneRepoWikiConfig）
//  4. 持久化（repo.config.Create 同步写入缓存）
//
// 参数说明:
//   - ctx: 上下文
//   - req:  创建配置请求 DTO（RepoURL 必填，SSH Key / 密码可选）
//
// 返回值:
//   - *entity.RepoWikiConfig: 创建后的配置实体（含生成的 ID）
//   - *xError.Error:           SSH Key 加密失败 / 密码哈希失败 / 持久化失败
func (l *RepoWikiLogic) CreateConfig(ctx context.Context, req *apiRepowiki.CreateConfigRequest) (*entity.RepoWikiConfig, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("CreateConfig - 创建 RepoWiki 配置 [repoURL=%s]", req.RepoURL))

	// SSH Key 加密（非空时）
	var sshKeyEncrypted string
	if req.SSHKey != "" {
		secret := xEnv.GetEnvString("REPOWIKI_HMAC_SECRET", "")
		encrypted, xErr := service.EncryptSSHKey(req.SSHKey, secret)
		if xErr != nil {
			return nil, xErr
		}
		sshKeyEncrypted = encrypted
	}

	// Wiki 密码哈希（非空时）
	var passwordHash string
	if req.WikiPassword != "" {
		hash, err := service.HashPassword(req.WikiPassword)
		if err != nil {
			return nil, xError.NewError(ctx, xError.ServerInternalError, "Wiki 密码哈希失败", false, err)
		}
		passwordHash = hash
	}

	// 默认值填充
	branch := req.DefaultBranch
	if branch == "" {
		branch = "main"
	}
	language := req.DefaultLanguage
	if language == "" {
		language = bConst.RepoWikiDefaultLanguage
	}

	// 构建实体
	configID := xSnowflake.GenerateID(bConst.GeneRepoWikiConfig)
	config := &entity.RepoWikiConfig{
		BaseEntity:       xModels.BaseEntity{ID: configID},
		ProjectID:        xSnowflake.SnowflakeID(req.ProjectID),
		Name:             req.Name,
		GitURL:           req.RepoURL,
		DefaultBranch:    branch,
		DefaultLanguage:  language,
		SSHKeyEncrypted:  sshKeyEncrypted,
		WikiPasswordHash: passwordHash,
		Status:           bConst.RepoWikiStatusPending,
	}

	// 持久化
	created, xErr := l.repo.config.Create(ctx, config)
	if xErr != nil {
		return nil, xErr
	}

	return created, nil
}

// GetConfig 根据 ID 获取配置详情
func (l *RepoWikiLogic) GetConfig(ctx context.Context, id xSnowflake.SnowflakeID) (*entity.RepoWikiConfig, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("GetConfig - 获取配置详情 [%d]", id.Int64()))
	return l.repo.config.GetByID(ctx, id)
}

// ListConfigs 分页获取配置列表
//
// 参数说明:
//   - ctx:  上下文
//   - page: 页码（从 1 开始）
//   - size: 每页数量
//
// 返回值:
//   - []*entity.RepoWikiConfig: 当前页配置列表
//   - int64:                     总记录数
//   - *xError.Error:             查询错误
func (l *RepoWikiLogic) ListConfigs(ctx context.Context, page, size int) ([]*entity.RepoWikiConfig, int64, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("ListConfigs - 分页获取配置列表 [page=%d, size=%d]", page, size))

	// 分页参数规范化
	pageReq := xModels.PageRequest{Page: int64(page), Size: int64(size)}.Normalize()
	return l.repo.config.List(ctx, int(pageReq.Page), int(pageReq.Size))
}

// UpdateConfig 更新配置（仅更新提供的字段，指针 nil = 不更新）
//
// 业务规则：
//   - RepoURL / DefaultBranch / DefaultLanguage：非 nil 时直接更新
//   - SSHKey：非 nil 时加密后更新（空字符串表示清除 SSH Key）
//   - WikiPassword：非 nil 时哈希后更新（空字符串表示清除密码）
//
// 参数说明:
//   - ctx: 上下文
//   - id:  配置雪花 ID
//   - req: 更新请求 DTO（指针字段 nil = 不更新）
func (l *RepoWikiLogic) UpdateConfig(ctx context.Context, id xSnowflake.SnowflakeID, req *apiRepowiki.UpdateConfigRequest) *xError.Error {
	l.log.Info(ctx, fmt.Sprintf("UpdateConfig - 更新配置 [%d]", id.Int64()))

	// 获取现有配置（全量）
	config, xErr := l.repo.config.GetByID(ctx, id)
	if xErr != nil {
		return xErr
	}

	// 逐字段按需更新
	if req.RepoURL != nil {
		config.GitURL = *req.RepoURL
	}
	if req.DefaultBranch != nil {
		config.DefaultBranch = *req.DefaultBranch
	}
	if req.DefaultLanguage != nil {
		config.DefaultLanguage = *req.DefaultLanguage
	}

	// SSH Key 更新（空字符串 = 清除）
	if req.SSHKey != nil {
		if *req.SSHKey == "" {
			config.SSHKeyEncrypted = ""
		} else {
			secret := xEnv.GetEnvString("REPOWIKI_HMAC_SECRET", "")
			encrypted, xErr := service.EncryptSSHKey(*req.SSHKey, secret)
			if xErr != nil {
				return xErr
			}
			config.SSHKeyEncrypted = encrypted
		}
	}

	// Wiki 密码更新（空字符串 = 清除）
	if req.WikiPassword != nil {
		if *req.WikiPassword == "" {
			config.WikiPasswordHash = ""
		} else {
			hash, err := service.HashPassword(*req.WikiPassword)
			if err != nil {
				return xError.NewError(ctx, xError.ServerInternalError, "Wiki 密码哈希失败", false, err)
			}
			config.WikiPasswordHash = hash
		}
	}

	return l.repo.config.Update(ctx, config)
}

// DeleteConfig 删除配置
//
// 注意：仅删除数据库记录，不清理文件系统中的克隆仓库和版本数据。
// 文件清理应由上层（Handler）在删除前显式调用 storage 清理方法。
func (l *RepoWikiLogic) DeleteConfig(ctx context.Context, id xSnowflake.SnowflakeID) *xError.Error {
	l.log.Info(ctx, fmt.Sprintf("DeleteConfig - 删除配置 [%d]", id.Int64()))
	return l.repo.config.Delete(ctx, id)
}

// ──────────────────────────────────────────────────────────────────────
// 版本管理方法
// ──────────────────────────────────────────────────────────────────────

// AnalyzeRepo 触发仓库分析（异步执行）
//
// 业务流程：
//  1. 获取配置详情，校验存在性
//  2. 检查 AgentPassRunner 是否就绪（LLM Provider 已配置）
//  3. 非阻塞获取并发信号量（满则返回 BusinessError）
//  4. 确定分支和语言（优先请求参数，其次配置默认值）
//  5. 创建 WikiVersion 记录（status=pending）
//  6. 启动后台 goroutine 执行 AnalysisPipeline.Execute
//  7. 立即返回版本记录（status=pending），调用方可轮询 GetVersionStatus
//
// 并发控制：
//   - 信号量在 goroutine 内 defer 释放，确保异常退出也能释放
//   - 信号量容量 = REPOWIKI_MAX_CONCURRENT（默认 1）
//
// 参数说明:
//   - ctx:      上下文（用于同步阶段的 DB 操作；后台 goroutine 使用 context.Background()）
//   - configID: 配置雪花 ID
//   - req:      分析请求（Language / Branch 可选，默认使用配置值）
//
// 返回值:
//   - *entity.WikiVersion: 创建的版本记录（status=pending）
//   - *xError.Error:        配置不存在 / LLM 未就绪 / 并发上限 / 持久化失败
func (l *RepoWikiLogic) AnalyzeRepo(ctx context.Context, configID xSnowflake.SnowflakeID, req *apiRepowiki.AnalyzeRequest) (*entity.WikiVersion, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("AnalyzeRepo - 触发仓库分析 [configID=%d]", configID.Int64()))

	// Step 1: 获取配置
	config, xErr := l.repo.config.GetByID(ctx, configID)
	if xErr != nil {
		return nil, xErr
	}

	// Step 2: 检查 AgentPassRunner 就绪状态
	if l.runner == nil {
		return nil, xError.NewError(ctx, xError.ServerInternalError,
			"LLM Provider 未就绪，无法执行分析（请检查 LLM_API_KEY 配置）", false, nil)
	}

	// Step 3: 非阻塞获取并发信号量
	select {
	case l.semaphore <- struct{}{}:
		// 获取成功
	default:
		return nil, xError.NewError(ctx, xError.BusinessError,
			"分析任务已达并发上限，请稍后重试", false, nil)
	}

	// Step 4: 确定分支和语言
	branch := req.Branch
	if branch == "" {
		branch = config.DefaultBranch
	}
	if branch == "" {
		branch = "main"
	}

	language := req.Language
	if language == "" {
		language = config.DefaultLanguage
	}
	if language == "" {
		language = bConst.RepoWikiDefaultLanguage
	}

	// Step 5: 创建版本记录
	versionID := xSnowflake.GenerateID(bConst.GeneWikiVersion)
	version := &entity.WikiVersion{
		BaseEntity: xModels.BaseEntity{ID: versionID},
		ConfigID:   configID,
		Branch:     branch,
		Language:   language,
		Status:     bConst.RepoWikiStatusPending,
	}

	created, xErr := l.repo.version.Create(ctx, version)
	if xErr != nil {
		<-l.semaphore // 释放信号量
		return nil, xErr
	}

	// Step 6: 启动后台 goroutine 执行分析管道
	pipeline := NewAnalysisPipeline(l, l.log)

	go func() {
		defer func() {
			<-l.semaphore
		}()

		// 使用独立 context，不绑定请求生命周期
		bgCtx := context.Background()
		if pErr := pipeline.Execute(bgCtx, config, created); pErr != nil {
			l.log.Error(bgCtx, "分析管道执行失败",
				slog.Int64("versionID", created.ID.Int64()),
				slog.String("err", pErr.Error()))
		} else {
			l.log.Info(bgCtx, "分析管道执行完成",
				slog.Int64("versionID", created.ID.Int64()))
		}
	}()

	return created, nil
}

// GetVersionStatus 获取版本分析状态
func (l *RepoWikiLogic) GetVersionStatus(ctx context.Context, versionID xSnowflake.SnowflakeID) (*entity.WikiVersion, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("GetVersionStatus - 获取版本状态 [%d]", versionID.Int64()))
	return l.repo.version.GetByID(ctx, versionID)
}

// ListVersions 按配置 ID 分页获取版本列表
//
// 参数说明:
//   - ctx:      上下文
//   - configID: 配置雪花 ID
//   - page:     页码（从 1 开始）
//   - size:     每页数量
func (l *RepoWikiLogic) ListVersions(ctx context.Context, configID xSnowflake.SnowflakeID, page, size int) ([]*entity.WikiVersion, int64, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("ListVersions - 分页获取版本列表 [configID=%d, page=%d, size=%d]", configID.Int64(), page, size))

	pageReq := xModels.PageRequest{Page: int64(page), Size: int64(size)}.Normalize()
	return l.repo.version.ListByConfigID(ctx, configID, int(pageReq.Page), int(pageReq.Size))
}

// ──────────────────────────────────────────────────────────────────────
// Wiki 查询方法（MCP 工具入口）
// ──────────────────────────────────────────────────────────────────────

// QueryWiki 查询 Wiki 内容（MCP 工具用）
//
// 根据 wikiID 定位项目的 Wiki 目录，返回 manifest 或首页 Markdown 内容。
// query 参数保留用于未来全文检索扩展（当前 v1 仅返回首页摘要）。
//
// 参数说明:
//   - ctx:    上下文
//   - wikiID: Wiki 版本 ID（用于定位 Wiki 文档目录）
//   - query:  查询关键词（v1 未使用，保留扩展）
//
// 返回值:
//   - string:       Wiki 内容（Markdown 或 manifest JSON 摘要）
//   - *xError.Error: 版本不存在 / Wiki 未生成
func (l *RepoWikiLogic) QueryWiki(ctx context.Context, wikiID int64, query string) (string, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("QueryWiki - 查询 Wiki 内容 [wikiID=%d, query=%s]", wikiID, query))

	// 获取版本记录，确认 Wiki 已生成
	version, xErr := l.repo.version.GetByID(ctx, xSnowflake.SnowflakeID(wikiID))
	if xErr != nil {
		return "", xErr
	}

	if version.Status != bConst.RepoWikiStatusCompleted {
		return "", xError.NewError(ctx, xError.BusinessError,
			xError.ErrMessage(fmt.Sprintf("Wiki 尚未生成完成（当前状态: %s）", version.Status)), false, nil)
	}

	// 尝试读取 manifest
	manifestPath := l.svc.storage.GetManifestPath(version.ConfigID.Int64())
	if content, xErr := l.svc.storage.ReadMarkdown(manifestPath); xErr == nil {
		return content, nil
	}

	// manifest 不存在，尝试读取首页 index.md
	wikiPath := l.svc.storage.GetWikiPath(version.ConfigID.Int64())
	indexPath := filepath.Join(wikiPath, "index.md")
	if content, xErr := l.svc.storage.ReadMarkdown(indexPath); xErr == nil {
		return content, nil
	}

	return "", xError.NewError(ctx, xError.NotFound,
		"Wiki 文档文件不存在，可能文档组装尚未完成", false, nil)
}

// ──────────────────────────────────────────────────────────────────────
// Wiki 访问认证辅助方法
// ──────────────────────────────────────────────────────────────────────

// GenerateWikiToken 生成 Wiki 访问 Cookie Token
//
// 供 Handler 层在 Wiki 密码验证通过后调用，设置 HMAC 签名 Cookie。
// Token 有效期由 bConst.RepoWikiCookieMaxAge 控制（默认 2 小时）。
func (l *RepoWikiLogic) GenerateWikiToken(wikiID int64, password string) (string, error) {
	return l.svc.authToken.GenerateToken(wikiID, password)
}

// ValidateWikiToken 校验 Wiki 访问 Cookie Token
func (l *RepoWikiLogic) ValidateWikiToken(cookieValue string, wikiID int64) bool {
	return l.svc.authToken.ValidateToken(cookieValue, wikiID)
}

// HasWikiPassword 检查配置是否设置了 Wiki 访问密码
func (l *RepoWikiLogic) HasWikiPassword(config *entity.RepoWikiConfig) bool {
	return config.WikiPasswordHash != ""
}

// VerifyWikiPassword 校验 Wiki 访问密码
func (l *RepoWikiLogic) VerifyWikiPassword(config *entity.RepoWikiConfig, password string) bool {
	if config.WikiPasswordHash == "" {
		return true // 未设置密码，直接放行
	}
	return service.VerifyPassword(password, config.WikiPasswordHash)
}

// ──────────────────────────────────────────────────────────────────────
// 版本清理辅助方法
// ──────────────────────────────────────────────────────────────────────

// CleanVersionData 清理指定版本的文件系统数据
//
// 删除 {basePath}/versions/{versionID}/ 目录（含 raw/passes/sessions/ 全部子目录）。
// 供 Handler 在删除版本记录时调用，确保文件系统与数据库一致。
func (l *RepoWikiLogic) CleanVersionData(ctx context.Context, versionID xSnowflake.SnowflakeID) *xError.Error {
	l.log.Info(ctx, fmt.Sprintf("CleanVersionData - 清理版本数据 [%d]", versionID.Int64()))
	return l.svc.storage.CleanVersion(versionID.Int64())
}

// TouchLastAccessed 更新配置的最后访问时间（用于活跃度统计）
func (l *RepoWikiLogic) TouchLastAccessed(ctx context.Context, configID xSnowflake.SnowflakeID) {
	config, xErr := l.repo.config.GetByID(ctx, configID)
	if xErr != nil {
		return
	}
	now := time.Now()
	config.LastAccessedAt = &now
	if xErr := l.repo.config.Update(ctx, config); xErr != nil {
		l.log.Warn(ctx, "TouchLastAccessed - 更新最后访问时间失败", slog.Any("configID", configID.Int64()), slog.Any("err", xErr))
	}
}
