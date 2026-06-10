package repository

import (
	"context"
	"fmt"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"gorm.io/gorm"
)

// QaSupplementRepo QA补充数据访问层，提供 1:1 覆写写入与按目标/会话查询
//
// 字段说明:
//   - db:  GORM 数据库实例，执行持久化操作
//   - log: 带命名空间的结构化日志记录器
type QaSupplementRepo struct {
	db  *gorm.DB
	log *xLog.LogNamedLogger
}

// NewQaSupplementRepo 创建 QaSupplementRepo 实例
//
// 参数说明:
//   - db: 已初始化的 GORM 数据库实例
//
// 返回值:
//   - *QaSupplementRepo: 配置完成的 QaSupplementRepo 实例指针
func NewQaSupplementRepo(db *gorm.DB) *QaSupplementRepo {
	return &QaSupplementRepo{
		db:  db,
		log: xLog.WithName(xLog.NamedREPO, "QaSupplementRepo"),
	}
}

// CreateOrUpdate 创建或更新补充说明（1:1 覆写）
//
// 按 (target_type, target_id) 唯一约束进行覆写：
//   - 已存在：更新 Content、ContentType、UpdatedAt 字段
//   - 不存在：创建新记录
//
// 参数:
//   - ctx:       上下文对象
//   - supplement: 待创建或更新的补充说明实体
//
// 返回值:
//   - *entity.QaSupplement: 持久化后的实体
//   - *xError.Error:         操作过程中的错误
func (r *QaSupplementRepo) CreateOrUpdate(ctx context.Context, supplement *entity.QaSupplement) (*entity.QaSupplement, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("CreateOrUpdate - 创建或更新补充说明 [%s/%d]", supplement.TargetType, supplement.TargetID))

	var existing entity.QaSupplement
	err := r.db.WithContext(ctx).
		Where("target_type = ? AND target_id = ?", supplement.TargetType, supplement.TargetID).
		First(&existing).Error

	if err == nil {
		// 已存在 → 覆写
		existing.Content = supplement.Content
		existing.ContentType = supplement.ContentType
		if saveErr := r.db.WithContext(ctx).Save(&existing).Error; saveErr != nil {
			r.log.Warn(ctx, saveErr.Error())
			return nil, xError.NewError(ctx, xError.DatabaseError, "更新补充说明失败", false, saveErr)
		}
		return &existing, nil
	}

	if err == gorm.ErrRecordNotFound {
		// 不存在 → 创建
		if createErr := r.db.WithContext(ctx).Create(supplement).Error; createErr != nil {
			r.log.Warn(ctx, createErr.Error())
			return nil, xError.NewError(ctx, xError.DatabaseError, "创建补充说明失败", false, createErr)
		}
		return supplement, nil
	}

	// 其他数据库错误
	r.log.Warn(ctx, err.Error())
	return nil, xError.NewError(ctx, xError.DatabaseError, "查询补充说明失败", false, err)
}

// GetByTarget 根据目标类型和目标ID获取补充说明（1:1）
//
// 参数:
//   - ctx:       上下文对象
//   - targetType: 目标类型（question/option）
//   - targetID:   目标雪花 ID
//
// 返回值:
//   - *entity.QaSupplement: 查询到的补充说明实体
//   - *xError.Error:         查询过程中的错误（NotFound 表示未找到）
func (r *QaSupplementRepo) GetByTarget(ctx context.Context, targetType string, targetID xSnowflake.SnowflakeID) (*entity.QaSupplement, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("GetByTarget - 根据目标获取补充说明 [%s/%d]", targetType, targetID.Int64()))

	var supplement entity.QaSupplement
	if err := r.db.WithContext(ctx).
		Where("target_type = ? AND target_id = ?", targetType, targetID).
		First(&supplement).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, xError.NewError(ctx, xError.NotFound, "补充说明不存在", false, nil)
		}
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询补充说明失败", false, err)
	}

	return &supplement, nil
}

// GetBySessionID 根据会话ID获取所有补充说明列表
//
// 按 target_type、target_id 升序排列。
//
// 参数:
//   - ctx:      上下文对象
//   - sessionID: 会话雪花 ID
//
// 返回值:
//   - []*entity.QaSupplement: 补充说明列表（无记录时返回空切片）
//   - *xError.Error:           查询过程中的错误
func (r *QaSupplementRepo) GetBySessionID(ctx context.Context, sessionID xSnowflake.SnowflakeID) ([]*entity.QaSupplement, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("GetBySessionID - 根据会话获取补充说明列表 [%d]", sessionID.Int64()))

	supplements := make([]*entity.QaSupplement, 0)
	if err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("target_type ASC, target_id ASC").
		Find(&supplements).Error; err != nil {
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询会话补充说明列表失败", false, err)
	}

	return supplements, nil
}
