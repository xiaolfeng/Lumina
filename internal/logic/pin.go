package logic

import (
	"context"
	"fmt"
	"strings"
	"time"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xCtxUtil "github.com/bamboo-services/bamboo-base-go/common/utility/context"
	apiPin "github.com/xiaolfeng/Lumina/api/pin"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"github.com/xiaolfeng/Lumina/internal/repository"
)

// pinRepo Pin 模块依赖的仓储集合
//
// 刻意直接持有 *repository.ProjectRepo 而非 *ProjectLogic（D1 设计决策）：
// Pin 仅需要项目别名解析能力（resolveProject），属于数据访问层职责，
// 不涉及项目业务编排，直接复用 ProjectRepo 避免跨模块 Logic 耦合。
type pinRepo struct {
	pin     *repository.PinRepo
	project *repository.ProjectRepo
}

// PinLogic Pin 业务逻辑层，负责跨项目约束推送、消费与查询编排
type PinLogic struct {
	logic
	repo pinRepo
}

// NewPinLogic 创建 PinLogic 实例
//
// 通过上下文获取 db/rdb，构造 PinRepo 和 ProjectRepo 注入到 pinRepo 聚合结构中。
func NewPinLogic(ctx context.Context) *PinLogic {
	db := xCtxUtil.MustGetDB(ctx)
	rdb := xCtxUtil.MustGetRDB(ctx)

	return &PinLogic{
		logic: logic{
			log: xLog.WithName(xLog.NamedLOGC, "PinLogic"),
		},
		repo: pinRepo{
			pin:     repository.NewPinRepo(db),
			project: repository.NewProjectRepo(db, rdb),
		},
	}
}

// ResolveProject 根据名称/ID 双模式解析目标项目
//
// 双模式调度策略：
//  1. 优先尝试将输入解析为雪花 ID，命中则直接按 ID 查询
//  2. 解析失败时降级为别名查询（输入转小写以匹配 JSON @> 的区分大小写特性）
//  3. 两种方式均未命中时返回 NotFound 错误
//
// 导出方法供 MCP 工具处理器复用项目别名/ID 解析能力（如 pin_consume 工具
// 需将用户传入的 project_name 解析为 SnowflakeID 后再调用 Consume）。
func (l *PinLogic) ResolveProject(ctx context.Context, nameOrID string) (*entity.Project, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("ResolveProject - 解析目标项目 [%s]", nameOrID))

	// Step 1: 尝试解析为雪花 ID
	if parsedID, err := xSnowflake.ParseSnowflakeID(nameOrID); err == nil {
		if project, xErr := l.repo.project.GetByID(ctx, parsedID); xErr != nil {
			// ID 解析成功但查询失败（NotFound 或 DB 错误），直接透传
			return nil, xErr
		} else {
			return project, nil
		}
	}

	// Step 2: 降级为别名查询（输入转小写以匹配 JSON @> 区分大小写）
	project, xErr := l.repo.project.FindByAliasName(ctx, strings.ToLower(nameOrID))
	if xErr != nil {
		// 别名查询也失败，返回统一的 NotFound 错误
		if xErr.GetErrorCode() == xError.NotFound {
			return nil, xError.NewError(ctx, xError.NotFound, "项目不存在", false, nil)
		}
		return nil, xErr
	}

	return project, nil
}

// Push 创建 Pin 约束，解析来源/目标项目后构建实体并持久化
func (l *PinLogic) Push(ctx context.Context, req *apiPin.CreatePinRequest) (*apiPin.PinResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("Push - 创建 Pin 约束 [title=%s, toProject=%s]", req.Title, req.ToProjectID))

	// 解析目标项目
	toProject, xErr := l.ResolveProject(ctx, req.ToProjectID)
	if xErr != nil {
		return nil, xErr
	}

	// 解析来源项目（可选）
	var fromProjectID xSnowflake.SnowflakeID
	if req.FromProjectID != "" {
		fromProject, xErr := l.ResolveProject(ctx, req.FromProjectID)
		if xErr != nil {
			return nil, xErr
		}
		fromProjectID = fromProject.ID
	}

	// 分类默认值
	category := req.Category
	if category == "" {
		category = bConst.PinCategoryNotice
	}

	// 生成雪花 ID 并构建实体
	id := xSnowflake.GenerateID(bConst.GenePin)
	pinEntity := &entity.Pin{
		BaseEntity:    xModels.BaseEntity{ID: id},
		FromProjectID: fromProjectID,
		ToProjectID:   toProject.ID,
		Title:         req.Title,
		Content:       req.Content,
		Category:      category,
		Status:        bConst.PinStatusPending,
		Priority:      req.Priority,
	}

	// 持久化
	if xErr := l.repo.pin.Create(ctx, pinEntity); xErr != nil {
		return nil, xErr
	}

	return l.toResponse(pinEntity), nil
}

