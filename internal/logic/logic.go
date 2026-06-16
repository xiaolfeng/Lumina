package logic

import (
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
)

// logic 所有 Logic 结构体的公共基础，仅持有日志记录器
//
// 刻意不持有 db/rdb：业务编排层禁止直连数据源，所有持久化与缓存
// 读写必须经由 repository（+cache 子层），从结构上杜绝越界调用。
// 各 Logic 子结构体通过组合各自的 repo 字段获取数据访问能力。
type logic struct {
	log *xLog.LogNamedLogger
}
