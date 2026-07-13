package logic

import (
	"context"
	"fmt"
	"strconv"
	"time"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xCtxUtil "github.com/bamboo-services/bamboo-base-go/common/utility/context"
	xCache "github.com/bamboo-services/bamboo-base-go/major/cache"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	"github.com/xiaolfeng/Lumina/api/qa"
	"github.com/xiaolfeng/Lumina/internal/entity"
	qaQueue "github.com/xiaolfeng/Lumina/internal/qa"
	"github.com/xiaolfeng/Lumina/internal/repository"
	"github.com/xiaolfeng/Lumina/internal/repository/cache"
	"github.com/xiaolfeng/Lumina/internal/service"
)

// qaRepo QA模块依赖的仓储集合
type qaRepo struct {
	session       *repository.QaSessionRepo
	question      *repository.QaQuestionRepo
	supplement    *repository.QaSupplementRepo
	info          *repository.InfoRepo
	retryCache    *cache.QaRetryCache
	downloadToken *service.DownloadTokenService
	fileCache     *service.FileCacheService
}

// QaLogic QA问答业务编排层
type QaLogic struct {
	logic
	repo  qaRepo
	queue *qaQueue.AnswerQueue
}

// OnQuestionPushed 问题推送成功后的回调钩子，由 WebSocket 层设置以广播问题到在线设备
var OnQuestionPushed func(sessionID string, question *entity.QaQuestion)

// OnSupplementPushed 补充内容推送成功后的回调钩子，由 WebSocket 层设置以广播补充内容到在线设备
var OnSupplementPushed func(sessionID string, supplement *entity.QaSupplement)

// OnQuestionCancelled 问题取消后的回调钩子，由 WebSocket 层设置以广播取消通知到在线设备
//
// question 为 nil 时表示全部取消（cancelAll=true），非 nil 时表示单个问题取消
var OnQuestionCancelled func(sessionID string, question *entity.QaQuestion)

// OnSessionArchived 会话归档后的回调钩子，由 WebSocket 层设置以广播 session_end 到在线设备
var OnSessionArchived func(sessionID string)

// NewQaLogic 创建QaLogic实例
func NewQaLogic(ctx context.Context) *QaLogic {
	db := xCtxUtil.MustGetDB(ctx)
	rdb := xCtxUtil.MustGetRDB(ctx)

	return &QaLogic{
		logic: logic{
			log: xLog.WithName(xLog.NamedLOGC, "QaLogic"),
		},
		repo: qaRepo{
			session:       repository.NewQaSessionRepo(db, rdb),
			question:      repository.NewQaQuestionRepo(db),
			supplement:    repository.NewQaSupplementRepo(db),
			info:          repository.NewInfoRepo(db),
			retryCache:    &cache.QaRetryCache{Cache: &xCache.Cache{RDB: rdb}},
			downloadToken: service.NewDownloadTokenService(rdb),
			fileCache:     service.NewFileCacheService(),
		},
		queue: qaQueue.GetAnswerQueue(),
	}
}

// ─── Management API Methods ─────────────────────────────────────────────

// ListSessions 分页获取QA会话列表
func (l *QaLogic) ListSessions(ctx context.Context, req *qa.ListSessionRequest) (*qa.SessionListResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("ListSessions - 获取会话列表 [page=%d, size=%d, status=%s, type=%s]",
		req.Page, req.Size, req.Status, req.Type))

	// 分页参数规范化
	pageReq := xModels.PageRequest{Page: int64(req.Page), Size: int64(req.Size)}.Normalize()
	page := int(pageReq.Page)
	size := int(pageReq.Size)

	// 查询列表
	sessions, total, xErr := l.repo.session.List(ctx, page, size, req.Status, req.Type, req.Hash)
	if xErr != nil {
		return nil, xErr
	}

	// 映射响应
	items := make([]qa.SessionResponse, 0, len(sessions))
	for _, s := range sessions {
		items = append(items, toSessionResponse(s))
	}

	return &qa.SessionListResponse{
		Items: items,
		Total: total,
	}, nil
}

