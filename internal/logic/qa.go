package logic

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	xCache "github.com/bamboo-services/bamboo-base-go/major/cache"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xCtxUtil "github.com/bamboo-services/bamboo-base-go/common/utility/context"
	"github.com/xiaolfeng/Lumina/api/qa"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"github.com/xiaolfeng/Lumina/internal/repository"
	"github.com/xiaolfeng/Lumina/internal/repository/cache"
	"github.com/xiaolfeng/Lumina/internal/service"
	qaQueue "github.com/xiaolfeng/Lumina/internal/qa"
	"gorm.io/datatypes"
)

// qaRepo QA模块依赖的仓储集合
type qaRepo struct {
	session      *repository.QaSessionRepo
	question     *repository.QaQuestionRepo
	supplement   *repository.QaSupplementRepo
	info         *repository.InfoRepo
	retryCache   *cache.QaRetryCache
	downloadToken *service.DownloadTokenService
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
	page := req.Page
	size := req.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 20
	}

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
		key := strconv.FormatInt(s.TargetID.Int64(), 10)
		supplementMap[key] = append(supplementMap[key], qa.SupplementResponse{
			ID:          s.ID.String(),
			TargetType:  s.TargetType,
			TargetID:    key,
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
		ID:            session.ID.String(),
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
				ID:          supplement.ID.String(),
				TargetType:  supplement.TargetType,
				TargetID:    strconv.FormatInt(supplement.TargetID.Int64(), 10),
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
		ID:          question.ID.String(),
		SessionID:   sessionID,
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

// ─── MCP Tool Business Methods ──────────────────────────────────────────

// CreateSession 创建QA会话（MCP工具）
func (l *QaLogic) CreateSession(ctx context.Context, title, agent, sessionType, projectID string) (string, string, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("CreateSession - 创建QA会话 [%s]", title))

	// 解析项目ID
	var prjID xSnowflake.SnowflakeID
	if projectID != "" {
		parsedPrjID, parseErr := xSnowflake.ParseSnowflakeID(projectID)
		if parseErr != nil {
			return "", "", xError.NewError(ctx, xError.BusinessError, "无效的项目ID格式", false, nil)
		}
		prjID = parsedPrjID
	}

	// 生成雪花ID
	id := xSnowflake.GenerateID(bConst.GeneQaSession)

	// 生成 Hash（SHA256 前 16 位 hex，碰撞检测重试）
	hash := generateSessionHash(id)
	for {
		// 碰撞检测
		if _, xErr := l.repo.session.GetByHash(ctx, hash); xErr != nil {
			// 未找到说明无碰撞，跳出
			break
		}
		// 碰撞：重新生成 ID 和 Hash
		id = xSnowflake.GenerateID(bConst.GeneQaSession)
		hash = generateSessionHash(id)
	}

	// 处理会话类型和过期时间
	sessionType = strings.ToLower(strings.TrimSpace(sessionType))
	if sessionType == "" || sessionType == "temporary" {
		sessionType = "temporary"
		expiresAt := time.Now().Add(48 * time.Hour)
		entity := &entity.QaSession{
			BaseEntity: xModels.BaseEntity{ID: id},
			Title:      title,
			Agent:      agent,
			Owner:      "system",
			Type:       sessionType,
			Status:     "active",
			ProjectID:  prjID,
			Hash:       hash,
			ExpiresAt:  &expiresAt,
		}
		if xErr := l.repo.session.Create(ctx, entity); xErr != nil {
			return "", "", xErr
		}
	} else {
		// permanent 类型无过期时间
		entity := &entity.QaSession{
			BaseEntity: xModels.BaseEntity{ID: id},
			Title:      title,
			Agent:      agent,
			Owner:      "system",
			Type:       sessionType,
			Status:     "active",
			ProjectID:  prjID,
			Hash:       hash,
			ExpiresAt:  nil,
		}
		if xErr := l.repo.session.Create(ctx, entity); xErr != nil {
			return "", "", xErr
		}
	}

	// 读取运行时域名配置，生成浏览器链接
	domain, xErr := l.repo.info.GetByKey(ctx, "runtime.domain")
	if xErr != nil || domain == "" {
		domain = "http://localhost:3000"
	}
	link := fmt.Sprintf("%s/interact?session=%s", strings.TrimRight(domain, "/"), hash)

	return id.String(), link, nil
}

// PushQuestion 推送问题到会话（MCP工具）
func (l *QaLogic) PushQuestion(ctx context.Context, sessionID, qType, title, description string, options, config, batch any, groupLabel string, supplement bool) (string, map[string]string, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("PushQuestion - 推送问题 [session=%s, title=%s]", sessionID, title))

	// 解析并验证会话
	parsedSID, err := xSnowflake.ParseSnowflakeID(sessionID)
	if err != nil {
		return "", nil, xError.NewError(ctx, xError.BusinessError, "无效的会话ID", false, nil)
	}

	// 验证会话存在且为活跃状态
	session, xErr := l.repo.session.GetByID(ctx, parsedSID)
	if xErr != nil {
		return "", nil, xErr
	}
	if session.Status != "active" {
		return "", nil, xError.NewError(ctx, xError.BusinessError, "会话不是活跃状态，无法推送问题", false, nil)
	}

	// ── 题型参数校验（P-01, P-15）──
	requiresOptions := map[string]bool{
		"select": true, "multi-select": true, "rank": true,
		"rate": true, "options": true,
	}
	if requiresOptions[qType] {
		if options == nil {
			return "", nil, xError.NewError(ctx, xError.BusinessError,
				xError.ErrMessage(fmt.Sprintf("题型 %s 必须提供 options 参数", qType)), false, nil)
		}
		var optList []map[string]interface{}
		switch v := options.(type) {
		case string:
			_ = json.Unmarshal([]byte(v), &optList)
		default:
			optBytes, _ := json.Marshal(options)
			_ = json.Unmarshal(optBytes, &optList)
		}
		if len(optList) == 0 {
			return "", nil, xError.NewError(ctx, xError.BusinessError,
				xError.ErrMessage(fmt.Sprintf("题型 %s 的 options 参数不能为空", qType)), false, nil)
		}
		// options 题型额外校验 pros/cons（P-01）
		if qType == "options" {
			for _, opt := range optList {
				pros, _ := opt["pros"].([]interface{})
				cons, _ := opt["cons"].([]interface{})
				if len(pros) == 0 || len(cons) == 0 {
					return "", nil, xError.NewError(ctx, xError.BusinessError,
						"options 题型的每个选项必须包含 pros 和 cons（用于差异化对比）", false, nil)
				}
			}
		}
	}

	// 生成问题ID
	qID := xSnowflake.GenerateID(bConst.GeneQaQuestion)

	// 构建问题实体
	questionEntity := &entity.QaQuestion{
		BaseEntity:  xModels.BaseEntity{ID: qID},
		SessionID:   parsedSID,
		Type:        qType,
		Title:       title,
		Description: description,
		Options:     toJSON(options),
		Config:      toJSON(config),
		Batch:       toJSON(batch),
		GroupLabel:  groupLabel,
		Supplement:  supplement,
		Status:      "pending",
	}

	// 为每个 option 生成雪花 ID 并构建映射
	optionIDMap := make(map[string]string)
	if options != nil {
		var rawOpts string
		switch v := options.(type) {
		case string:
			rawOpts = v
		default:
			if optBytes, jsonErr := json.Marshal(options); jsonErr == nil {
				rawOpts = string(optBytes)
			}
		}
		if rawOpts != "" {
			var optList []map[string]interface{}
			if json.Unmarshal([]byte(rawOpts), &optList) == nil {
				for i, opt := range optList {
					optID := xSnowflake.GenerateID(bConst.GeneQaQuestion).String()
					opt["id"] = optID
					optList[i] = opt
					if label, ok := opt["label"].(string); ok {
						optionIDMap[label] = optID
					}
				}
				if updated, jsonErr := json.Marshal(optList); jsonErr == nil {
					questionEntity.Options = datatypes.JSON(updated)
				}
			}
		}
	}

	// 持久化
	if xErr := l.repo.question.Create(ctx, questionEntity); xErr != nil {
		return "", nil, xErr
	}

	// 通知 WebSocket 层广播新问题到在线设备
	if OnQuestionPushed != nil {
		OnQuestionPushed(sessionID, questionEntity)
	}

	return qID.String(), optionIDMap, nil
}

