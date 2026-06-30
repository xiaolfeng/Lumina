package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	apiPin "github.com/xiaolfeng/Lumina/api/pin"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// errPinNoPending 是 ConsumeOldestPending 内部哨兵错误，表示当前项目无待消费 Pin。
// 仅在事务闭包内使用，事务提交后会通过 errors.Is 转换为 (nil, nil) 语义返回。
var errPinNoPending = errors.New("no pending pin available")

// errPinNotConsumable 是 ConsumeByID 内部哨兵错误，表示指定 Pin 不属于该项目或不在 pending 状态。
var errPinNotConsumable = errors.New("pin not found or not consumable")

// PinRepo Pin 数据访问层，提供完整 CRUD 操作与原子 FIFO 消费能力。
//
// 与 Project / QaSession 不同，PinRepo 不使用 Redis 缓存（D1 设计决策），
// 因为 Pin 的核心消费路径（ConsumeOldestPending）依赖 PostgreSQL 行级锁保证原子性，
// 缓存层会引入数据一致性问题。Pin 全量存储在 PostgreSQL。
//
// 字段说明:
//   - db:  GORM 数据库实例，执行持久化与事务操作
//   - log: 带命名空间的结构化日志记录器
type PinRepo struct {
	db  *gorm.DB
	log *xLog.LogNamedLogger
}

// NewPinRepo 创建 PinRepo 实例
//
// 参数说明:
//   - db: 已初始化的 GORM 数据库实例
//
// 返回值:
//   - *PinRepo: 配置完成的 PinRepo 实例指针
func NewPinRepo(db *gorm.DB) *PinRepo {
	return &PinRepo{
		db:  db,
		log: xLog.WithName(xLog.NamedREPO, "PinRepo"),
	}
}

// Create 创建 Pin 记录
//
// 参数:
//   - ctx:  上下文对象
//   - pin:  待创建的 Pin 实体（ID 由雪花算法自动生成）
//
// 返回值:
//   - *xError.Error: 创建过程中的错误
func (r *PinRepo) Create(ctx context.Context, pin *entity.Pin) *xError.Error {
	r.log.Info(ctx, fmt.Sprintf("Create - 创建 Pin [%s]", pin.Title))

	if err := r.db.WithContext(ctx).Create(pin).Error; err != nil {
		r.log.Warn(ctx, err.Error())
		return xError.NewError(ctx, xError.DatabaseError, "创建 Pin 失败", false, err)
	}
	return nil
}

// GetByID 根据 ID 获取 Pin
//
// 参数:
//   - ctx: 上下文对象
//   - id:  Pin 雪花 ID
//
// 返回值:
//   - *entity.Pin:  查询到的 Pin 实体
//   - *xError.Error: 查询过程中的错误
func (r *PinRepo) GetByID(ctx context.Context, id xSnowflake.SnowflakeID) (*entity.Pin, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("GetByID - 根据 ID 获取 Pin [%d]", id.Int64()))

	var pin entity.Pin
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&pin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, xError.NewError(ctx, xError.NotFound, "Pin 不存在", false, nil)
		}
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询 Pin 失败", false, err)
	}
	return &pin, nil
}

