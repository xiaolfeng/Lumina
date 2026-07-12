package repository

import (
	"context"
	"errors"
	"fmt"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xCache "github.com/bamboo-services/bamboo-base-go/major/cache"
	"github.com/redis/go-redis/v9"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"github.com/xiaolfeng/Lumina/internal/repository/cache"
	"gorm.io/gorm"
)

// SshKeyRepo SSH 密钥数据访问层，提供完整 CRUD 操作与 Redis 缓存层
//
// 缓存策略采用 Cache-Aside 模式：读取时优先命中缓存，未命中则回源数据库并回填缓存；
// 写入时同步刷新缓存；更新/删除时清除关联缓存键。
// 缓存读写委托给 SshKeyCache（位于 repository/cache 子层），通过 constant.RedisKey
// 统一管理缓存键（TTL 30 分钟）。
type SshKeyRepo struct {
	db    *gorm.DB
	cache *cache.SshKeyCache
	log   *xLog.LogNamedLogger
}

// NewSshKeyRepo 创建 SshKeyRepo 实例
func NewSshKeyRepo(db *gorm.DB, rdb *redis.Client) *SshKeyRepo {
	return &SshKeyRepo{
		db:    db,
		cache: cache.NewSshKeyCache(&xCache.Cache{RDB: rdb}),
		log:   xLog.WithName(xLog.NamedREPO, "SshKeyRepo"),
	}
}

// Create 创建 SSH 密钥，成功后同步写入缓存
func (r *SshKeyRepo) Create(ctx context.Context, sshKey *entity.SshKey) (*entity.SshKey, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("Create - 创建 SSH 密钥 [name=%s, fingerprint=%s]", sshKey.Name, sshKey.Fingerprint))

	if err := r.db.WithContext(ctx).Create(sshKey).Error; err != nil {
		r.log.Warn(ctx, err.Error())
		return nil, xError.NewError(ctx, xError.DatabaseError, "创建 SSH 密钥失败", false, err)
	}

	if xErr := r.cache.SetConfig(ctx, sshKey); xErr != nil {
		r.log.Warn(ctx, xErr.Error())
	}
	return sshKey, nil
}

// GetByID 根据 ID 获取 SSH 密钥，优先读取缓存（Cache-First）
//
// 缓存命中时直接返回；未命中则查询数据库并回填缓存。
//
// 返回值:
//   - *entity.SshKey: 查询到的密钥实体（含 PrivateKey）
//   - bool:            是否找到
//   - *xError.Error:    查询过程中的错误
func (r *SshKeyRepo) GetByID(ctx context.Context, id xSnowflake.SnowflakeID) (*entity.SshKey, bool, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("GetByID - 根据 ID 获取 SSH 密钥 [%d]", id.Int64()))

	if sshKey, ok, _ := r.cache.GetConfig(ctx, id.Int64()); ok {
		r.log.Info(ctx, fmt.Sprintf("GetByID - 缓存命中 [%d]", id.Int64()))
		return sshKey, true, nil
	}

	var sshKey entity.SshKey
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&sshKey).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, xError.NewError(ctx, xError.DatabaseError, "查询 SSH 密钥失败", false, err)
	}

	if xErr := r.cache.SetConfig(ctx, &sshKey); xErr != nil {
		r.log.Warn(ctx, xErr.Error())
	}
	return &sshKey, true, nil
}

// GetByFingerprint 根据指纹查询 SSH 密钥（用于查重）
//
// 返回值:
//   - *entity.SshKey: 查询到的密钥实体（含 PrivateKey）
//   - bool:            是否找到
//   - *xError.Error:    查询过程中的错误
func (r *SshKeyRepo) GetByFingerprint(ctx context.Context, fingerprint string) (*entity.SshKey, bool, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("GetByFingerprint - 根据指纹查询 SSH 密钥 [%s]", fingerprint))

	var sshKey entity.SshKey
	if err := r.db.WithContext(ctx).Where("fingerprint = ?", fingerprint).First(&sshKey).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, xError.NewError(ctx, xError.DatabaseError, "根据指纹查询 SSH 密钥失败", false, err)
	}

	return &sshKey, true, nil
}

// List 分页获取 SSH 密钥列表（按创建时间降序）
//
// 返回值:
//   - []*entity.SshKey: 当前页的密钥列表
//   - int64:             符合条件的总记录数
//   - *xError.Error:     查询过程中的错误
func (r *SshKeyRepo) List(ctx context.Context, page, size int) ([]*entity.SshKey, int64, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("List - 分页获取 SSH 密钥列表 [page=%d, size=%d]", page, size))

	var total int64
	if err := r.db.WithContext(ctx).Model(&entity.SshKey{}).Count(&total).Error; err != nil {
		return nil, 0, xError.NewError(ctx, xError.DatabaseError, "统计 SSH 密钥数量失败", false, err)
	}

	var sshKeys []*entity.SshKey
	offset := (page - 1) * size
	if err := r.db.WithContext(ctx).
		Offset(offset).
		Limit(size).
		Order("created_at DESC").
		Find(&sshKeys).Error; err != nil {
		return nil, 0, xError.NewError(ctx, xError.DatabaseError, "查询 SSH 密钥列表失败", false, err)
	}

	return sshKeys, total, nil
}

// Update 更新 SSH 密钥的 name 和 description 字段，成功后清除缓存
func (r *SshKeyRepo) Update(ctx context.Context, sshKey *entity.SshKey) *xError.Error {
	r.log.Info(ctx, fmt.Sprintf("Update - 更新 SSH 密钥 [id=%d]", sshKey.ID.Int64()))

	result := r.db.WithContext(ctx).
		Model(&entity.SshKey{}).
		Where("id = ?", sshKey.ID).
		Select("name", "description").
		Updates(sshKey)
	if result.Error != nil {
		r.log.Warn(ctx, result.Error.Error())
		return xError.NewError(ctx, xError.DatabaseError, "更新 SSH 密钥失败", false, result.Error)
	}
	if result.RowsAffected == 0 {
		return xError.NewError(ctx, xError.NotFound, "SSH 密钥不存在", false, nil)
	}

	r.cache.DeleteConfig(ctx, sshKey.ID.Int64())
	return nil
}

// Delete 删除 SSH 密钥，成功后清除关联缓存
func (r *SshKeyRepo) Delete(ctx context.Context, id xSnowflake.SnowflakeID) *xError.Error {
	r.log.Info(ctx, fmt.Sprintf("Delete - 删除 SSH 密钥 [%d]", id.Int64()))

	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.SshKey{})
	if result.Error != nil {
		r.log.Warn(ctx, result.Error.Error())
		return xError.NewError(ctx, xError.DatabaseError, "删除 SSH 密钥失败", false, result.Error)
	}
	if result.RowsAffected == 0 {
		return xError.NewError(ctx, xError.NotFound, "SSH 密钥不存在", false, nil)
	}

	r.cache.DeleteConfig(ctx, id.Int64())
	return nil
}