// PushSupplement 推送补充内容（MCP工具）
func (l *QaLogic) PushSupplement(ctx context.Context, sessionID, targetType string, targetID xSnowflake.SnowflakeID, contentType, content string) *xError.Error {
	l.log.Info(ctx, fmt.Sprintf("PushSupplement - 推送补充内容 [session=%s, target=%s/%d]", sessionID, targetType, targetID))

	// 解析会话ID
	parsedSID, err := xSnowflake.ParseSnowflakeID(sessionID)
	if err != nil {
		return xError.NewError(ctx, xError.BusinessError, "无效的会话ID", false, nil)
	}

	// 生成补充ID
	sID := xSnowflake.GenerateID(bConst.GeneQaSupplement)

	supplementEntity := &entity.QaSupplement{
		BaseEntity:  xModels.BaseEntity{ID: sID},
		SessionID:   parsedSID,
		TargetType:  targetType,
		TargetID:    targetID,
		ContentType: contentType,
		Content:     content,
	}

	// 创建或覆写
	result, xErr := l.repo.supplement.CreateOrUpdate(ctx, supplementEntity)
	if xErr != nil {
		return xErr
	}

	// 通知 WebSocket 层广播补充内容到在线设备
	if OnSupplementPushed != nil {
		OnSupplementPushed(sessionID, result)
	}

	return nil
}

// GetAnswer 阻塞获取回答（MCP工具）
func (l *QaLogic) GetAnswer(ctx context.Context, sessionID string, timeout time.Duration) (string, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("GetAnswer - 阻塞获取回答 [session=%s, timeout=%s]", sessionID, timeout))

	// 从队列消费
	answers, err := l.queue.Consume(ctx, sessionID, timeout)
	if err != nil {
		return "", xError.NewError(ctx, xError.BusinessError, "获取回答被中断", false, err)
	}

	// 无回答
	if len(answers) == 0 {
		return "", nil
	}

	// 成功消费回答，重置该会话的重试计数器
	if resetErr := l.repo.retryCache.Reset(ctx, sessionID); resetErr != nil {
		l.log.Warn(ctx, fmt.Sprintf("GetAnswer - 重置重试计数器失败（忽略）: %s", resetErr.Error()))
	}

	// 格式化回答字符串（需要查询问题元数据）
	return l.formatAnswerString(ctx, answers), nil
}

// IncrementGetAnswerRetry 增加指定会话的 qa_get_answer 重试计数器
// 返回当前计数（递增后的值）。首次调用时自动设置过期时间为 Session TTL。
func (l *QaLogic) IncrementGetAnswerRetry(ctx context.Context, sessionID string, ttlSeconds int) (int, *xError.Error) {
	ttl := time.Duration(ttlSeconds) * time.Second
	count, err := l.repo.retryCache.Increment(ctx, sessionID, ttl)
	if err != nil {
		return 0, xError.NewError(ctx, xError.UnknownError, "增加重试计数器失败", false, err)
	}
	return int(count), nil
}

// RegetAnswer 重新获取已回答的问题（MCP工具）
func (l *QaLogic) RegetAnswer(ctx context.Context, sessionID, questionID string, base64Flag bool) (string, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("RegetAnswer - 重新获取回答 [session=%s, question=%s]", sessionID, questionID))

	// 解析问题ID
	parsedQID, err := xSnowflake.ParseSnowflakeID(questionID)
	if err != nil {
		return "", xError.NewError(ctx, xError.BusinessError, "无效的问题ID", false, nil)
	}

	// 查询问题
	question, xErr := l.repo.question.GetByID(ctx, parsedQID)
	if xErr != nil {
		return "", xErr
	}

	// 检查是否已回答
	if question.Status != "answered" {
		return "该问题用户暂未回答，请使用 qa_get_answer 来获取用户最近的一次回答。", nil
	}

	// 如果需要 base64 且有媒体数据
	if base64Flag && question.Media != nil {
		mediaBytes, jsonErr := json.Marshal(question.Media)
		if jsonErr == nil && len(mediaBytes) > 0 {
			return fmt.Sprintf("--- question:%s ---\n[MEDIA] %s", questionID, string(mediaBytes)), nil
		}
	}

	// 构造单条回答格式
	var answerData interface{}
	if question.Answer != nil {
		_ = json.Unmarshal(question.Answer, &answerData)
	}
	questionType, questionOptions, config, description := parseQuestionMetadata(question)
	fmtCtx := AnswerFormatContext{
		Description:       description,
		Config:            config,
		OptionSupplements: l.buildOptionSupplementMap(ctx, sessionID, questionID),
	}
	summary := formatAnswerData(questionID, answerData, questionType, questionOptions, fmtCtx)
	// image/file 类型：追加 OTP 下载令牌（P-11）
	summary = l.enhanceMediaAnswerWithOTP(ctx, questionType, sessionID, summary, answerData)
	return fmt.Sprintf("--- question:%s ---\n[ANSWER] %s", questionID, summary), nil
}

