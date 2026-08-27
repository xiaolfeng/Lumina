package bConst

import xCtx "github.com/bamboo-services/bamboo-base-go/defined/context"

const (
	CtxOwnerKey      xCtx.ContextKey = "business_owner"  // CtxOwnerKey 认证标记上下文键（单用户模式）
	RepoWikiLogicKey xCtx.ContextKey = "repo_wiki_logic" // RepoWikiLogicKey RepoWikiLogic 实例上下文键

	// WebAuthnOriginContextKey 浏览器 Origin 上下文键（WebAuthn RP 推导用）。
	// 刻意使用原生 string 而非 xCtx.ContextKey：gin 引擎未开启
	// ContextWithFallback 时 Context.Value 不回退 Request.Context()，且其 Keys
	// 查找仅命中动态类型为 string 的键；原生 string 可同时兼容 c.Set 与
	// context.WithValue 两条写入路径（见 middleware.WebAuthnOrigin）。
	WebAuthnOriginContextKey = "webauthn_origin"
)
