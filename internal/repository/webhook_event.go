package repository

import (
	"context"
	"fmt"
	"time"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	"github.com/redis/go-redis/v9"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"gorm.io/gorm"
)

// WebhookEventRepo Webhook事件持久化（仅追加审计日志，不更新不删除）
//
// 字段说明:
//   - db:  GORM数据库实例，执行持久化操作
//   - rdb: Redis客户端实例（预留，用于后续可能的缓存或限流扩展）
//   - log: 带命名空间的结构化日志记录器
type WebhookEventRepo struct {
	db  *gorm.DB
	rdb *redis.Client
	log *xLog.LogNamedLogger
}

// NewWebhookEventRepo 创建 WebhookEventRepo 实例
//
// 参数说明:
//   - db:  已初始化的 GORM数据库实例
//   - rdb: 已初始化的 Redis客户端实例
//
// 返回值:
//   - *WebhookEventRepo: 配置完成的 WebhookEventRepo 实例指针
func NewWebhookEventRepo(db *gorm.DB, rdb *redis.Client) *WebhookEventRepo {
	return &WebhookEventRepo{
		db:  db,
		rdb: rdb,
		log: xLog.WithName(xLog.NamedREPO, "WebhookEventRepo"),
	}
}

// Create 创建Webhook事件记录
//
// 参数:
//   - ctx:   上下文对象
//   - event: 待创建的事件实体（ID由雪花算法自动生成）
//
// 返回值:
//   - *entity.WebhookEvent: 创建后的事件实体（含生成的ID）
//   - *xError.Error:         创建过程中的错误
func (r *WebhookEventRepo) Create(ctx context.Context, event *entity.WebhookEvent) (*entity.WebhookEvent, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("Create - 创建 Webhook 事件 [provider=%s, eventType=%s, branch=%s]", event.Provider, event.EventType, event.Branch))

	if err := r.db.WithContext(ctx).Create(event).Error; err != nil {
		r.log.Warn(ctx, err.Error())
		return nil, xError.NewError(ctx, xError.DatabaseError, "创建 Webhook 事件失败", false, err)
	}

	return event, nil
}

// UpdateStatus 更新事件状态（用于 received → accepted/ignored/failed 转换）
//
// 参数:
//   - ctx:          上下文对象
//   - id:           事件雪花ID
//   - status:       新状态值
//   - reason:       跳过或失败原因
//   - versionID:    关联的Wiki版本ID
//   - responseCode: HTTP响应码
//   - processedAt:  处理完成时间
//
// 返回值:
//   - *xError.Error: 更新过程中的错误
func (r *WebhookEventRepo) UpdateStatus(ctx context.Context, id xSnowflake.SnowflakeID, status string, reason string, versionID xSnowflake.SnowflakeID, responseCode int, processedAt *time.Time) *xError.Error {
	r.log.Info(ctx, fmt.Sprintf("UpdateStatus - 更新 Webhook 事件状态 [%d] → %s", id.Int64(), status))

	result := r.db.WithContext(ctx).
		Model(&entity.WebhookEvent{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":         status,
			"reason":         reason,
			"version_id":     versionID,
			"response_code":  responseCode,
			"processed_at":   processedAt,
			"updated_at":     time.Now(),
		})
	if result.Error != nil {
		r.log.Warn(ctx, result.Error.Error())
		return xError.NewError(ctx, xError.DatabaseError, "更新 Webhook 事件状态失败", false, result.Error)
	}
	if result.RowsAffected == 0 {
		return xError.NewError(ctx, xError.NotFound, "Webhook 事件不存在", false, nil)
	}

	return nil
}

// ListByConfigID 按配置ID分页查询事件（按接收时间倒序）
//
// 参数:
//   - ctx:      上下文对象
//   - configID: 关联的配置雪花ID
//   - page:     页码（从1开始）
//   - size:     每页数量
//
// 返回值:
//   - []*entity.WebhookEvent: 当前页的事件列表
//   - int64:                   符合条件的总记录数
//   - *xError.Error:           查询过程中的错误
func (r *WebhookEventRepo) ListByConfigID(ctx context.Context, configID xSnowflake.SnowflakeID, page, size int) ([]*entity.WebhookEvent, int64, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("ListByConfigID - 按配置ID分页查询 Webhook 事件 [configID=%d, page=%d, size=%d]", configID.Int64(), page, size))

	var total int64
	if err := r.db.WithContext(ctx).Model(&entity.WebhookEvent{}).Where("config_id = ?", configID).Count(&total).Error; err != nil {
		return nil, 0, xError.NewError(ctx, xError.DatabaseError, "统计 Webhook 事件数量失败", false, err)
	}

	var events []*entity.WebhookEvent
	offset := (page - 1) * size
	if err := r.db.WithContext(ctx).
		Where("config_id = ?", configID).
		Offset(offset).
		Limit(size).
		Order("received_at DESC").
		Find(&events).Error; err != nil {
		return nil, 0, xError.NewError(ctx, xError.DatabaseError, "查询 Webhook 事件列表失败", false, err)
	}

	return events, total, nil
}