// RegetAnswers 批量重新获取回答（MCP工具）
//
// 遍历所有 questionIDs，对每个执行：
// - 已回答 → "[ANSWERED] {question_id}: {答案摘要}"
// - 未回答 → "[PENDING] {question_id}: 该问题用户暂未回答，请使用 qa_get_answer 来获取用户最近的一次回答。"
// - 无效ID → "[ERROR] {question_id}: 无效的问题ID格式"
// - 不存在 → "[ERROR] {question_id}: 问题不存在"
func (l *QaLogic) RegetAnswers(ctx context.Context, sessionID string, questionIDs []string) (string, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("RegetAnswers - 批量重新获取回答 [session=%s, count=%d]", sessionID, len(questionIDs)))

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("批量查询结果（%d 个问题）:\n\n", len(questionIDs)))

	for _, qid := range questionIDs {
		parsedQID, err := xSnowflake.ParseSnowflakeID(qid)
		if err != nil {
			sb.WriteString(fmt.Sprintf("--- question:%s ---\n[ERROR] 无效的问题ID格式\n", qid))
			continue
		}

		question, xErr := l.repo.question.GetByID(ctx, parsedQID)
		if xErr != nil {
			sb.WriteString(fmt.Sprintf("--- question:%s ---\n[ERROR] 问题不存在\n", qid))
			continue
		}

		var answerData interface{}
		if question.Answer != nil {
			_ = json.Unmarshal(question.Answer, &answerData)
		}
		questionType, questionOptions, config, description := parseQuestionMetadata(question)
		fmtCtx := AnswerFormatContext{
			Description:       description,
			Config:            config,
			OptionSupplements: l.buildOptionSupplementMap(ctx, sessionID, qid),
		}
		summary := formatAnswerData(qid, answerData, questionType, questionOptions, fmtCtx)
		// image/file 类型：追加 OTP 下载令牌（P-11）
		summary = l.enhanceMediaAnswerWithOTP(ctx, questionType, sessionID, summary, answerData)
		if question.Status == "answered" {
			sb.WriteString(fmt.Sprintf("--- question:%s ---\n[ANSWER] %s\n", qid, summary))
		} else {
			sb.WriteString(fmt.Sprintf("--- question:%s ---\n[PENDING] 该问题用户暂未回答，请使用 qa_get_answer 来获取用户最近的一次回答。\n", qid))
		}
	}

	return sb.String(), nil
}

// GetSessionMCP 获取会话信息（MCP工具，返回人类可读字符串）
func (l *QaLogic) GetSessionMCP(ctx context.Context, sessionID string) (string, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("GetSessionMCP - 获取会话信息 [%s]", sessionID))

	// 解析会话ID
	parsedSID, err := xSnowflake.ParseSnowflakeID(sessionID)
	if err != nil {
		return "", xError.NewError(ctx, xError.BusinessError, "无效的会话ID", false, nil)
	}

	// 查询会话及问题
	session, questions, xErr := l.repo.session.GetByIDWithQuestions(ctx, parsedSID)
	if xErr != nil {
		return "", xErr
	}

	// 构建人类可读字符串
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("会话ID: %s\n", session.ID.String()))
	sb.WriteString(fmt.Sprintf("标题: %s\n", session.Title))
	sb.WriteString(fmt.Sprintf("Agent: %s\n", session.Agent))
	sb.WriteString(fmt.Sprintf("类型: %s\n", session.Type))
	sb.WriteString(fmt.Sprintf("状态: %s\n", session.Status))
	sb.WriteString(fmt.Sprintf("关联项目: %d\n", session.ProjectID.Int64()))
	sb.WriteString(fmt.Sprintf("在线设备: %d\n", session.OnlineDevices))
	if session.ExpiresAt != nil {
		sb.WriteString(fmt.Sprintf("过期时间: %s\n", session.ExpiresAt.Format(time.RFC3339)))
	}
	sb.WriteString(fmt.Sprintf("创建时间: %s\n", session.CreatedAt.Format(time.RFC3339)))
	sb.WriteString("\n问题列表:\n")

	if len(questions) == 0 {
		sb.WriteString("  （暂无问题）\n")
	} else {
		for i, q := range questions {
			sb.WriteString(fmt.Sprintf("  %d. [%s] %s - %s\n", i+1, q.Type, q.Title, q.Status))
		}
	}

	return sb.String(), nil
}

// DeleteSessionMCP 删除会话（MCP工具）
func (l *QaLogic) DeleteSessionMCP(ctx context.Context, sessionID string) *xError.Error {
	return l.DeleteSession(ctx, sessionID)
}

// ArchiveSession 归档会话（MCP工具），将 active 会话转为 expired 只读状态。
func (l *QaLogic) ArchiveSession(ctx context.Context, sessionID string) *xError.Error {
	l.log.Info(ctx, fmt.Sprintf("ArchiveSession - 归档会话 [%s]", sessionID))

	parsedID, err := xSnowflake.ParseSnowflakeID(sessionID)
	if err != nil {
		return xError.NewError(ctx, xError.BusinessError, "无效的会话ID", false, nil)
	}

	session, xErr := l.repo.session.GetByID(ctx, parsedID)
	if xErr != nil {
		return xErr
	}
	if session.Status != "active" {
		return xError.NewError(ctx, xError.BusinessError, "仅 active 状态的会话可以归档", false, nil)
	}

	if xErr := l.repo.session.UpdateStatus(ctx, parsedID, "expired"); xErr != nil {
		return xError.NewError(ctx, xError.UnknownError, "归档会话失败", false, xErr)
	}

	// P-16: 归档后清除会话 Redis 缓存，避免后续读取到 active 状态的脏数据
	if cacheErr := l.repo.session.ClearCache(ctx, parsedID); cacheErr != nil {
		l.log.Warn(ctx, fmt.Sprintf("ArchiveSession - 归档后清除缓存失败（忽略）: %s", cacheErr.Error()))
	}

	// P-16: 清理会话级文件缓存（base64 媒体文件等）
	fileCacheSvc := service.NewFileCacheService()
	if cleanErr := fileCacheSvc.CleanSession(ctx, sessionID); cleanErr != nil {
		l.log.Warn(ctx, fmt.Sprintf("ArchiveSession - 归档后清理文件缓存失败（忽略）: %s", cleanErr.Error()))
	}

	l.queue.RemoveQueue(sessionID)
	if resetErr := l.repo.retryCache.Reset(ctx, sessionID); resetErr != nil {
		l.log.Warn(ctx, fmt.Sprintf("ArchiveSession - 重置重试计数器失败（忽略）: %s", resetErr.Error()))
	}

	if OnSessionArchived != nil {
		OnSessionArchived(sessionID)
	}
	return nil
}

