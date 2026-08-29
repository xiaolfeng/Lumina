package repository

import (
	"context"
	"errors"
	"fmt"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"gorm.io/gorm"
)

// OAuthClientRepo OAuth 动态注册客户端持久化层（RFC 7591）。
//
// 客户端注册数量少、写入后基本只读，不做缓存。
type OAuthClientRepo struct {
	db  *gorm.DB
	log *xLog.LogNamedLogger
}

// NewOAuthClientRepo 创建 OAuthClientRepo 实例
func NewOAuthClientRepo(db *gorm.DB) *OAuthClientRepo {
	return &OAuthClientRepo{
		db:  db,
		log: xLog.WithName(xLog.NamedREPO, "OAuthClientRepo"),
	}
}

// Create 创建动态注册客户端，ID 由雪花算法生成并作为 client_id
func (r *OAuthClientRepo) Create(ctx context.Context, client *entity.OAuthClient) (*entity.OAuthClient, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("Create - 注册 OAuth 客户端 [name=%s]", client.Name))

	if err := r.db.WithContext(ctx).Create(client).Error; err != nil {
		r.log.Warn(ctx, err.Error())
		return nil, xError.NewError(ctx, xError.DatabaseError, "注册 OAuth 客户端失败", false, err)
	}
	return client, nil
}

// GetByID 按客户端 ID（雪花 ID）查询注册客户端
//
// 返回值:
//   - *entity.OAuthClient: 查询到的客户端实体
//   - bool:                是否找到
//   - *xError.Error:       查询过程中的错误
func (r *OAuthClientRepo) GetByID(ctx context.Context, id xSnowflake.SnowflakeID) (*entity.OAuthClient, bool, *xError.Error) {
	var client entity.OAuthClient
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&client).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		r.log.Warn(ctx, err.Error())
		return nil, false, xError.NewError(ctx, xError.DatabaseError, "查询 OAuth 客户端失败", false, err)
	}
	return &client, true, nil
}
