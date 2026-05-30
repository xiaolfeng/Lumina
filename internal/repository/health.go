package repository

import (
	"context"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type HealthRepo struct {
	db  *gorm.DB
	rdb *redis.Client
	log *xLog.LogNamedLogger
}

func NewHealthRepo(db *gorm.DB, rdb *redis.Client) *HealthRepo {
	return &HealthRepo{
		db:  db,
		rdb: rdb,
		log: xLog.WithName(xLog.NamedREPO, "HealthRepo"),
	}
}

func (r *HealthRepo) DatabaseReady(ctx context.Context) (bool, *xError.Error) {
	r.log.Info(ctx, "DatabaseReady - 检查数据库连接")

	sqlDB, err := r.db.WithContext(ctx).DB()
	if err != nil {
		return false, xError.NewError(nil, xError.DatabaseError, "获取数据库连接失败", false, err)
	}

	if err = sqlDB.PingContext(ctx); err != nil {
		return false, xError.NewError(nil, xError.DatabaseError, "数据库健康检查失败", false, err)
	}

	return true, nil
}

func (r *HealthRepo) RedisReady(ctx context.Context) (bool, *xError.Error) {
	r.log.Info(ctx, "RedisReady - 检查 Redis 连接")

	if err := r.rdb.Ping(ctx).Err(); err != nil {
		return false, xError.NewError(nil, xError.CacheError, "缓存健康检查失败", false, err)
	}

	return true, nil
}