// CancelQuestion 取消指定问题或会话的全部待回答问题（MCP工具 qa_cancel_question）
//
// 行为：
//   - cancelAll=false：取消单个问题（questionID 指定），状态 pending → cancelled
//   - cancelAll=true：取消该会话全部 pending 问题，清空回答队列
//
// 已回答（answered/skipped/cancelled）的问题跳过，仅 pending 状态可取消。
// 取消后通过 OnQuestionCancelled 回调广播通知在线设备（回调由 WebSocket 层注入）。
//
// 参数:
//   - ctx:        上下文
//   - sessionID:  会话 ID
//   - questionID: 问题 ID（cancelAll=false 时使用，cancelAll=true 时忽略）
//   - cancelAll:  true=取消全部 pending，false=取消指定问题
//
// 返回值:
//   - cancelled: 成功取消的问题数
//   - skipped:   跳过的问题数（非 pending 状态）
//   - *xError.Error: 操作错误
func (l *QaLogic) CancelQuestion(ctx context.Context, sessionID, questionID string, cancelAll bool) (int, int, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("CancelQuestion - 取消问题 [session=%s, question=%s, all=%v]", sessionID, questionID, cancelAll))

	// 解析会话ID
	parsedSID, err := xSnowflake.ParseSnowflakeID(sessionID)
	if err != nil {
		return 0, 0, xError.NewError(ctx, xError.BusinessError, "无效的会话ID", false, nil)
	}

	// 验证会话存在且为活跃状态
	session, xErr := l.repo.session.GetByID(ctx, parsedSID)
	if xErr != nil {
		return 0, 0, xErr
	}
	if session.Status != "active" {
		return 0, 0, xError.NewError(ctx, xError.BusinessError, "会话不是活跃状态，无法取消问题", false, nil)
	}

	cancelled := 0
	skipped := 0

	if cancelAll {
		// 取消全部 pending 问题
		questions, xErr := l.repo.question.GetPendingBySessionID(ctx, parsedSID)
		if xErr != nil {
			return 0, 0, xError.NewError(ctx, xError.UnknownError, "查询待回答问题失败", false, xErr)
		}
		for _, q := range questions {
			if q.Status == "pending" {
				if updateErr := l.repo.question.UpdateStatus(ctx, q.ID, "cancelled"); updateErr != nil {
					l.log.Warn(ctx, fmt.Sprintf("CancelQuestion - 取消问题失败 [id=%s]: %s", q.ID.String(), updateErr.Error()))
					skipped++
				} else {
					cancelled++
				}
			} else {
				skipped++
			}
		}

		// 清空回答队列（移除所有待消费回答）
		l.queue.RemoveQueue(sessionID)

		// 清除重试计数器
		if resetErr := l.repo.retryCache.Reset(ctx, sessionID); resetErr != nil {
			l.log.Warn(ctx, fmt.Sprintf("CancelQuestion - 重置重试计数器失败（忽略）: %s", resetErr.Error()))
		}

		// WebSocket 通知：问题全部取消
		if OnQuestionCancelled != nil {
			OnQuestionCancelled(sessionID, nil)
		}
	} else {
		// 取消单个问题
		parsedQID, err := xSnowflake.ParseSnowflakeID(questionID)
		if err != nil {
			return 0, 0, xError.NewError(ctx, xError.BusinessError, "无效的问题ID", false, nil)
		}
		question, xErr := l.repo.question.GetByID(ctx, parsedQID)
		if xErr != nil {
			return 0, 0, xErr
		}
		if question.Status != "pending" {
			return 0, 1, nil // 非 pending 状态跳过
		}
		if updateErr := l.repo.question.UpdateStatus(ctx, parsedQID, "cancelled"); updateErr != nil {
			return 0, 0, xError.NewError(ctx, xError.UnknownError, "取消问题失败", false, updateErr)
		}
		cancelled = 1

		// WebSocket 通知：单个问题取消
		if OnQuestionCancelled != nil {
			OnQuestionCancelled(sessionID, question)
		}
	}

	return cancelled, skipped, nil
}

// ─── Helper Methods ─────────────────────────────────────────────────────

// toSessionResponse 将会话实体映射为响应 DTO
func toSessionResponse(session *entity.QaSession) qa.SessionResponse {
	return qa.SessionResponse{
		ID:            session.ID.String(),
		Hash:          session.Hash,
		Title:         session.Title,
		Agent:         session.Agent,
		Type:          session.Type,
		Status:        session.Status,
		OnlineDevices: session.OnlineDevices,
		ExpiresAt:     formatTimePtr(session.ExpiresAt),
		CreatedAt:     session.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     session.UpdatedAt.Format(time.RFC3339),
	}
}

// toQuestionSummary 将问题实体映射为摘要 DTO
//
// supplements 为该问题关联的补充内容（按 question 维度分组后注入），
// 支持 Interact 页面刷新后恢复完整历史问答（含回答、选项、补充内容）。
//
// P-21 安全过滤：HTML 格式 supplement 仅用于浏览器渲染，不返回给 AI；
// markdown 格式 supplement 检测危险 HTML 标签（<script>/<iframe> 等），含危险标签则跳过。
func toQuestionSummary(q *entity.QaQuestion, supplements []qa.SupplementResponse) qa.QuestionSummaryResponse {
	// P-21: 过滤 supplement — 仅保留安全的 markdown 格式内容给 AI
	filtered := make([]qa.SupplementResponse, 0, len(supplements))
	for _, s := range supplements {
		if s.ContentType == "html" {
			continue
		}
		if containsDangerousTags(s.Content) {
			continue
		}
		filtered = append(filtered, s)
	}

	return qa.QuestionSummaryResponse{
		ID:          q.ID.String(),
		Type:        q.Type,
		Title:       q.Title,
		Options:     jsonOrNull(q.Options),
		Config:      jsonOrNull(q.Config),
		Batch:       jsonOrNull(q.Batch),
		GroupLabel:  q.GroupLabel,
		Supplement:  q.Supplement,
		Status:      q.Status,
		Answer:      jsonOrNull(q.Answer),
		Media:       jsonOrNull(q.Media),
		Supplements: filtered,
		CreatedAt:   q.CreatedAt.Format(time.RFC3339),
		AnsweredAt:  formatTimePtr(q.AnsweredAt),
	}
}

// containsDangerousTags 检测内容中是否包含危险的 HTML 标签
//
// 检测 <script>、<style>、<iframe>、<object>、<embed>、<link>、<meta> 等
// 可能导致 XSS 或内容注入的标签。大小写不敏感，匹配标签前缀（如 <script 可匹配 <script>、<script src=...>）。
func containsDangerousTags(content string) bool {
	lower := strings.ToLower(content)
	dangerousTags := []string{"<script", "<style", "<iframe", "<object", "<embed", "<link", "<meta"}
	for _, tag := range dangerousTags {
		if strings.Contains(lower, tag) {
			return true
		}
	}
	return false
}

