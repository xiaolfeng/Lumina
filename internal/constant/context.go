package bConst

import xCtx "github.com/bamboo-services/bamboo-base-go/defined/context"

const (
	CtxOwnerKey      xCtx.ContextKey = "business_owner"  // CtxOwnerKey 认证标记上下文键（单用户模式）
	RepoWikiLogicKey xCtx.ContextKey = "repo_wiki_logic" // RepoWikiLogicKey RepoWikiLogic 实例上下文键
)