// Consume 消费 Pin 约束（双模式：FIFO 队首 / 精确 ID）
//
// 当 pinID 为 nil 时消费该项目最旧的待处理 Pin（FIFO 语义）；
// 当 pinID 非 nil 时精确消费指定 ID 的 Pin。
func (l *PinLogic) Consume(ctx context.Context, projectID xSnowflake.SnowflakeID, pinID *xSnowflake.SnowflakeID) (*apiPin.PinResponse, *xError.Error) {
	if pinID == nil {
		l.log.Info(ctx, fmt.Sprintf("Consume - FIFO 队首消费 [projectID=%d]", projectID.Int64()))
		pin, xErr := l.repo.pin.ConsumeOldestPending(ctx, projectID)
		if xErr != nil {
			return nil, xErr
		}
		if pin == nil {
			// 无待处理约束，返回 NotFound 语义错误供上层判定
			return nil, xError.NewError(ctx, xError.NotFound, "暂无待处理约束", false, nil)
		}
		return l.toResponse(pin), nil
	}

	l.log.Info(ctx, fmt.Sprintf("Consume - 精确消费 [projectID=%d, pinID=%d]", projectID.Int64(), pinID.Int64()))
	pin, xErr := l.repo.pin.ConsumeByID(ctx, projectID, *pinID)
	if xErr != nil {
		return nil, xErr
	}
	return l.toResponse(pin), nil
}

// List 分页获取 Pin 列表，支持多条件动态过滤
func (l *PinLogic) List(ctx context.Context, req *apiPin.PinListRequest) (*apiPin.PinListResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("List - 分页获取 Pin 列表 [toProject=%s, status=%s]", req.ToProjectID, req.Status))

	pins, total, xErr := l.repo.pin.List(ctx, req)
	if xErr != nil {
		return nil, xErr
	}

	items := make([]apiPin.PinResponse, 0, len(pins))
	for _, p := range pins {
		items = append(items, *l.toResponse(p))
	}

	return &apiPin.PinListResponse{
		Items: items,
		Total: total,
	}, nil
}

// Update 更新 Pin 的元数据字段（仅 priority / category），不支持更新 status
//
// 接收指针字段实现可选更新语义：nil 字段不更新。
// 状态流转（pending → consumed）由 Consume 方法独占控制，禁止通过 Update 修改。
func (l *PinLogic) Update(ctx context.Context, id xSnowflake.SnowflakeID, req *apiPin.UpdatePinRequest) (*apiPin.PinResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("Update - 更新 Pin 元数据 [%d]", id.Int64()))

	updates := make(map[string]any)
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}
	if req.Category != nil {
		updates["category"] = *req.Category
	}

	// 无更新字段，直接返回当前实体
	if len(updates) == 0 {
		pin, xErr := l.repo.pin.GetByID(ctx, id)
		if xErr != nil {
			return nil, xErr
		}
		return l.toResponse(pin), nil
	}

	updated, xErr := l.repo.pin.Update(ctx, id, updates)
	if xErr != nil {
		return nil, xErr
	}

	return l.toResponse(updated), nil
}

// Delete 删除 Pin 记录
func (l *PinLogic) Delete(ctx context.Context, id xSnowflake.SnowflakeID) *xError.Error {
	l.log.Info(ctx, fmt.Sprintf("Delete - 删除 Pin [%d]", id.Int64()))
	return l.repo.pin.Delete(ctx, id)
}

// Peek 查看 Pin 详情（只读，不改变状态）
//
// ConsumedAt 字段非空表示该 Pin 已被消费。
func (l *PinLogic) Peek(ctx context.Context, id xSnowflake.SnowflakeID) (*apiPin.PinResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("Peek - 查看 Pin 详情 [%d]", id.Int64()))

	pin, xErr := l.repo.pin.GetByID(ctx, id)
	if xErr != nil {
		return nil, xErr
	}
	return l.toResponse(pin), nil
}

// toResponse 将 Pin 实体映射为响应 DTO
//
// 时间格式遵循 RFC3339（与项目模块 toResponse 保持一致）。
// FromProjectID 为零值时返回空字符串，避免前端展示无意义的 "0"。
func (l *PinLogic) toResponse(pin *entity.Pin) *apiPin.PinResponse {
	fromProjectID := ""
	if pin.FromProjectID != 0 {
		fromProjectID = pin.FromProjectID.String()
	}

	consumedAt := ""
	if pin.ConsumedAt != nil {
		consumedAt = pin.ConsumedAt.Format(time.RFC3339)
	}

	return &apiPin.PinResponse{
		ID:            pin.ID.String(),
		Title:         pin.Title,
		Content:       pin.Content,
		Category:      pin.Category,
		Status:        pin.Status,
		Priority:      pin.Priority,
		FromProjectID: fromProjectID,
		ToProjectID:   pin.ToProjectID.String(),
		ConsumedAt:    consumedAt,
		CreatedAt:     pin.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     pin.UpdatedAt.Format(time.RFC3339),
	}
}
