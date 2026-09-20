package bConst

import "regexp"

// Page 状态常量
const (
	PageStatusPublished = "published" // 已发布
	PageStatusArchived  = "archived"  // 已归档
)

// Page 访问权限策略常量
const (
	PageAccessModePublic   = "public"   // 公开访问
	PageAccessModePassword = "password" // 密码保护
)

// Pages 文件与 Cookie 约束
const (
	// PagesFileMaxSize 页面快照单文件大小上限，与 Preview 保持一致（256KB）
	PagesFileMaxSize = PreviewFileMaxSize
	// PagesCookieMaxAge 密码门 Cookie 默认有效期（秒），24 小时
	PagesCookieMaxAge = 24 * 60 * 60
	// PagesDefaultEntryFilename 缺省入口文件名
	PagesDefaultEntryFilename = "index.html"
	// PagesInitialVersion 首次晋升默认版本号
	PagesInitialVersion = "v1.0.0"
	// PagesCookieNamePrefix 密码门 Cookie 名前缀，完整名为 lum_pages_auth_<pageID>
	PagesCookieNamePrefix = "lum_pages_auth_"
)

// PageSlugPattern 项目内 Slug 合法正则（小写字母、数字、短横线）
var PageSlugPattern = regexp.MustCompile(`^[a-z0-9-]+$`)
