// Package handler Wiki 阅读器 HTTP 处理器。
//
// WikiReaderHandler 是面向公开用户的 Wiki 内容读取入口，与 RepoWikiHandler
// （管理端 CRUD，Bearer Token 认证）不同：
//   - 认证方式：HMAC Cookie（POST /wiki/:id/auth 获取，非 Bearer Token）
//   - 无密码 Wiki：完全公开，无需任何认证
//   - 有密码 Wiki：Cookie 校验通过后可读页面和清单
//
// 安全要点：
//   - GetWikiPage 对用户提供的 page path 执行双重路径遍历防护
//   - 不暴露文件系统原始路径，仅返回 Markdown 内容

package handler

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xResult "github.com/bamboo-services/bamboo-base-go/major/result"
	"github.com/gin-gonic/gin"

	apiCommon "github.com/xiaolfeng/Lumina/api/common"
	apiRepowiki "github.com/xiaolfeng/Lumina/api/repowiki"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/logic"
	wikiService "github.com/xiaolfeng/Lumina/internal/service"
)

// 确保 apiCommon 包被编译器识别（swag 注释依赖此导入）
var _ = apiCommon.BaseResponse{}

// ──────────────────────────────────────────────────────────────────────
// WikiReaderHandler
//
// 独立于 NewHandler[T] 模式的专用阅读器处理器。
// 持有 RepoWikiLogic（配置/密码查询）、WikiAuthTokenService（Cookie 签名）
// 和 WikiStorageService（文件 I/O），由路由注册层通过 NewWikiReaderHandler 构造。
// ──────────────────────────────────────────────────────────────────────

// WikiReaderHandler Wiki 公开阅读器处理器
type WikiReaderHandler struct {
	name             string
	log              *xLog.LogNamedLogger
	logic            *logic.RepoWikiLogic
	authTokenService *wikiService.WikiAuthTokenService
	storage          *wikiService.WikiStorageService
}

// NewWikiReaderHandler 创建 WikiReaderHandler 实例
//
// 从 context 获取依赖，构造 Logic 和 Service 实例。
// RepoWikiLogic 构造时 LLM Provider 初始化失败不阻塞（阅读器不需要分析功能）。
// ctx 必须包含 DB/Redis 注入（由启动阶段注册到 context 的基础设施）。
func NewWikiReaderHandler(ctx context.Context) *WikiReaderHandler {
	return &WikiReaderHandler{
		name:             "WikiReaderHandler",
		log:              xLog.WithName(xLog.NamedCONT, "WikiReaderHandler"),
		logic:            logic.NewRepoWikiLogic(ctx),
		authTokenService: wikiService.NewWikiAuthTokenService(),
		storage:          wikiService.NewWikiStorageService(),
	}
}

// ──────────────────────────────────────────────────────────────────────
// 1. GetWikiPage — 读取 Wiki 页面 Markdown
// ──────────────────────────────────────────────────────────────────────

// GetWikiPage 获取 Wiki 页面内容
//
// GET /api/v1/wiki/:id/page/*path
//
// 路径遍历防护双层校验：
//  1. filepath.Clean + ".." 前缀检测 → 阻止显式上跳
//  2. filepath.Rel 包含校验 → 确保 fullPath 在 wikiPath 目录内
//
// @Summary     [公开] 获取 Wiki 页面
// @Description 根据 Wiki ID 和页面路径读取 Markdown 内容，受密码保护的 Wiki 需携带有效的 HMAC Cookie
// @Tags        Wiki阅读器接口
// @Produce     json
// @Param       id    path  string  true  "Wiki 配置 ID"
// @Param       path  path  string  true  "页面路径（如 content/项目概览.md）"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiRepowiki.WikiPageResponse}  "获取成功"
// @Failure     400  {object}  apiCommon.BaseResponse  "无效的 Wiki ID 或页面路径"
// @Failure     401  {object}  apiCommon.BaseResponse  "Wiki 需要认证（Cookie 缺失或无效）"
// @Failure     404  {object}  apiCommon.BaseResponse  "页面文件不存在"
// @Router      /api/v1/wiki/{id}/page/{path} [GET]
func (h *WikiReaderHandler) GetWikiPage(ctx *gin.Context) {
	h.log.Info(ctx, fmt.Sprintf("GetWikiPage - path=[%s%s]", ctx.Param("id"), ctx.Param("path")))

	configID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		xResult.AbortError(ctx, xError.BadRequest, "无效的 Wiki ID", false)
		return
	}

	pagePath := ctx.Param("path") // catch-all 参数，以 "/" 开头

	// 获取 Wiki 文档目录
	wikiPath := h.storage.GetWikiPath(configID)

	// 路径遍历防护：双层校验
	safePath, err := sanitizeWikiPath(wikiPath, pagePath)
	if err != nil {
		h.log.Info(ctx, fmt.Sprintf("GetWikiPage - 路径遍历拦截 [%s]", pagePath))
		xResult.AbortError(ctx, xError.BadRequest, "无效的页面路径", false)
		return
	}

	// 读取 Markdown 文件
	content, xErr := h.storage.ReadMarkdown(safePath)
	if xErr != nil {
		xResult.AbortError(ctx, xError.FileNotFound, "Wiki 页面不存在", false)
		return
	}

	// 构建响应
	resp := apiRepowiki.WikiPageResponse{
		Title:    extractTitleFromPath(pagePath),
		Content:  content,
		Path:     strings.TrimPrefix(pagePath, "/"),
		Language: bConst.RepoWikiDefaultLanguage,
	}

	xResult.SuccessHasData(ctx, "获取成功", resp)
}

