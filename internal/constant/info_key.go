package bConst

// ── Info 表配置键常量 ──
//
// 所有 Info 表键名统一在此定义，禁止在业务代码中写死键名字符串。
// 键名规范：层级用 . 分隔，同层多词用 - 连接，禁止使用 _。
const (
	// ── 站点信息（site）──
	InfoKeySiteName        = "site.name"        // 站点名称
	InfoKeySiteDescription = "site.description" // 站点描述
	InfoKeySiteLogoURL     = "site.logo-url"    // 站点 Logo URL 地址
	InfoKeySiteDomain      = "site.domain"      // 对外访问域名（生成交互/下载链接）
	InfoKeySiteFooterText  = "site.footer-text" // 站点页脚文本

	// ── Q&A 配置（qa）──
	InfoKeyQaSessionTTL          = "qa.session.ttl"            // Session 默认 TTL（秒）
	InfoKeyQaGetAnswerPollSlice  = "qa.get-answer.poll-slice"  // qa_get_answer 单次阻塞上限（秒）
	InfoKeyQaGetAnswerMaxRetries = "qa.get-answer.max-retries" // qa_get_answer 最大重试次数
	InfoKeyQaMaxActiveSessions   = "qa.max-active-sessions"    // Q&A 最大活跃会话数
	InfoKeyQaEnableFileUpload    = "qa.enable-file-upload"     // 是否启用 Q&A 文件上传

	// ── RepoWiki 配置（repowiki）──
	InfoKeyRepoWikiDefaultLanguage = "repowiki.default-language"    // RepoWiki 默认 Wiki 语言
	InfoKeyRepoWikiDefaultBranch   = "repowiki.default-branch"      // RepoWiki 默认 Git 分支
	InfoKeyRepoWikiCookieMaxAge    = "repowiki.wiki-cookie-max-age" // Wiki Cookie 最大有效期（秒）

	// ── 安全认证配置（security）──
	InfoKeySecurityAccessTokenTTL  = "security.access-token-ttl"  // 访问令牌有效期（秒）
	InfoKeySecurityRefreshTokenTTL = "security.refresh-token-ttl" // 刷新令牌有效期（秒）
	InfoKeySecurityMaxAPIKeys      = "security.max-api-keys"      // 单个用户最大 API Key 数量
	InfoKeySecurityWebAuthnTimeout = "security.webauthn-timeout"  // WebAuthn 操作超时时间（毫秒）

	// ── 认证内部状态（owner/auth，不对外暴露为设置项）──
	InfoKeyOwnerUsername         = "owner.username"          // 站主用户名
	InfoKeyOwnerEmail            = "owner.email"             // 站主邮箱
	InfoKeyOwnerPassword         = "owner.password"          // 站主密码（加密存储）
	InfoKeyOwnerBiometricEnabled = "owner.biometric-enabled" // 站主生物特征是否启用
	InfoKeyAuthIsInitial         = "auth.is-initial"         // 系统是否已初始化（true=未初始化）
)

// ── 认证环境变量键名 ──
const (
	EnvInitializeToken = "XLF_INITIALIZE_TOKEN" // 一次性初始化令牌（非空时初始化必须携带匹配值）
)