// GetSessionDetail 获取QA会话详情（含问题列表）
func (l *QaLogic) GetSessionDetail(ctx context.Context, id string) (*qa.SessionDetailResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("GetSessionDetail - 获取会话详情 [%s]", id))

	// 解析雪花 ID
	parsedID, err := xSnowflake.ParseSnowflakeID(id)
	if err != nil {
		return nil, xError.NewError(ctx, xError.BusinessError, "无效的会话ID", false, nil)
	}

	// 查询会话及问题列表
	session, questions, xErr := l.repo.session.GetByIDWithQuestions(ctx, parsedID)
	if xErr != nil {
		return nil, xErr
	}

	// 批量查询会话级全部补充内容（一次查询，避免 N+1），按问题 ID 分组
	supplementMap := make(map[string][]qa.SupplementResponse)
	supplements, suppErr := l.repo.supplement.GetBySessionID(ctx, parsedID)
	if suppErr != nil {
		l.log.Warn(ctx, fmt.Sprintf("GetSessionDetail - 查询补充内容失败（忽略，返回空补充）: %s", suppErr.Error()))
	}
	for _, s := range supplements {
		// 仅聚合问题级补充（target_type=question），选项级补充由单题详情接口提供
		if s.TargetType != "question" {
			continue
		}
		key := s.TargetID.String()
		supplementMap[key] = append(supplementMap[key], qa.SupplementResponse{
			ID:          s.ID,
			TargetType:  s.TargetType,
			TargetID:    s.TargetID,
			ContentType: s.ContentType,
			Content:     s.Content,
			CreatedAt:   s.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   s.UpdatedAt.Format(time.RFC3339),
		})
	}

	// 映射问题摘要（含完整历史数据：回答、选项、补充内容）
	questionSummaries := make([]qa.QuestionSummaryResponse, 0, len(questions))
	for _, q := range questions {
		questionSummaries = append(questionSummaries, toQuestionSummary(q, supplementMap[q.ID.String()]))
	}

	return &qa.SessionDetailResponse{
		ID:            session.ID,
		Hash:          session.Hash,
		Title:         session.Title,
		Agent:         session.Agent,
		Type:          session.Type,
		Status:        session.Status,
		OnlineDevices: session.OnlineDevices,
		ExpiresAt:     formatTimePtr(session.ExpiresAt),
		CreatedAt:     session.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     session.UpdatedAt.Format(time.RFC3339),
		Questions:     questionSummaries,
	}, nil
}

// GetQuestionDetail 获取QA问题详情（含补充内容）
func (l *QaLogic) GetQuestionDetail(ctx context.Context, sessionID, questionID string) (*qa.QuestionDetailResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("GetQuestionDetail - 获取问题详情 [session=%s, question=%s]", sessionID, questionID))

	// 解析会话ID
	parsedSID, err := xSnowflake.ParseSnowflakeID(sessionID)
	if err != nil {
		return nil, xError.NewError(ctx, xError.BusinessError, "无效的会话ID", false, nil)
	}

	// 解析问题ID
	parsedQID, err := xSnowflake.ParseSnowflakeID(questionID)
	if err != nil {
		return nil, xError.NewError(ctx, xError.BusinessError, "无效的问题ID", false, nil)
	}

	// 查询问题
	question, xErr := l.repo.question.GetByID(ctx, parsedQID)
	if xErr != nil {
		return nil, xErr
	}

	// 查询问题级别的补充内容
	supplement, suppErr := l.repo.supplement.GetByTarget(ctx, "question", parsedQID)
	var supplements []qa.SupplementResponse
	if suppErr == nil && supplement != nil {
		supplements = []qa.SupplementResponse{
			{
				ID:          supplement.ID,
				TargetType:  supplement.TargetType,
				TargetID:    supplement.TargetID,
				ContentType: supplement.ContentType,
				Content:     supplement.Content,
				CreatedAt:   supplement.CreatedAt.Format(time.RFC3339),
				UpdatedAt:   supplement.UpdatedAt.Format(time.RFC3339),
			},
		}
	} else {
		supplements = make([]qa.SupplementResponse, 0)
	}

	return &qa.QuestionDetailResponse{
		ID:          question.ID,
		SessionID:   parsedSID,
		Type:        question.Type,
		Title:       question.Title,
		Description: question.Description,
		Options:     jsonOrNull(question.Options),
		Config:      jsonOrNull(question.Config),
		Batch:       jsonOrNull(question.Batch),
		GroupLabel:  question.GroupLabel,
		Status:      question.Status,
		Answer:      jsonOrNull(question.Answer),
		Supplements: supplements,
		CreatedAt:   question.CreatedAt.Format(time.RFC3339),
		AnsweredAt:  formatTimePtr(question.AnsweredAt),
	}, nil
}

// DeleteSession 删除QA会话（软删除：状态设为deleted）
func (l *QaLogic) DeleteSession(ctx context.Context, id string) *xError.Error {
	l.log.Info(ctx, fmt.Sprintf("DeleteSession - 删除会话 [%s]", id))

	// 解析雪花 ID
	parsedID, err := xSnowflake.ParseSnowflakeID(id)
	if err != nil {
		return xError.NewError(ctx, xError.BusinessError, "无效的会话ID", false, nil)
	}

	// 执行软删除
	if xErr := l.repo.session.Delete(ctx, parsedID); xErr != nil {
		return xErr
	}

	// P-16: 删除后同步清理会话级文件缓存（base64 媒体文件等）
	fileCacheSvc := service.NewFileCacheService()
	if cleanErr := fileCacheSvc.CleanSession(ctx, id); cleanErr != nil {
		l.log.Warn(ctx, fmt.Sprintf("DeleteSession - 删除后清理文件缓存失败（忽略）: %s", cleanErr.Error()))
	}

	// 清理队列和重试计数器
	l.queue.RemoveQueue(id)
	if resetErr := l.repo.retryCache.Reset(ctx, id); resetErr != nil {
		l.log.Warn(ctx, fmt.Sprintf("DeleteSession - 重置重试计数器失败（忽略）: %s", resetErr.Error()))
	}

	return nil
}