// formatAnswerString 将多条回答格式化为人类可读字符串（P-18 优化）
//
// 优化要点：skip/supplement 标记由 WebSocket handler 直接推入字符串（"[SKIPPED]" /
// "[NEED_SUPPLEMENT] ..."），正常回答推入 map。通过类型判断直接区分两种路径，
// 消除原来对每个回答先走一次 default 格式化再查 DB 再格式化的性能浪费。
// 同时为 select/multi-select 题型注入选项级 supplement 映射。
func (l *QaLogic) formatAnswerString(ctx context.Context, answers []qaQueue.Answer) string {
	var sb strings.Builder
	for i, a := range answers {
		if i > 0 {
			sb.WriteString("\n")
		}

		var marker, content string

		// skip/supplement 标记是字符串类型（WebSocket handler 直接推入字符串）
		if dataStr, ok := a.Data.(string); ok {
			switch {
			case dataStr == "[SKIPPED]":
				marker = "[SKIPPED]"
				content = "用户跳过了此问题"
			case strings.HasPrefix(dataStr, "[NEED_SUPPLEMENT]"):
				marker = "[NEED_SUPPLEMENT]"
				content = strings.TrimPrefix(dataStr, "[NEED_SUPPLEMENT] ")
			default:
				marker = "[ANSWER]"
				content = dataStr
			}
		} else {
			// 正常回答（map 类型）— 查询 DB 获取题型元数据后格式化
			marker = "[ANSWER]"
			parsedQID, err := xSnowflake.ParseSnowflakeID(a.QuestionID)
			if err == nil {
				question, xErr := l.repo.question.GetByID(ctx, parsedQID)
				if xErr == nil {
					questionType, questionOptions, config, description := parseQuestionMetadata(question)
					fmtCtx := AnswerFormatContext{
						Description:       description,
						Config:            config,
						OptionSupplements: l.buildOptionSupplementMap(ctx, question.SessionID.String(), a.QuestionID),
					}
					content = formatAnswerData(a.QuestionID, a.Data, questionType, questionOptions, fmtCtx)
					// image/file 类型：追加 OTP 下载令牌（P-11）
					content = l.enhanceMediaAnswerWithOTP(ctx, questionType, question.SessionID.String(), content, a.Data)
				} else {
					content = formatAnswerData(a.QuestionID, a.Data, "", nil, AnswerFormatContext{})
				}
			} else {
				content = formatAnswerData(a.QuestionID, a.Data, "", nil, AnswerFormatContext{})
			}
		}

		sb.WriteString(fmt.Sprintf("--- question:%s ---\n%s %s\n", a.QuestionID, marker, content))
	}
	return sb.String()
}

// AnswerFormatContext 封装格式化回答所需的 question 级上下文
type AnswerFormatContext struct {
	Description       string                // 问题级描述（question.description）
	Config            map[string]interface{} // 问题配置（plan sections / diff before-after 等）
	OptionSupplements map[string]string      // 选项级 supplement 映射（key=optionID, value=markdown 内容，HTML 已过滤）
}

// parseQuestionMetadata 从 QaQuestion 实体提取题型、选项列表、配置和描述
func parseQuestionMetadata(q *entity.QaQuestion) (string, []map[string]interface{}, map[string]interface{}, string) {
	if q == nil {
		return "", nil, nil, ""
	}
	questionType := q.Type
	var options []map[string]interface{}
	if q.Options != nil {
		_ = json.Unmarshal(q.Options, &options)
	}
	var config map[string]interface{}
	if q.Config != nil {
		_ = json.Unmarshal(q.Config, &config)
	}
	return questionType, options, config, q.Description
}

// buildOptionSupplementMap 查询问题关联的选项级 supplement，构建 map[optionID]content
// 仅包含 content_type=markdown 的 supplement（HTML 格式不返回给 AI）
func (l *QaLogic) buildOptionSupplementMap(ctx context.Context, sessionID string, questionID string) map[string]string {
	parsedSID, err := xSnowflake.ParseSnowflakeID(sessionID)
	if err != nil {
		return nil
	}
	supplements, xErr := l.repo.supplement.GetBySessionID(ctx, parsedSID)
	if xErr != nil {
		return nil
	}
	result := make(map[string]string)
	for _, s := range supplements {
		if s.TargetType != "option" || s.ContentType != "markdown" {
			continue
		}
		result[s.TargetID.String()] = s.Content
	}
	return result
}

// formatAnswerData 格式化单条回答数据
func formatAnswerData(questionID string, data any, questionType string, options []map[string]interface{}, fmtCtx AnswerFormatContext) string {
	if data == nil {
		return ""
	}

	m, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Sprintf("%v", data)
	}

	switch questionType {
	case "select":
		return formatSelectAnswer(m, options, fmtCtx)
	case "multi-select":
		return formatMultiSelectAnswer(m, options, fmtCtx)
	case "text":
		return formatTextAnswer(m)
	case "boolean":
		return formatBooleanAnswer(m)
	case "code":
		return formatCodeAnswer(m)
	case "image":
		return formatImageAnswer(m)
	case "file":
		return formatFileAnswer(m)
	case "plan":
		return formatPlanAnswer(m, fmtCtx.Config)
	case "diff":
		return formatDiffAnswer(m, fmtCtx.Config)
	case "review":
		return formatReviewAnswer(m)
case "options":
			return formatOptionsAnswer(m, options, fmtCtx)
	case "slider":
		return formatSliderAnswer(m)
	case "rank":
		return formatRankAnswer(m, options)
	case "rate":
		return formatRateAnswer(m, options)
	default:
		return formatGenericAnswer(m)
	}
}

// formatSelectAnswer 格式化单选回答
//
// 输出格式（P-05）：
//   - 用户选择：<label>
//   - [DESCRIPTION] <question.description>          （问题级，可选）
//   - [OPTION_DESCRIPTION] <option.description>      （选项级，可选）
//   - [SUPPLEMENT] <agent option supplement | other> （选项级补充或自定义输入，可选）
//
// 每级信息为空时整行省略。
func formatSelectAnswer(m map[string]interface{}, options []map[string]interface{}, fmtCtx AnswerFormatContext) string {
	selected, exists := m["selected"]
	if !exists {
		return fmt.Sprintf("%v", m)
	}

	selectedID := getOptionID(selected)
	label, optionDesc := resolveOptionLabel(selected, options)
	if label == "" {
		label = selectedID
	}

	var sb strings.Builder
	sb.WriteString("用户选择：" + label)

	if strings.TrimSpace(fmtCtx.Description) != "" {
		sb.WriteString("\n[DESCRIPTION] ")
		sb.WriteString(fmtCtx.Description)
	}
	if optionDesc != "" {
		sb.WriteString("\n[OPTION_DESCRIPTION] ")
		sb.WriteString(optionDesc)
	}
	if supp, ok := fmtCtx.OptionSupplements[selectedID]; ok && strings.TrimSpace(supp) != "" {
		sb.WriteString("\n[SUPPLEMENT] ")
		sb.WriteString(supp)
	}
	if other, ok := m["other"].(string); ok && strings.TrimSpace(other) != "" {
		sb.WriteString("\n[SUPPLEMENT] ")
		sb.WriteString(other)
	}

	return sb.String()
}