// ──────────────────────────────────────────────────────────────────────
// 2. GetWikiManifest — 读取 Wiki 导航清单
// ──────────────────────────────────────────────────────────────────────

// GetWikiManifest 获取 Wiki 导航清单
//
// GET /api/v1/wiki/:id/manifest
//
// 读取 {wikiPath}/meta/repowiki-metadata.json 并返回导航结构。
//
// @Summary     [公开] 获取 Wiki 导航清单
// @Description 读取 Wiki 元数据清单（含侧边栏导航、首页路径、项目名称），受密码保护的 Wiki 需 Cookie 认证
// @Tags        Wiki阅读器接口
// @Produce     json
// @Param       id  path  string  true  "Wiki 配置 ID"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiRepowiki.WikiManifestResponse}  "获取成功"
// @Failure     400  {object}  apiCommon.BaseResponse  "无效的 Wiki ID"
// @Failure     401  {object}  apiCommon.BaseResponse  "Wiki 需要认证"
// @Failure     404  {object}  apiCommon.BaseResponse  "清单文件不存在"
// @Router      /api/v1/wiki/{id}/manifest [GET]
func (h *WikiReaderHandler) GetWikiManifest(ctx *gin.Context) {
	h.log.Info(ctx, fmt.Sprintf("GetWikiManifest - id=[%s]", ctx.Param("id")))

	configID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		xResult.AbortError(ctx, xError.BadRequest, "无效的 Wiki ID", false)
		return
	}

	manifestPath := h.storage.GetManifestPath(configID)

	// 直接反序列化到响应 DTO（JSON 字段名与 DTO tag 对齐）
	var manifest apiRepowiki.WikiManifestResponse
	if xErr := h.storage.ReadJSON(manifestPath, &manifest); xErr != nil {
		xResult.AbortError(ctx, xError.FileNotFound, "Wiki 清单不存在，可能文档尚未生成", false)
		return
	}

	xResult.SuccessHasData(ctx, "获取成功", manifest)
}

// ──────────────────────────────────────────────────────────────────────
// 3. WikiAuth — Wiki 密码验证 + 设置 Cookie
// ──────────────────────────────────────────────────────────────────────

// WikiAuth Wiki 密码验证
//
// POST /api/v1/wiki/:id/auth
//
// 业务流程：
//  1. 查询配置 → 检查是否设置了密码
//  2. 未设密码 → 直接返回成功（无需认证）
//  3. 已设密码 → bcrypt 校验 → 正确则生成 HMAC Token 并 Set-Cookie
//  4. 密码错误 → 401
//
// Cookie 配置：
//   - name: `wiki_auth_{configID}`
//   - maxAge: bConst.RepoWikiCookieMaxAge（2 小时）
//   - httpOnly: true
//   - path: "/"
//
// @Summary     [公开] Wiki 密码验证
// @Description 提交 Wiki 访问密码，验证通过后设置 HMAC 签名 Cookie（有效期 2 小时）
// @Tags        Wiki阅读器接口
// @Accept      json
// @Produce     json
// @Param       id       path  string                     true  "Wiki 配置 ID"
// @Param       request  body  apiRepowiki.WikiAuthRequest  true  "Wiki 密码验证请求"
// @Success     200  {object}  apiCommon.BaseResponse  "验证成功"
// @Failure     400  {object}  apiCommon.BaseResponse  "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse  "密码错误"
// @Failure     404  {object}  apiCommon.BaseResponse  "Wiki 配置不存在"
// @Router      /api/v1/wiki/{id}/auth [POST]
func (h *WikiReaderHandler) WikiAuth(ctx *gin.Context) {
	h.log.Info(ctx, fmt.Sprintf("WikiAuth - id=[%s]", ctx.Param("id")))

	configID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		xResult.AbortError(ctx, xError.BadRequest, "无效的 Wiki ID", false)
		return
	}

	var req apiRepowiki.WikiAuthRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(err)
		return
	}

	// 获取配置
	config, xErr := h.logic.GetConfig(ctx.Request.Context(), xSnowflake.SnowflakeID(configID))
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	// 未设置密码 → 无需认证
	if !h.logic.HasWikiPassword(config) {
		xResult.Success(ctx, "该 Wiki 无需密码认证")
		return
	}

	// 校验密码
	if !h.logic.VerifyWikiPassword(config, req.Password) {
		xResult.AbortError(ctx, xError.Unauthorized, "Wiki 密码错误", false)
		return
	}

	// 生成 HMAC Token 并设置 Cookie
	token, err := h.logic.GenerateWikiToken(configID, req.Password)
	if err != nil {
		xResult.AbortError(ctx, xError.ServerInternalError, "Token 生成失败", false)
		return
	}

	ctx.SetCookie(
		fmt.Sprintf("wiki_auth_%d", configID), // name
		token,                                 // value
		bConst.RepoWikiCookieMaxAge,           // maxAge (7200s = 2h)
		"/",                                   // path
		"",                                    // domain
		false,                                 // secure（生产环境建议 true）
		true,                                  // httpOnly
	)

	xResult.Success(ctx, "Wiki 认证成功")
}

