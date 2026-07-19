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
// 路径遍历防护双层校验：
//  1. filepath.Clean + ".." 前缀检测 → 阻止显式上跳
//  2. filepath.Rel 包含校验 → 确保 fullPath 在 wikiPath 目录内
//
// BREAKING: 仅读取 .mdx 文件，不再支持 .md 回退。page path 来自 manifest（无扩展名），
// 磁盘文件为 {wikiPath}/{path}.mdx。页面 YAML frontmatter 中的 title/description/icon
// 会被提取到响应 DTO；title 缺失时回退到路径 basename。
//
// @Summary     [公开] 获取 Wiki 页面
// @Description 根据 Wiki ID 和页面路径读取 .mdx 页面内容（含 YAML frontmatter 解析），受密码保护的 Wiki 需携带有效的 HMAC Cookie
// @Tags        Wiki阅读器接口
// @Produce     json
// @Param       id    path  string  true  "Wiki 配置 ID"
// @Param       path  path  string  true  "页面路径（无扩展名，如 content/项目概览）"
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

	// 经 config.SelectedVersionID 间接定位版本目录
	versionID, ok := h.resolveSelectedVersion(ctx, configID)
	if !ok {
		return // resolveSelectedVersion 已写入响应
	}

	pagePath := ctx.Param("path") // catch-all 参数，以 "/" 开头

	// 获取 Wiki 文档目录
	wikiPath := h.storage.GetWikiPath(versionID)

	resp, xErr := buildWikiPageResponse(h.storage, wikiPath, pagePath)
	if xErr != nil {
		h.log.Info(ctx, fmt.Sprintf("GetWikiPage - 读取页面失败 [%s] %s", pagePath, xErr.Error()))
		xResult.AbortError(ctx, xError.FileNotFound, "Wiki 页面不存在", false)
		return
	}

	// 读取 manifest 计算 prev/next/breadcrumb（失败不阻塞页面返回）
	manifestPath := h.storage.GetManifestPath(versionID)
	var manifest apiRepowiki.WikiManifestResponse
	if xErr := h.storage.ReadJSON(manifestPath, &manifest); xErr != nil {
		h.log.Info(ctx, fmt.Sprintf("GetWikiPage - 读取 manifest 失败，prev/next/breadcrumb 置空: %s", xErr.Error()))
	} else {
		currentPagePath := strings.TrimPrefix(pagePath, "/")
		prev, next, breadcrumb := computeNav(&manifest, currentPagePath)
		resp.Prev = prev
		resp.Next = next
		resp.Breadcrumb = breadcrumb
	}

	xResult.SuccessHasData(ctx, "获取成功", resp)
}

// buildWikiPageResponse 读取 .mdx 页面并构建 WikiPageResponse
//
// 流程：
//  1. sanitizeWikiPath 双层路径遍历防护
//  2. 强制 .mdx 扩展名（BREAKING：无 .md 回退）
//  3. storage.ReadPage 解析 YAML frontmatter
//  4. 提取 title/description/icon；title 缺失时回退到 extractTitleFromPath
//
// 失败时返回 *xError.Error（FileNotFound 表示文件不存在，BadRequest 表示路径非法）。
func buildWikiPageResponse(storage *wikiService.WikiStorageService, wikiPath, pagePath string) (apiRepowiki.WikiPageResponse, *xError.Error) {
	// 路径遍历防护：双层校验
	safePath, err := sanitizeWikiPath(wikiPath, pagePath)
	if err != nil {
		return apiRepowiki.WikiPageResponse{}, xError.NewError(context.Background(), xError.BadRequest,
			xError.ErrMessage("无效的页面路径"), false, err)
	}

	// BREAKING: 仅读取 .mdx，manifest path 无扩展名，磁盘文件为 {path}.mdx
	if !strings.HasSuffix(safePath, ".mdx") {
		safePath = safePath + ".mdx"
	}

	// 读取并解析 .mdx 页面（含 frontmatter）
	page, xErr := storage.ReadPage(safePath)
	if xErr != nil {
		return apiRepowiki.WikiPageResponse{}, xErr
	}

	// 提取 frontmatter 元信息
	title := frontmatterString(page.Frontmatter, "title")
	description := frontmatterString(page.Frontmatter, "description")
	icon := frontmatterString(page.Frontmatter, "icon")

	// title 回退：frontmatter.title → extractTitleFromPath（basename 去扩展名）
	if title == "" {
		title = extractTitleFromPath(pagePath)
	}

	return apiRepowiki.WikiPageResponse{
		Title:       title,
		Content:     page.Body,
		Path:        strings.TrimPrefix(pagePath, "/"),
		Language:    bConst.RepoWikiDefaultLanguage,
		Description: description,
		Icon:        icon,
		LastUpdated: page.ModTime.Unix(),
	}, nil
}