// formatMultiSelectAnswer 格式化多选回答
//
// 输出格式（P-08）：
//   - 用户选择 N 项
//   - [DESCRIPTION] <question.description>                  （问题级，可选）
//   - ---
//   - [OPTION] <label>                                      （逐项）
//   - [OPTION_DESCRIPTION] <option.description>             （选项级，可选）
//   - [SUPPLEMENT] <agent option supplement>                （选项级补充，可选）
//
// 每级信息为空时整行省略；选项之间以 `---` 分隔。
func formatMultiSelectAnswer(m map[string]interface{}, options []map[string]interface{}, fmtCtx AnswerFormatContext) string {
	selected, exists := m["selected"]
	if !exists {
		return fmt.Sprintf("%v", m)
	}

	var selectedList []interface{}
	switch v := selected.(type) {
	case []interface{}:
		selectedList = v
	default:
		selectedList = []interface{}{selected}
	}

	otherCount := 0
	if others, ok := m["other"].([]interface{}); ok {
		otherCount = len(others)
	}
	totalCount := len(selectedList) + otherCount

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("用户选择 %d 项", totalCount))

	if strings.TrimSpace(fmtCtx.Description) != "" {
		sb.WriteString("\n[DESCRIPTION] ")
		sb.WriteString(fmtCtx.Description)
	}

	for _, item := range selectedList {
		selectedID := getOptionID(item)
		label, optionDesc := resolveOptionLabel(item, options)
		if label == "" {
			label = selectedID
		}

		sb.WriteString("\n---")
		sb.WriteString("\n[OPTION] ")
		sb.WriteString(label)

		if optionDesc != "" {
			sb.WriteString("\n[OPTION_DESCRIPTION] ")
			sb.WriteString(optionDesc)
		}
		if supp, ok := fmtCtx.OptionSupplements[selectedID]; ok && strings.TrimSpace(supp) != "" {
			sb.WriteString("\n[SUPPLEMENT] ")
			sb.WriteString(supp)
		}
	}

	if others, ok := m["other"].([]interface{}); ok && len(others) > 0 {
		for _, other := range others {
			s, ok := other.(string)
			if !ok || strings.TrimSpace(s) == "" {
				continue
			}
			sb.WriteString("\n---")
			sb.WriteString("\n[OPTION] __other__")
			sb.WriteString("\n[SUPPLEMENT] ")
			sb.WriteString(s)
		}
	}

	return sb.String()
}

// resolveOptionLabel 从选项列表中解析选项的标签和描述
func resolveOptionLabel(selected any, options []map[string]interface{}) (string, string) {
	// 直接是 map（含 id + label）
	if selMap, ok := selected.(map[string]interface{}); ok {
		label, _ := selMap["label"].(string)
		desc, _ := selMap["description"].(string)
		return label, desc
	}

	// 纯 ID 字符串 → 从 options 反查
	selStr, ok := selected.(string)
	if !ok || len(options) == 0 {
		return fmt.Sprintf("%v", selected), ""
	}

	for _, opt := range options {
		if optID, _ := opt["id"].(string); optID == selStr {
			label, _ := opt["label"].(string)
			desc, _ := opt["description"].(string)
			return label, desc
		}
	}
	return selStr, ""
}

// resolveLabelByID 从选项列表反查 label，找不到则返回 ID 本身（降级）
//
// 用于 rank/rate 等仅持有 ID 列表的场景，避免直接向 Agent 暴露雪花 ID。
func resolveLabelByID(id string, options []map[string]interface{}) string {
	for _, opt := range options {
		if optID, _ := opt["id"].(string); optID == id {
			if label, _ := opt["label"].(string); label != "" {
				return label
			}
		}
	}
	return id
}

// getOptionID 从 selected 值中提取 optionID
//
// 兼容两种前端提交格式：
//   - string（纯 ID，主流格式）
//   - map[string]interface{}（含 id 字段，历史兼容）
func getOptionID(selected any) string {
	if idStr, ok := selected.(string); ok {
		return idStr
	}
	if selMap, ok := selected.(map[string]interface{}); ok {
		if id, ok := selMap["id"].(string); ok {
			return id
		}
	}
	return fmt.Sprintf("%v", selected)
}

// formatTextAnswer 格式化文本回答
func formatTextAnswer(m map[string]interface{}) string {
	if text, ok := m["text"].(string); ok {
		return text
	}
	return fmt.Sprintf("%v", m)
}

// formatBooleanAnswer 格式化布尔回答
func formatBooleanAnswer(m map[string]interface{}) string {
	if choice, ok := m["choice"].(string); ok {
		if choice == "yes" {
			return "是"
		}
		return "否"
	}
	if choice, ok := m["choice"].(bool); ok {
		if choice {
			return "是"
		}
		return "否"
	}
	return fmt.Sprintf("%v", m)
}

// formatCodeAnswer 格式化代码回答（P-10，从 formatMediaAnswer 拆分）
//
// 输出 code + 可选的 [LANGUAGE] 标记，便于 Agent 识别代码语言。
func formatCodeAnswer(m map[string]interface{}) string {
	code, ok := m["code"].(string)
	if !ok {
		return fmt.Sprintf("%v", m)
	}
	var sb strings.Builder
	sb.WriteString(code)
	if lang, ok := m["language"].(string); ok && lang != "" {
		sb.WriteString("\n[LANGUAGE] ")
		sb.WriteString(lang)
	}
	return sb.String()
}

// enhanceMediaAnswerWithOTP 为 image/file 类型的格式化回答追加 OTP 下载令牌链接
//
// 后处理注入策略：formatImageAnswer/formatFileAnswer 是纯函数，无法访问 Redis，
// 后处理注入策略：formatImageAnswer/formatFileAnswer 是纯函数，无法访问缓存层，
// 故在 QaLogic 方法层（通过注入的 downloadToken 服务）对格式化结果做后处理，为每个文件生成一次性令牌。
//
// 输出标记（P-11 核心交付物）：
//   - [DOWNLOAD_URL]  每行一个 OTP 下载链接（<domain>/api/v1/qa/download/<token>）
//   - [IMPORTANT]     一次性令牌提示
//   - [TIP]           curl/wget 下载指引
//   - [GIT_TIP]       .lumina/cache 忽略建议
//
// 参数:
//   - ctx: 上下文（用于 Redis 读写和 DB 查询）
//   - questionType: 题型（仅 "image"/"file" 触发增强，其他类型直接返回原字符串）
//   - sessionID: 会话 ID（当前未参与令牌生成，预留扩展）
//   - formatted: formatAnswerData 已格式化的回答字符串
//   - answerData: 原始回答数据（用于提取 images/files 列表）
//
// 返回增强后的字符串；若无需增强（非 media 类型 / 无文件 / 令牌生成失败）则原样返回。
func (l *QaLogic) enhanceMediaAnswerWithOTP(ctx context.Context, questionType, sessionID, formatted string, answerData any) string {
	// 仅 image/file 类型需要追加 OTP 令牌
	if questionType != "image" && questionType != "file" {
		return formatted
	}

	m, ok := answerData.(map[string]interface{})
	if !ok {
		return formatted
	}

	fieldKey := "images"
	if questionType == "file" {
		fieldKey = "files"
	}
	items, ok := m[fieldKey].([]interface{})
	if !ok || len(items) == 0 {
		return formatted
	}

	// 读取运行时域名配置（Info 表 key=runtime.domain）
	domain := "http://localhost:8080"
	if domainVal, xErr := l.repo.info.GetByKey(ctx, "runtime.domain"); xErr == nil && domainVal != "" {
		domain = strings.TrimRight(domainVal, "/")
	}

	// 为每个文件生成 OTP 令牌（使用注入的 downloadToken 服务，避免 logic 直连 Redis）
	var tokenUrls []string
	for _, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		filePath, _ := itemMap["filePath"].(string)
		if filePath == "" {
			continue
		}
		filename, _ := itemMap["filename"].(string)
		if filename == "" {
			filename, _ = itemMap["name"].(string)
		}
		mimeType, _ := itemMap["mimeType"].(string)

		token, err := l.repo.downloadToken.GenerateToken(ctx, filePath, filename, mimeType)
		if err != nil {
			l.log.Warn(ctx, fmt.Sprintf("enhanceMediaAnswerWithOTP - 生成下载令牌失败 [file=%s]: %v", filePath, err))
			continue
		}
		tokenUrls = append(tokenUrls, fmt.Sprintf("    - %s/api/v1/qa/download/%s", domain, token))
	}

	if len(tokenUrls) == 0 {
		return formatted
	}

	var sb strings.Builder
	sb.WriteString(formatted)
	sb.WriteString("\n[DOWNLOAD_URL]")
	for _, url := range tokenUrls {
		sb.WriteString("\n")
		sb.WriteString(url)
	}
	sb.WriteString("\n[IMPORTANT] 下载链接为一次性令牌，使用后立即失效。若下载失败需重新下载，请重新调用 qa_reget_answer 获取新的下载链接。")
	sb.WriteString("\n[TIP] 下载并保存文件时，最终路径 = DOWNLOAD_PATH + FILE_NAME：")
	sb.WriteString("\n  - Mac/Linux: curl -o \"<DOWNLOAD_PATH><FILE_NAME>\" <url>")
	sb.WriteString("\n  - Windows:   Invoke-WebRequest -Uri <url> -OutFile \"<DOWNLOAD_PATH><FILE_NAME>\"")
	sb.WriteString("\n  AI 引用该文件时，使用 <DOWNLOAD_PATH><FILE_NAME> 作为完整路径。")
	sb.WriteString("\n[GIT_TIP] 若存在 git 等版本管理项目，需要把 .lumina/cache/* 加入忽略上传（如 .gitignore 使用分类模式：\n    ### Lumina ###\n    .lumina/cache/）")

	return sb.String()
}

