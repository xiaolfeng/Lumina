package logic

import (
	"context"
	"time"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xCtxUtil "github.com/bamboo-services/bamboo-base-go/common/utility/context"
	xEnv "github.com/bamboo-services/bamboo-base-go/defined/env"
	"github.com/gin-gonic/gin"
	apiHealth "github.com/xiaolfeng/Lumina/api/health"
	"github.com/xiaolfeng/Lumina/internal/repository"
)

type healthRepo struct {
	health *repository.HealthRepo
}

type HealthLogic struct {
	logic
	repo healthRepo
}

func NewHealthLogic(ctx context.Context) *HealthLogic {
	db := xCtxUtil.MustGetDB(ctx)
	rdb := xCtxUtil.MustGetRDB(ctx)

	return &HealthLogic{
		logic: logic{
			db:  db,
			rdb: rdb,
			log: xLog.WithName(xLog.NamedLOGC, "HealthLogic"),
		},
		repo: healthRepo{
			health: repository.NewHealthRepo(db, rdb),
		},
	}
}

func (l *HealthLogic) Ping(ctx *gin.Context) (*apiHealth.PingResponse, *xError.Error) {
	l.log.Info(ctx, "Ping - 执行健康检查")

	databaseReady, xErr := l.repo.health.DatabaseReady(ctx.Request.Context())
	if xErr != nil {
		return nil, xErr
	}

	redisReady, xErr := l.repo.health.RedisReady(ctx.Request.Context())
	if xErr != nil {
		return nil, xErr
	}

	return &apiHealth.PingResponse{
		App:           xEnv.GetEnvString(xEnv.AppName, "bamboo-base-go-template"),
		Version:       xEnv.GetEnvString(xEnv.AppVersion, "v0.1.0"),
		DatabaseReady: databaseReady,
		RedisReady:    redisReady,
		Timestamp:     time.Now().UnixMilli(),
	}, nil
}
