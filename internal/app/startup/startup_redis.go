package startup

import (
	"context"

	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xEnv "github.com/bamboo-services/bamboo-base-go/defined/env"
	"github.com/redis/go-redis/v9"
)

func (r *reg) nosqlInit(ctx context.Context) (any, error) {
	log := xLog.WithName(xLog.NamedINIT)
	log.Debug(ctx, "正在连接缓存...")

	rdb := redis.NewClient(&redis.Options{
		Addr:     xEnv.GetEnvString(xEnv.NoSqlHost, "localhost") + ":" + xEnv.GetEnvString(xEnv.NoSqlPort, "6379"),
		Password: xEnv.GetEnvString(xEnv.NoSqlPass, ""),
		DB:       xEnv.GetEnvInt(xEnv.NoSqlDatabase, 0),
		PoolSize: xEnv.GetEnvInt(xEnv.NoSqlPoolSize, 10),
	})

	log.Info(ctx, "缓存连接成功")
	return rdb, nil
}