// List 分页获取 Pin 列表，支持多条件动态过滤
//
// 根据 PinListRequest 中的非空字段动态构建 WHERE 条件，按 created_at ASC 排序
// （FIFO 顺序，便于消费场景下展示队列顺序）。分页参数通过 xModels.PageRequest.Normalize() 规范化。
//
// 参数:
//   - ctx: 上下文对象
//   - req:  列表查询请求（含过滤字段 + 分页参数）
//
// 返回值:
//   - []*entity.Pin: 当前页的 Pin 列表
//   - int64:         符合条件的总记录数
//   - *xError.Error: 查询过程中的错误
func (r *PinRepo) List(ctx context.Context, req *apiPin.PinListRequest) ([]*entity.Pin, int64, *xError.Error) {
	// 分页参数规范化
	pageReq := xModels.PageRequest{Page: int64(req.Page), Size: int64(req.Size)}.Normalize()
	page := int(pageReq.Page)
	size := int(pageReq.Size)
	r.log.Info(ctx, fmt.Sprintf("List - 分页获取 Pin 列表 [page=%d, size=%d, toProject=%s, fromProject=%s, status=%s, category=%s, priority=%s]",
		page, size, req.ToProjectID, req.FromProjectID, req.Status, req.Category, req.Priority))

	// 构建基础查询
	query := r.db.WithContext(ctx).Model(&entity.Pin{})

	// 动态追加过滤条件
	if req.ToProjectID != "" {
		query = query.Where("to_project_id = ?", req.ToProjectID)
	}
	if req.FromProjectID != "" {
		query = query.Where("from_project_id = ?", req.FromProjectID)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.Category != "" {
		query = query.Where("category = ?", req.Category)
	}
	if req.Priority != "" {
		query = query.Where("priority = ?", req.Priority)
	}

	// 统计总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, xError.NewError(ctx, xError.DatabaseError, "统计 Pin 数量失败", false, err)
	}

	// 分页查询（FIFO 顺序：created_at ASC）
	var pins []*entity.Pin
	offset := (page - 1) * size
	if err := query.
		Offset(offset).
		Limit(size).
		Order("created_at ASC").
		Find(&pins).Error; err != nil {
		return nil, 0, xError.NewError(ctx, xError.DatabaseError, "查询 Pin 列表失败", false, err)
	}

	return pins, total, nil
}

// Update 更新 Pin 的元数据字段（priority / category），不支持更新 status 字段
//
// Logic 层确保 updates 仅包含 priority 和/或 category 键，
// 此方法直接透传给 GORM Updates。
//
// 参数:
//   - ctx:     上下文对象
//   - id:      待更新的 Pin 雪花 ID
//   - updates: 待更新的字段映射（仅 priority / category）
//
// 返回值:
//   - *entity.Pin:  更新后的 Pin 实体
//   - *xError.Error: 更新过程中的错误
func (r *PinRepo) Update(ctx context.Context, id xSnowflake.SnowflakeID, updates map[string]any) (*entity.Pin, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("Update - 更新 Pin [%d]", id.Int64()))

	result := r.db.WithContext(ctx).
		Model(&entity.Pin{}).
		Where("id = ?", id).
		Updates(updates)
	if result.Error != nil {
		r.log.Warn(ctx, result.Error.Error())
		return nil, xError.NewError(ctx, xError.DatabaseError, "更新 Pin 失败", false, result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, xError.NewError(ctx, xError.NotFound, "Pin 不存在", false, nil)
	}

	// 查询返回更新后的完整实体
	var pin entity.Pin
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&pin).Error; err != nil {
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询更新后的 Pin 失败", false, err)
	}
	return &pin, nil
}

// Delete 物理删除 Pin 记录
//
// 参数:
//   - ctx: 上下文对象
//   - id:  待删除的 Pin 雪花 ID
//
// 返回值:
//   - *xError.Error: 删除过程中的错误
func (r *PinRepo) Delete(ctx context.Context, id xSnowflake.SnowflakeID) *xError.Error {
	r.log.Info(ctx, fmt.Sprintf("Delete - 删除 Pin [%d]", id.Int64()))

	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.Pin{})
	if result.Error != nil {
		r.log.Warn(ctx, result.Error.Error())
		return xError.NewError(ctx, xError.DatabaseError, "删除 Pin 失败", false, result.Error)
	}
	if result.RowsAffected == 0 {
		return xError.NewError(ctx, xError.NotFound, "Pin 不存在", false, nil)
	}
	return nil
}