// GetQaConfig 获取Q&A配置
func (l *QaLogic) GetQaConfig(ctx context.Context) (*qa.QaConfigResponse, *xError.Error) {
	l.log.Info(ctx, "GetQaConfig - 获取Q&A配置")

	// 读取 Session TTL（读失败兜底默认 7 天）
	ttlStr, xErr := l.repo.info.GetByKey(ctx, "qa.session.ttl")
	if xErr != nil {
		l.log.Warn(ctx, fmt.Sprintf("读取qa.session.ttl失败: %s，使用默认值", xErr.GetMessage()))
		ttlStr = "604800"
	}
	ttl, _ := strconv.Atoi(ttlStr)
	if ttl <= 0 {
		ttl = 604800
	}

	// 读取运行时域名（读失败兜底 localhost）
	domain, xErr := l.repo.info.GetByKey(ctx, "runtime.domain")
	if xErr != nil {
		l.log.Warn(ctx, fmt.Sprintf("读取runtime.domain失败: %s，使用默认值", xErr.GetMessage()))
		domain = "http://localhost:3000"
	}

	// 读取 qa_get_answer 单次阻塞上限（分片超时）
	sliceStr, xErr := l.repo.info.GetByKey(ctx, "qa.get_answer.poll_slice")
	if xErr != nil {
		l.log.Warn(ctx, fmt.Sprintf("读取qa.get_answer.poll_slice失败: %s，使用默认值", xErr.GetMessage()))
		sliceStr = "25"
	}
	pollSlice, _ := strconv.Atoi(sliceStr)
	if pollSlice <= 0 {
		pollSlice = 25
	}

	// 读取 qa_get_answer 最大重试次数
	retryStr, xErr := l.repo.info.GetByKey(ctx, "qa.get_answer.max_retries")
	if xErr != nil {
		l.log.Warn(ctx, fmt.Sprintf("读取qa.get_answer.max_retries失败: %s，使用默认值", xErr.GetMessage()))
		retryStr = "36"
	}
	maxRetries, _ := strconv.Atoi(retryStr)
	if maxRetries <= 0 {
		maxRetries = 36
	}

	return &qa.QaConfigResponse{
		SessionTTL:    ttl,
		RuntimeDomain: domain,
		PollSlice:     pollSlice,
		MaxRetries:    maxRetries,
	}, nil
}

// UpdateQaConfig 更新Q&A配置
func (l *QaLogic) UpdateQaConfig(ctx context.Context, req *qa.UpdateQaConfigRequest) (*qa.QaConfigResponse, *xError.Error) {
	l.log.Info(ctx, "UpdateQaConfig - 更新Q&A配置")

	// 更新 Session TTL
	if req.SessionTTL != nil {
		if xErr := l.repo.info.UpdateValue(ctx, "qa.session.ttl", strconv.Itoa(*req.SessionTTL)); xErr != nil {
			return nil, xError.NewError(ctx, xError.DatabaseError, "更新Session TTL失败", false, nil)
		}
	}

	// 更新运行时域名
	if req.RuntimeDomain != nil {
		if xErr := l.repo.info.UpdateValue(ctx, "runtime.domain", *req.RuntimeDomain); xErr != nil {
			return nil, xError.NewError(ctx, xError.DatabaseError, "更新运行时域名失败", false, nil)
		}
	}

	// 更新 qa_get_answer 单次阻塞上限（必须小于 MCP 客户端 tool 超时约 30s）
	if req.PollSlice != nil {
		if *req.PollSlice < 1 || *req.PollSlice > 28 {
			return nil, xError.NewError(ctx, xError.BusinessError, "poll_slice 必须在 1-28 秒之间（需小于客户端 tool 超时）", false, nil)
		}
		if xErr := l.repo.info.UpdateValue(ctx, "qa.get_answer.poll_slice", strconv.Itoa(*req.PollSlice)); xErr != nil {
			return nil, xError.NewError(ctx, xError.DatabaseError, "更新poll_slice失败", false, nil)
		}
	}

	// 更新 qa_get_answer 最大重试次数（必须 ≥1）
	if req.MaxRetries != nil {
		if *req.MaxRetries < 1 {
			return nil, xError.NewError(ctx, xError.BusinessError, "max_retries 必须至少为 1", false, nil)
		}
		if xErr := l.repo.info.UpdateValue(ctx, "qa.get_answer.max_retries", strconv.Itoa(*req.MaxRetries)); xErr != nil {
			return nil, xError.NewError(ctx, xError.DatabaseError, "更新max_retries失败", false, nil)
		}
	}

	// 返回更新后的完整配置
	return l.GetQaConfig(ctx)
}

// GetQuestionByID 根据 ID 获取问题实体，供 MCP 层查询问题选项等场景使用
//
// 参数:
//   - ctx:       上下文对象
//   - questionID: 问题雪花 ID
//
// 返回值:
//   - *entity.QaQuestion: 问题实体
//   - *xError.Error:      查询过程中的错误
func (l *QaLogic) GetQuestionByID(ctx context.Context, questionID xSnowflake.SnowflakeID) (*entity.QaQuestion, *xError.Error) {
	return l.repo.question.GetByID(ctx, questionID)
}