// formatImageAnswer 格式化图片回答（P-11，从 formatMediaAnswer 拆分）
//
// 注意：OTP 令牌生成（[DOWNLOAD_URL]/[TIP]/[IMPORTANT]）在 enhanceMediaAnswerWithOTP 中后处理注入，
// 本函数仅输出文件名和内部存储路径信息。
func formatImageAnswer(m map[string]interface{}) string {
	images, ok := m["images"].([]interface{})
	if !ok || len(images) == 0 {
		return fmt.Sprintf("%v", m)
	}

	var sb strings.Builder
	sb.WriteString("用户已上传内容")
	for _, img := range images {
		imgMap, ok := img.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := imgMap["filename"].(string)
		filePath, _ := imgMap["filePath"].(string)
		sb.WriteString(fmt.Sprintf("\n---\n[FILE_NAME] %s", name))
		if filePath != "" {
			// 输出目录级路径，去掉末尾 UUID 文件名，便于 Agent 拼接 FILE_NAME
			dir := filepath.Dir(filePath)
			if !strings.HasSuffix(dir, "/") {
				dir += "/"
			}
			sb.WriteString(fmt.Sprintf("\n[DOWNLOAD_PATH] %s", dir))
		}
	}
	return sb.String()
}

// formatFileAnswer 格式化文件回答（P-11，从 formatMediaAnswer 拆分）
//
// 注意：OTP 令牌生成（[DOWNLOAD_URL]/[TIP]/[IMPORTANT]）在 Task 12 完整集成，
// 当前 Wave 2 仅输出文件名和内部存储路径信息。兼容 filename 和 name 两种字段。
func formatFileAnswer(m map[string]interface{}) string {
	files, ok := m["files"].([]interface{})
	if !ok || len(files) == 0 {
		return fmt.Sprintf("%v", m)
	}

	var sb strings.Builder
	sb.WriteString("用户已上传内容")
	for _, f := range files {
		fMap, ok := f.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := fMap["filename"].(string)
		if name == "" {
			name, _ = fMap["name"].(string)
		}
		filePath, _ := fMap["filePath"].(string)
		sb.WriteString(fmt.Sprintf("\n---\n[FILE_NAME] %s", name))
		if filePath != "" {
			// 输出目录级路径，去掉末尾 UUID 文件名，便于 Agent 拼接 FILE_NAME
			dir := filepath.Dir(filePath)
			if !strings.HasSuffix(dir, "/") {
				dir += "/"
			}
			sb.WriteString(fmt.Sprintf("\n[DOWNLOAD_PATH] %s", dir))
		}
	}
	return sb.String()
}

// formatSliderAnswer 格式化滑块回答
func formatSliderAnswer(m map[string]interface{}) string {
	if value, ok := m["value"]; ok {
		return fmt.Sprintf("%v", value)
	}
	return fmt.Sprintf("%v", m)
}

// formatRankAnswer 格式化排名回答
//
// 兼容前端 "ranking"（主）和 "ranked"（向后兼容）两种字段名；
// 纯 ID 列表会通过 options 反查 label，避免向 Agent 暴露雪花 ID。
// 输出形如 `1. A → 2. B → 3. C`。
func formatRankAnswer(m map[string]interface{}, options []map[string]interface{}) string {
	var ranked []interface{}
	if v, ok := m["ranking"].([]interface{}); ok {
		ranked = v
	} else if v, ok := m["ranked"].([]interface{}); ok {
		ranked = v
	}
	if len(ranked) == 0 {
		return fmt.Sprintf("%v", m)
	}

	items := make([]string, 0, len(ranked))
	for i, item := range ranked {
		var label string
		if idStr, ok := item.(string); ok {
			label = resolveLabelByID(idStr, options)
		} else if rMap, ok := item.(map[string]interface{}); ok {
			if l, ok := rMap["label"].(string); ok && l != "" {
				label = l
			} else {
				label = fmt.Sprintf("%v", item)
			}
		} else {
			label = fmt.Sprintf("%v", item)
		}
		items = append(items, fmt.Sprintf("%d. %s", i+1, label))
	}
	return strings.Join(items, " → ")
}

// formatRateAnswer 格式化评分回答
//
// 按 options 顺序输出评分（保持稳定输出顺序，避免 map 随机迭代）；
// 空 ratings 或无 options 匹配时降级为原始 map 输出。
// 无评分数据时返回友好提示而非原始 JSON。
func formatRateAnswer(m map[string]interface{}, options []map[string]interface{}) string {
	ratings, ok := m["ratings"].(map[string]interface{})
	if !ok || len(ratings) == 0 {
		return "用户未提供评分"
	}

	// 按选项顺序输出（稳定排序），避免 map 随机迭代
	parts := make([]string, 0, len(ratings))
	for _, opt := range options {
		optID, _ := opt["id"].(string)
		if optID == "" {
			continue
		}
		if score, exists := ratings[optID]; exists {
			label, _ := opt["label"].(string)
			if label == "" {
				label = optID
			}
			parts = append(parts, fmt.Sprintf("%s: %v", label, score))
		}
	}

	// 降级：没有 options 或无匹配项时直接输出原始 map
	if len(parts) == 0 {
		for k, v := range ratings {
			parts = append(parts, fmt.Sprintf("%s: %v", k, v))
		}
	}
	return strings.Join(parts, ", ")
}