// ──────────────────────────────────────────────────────────────────────
// 4. CheckWikiAuth — 检查当前认证状态
// ──────────────────────────────────────────────────────────────────────

// CheckWikiAuth 检查 Wiki 认证状态
//
// GET /api/v1/wiki/:id/auth-check
//
// 返回当前请求的认证状态，供前端判断是否需要显示密码输入框：
//   - password_required=false → 公开 Wiki，无需认证
//   - password_required=true + authenticated=true → Cookie 有效
//   - password_required=true + authenticated=false → 需要输入密码
//
// @Summary     [公开] 检查 Wiki 认证状态
// @Description 检查 Wiki 是否需要密码保护以及当前 Cookie 是否有效
// @Tags        Wiki阅读器接口
// @Produce     json
// @Param       id  path  string  true  "Wiki 配置 ID"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiRepowiki.WikiAuthCheckResponse}  "认证状态"
// @Failure     400  {object}  apiCommon.BaseResponse  "无效的 Wiki ID"
// @Failure     404  {object}  apiCommon.BaseResponse  "Wiki 配置不存在"
// @Router      /api/v1/wiki/{id}/auth-check [GET]
func (h *WikiReaderHandler) CheckWikiAuth(ctx *gin.Context) {
	h.log.Info(ctx, fmt.Sprintf("CheckWikiAuth - id=[%s]", ctx.Param("id")))

	configID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		xResult.AbortError(ctx, xError.BadRequest, "无效的 Wiki ID", false)
		return
	}

	// 获取配置
	config, xErr := h.logic.GetConfig(ctx.Request.Context(), xSnowflake.SnowflakeID(configID))
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	passwordRequired := h.logic.HasWikiPassword(config)
	authenticated := false

	if !passwordRequired {
		// 公开 Wiki → 始终"已认证"
		authenticated = true
	} else {
		// 受保护 Wiki → 检查 Cookie
		cookieName := fmt.Sprintf("wiki_auth_%d", configID)
		cookieValue, cookieErr := ctx.Cookie(cookieName)
		if cookieErr == nil {
			authenticated = h.authTokenService.ValidateToken(cookieValue, configID)
		}
	}

	xResult.SuccessHasData(ctx, "OK", apiRepowiki.WikiAuthCheckResponse{
		Authenticated:    authenticated,
		PasswordRequired: passwordRequired,
	})
}

// GetConfigPasswordHash 获取 Wiki 配置的密码哈希（供 WikiAuth 中间件回调，空串=无密码）
func (h *WikiReaderHandler) GetConfigPasswordHash(ctx context.Context, wikiID int64) (string, error) {
	config, xErr := h.logic.GetConfig(ctx, xSnowflake.SnowflakeID(wikiID))
	if xErr != nil {
		return "", xErr
	}
	return config.WikiPasswordHash, nil
}

// ──────────────────────────────────────────────────────────────────────
// 路径安全辅助函数
// ──────────────────────────────────────────────────────────────────────

// sanitizeWikiPath 对用户提供的 Wiki 页面路径进行安全校验
//
// 双层防护：
//  1. filepath.Clean 归一化 + ".." 前缀检测 → 阻止显式上跳
//  2. filepath.Rel 包含校验 → 确保 fullPath 在 wikiPath 目录内
//
// 参数:
//   - wikiPath: Wiki 文档根目录（如 {basePath}/wiki/{configID}/{language}/）
//   - userPath: 用户请求的页面路径（catch-all 参数，以 "/" 开头）
//
// 返回值:
//   - string: 安全的绝对文件路径
//   - error:  路径包含遍历或超出 wikiPath 范围
func sanitizeWikiPath(wikiPath, userPath string) (string, error) {
	// 去掉前导 "/"，Clean 归一化
	cleaned := filepath.Clean(strings.TrimPrefix(userPath, "/"))

	// 检查显式上跳（".." 或 "../..."）
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path traversal detected: %s", userPath)
	}

	// 拼接到 wiki 目录
	fullPath := filepath.Join(wikiPath, cleaned)

	// 包含校验：确认 fullPath 在 wikiPath 之下
	rel, err := filepath.Rel(wikiPath, fullPath)
	if err != nil {
		return "", fmt.Errorf("path escape detected: %s", userPath)
	}
	if strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("path escapes wiki directory: %s", userPath)
	}

	return fullPath, nil
}

// extractTitleFromPath 从页面路径中提取显示标题
//
// 规则：取最后一段文件名，去掉扩展名。
// 例："/content/项目概览.md" → "项目概览"
func extractTitleFromPath(path string) string {
	cleaned := strings.TrimPrefix(path, "/")
	base := filepath.Base(cleaned)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}