// frontmatterString 从 YAML frontmatter 中安全提取字符串字段
//
// yaml.v3 解析到 map[string]interface{} 时，标量值类型可能为 string/bool/int/float64。
// 此函数对非字符串值使用 fmt.Sprintf 兜底，nil 或缺失键返回空串。
func frontmatterString(fm map[string]interface{}, key string) string {
	if fm == nil {
		return ""
	}
	v, ok := fm[key]
	if !ok || v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
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

	// 经 config.SelectedVersionID 间接定位版本目录
	versionID, ok := h.resolveSelectedVersion(ctx, configID)
	if !ok {
		return
	}

	manifestPath := h.storage.GetManifestPath(versionID)

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
	if !BindJSON(ctx, &req) {
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

// resolveSelectedVersion 经 config.SelectedVersionID 解析当前 Wiki 版本目录 ID。
//
// WikiReader 通过配置选中的版本间接定位 Wiki 文件，而非直接以 configID 作为目录。
// 失败时已写入 HTTP 响应（404），调用方仅需 return。
func (h *WikiReaderHandler) resolveSelectedVersion(ctx *gin.Context, configID int64) (int64, bool) {
	config, xErr := h.logic.GetConfig(ctx.Request.Context(), xSnowflake.SnowflakeID(configID))
	if xErr != nil {
		_ = ctx.Error(xErr)
		return 0, false
	}
	if config.SelectedVersionID == nil {
		xResult.AbortError(ctx, xError.NotFound, "Wiki 尚未生成或未选择版本", false)
		return 0, false
	}
	return config.SelectedVersionID.Int64(), true
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

// ──────────────────────────────────────────────────────────────────────
// 导航计算（prev / next / breadcrumb）
// ──────────────────────────────────────────────────────────────────────

// computeNav 根据 manifest 导航树计算当前页面的 prev/next/breadcrumb。
//
// 参数:
//   - manifest: Wiki 清单（含 Navigation 树），nil 时返回全 nil
//   - currentPagePath: 当前页面路径（无扩展名，与 manifest 中叶子 Path 字段对齐）
//
// 返回:
//   - prev: DFS 叶子序列中当前页的前一个叶子（无则 nil）
//   - next: DFS 叶子序列中当前页的后一个叶子（无则 nil）
//   - breadcrumb: 从根到当前页的目录节点 + 当前页本身（WikiNavRef 列表）
//
// 规则:
//   - DFS 遍历 manifest.Navigation，按 manifest 顺序收集叶子节点（无 children 的节点）
//   - 跳过 Separator 节点（Separator 字段非空，Title/Path 为空）
//   - prev/next 通过 Path 精确匹配定位当前页在叶子序列中的位置
//   - breadcrumb: 从根遍历到当前页，收集所有目录节点（有 children 的节点）+ 当前页本身
//   - 当前页未在叶子序列中找到时返回全 nil（不阻塞页面返回）
func computeNav(manifest *apiRepowiki.WikiManifestResponse, currentPagePath string) (prev, next *apiRepowiki.WikiNavRef, breadcrumb []apiRepowiki.WikiNavRef) {
	if manifest == nil || currentPagePath == "" {
		return nil, nil, nil
	}

	// DFS 收集叶子节点（跳过 Separator）
	var leaves []apiRepowiki.WikiNavItem
	var walk func(items []apiRepowiki.WikiNavItem)
	walk = func(items []apiRepowiki.WikiNavItem) {
		for _, item := range items {
			if item.Separator != "" {
				continue
			}
			if len(item.Children) > 0 {
				walk(item.Children)
				continue
			}
			leaves = append(leaves, item)
		}
	}
	walk(manifest.Navigation)

	// 定位当前页在叶子序列中的位置
	idx := -1
	for i, leaf := range leaves {
		if leaf.Path == currentPagePath {
			idx = i
			break
		}
	}
	if idx == -1 {
		return nil, nil, nil
	}

	if idx > 0 {
		prev = navRefPtrFromItem(leaves[idx-1])
	}
	if idx < len(leaves)-1 {
		next = navRefPtrFromItem(leaves[idx+1])
	}
	breadcrumb = findBreadcrumb(manifest.Navigation, currentPagePath)
	return prev, next, breadcrumb
}

// findBreadcrumb 在导航树中查找从根到目标页面的路径，返回路径上的目录节点 + 当前页本身。
//
// 跳过 Separator 节点。目标未找到时返回 nil。
func findBreadcrumb(items []apiRepowiki.WikiNavItem, targetPath string) []apiRepowiki.WikiNavRef {
	var result []apiRepowiki.WikiNavRef
	var search func(nodes []apiRepowiki.WikiNavItem, ancestors []apiRepowiki.WikiNavRef) bool
	search = func(nodes []apiRepowiki.WikiNavItem, ancestors []apiRepowiki.WikiNavRef) bool {
		for _, node := range nodes {
			if node.Separator != "" {
				continue
			}
			if len(node.Children) > 0 {
				dirRef := navRefFromItem(node)
				if search(node.Children, append(ancestors, dirRef)) {
					return true
				}
				continue
			}
			if node.Path == targetPath {
				result = make([]apiRepowiki.WikiNavRef, 0, len(ancestors)+1)
				result = append(result, ancestors...)
				result = append(result, navRefFromItem(node))
				return true
			}
		}
		return false
	}
	if search(items, nil) {
		return result
	}
	return nil
}

// navRefFromItem 将 WikiNavItem 转换为 WikiNavRef（提取 Title/Path/Icon）
func navRefFromItem(item apiRepowiki.WikiNavItem) apiRepowiki.WikiNavRef {
	return apiRepowiki.WikiNavRef{
		Title: item.Title,
		Path:  item.Path,
		Icon:  item.Icon,
	}
}

// navRefPtrFromItem 将 WikiNavItem 转换为 *WikiNavRef（用于 prev/next 指针字段）
func navRefPtrFromItem(item apiRepowiki.WikiNavItem) *apiRepowiki.WikiNavRef {
	return &apiRepowiki.WikiNavRef{
		Title: item.Title,
		Path:  item.Path,
		Icon:  item.Icon,
	}
}
