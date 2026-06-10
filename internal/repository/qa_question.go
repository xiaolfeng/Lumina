package repository

import (
	"context"
	"fmt"
	"time"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"gorm.io/gorm"
)

// QaQuestionRepo QA问题数据访问层，提供问题持久化与查询操作
//
// 字段说明:
//   - db:  GORM 数据库实例，执行持久化操作
//   - log: 带命名空间的结构化日志记录器
type QaQuestionRepo struct {
	db  *gorm.DB
	log *xLog.LogNamedLogger
}

// NewQaQuestionRepo 创建 QaQuestionRepo 实例
//
// 参数说明:
//   - db: 已初始化的 GORM 数据库实例
//
// 返回值:
//   - *QaQuestionRepo: 配置完成的 QaQuestionRepo 实例指针
func NewQaQuestionRepo(db *gorm.DB) *QaQuestionRepo {
	return &QaQuestionRepo{
		db:  db,
		log: xLog.WithName(xLog.NamedREPO, "QaQuestionRepo"),
	}
}

// Create 创建QA问题记录
//
// 参数:
//   - ctx:      上下文对象
//   - question: 待创建的问题实体（ID 由雪花算法自动生成）
//
// 返回值:
//   - *xError.Error: 创建过程中的错误
func (r *QaQuestionRepo) Create(ctx context.Context, question *entity.QaQuestion) *xError.Error {
	r.log.Info(ctx, fmt.Sprintf("Create - 创建QA问题 [session_id=%d]", question.SessionID))

	if err := r.db.WithContext(ctx).Create(question).Error; err != nil {
		r.log.Warn(ctx, err.Error())
		return xError.NewError(ctx, xError.DatabaseError, "创建QA问题失败", false, err)
	}

	return nil
}

// GetByID 根据ID获取QA问题
//
// 参数:
//   - ctx: 上下文对象
//   - id:  问题雪花 ID
//
// 返回值:
//   - *entity.QaQuestion: 查询到的问题实体
//   - *xError.Error:      查询过程中的错误
func (r *QaQuestionRepo) GetByID(ctx context.Context, id xSnowflake.SnowflakeID) (*entity.QaQuestion, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("GetByID - 根据ID获取QA问题 [%d]", id.Int64()))

	var question entity.QaQuestion
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&question).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, xError.NewError(ctx, xError.NotFound, "QA问题不存在", false, nil)
		}
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询QA问题失败", false, err)
	}

	return &question, nil
}

// GetBySessionID 根据会话ID获取该会话下的所有问题（按创建时间升序）
//
// 参数:
//   - ctx:       上下文对象
//   - sessionID: 会话 ID
//
// 返回值:
//   - []*entity.QaQuestion: 问题列表（无结果时返回空切片）
//   - *xError.Error:        查询过程中的错误
func (r *QaQuestionRepo) GetBySessionID(ctx context.Context, sessionID int64) ([]*entity.QaQuestion, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("GetBySessionID - 根据会话ID获取问题列表 [session_id=%d]", sessionID))

	var questions []*entity.QaQuestion
	if err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Find(&questions).Error; err != nil {
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询会话问题列表失败", false, err)
	}

	// 保证返回空切片而非 nil
	if questions == nil {
		questions = make([]*entity.QaQuestion, 0)
	}

	return questions, nil
}

// UpdateStatus 更新问题状态，当状态为 answered 时同步设置 answered_at
//
// 参数:
//   - ctx:    上下文对象
//   - id:     问题雪花 ID
//   - status: 目标状态（pending/answered/skipped）
//
// 返回值:
//   - *xError.Error: 更新过程中的错误
func (r *QaQuestionRepo) UpdateStatus(ctx context.Context, id xSnowflake.SnowflakeID, status string) *xError.Error {
	r.log.Info(ctx, fmt.Sprintf("UpdateStatus - 更新问题状态 [id=%d, status=%s]", id.Int64(), status))

	updates := map[string]interface{}{
		"status": status,
	}

	// 状态为 answered 时同步设置回答时间
	if status == "answered" {
		now := time.Now()
		updates["answered_at"] = &now
	}

	result := r.db.WithContext(ctx).
		Model(&entity.QaQuestion{}).
		Where("id = ?", id).
		Updates(updates)
	if result.Error != nil {
		r.log.Warn(ctx, result.Error.Error())
		return xError.NewError(ctx, xError.DatabaseError, "更新问题状态失败", false, result.Error)
	}

	if result.RowsAffected == 0 {
		return xError.NewError(ctx, xError.NotFound, "QA问题不存在", false, nil)
	}

	return nil
}

// UpdateAnswer 更新问题回答，同时更新状态与回答时间
//
// 参数:
//   - ctx:    上下文对象
//   - id:     问题雪花 ID
//   - status: 目标状态（pending/answered/skipped）
//   - answer: 回答数据（调用方负责序列化为 datatypes.JSON）
//
// 返回值:
//   - *xError.Error: 更新过程中的错误
func (r *QaQuestionRepo) UpdateAnswer(ctx context.Context, id xSnowflake.SnowflakeID, status string, answer any) *xError.Error {
	r.log.Info(ctx, fmt.Sprintf("UpdateAnswer - 更新问题回答 [id=%d, status=%s]", id.Int64(), status))

	updates := map[string]interface{}{
		"status": status,
		"answer": answer,
	}

	// 状态为 answered 时同步设置回答时间
	if status == "answered" {
		now := time.Now()
		updates["answered_at"] = &now
	}

	result := r.db.WithContext(ctx).
		Model(&entity.QaQuestion{}).
		Where("id = ?", id).
		Updates(updates)
	if result.Error != nil {
		r.log.Warn(ctx, result.Error.Error())
		return xError.NewError(ctx, xError.DatabaseError, "更新问题回答失败", false, result.Error)
	}

	if result.RowsAffected == 0 {
		return xError.NewError(ctx, xError.NotFound, "QA问题不存在", false, nil)
	}

	return nil
}