// formatGenericAnswer 未知题型的通用格式化
func formatGenericAnswer(m map[string]interface{}) string {
	if selected, ok := m["selected"]; ok {
		return fmt.Sprintf("%v", selected)
	}
	if text, ok := m["text"].(string); ok {
		return text
	}
	return fmt.Sprintf("%v", m)
}

// formatPlanAnswer 格式化 Plan 计划题回答（P-07 重写）
//
// 回答结构：{ decision: "approve"|"reject"|"revise", annotations?: [...], feedback?: "..." }
// - approve → 输出 [PLAN_DETAIL]，让 Agent 知道用户批准了什么
// - reject  → 仅提示拒绝
// - revise  → 输出 [REVISIONS] 逐章节修订意见
// config 来源：question.config（含 sections）
func formatPlanAnswer(m map[string]interface{}, config map[string]interface{}) string {
	decision, _ := m["decision"].(string)
	var sb strings.Builder

	switch decision {
	case "approve":
		sb.WriteString("用户已批准该计划")
		if sections, ok := config["sections"].([]interface{}); ok && len(sections) > 0 {
			sb.WriteString("\n[PLAN_DETAIL]")
			for i, sec := range sections {
				if secMap, ok := sec.(map[string]interface{}); ok {
					title, _ := secMap["title"].(string)
					content, _ := secMap["content"].(string)
					sb.WriteString(fmt.Sprintf("\n%d. %s\n   %s", i+1, title, content))
				}
			}
		}
	case "reject":
		sb.WriteString("用户已拒绝该计划")
	case "revise":
		sb.WriteString("用户要求修改该计划")
		if annotations, ok := m["annotations"].([]interface{}); ok && len(annotations) > 0 {
			sb.WriteString("\n[REVISIONS]")
			for i, ann := range annotations {
				if annMap, ok := ann.(map[string]interface{}); ok {
					sectionID, _ := annMap["sectionId"].(string)
					content, _ := annMap["content"].(string)
					if content != "" {
						sb.WriteString(fmt.Sprintf("\n%d. [%s] %s", i+1, sectionID, content))
					}
				}
			}
		}
	default:
		sb.WriteString(fmt.Sprintf("decision=%s", decision))
	}

	if feedback, ok := m["feedback"].(string); ok && strings.TrimSpace(feedback) != "" {
		sb.WriteString(fmt.Sprintf("\n[FEEDBACK] %s", feedback))
	}

	return sb.String()
}

// formatOptionsAnswer 格式化通用 options 题型回答（P-03 新增）
//
// options 题型与 select 行为相似，但无 [DESCRIPTION] 问题级描述输出，
// 主要输出选中选项的 label + 可选的选项级 description + 用户反馈。
func formatOptionsAnswer(m map[string]interface{}, options []map[string]interface{}, fmtCtx AnswerFormatContext) string {
	selected, exists := m["selected"]
	if !exists {
		return fmt.Sprintf("%v", m)
	}

	selectedID := getOptionID(selected)
	label, desc := resolveOptionLabel(selected, options)
	if label == "" {
		label = selectedID
	}

	var sb strings.Builder
	sb.WriteString(label)
	if desc != "" {
		sb.WriteString("\n[DESCRIPTION] ")
		sb.WriteString(desc)
	}
	// 选项级 supplement（Agent 为该选项推送的 markdown 内容）
	if supp, ok := fmtCtx.OptionSupplements[selectedID]; ok && strings.TrimSpace(supp) != "" {
		sb.WriteString("\n[SUPPLEMENT] ")
		sb.WriteString(supp)
	}
	// 用户的选择理由（feedback）
	if feedback, ok := m["feedback"].(string); ok && strings.TrimSpace(feedback) != "" {
		sb.WriteString("\n[SUPPLEMENT] ")
		sb.WriteString(feedback)
	}
	return sb.String()
}

// formatDiffAnswer 格式化 Diff 决策题回答（P-09，从 formatDecisionAnswer 拆分）
//
// 回答结构：{ decision: "approve"|"reject"|"edit", edited?: "...", feedback?: "..." }
// - approve → 输出 config.after 作为 [FINAL] 最终代码，让 Agent 知道实际写入内容
// - reject  → 仅提示拒绝
// - edit    → 输出用户编辑后的 m.edited 作为 [FINAL]
func formatDiffAnswer(m map[string]interface{}, config map[string]interface{}) string {
	decision, _ := m["decision"].(string)
	var sb strings.Builder

	switch decision {
	case "approve":
		sb.WriteString("用户已批准该修改")
		if after, ok := config["after"].(string); ok && after != "" {
			sb.WriteString("\n[FINAL]\n")
			sb.WriteString(after)
		}
	case "reject":
		sb.WriteString("用户已拒绝该修改")
	case "edit":
		sb.WriteString("用户修改后提交")
		if edited, ok := m["edited"].(string); ok && edited != "" {
			sb.WriteString("\n[FINAL]\n")
			sb.WriteString(edited)
		}
	default:
		sb.WriteString(fmt.Sprintf("decision=%s", decision))
	}

	if feedback, ok := m["feedback"].(string); ok && strings.TrimSpace(feedback) != "" {
		sb.WriteString(fmt.Sprintf("\n[FEEDBACK] %s", feedback))
	}

	return sb.String()
}

// formatReviewAnswer 格式化 Review 决策题回答（P-12，从 formatDecisionAnswer 拆分）
//
// 回答结构：{ decision: "approve"|"revise", feedback?: "..." }
// 去除 [已批准]/[已拒绝] 前缀（外层 [ANSWER] 已有标记，避免语义重复）。
func formatReviewAnswer(m map[string]interface{}) string {
	decision, _ := m["decision"].(string)
	var sb strings.Builder

	switch decision {
	case "approve":
		sb.WriteString("用户批准了该修改")
	case "revise":
		sb.WriteString("用户要求修改")
	default:
		sb.WriteString(fmt.Sprintf("decision=%s", decision))
	}

	if feedback, ok := m["feedback"].(string); ok && strings.TrimSpace(feedback) != "" {
		sb.WriteString(fmt.Sprintf("\n[FEEDBACK] %s", feedback))
	}

	return sb.String()
}

// formatTimePtr 格式化时间指针，nil 返回空字符串
func formatTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

// toJSON 将任意值序列化为 datatypes.JSON，nil 返回 nil
func toJSON(v any) datatypes.JSON {
	if v == nil {
		return nil
	}
	bytes, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return datatypes.JSON(bytes)
}

// jsonOrNull 将 datatypes.JSON 转换为 any（用于 DTO 赋值），nil 返回 nil
func jsonOrNull(data datatypes.JSON) any {
	if data == nil {
		return nil
	}
	return data
}

// generateSessionHash 基于雪花 ID 生成 16 位 hex 哈希标识
func generateSessionHash(id xSnowflake.SnowflakeID) string {
	sum := sha256.Sum256([]byte(id.String()))
	return hex.EncodeToString(sum[:])[:16]
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
