package startup

import (
	"context"

	xCtx "github.com/bamboo-services/bamboo-base-go/defined/context"
	xRegNode "github.com/bamboo-services/bamboo-base-go/major/register/node"

	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

type reg struct {
	ctx context.Context
}

func newInit() *reg {
	return &reg{ctx: context.Background()}
}

func Init() (context.Context, []xRegNode.RegNodeList) {
	businessReg := newInit()
	regNode := []xRegNode.RegNodeList{
		// Database / Cache 已由 main.go 的 xOption.WithDatabase / WithCache 声明式装配
		// （框架在业务节点之前注册 DatabaseKey / CacheManagerKey / RedisClientKey），
		// 此处仅注册业务节点（依赖 db/rdb 的 RepoWiki / MCP / 种子数据）。
		{Key: bConst.RepoWikiLogicKey, Node: businessReg.repoWikiInit},
		{Key: MCPHandlerKey, Node: businessReg.mcpInit},
		{Key: xCtx.Exec, Node: businessReg.businessDataPrepare},
	}

	return businessReg.ctx, regNode
}
