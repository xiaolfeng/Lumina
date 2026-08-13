package bConst

// ── 系统设置分类常量 ──
const (
	SettingCategorySite     = "site"     // 站点外观设置分类
	SettingCategoryQa       = "qa"       // Q&A 模块设置分类
	SettingCategoryRepoWiki = "repowiki" // RepoWiki 模块设置分类
	SettingCategorySecurity = "security" // 安全认证设置分类
)

// SettingKeyDef 定义单个配置项的元数据
type SettingKeyDef struct {
	Key         string // 配置键名
	Category    string // 所属分类
	Type        string // 值类型：string / int / bool
	Default     string // 默认值字符串表示
	Description string // 配置说明
}

// SettingKeyDefs 系统设置全量元数据定义（共 17 项）
//
// 覆盖站点信息、Q&A 模块、RepoWiki 模块、安全认证四大分类。
// 新增设置项时只需在此追加定义并在 KeysByCategory 中自动归类。
// 键名规范：层级用 . 分隔，同层多词用 - 连接，禁止使用 _。
var SettingKeyDefs = []SettingKeyDef{
	// ── site 分类（5 项）──
	{Key: InfoKeySiteName, Category: SettingCategorySite, Type: "string", Default: "Lumina", Description: "站点名称"},
	{Key: InfoKeySiteDescription, Category: SettingCategorySite, Type: "string", Default: "赋予 AI 深度代码认知与长期记忆的知识中枢", Description: "站点描述"},
	{Key: InfoKeySiteLogoURL, Category: SettingCategorySite, Type: "string", Default: "", Description: "站点 Logo URL 地址"},
	{Key: InfoKeySiteDomain, Category: SettingCategorySite, Type: "string", Default: "", Description: "对外访问域名（生成交互/下载链接，部署时必须配置）"},
	{Key: InfoKeySiteFooterText, Category: SettingCategorySite, Type: "string", Default: "Lumina · 微明", Description: "站点页脚文本"},

	// ── qa 分类（5 项）──
	{Key: InfoKeyQaSessionTTL, Category: SettingCategoryQa, Type: "int", Default: "172800", Description: "Session默认TTL（秒）"},
	{Key: InfoKeyQaGetAnswerPollSlice, Category: SettingCategoryQa, Type: "int", Default: "25", Description: "qa_get_answer 单次阻塞上限（秒），到点无回答返回 PENDING 引导重试"},
	{Key: InfoKeyQaGetAnswerMaxRetries, Category: SettingCategoryQa, Type: "int", Default: "36", Description: "qa_get_answer 最大重试次数（默认36次≈15分钟），达到后返回 STOPPED 提示用户主动触发"},
	{Key: InfoKeyQaMaxActiveSessions, Category: SettingCategoryQa, Type: "int", Default: "100", Description: "Q&A 最大活跃会话数"},
	{Key: InfoKeyQaEnableFileUpload, Category: SettingCategoryQa, Type: "bool", Default: "true", Description: "是否启用 Q&A 文件上传功能"},

	// ── repowiki 分类（3 项）──
	{Key: InfoKeyRepoWikiDefaultLanguage, Category: SettingCategoryRepoWiki, Type: "string", Default: "zh", Description: "RepoWiki 默认 Wiki 语言"},
	{Key: InfoKeyRepoWikiDefaultBranch, Category: SettingCategoryRepoWiki, Type: "string", Default: "main", Description: "RepoWiki 默认 Git 分支"},
	{Key: InfoKeyRepoWikiCookieMaxAge, Category: SettingCategoryRepoWiki, Type: "int", Default: "7200", Description: "Wiki Cookie 最大有效期（秒）"},

	// ── security 分类（4 项）──
	{Key: InfoKeySecurityAccessTokenTTL, Category: SettingCategorySecurity, Type: "int", Default: "3600", Description: "访问令牌有效期（秒）"},
	{Key: InfoKeySecurityRefreshTokenTTL, Category: SettingCategorySecurity, Type: "int", Default: "604800", Description: "刷新令牌有效期（秒）"},
	{Key: InfoKeySecurityMaxAPIKeys, Category: SettingCategorySecurity, Type: "int", Default: "10", Description: "单个用户最大 API Key 数量"},
	{Key: InfoKeySecurityWebAuthnTimeout, Category: SettingCategorySecurity, Type: "int", Default: "300000", Description: "WebAuthn 操作超时时间（毫秒）"},
}

// KeysByCategory 按分类索引的设置项元数据
//
// 在 init() 中从 SettingKeyDefs 自动构建，外部逻辑可按分类快速检索。
var KeysByCategory map[string][]SettingKeyDef

func init() {
	KeysByCategory = make(map[string][]SettingKeyDef, 4) // 4 个分类 // 4 个分类
	for _, def := range SettingKeyDefs {
		KeysByCategory[def.Category] = append(KeysByCategory[def.Category], def)
	}
}