// ConsumeOldestPending 原子消费指定项目的最旧待处理 Pin（FIFO 队首消费）
//
// 使用 PostgreSQL FOR UPDATE SKIP LOCKED 行级锁在事务中完成「查询最旧 pending → 原子更新为 consumed」操作，
// 保证多个并发消费者不会重复消费同一条 Pin。当无待消费 Pin 时返回 (nil, nil)，不视为错误。
//
// 事务流程:
//  1. SELECT ... FOR UPDATE SKIP LOCKED 锁定并查询该项目的最旧 pending Pin（按 created_at ASC）
//  2. 原子 UPDATE 将其状态置为 consumed 并记录 consumed_at
//  3. 返回消费后的 Pin 实体
//
// 参数:
//   - ctx:        上下文对象
//   - projectID:  目标项目 ID（消费 to_project_id 等于此值的 Pin）
//
// 返回值:
//   - *entity.Pin:  被消费的 Pin 实体；无待消费 Pin 时返回 nil
//   - *xError.Error: 消费过程中的错误；无待消费 Pin 时返回 nil（非错误）
func (r *PinRepo) ConsumeOldestPending(ctx context.Context, projectID xSnowflake.SnowflakeID) (*entity.Pin, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("ConsumeOldestPending - 原子消费最旧待处理 Pin [projectID=%d]", projectID.Int64()))

	var pin entity.Pin
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Step 1: 加锁查询该项目最旧的 pending Pin（FIFO 队首）
		if err := tx.
			Where("to_project_id = ? AND status = ?", projectID, bConst.PinStatusPending).
			Order("created_at ASC").
			Limit(1).
			Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			First(&pin).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errPinNoPending // 哨兵：无待消费 Pin
			}
			return err
		}

		// Step 2: 原子更新为 consumed
		now := time.Now()
		if err := tx.Model(&pin).Updates(map[string]any{
			"status":      bConst.PinStatusConsumed,
			"consumed_at": now,
		}).Error; err != nil {
			return err
		}

		// Step 3: 回填内存对象，避免二次查询
		pin.Status = bConst.PinStatusConsumed
		pin.ConsumedAt = &now
		return nil
	})

	if err != nil {
		if errors.Is(err, errPinNoPending) {
			// 无待消费 Pin，非错误
			return nil, nil
		}
		r.log.Warn(ctx, err.Error())
		return nil, xError.NewError(ctx, xError.DatabaseError, "原子消费 Pin 失败", false, err)
	}

	r.log.Info(ctx, fmt.Sprintf("ConsumeOldestPending - 消费成功 [pinID=%d]", pin.ID.Int64()))
	return &pin, nil
}

// ConsumeByID 精确消费指定 ID 的 Pin（要求该项目归属匹配且状态为 pending）
//
// 通过条件 UPDATE 实现：仅当 id + to_project_id + status=pending 三条件全部匹配时才会更新成功，
// 否则 RowsAffected=0 表示 Pin 不存在 / 不属于此项目 / 已被消费。
//
// 参数:
//   - ctx:        上下文对象
//   - projectID:  目标项目 ID（校验 to_project_id 匹配）
//   - pinID:      待消费的 Pin ID
//
// 返回值:
//   - *entity.Pin:  被消费的 Pin 实体
//   - *xError.Error: 消费过程中的错误
func (r *PinRepo) ConsumeByID(ctx context.Context, projectID, pinID xSnowflake.SnowflakeID) (*entity.Pin, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("ConsumeByID - 精确消费 Pin [projectID=%d, pinID=%d]", projectID.Int64(), pinID.Int64()))

	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&entity.Pin{}).
		Where("id = ? AND to_project_id = ? AND status = ?", pinID, projectID, bConst.PinStatusPending).
		Updates(map[string]any{
			"status":      bConst.PinStatusConsumed,
			"consumed_at": now,
		})
	if result.Error != nil {
		r.log.Warn(ctx, result.Error.Error())
		return nil, xError.NewError(ctx, xError.DatabaseError, "消费 Pin 失败", false, result.Error)
	}
	if result.RowsAffected == 0 {
		// Pin 不存在 / 不属于此项目 / 状态非 pending（可能已被并发消费）
		return nil, xError.NewError(ctx, xError.NotFound, "Pin 不存在或不可消费", false, nil)
	}

	// 查询返回消费后的完整实体
	var pin entity.Pin
	if err := r.db.WithContext(ctx).Where("id = ?", pinID).First(&pin).Error; err != nil {
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询消费后的 Pin 失败", false, err)
	}

	r.log.Info(ctx, fmt.Sprintf("ConsumeByID - 消费成功 [pinID=%d]", pinID.Int64()))
	return &pin, nil
}
