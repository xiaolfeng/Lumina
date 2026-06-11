package prepare

import (
	"log/slog"

	xCtxUtil "github.com/bamboo-services/bamboo-base-go/common/utility/context"
	xEnv "github.com/bamboo-services/bamboo-base-go/defined/env"
)

func (p *Prepare) prepareProject() {
	rdb := xCtxUtil.MustGetRDB(p.ctx)
	prefix := xEnv.GetEnvString(xEnv.NoSqlPrefix, "lum:")
	pattern := prefix + "project:*"
	var cursor uint64
	var deletedCount int
	for {
		keys, nextCursor, err := rdb.Scan(p.ctx, cursor, pattern, 100).Result()
		if err != nil {
			p.log.Warn(p.ctx, "扫描项目缓存键失败: "+err.Error())
			return
		}
		if len(keys) > 0 {
			deleted, delErr := rdb.Del(p.ctx, keys...).Result()
			if delErr != nil {
				p.log.Warn(p.ctx, "删除项目缓存键失败: "+delErr.Error())
			} else {
				deletedCount += int(deleted)
			}
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	if deletedCount > 0 {
		p.log.Info(p.ctx, "已清理项目缓存（字段类型变更）", slog.Int("count", deletedCount))
	}
}
