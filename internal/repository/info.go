package repository

import (
	"context"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// InfoRepo 键值配置数据访问层，统一承接 logic 层对 entity.Info 表的读写
//
// 该仓库将原先散落在 auth/qa logic 层的 Info 表直查收口到 repository，
// 消除 logic 直连 DB 的越界调用。Info 表以 key 为主键（非雪花 ID），
// 因此不提供 GetByID/Create/List/Delete，仅提供按 key 的点查与更新。
//
// 字段说明:
//   - db:  GORM 数据库实例
//   - log: 带命名空间的结构化日志记录器
type InfoRepo struct {
	db  *gorm.DB
	log *xLog.LogNamedLogger
}

// NewInfoRepo 创建 InfoRepo 实例
//
// 参数说明:
//   - db: 已初始化的 GORM 数据库实例
//
// 返回值:
//   - *InfoRepo: 配置完成的 InfoRepo 实例指针
func NewInfoRepo(db *gorm.DB) *InfoRepo {
	return &InfoRepo{
		db:  db,
		log: xLog.WithName(xLog.NamedREPO, "InfoRepo"),
	}
}

// GetByKey 根据 key 读取配置值
//
// 未命中时返回空字符串与 NotFound 错误，由调用方（logic）决定兜底策略。
//
// 参数:
//   - ctx: 上下文对象
//   - key:  配置键名（主键）
//
// 返回值:
//   - string:       配置值
//   - *xError.Error: 查询过程中的错误（含 NotFound）
func (r *InfoRepo) GetByKey(ctx context.Context, key string) (string, *xError.Error) {
	r.log.Info(ctx, "GetByKey - 根据key获取配置 ["+key+"]")

	var info entity.Info
	if err := r.db.WithContext(ctx).Where("\"key\" = ?", key).First(&info).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", xError.NewError(ctx, xError.NotFound, "配置项不存在", false, nil)
		}
		return "", xError.NewError(ctx, xError.DatabaseError, "查询配置项失败", false, err)
	}

	return info.Value, nil
}

// UpdateValue 更新单个 key 的配置值
//
// 参数:
//   - ctx:   上下文对象
//   - key:   配置键名
//   - value: 新配置值
//
// 返回值:
//   - *xError.Error: 更新过程中的错误
func (r *InfoRepo) UpdateValue(ctx context.Context, key, value string) *xError.Error {
	r.log.Info(ctx, "UpdateValue - 更新配置 ["+key+"]")

	if err := r.db.WithContext(ctx).
		Model(&entity.Info{}).
		Where("\"key\" = ?", key).
		Update("value", value).Error; err != nil {
		return xError.NewError(ctx, xError.DatabaseError, "更新配置项失败", false, err)
	}

	return nil
}

// UpdateValuesInTx 在单个事务内原子更新多个 key 的配置值
//
// 任一 key 更新失败则整体回滚。用于 auth.Initialize 等「全部成功或全部失败」场景。
//
// 参数:
//   - ctx: 上下文对象
//   - kv:  key→value 映射（map 遍历顺序不确定，但不影响结果正确性）
//
// 返回值:
//   - *xError.Error: 事务执行过程中的错误
func (r *InfoRepo) UpdateValuesInTx(ctx context.Context, kv map[string]string) *xError.Error {
	r.log.Info(ctx, "UpdateValuesInTx - 事务更新多个配置")

	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for key, value := range kv {
			if err := tx.WithContext(ctx).
				Model(&entity.Info{}).
				Where("\"key\" = ?", key).
				Update("value", value).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return xError.NewError(ctx, xError.DatabaseError, "事务更新配置项失败", false, err)
	}

	return nil
}

// UpsertValue 插入或更新配置项（key 存在则更新 value，不存在则插入）
//
// 使用 GORM clause.OnConflict 实现 PostgreSQL 的 INSERT ... ON CONFLICT 语义，
// 适用于 Agent 模型分配等「不存在则创建、存在则覆盖」场景。
//
// 参数:
//   - ctx:   上下文对象
//   - key:   配置键名（主键）
//   - value: 配置值
//
// 返回值:
//   - *xError.Error: 插入或更新过程中的错误
func (r *InfoRepo) UpsertValue(ctx context.Context, key, value string) *xError.Error {
	r.log.Info(ctx, "UpsertValue - 插入或更新配置 ["+key+"]")

	info := entity.Info{
		Key:   key,
		Value: value,
	}

	if err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "key"}},
			DoUpdates: clause.AssignmentColumns([]string{"value"}),
		}).
		Create(&info).Error; err != nil {
		return xError.NewError(ctx, xError.DatabaseError, "插入或更新配置项失败", false, err)
	}

	return nil
}

// InitializeIfNotInitialized 在事务内原子地检查初始化状态并写入凭据。
//
// 通过 SELECT ... FOR UPDATE 行锁锁住初始化标志行，保证「检查-写入」原子，
// 杜绝并发 TOCTOU（两个请求同时通过 GetInitialStatus 读旧值后重复初始化）。
//
// 参数:
//   - ctx:         上下文对象
//   - initFlagKey: 初始化状态键名（值为 "true" 表示未初始化，见 prepare 种子）
//   - kv:          待写入的 key→value 映射（应包含将 initFlagKey 置 "false"）
//
// 返回值:
//   - bool:         true 表示本次成功执行初始化；false 表示已被他人初始化
//   - *xError.Error: 事务执行过程中的错误
func (r *InfoRepo) InitializeIfNotInitialized(ctx context.Context, initFlagKey string, kv map[string]string) (bool, *xError.Error) {
	r.log.Info(ctx, "InitializeIfNotInitialized - 原子初始化")

	var initialized bool
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 行锁读取初始化标志，防止并发重复初始化
		var info entity.Info
		readErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("\"key\" = ?", initFlagKey).
			First(&info).Error

		switch {
		case readErr == gorm.ErrRecordNotFound:
			// 种子行缺失：视为未初始化，继续写入（下方 Upsert 会补全）
		case readErr != nil:
			return readErr
		default:
			// 已初始化（标志非 "true"）则放弃，交由调用方返回 RepeatOperation
			if info.Value != "true" {
				return nil
			}
		}

		// 未初始化，用 Upsert 原子写入全部凭据（幂等，兼容 owner 种子行缺失）
		for key, value := range kv {
			item := entity.Info{Key: key, Value: value}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "key"}},
				DoUpdates: clause.AssignmentColumns([]string{"value"}),
			}).Create(&item).Error; err != nil {
				return err
			}
		}
		initialized = true
		return nil
	})
	if err != nil {
		return false, xError.NewError(ctx, xError.DatabaseError, "事务更新配置项失败", false, err)
	}

	return initialized, nil
}
