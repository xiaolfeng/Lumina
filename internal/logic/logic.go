package logic

import (
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type logic struct {
	db  *gorm.DB
	rdb *redis.Client
	log *xLog.LogNamedLogger
}
