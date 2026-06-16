package repository

import (
	"context"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"gorm.io/gorm"
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
