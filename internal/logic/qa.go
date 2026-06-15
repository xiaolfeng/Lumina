package logic

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xCtxUtil "github.com/bamboo-services/bamboo-base-go/common/utility/context"
	"github.com/xiaolfeng/Lumina/api/qa"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"github.com/xiaolfeng/Lumina/internal/repository"
	qaQueue "github.com/xiaolfeng/Lumina/internal/qa"
	"gorm.io/datatypes"
)

// qaRepo QA模块依赖的仓储集合
type qaRepo struct {
	session    *repository.QaSessionRepo
	question   *repository.QaQuestionRepo
	supplement *repository.QaSupplementRepo
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

// NewQaLogic 创建QaLogic实例
func NewQaLogic(ctx context.Context) *QaLogic {
	db := xCtxUtil.MustGetDB(ctx)
	rdb := xCtxUtil.MustGetRDB(ctx)

	return &QaLogic{
		logic: logic{
			db:  db,
			rdb: rdb,
			log: xLog.WithName(xLog.NamedLOGC, "QaLogic"),
		},
		repo: qaRepo{
			session:    repository.NewQaSessionRepo(db, rdb),
			question:   repository.NewQaQuestionRepo(db),
			supplement: repository.NewQaSupplementRepo(db),
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

	// 清理队列和重试计数器
	l.queue.RemoveQueue(id)
	retryKey := bConst.CacheQaGetAnswerRetry.Get(id).String()
	_ = l.rdb.Del(ctx, retryKey).Err()

	return nil
}

// GetQaConfig 获取Q&A配置
func (l *QaLogic) GetQaConfig(ctx context.Context) (*qa.QaConfigResponse, *xError.Error) {
	l.log.Info(ctx, "GetQaConfig - 获取Q&A配置")

	// 读取 Session TTL
	var ttlInfo entity.Info
	if err := l.db.WithContext(ctx).Where("\"key\" = ?", "qa.session.ttl").First(&ttlInfo).Error; err != nil {
		l.log.Warn(ctx, fmt.Sprintf("读取qa.session.ttl失败: %s", err.Error()))
		ttlInfo.Value = "604800" // 默认7天
	}
	ttl, _ := strconv.Atoi(ttlInfo.Value)
	if ttl <= 0 {
		ttl = 604800
	}

	// 读取运行时域名
	var domainInfo entity.Info
	if err := l.db.WithContext(ctx).Where("\"key\" = ?", "runtime.domain").First(&domainInfo).Error; err != nil {
		l.log.Warn(ctx, fmt.Sprintf("读取runtime.domain失败: %s", err.Error()))
		domainInfo.Value = "http://localhost:3000"
	}

	// 读取 qa_get_answer 单次阻塞上限（分片超时）
	var sliceInfo entity.Info
	if err := l.db.WithContext(ctx).Where("\"key\" = ?", "qa.get_answer.poll_slice").First(&sliceInfo).Error; err != nil {
		l.log.Warn(ctx, fmt.Sprintf("读取qa.get_answer.poll_slice失败: %s", err.Error()))
		sliceInfo.Value = "25"
	}
	pollSlice, _ := strconv.Atoi(sliceInfo.Value)
	if pollSlice <= 0 {
		pollSlice = 25
	}

	// 读取 qa_get_answer 最大重试次数
	var retryInfo entity.Info
	if err := l.db.WithContext(ctx).Where("\"key\" = ?", "qa.get_answer.max_retries").First(&retryInfo).Error; err != nil {
		l.log.Warn(ctx, fmt.Sprintf("读取qa.get_answer.max_retries失败: %s", err.Error()))
		retryInfo.Value = "36"
	}
	maxRetries, _ := strconv.Atoi(retryInfo.Value)
	if maxRetries <= 0 {
		maxRetries = 36
	}

	return &qa.QaConfigResponse{
		SessionTTL:    ttl,
		RuntimeDomain: domainInfo.Value,
		PollSlice:     pollSlice,
		MaxRetries:    maxRetries,
	}, nil
}

// UpdateQaConfig 更新Q&A配置
func (l *QaLogic) UpdateQaConfig(ctx context.Context, req *qa.UpdateQaConfigRequest) (*qa.QaConfigResponse, *xError.Error) {
	l.log.Info(ctx, "UpdateQaConfig - 更新Q&A配置")

	// 更新 Session TTL
	if req.SessionTTL != nil {
		if err := l.db.WithContext(ctx).
			Model(&entity.Info{}).
			Where("\"key\" = ?", "qa.session.ttl").
			Update("value", strconv.Itoa(*req.SessionTTL)).Error; err != nil {
			return nil, xError.NewError(ctx, xError.DatabaseError, "更新Session TTL失败", false, err)
		}
	}

	// 更新运行时域名
	if req.RuntimeDomain != nil {
		if err := l.db.WithContext(ctx).
			Model(&entity.Info{}).
			Where("\"key\" = ?", "runtime.domain").
			Update("value", *req.RuntimeDomain).Error; err != nil {
			return nil, xError.NewError(ctx, xError.DatabaseError, "更新运行时域名失败", false, err)
		}
	}

	// 更新 qa_get_answer 单次阻塞上限（必须小于 MCP 客户端 tool 超时约 30s）
	if req.PollSlice != nil {
		if *req.PollSlice < 1 || *req.PollSlice > 28 {
			return nil, xError.NewError(ctx, xError.BusinessError, "poll_slice 必须在 1-28 秒之间（需小于客户端 tool 超时）", false, nil)
		}
		if err := l.db.WithContext(ctx).
			Model(&entity.Info{}).
			Where("\"key\" = ?", "qa.get_answer.poll_slice").
			Update("value", strconv.Itoa(*req.PollSlice)).Error; err != nil {
			return nil, xError.NewError(ctx, xError.DatabaseError, "更新poll_slice失败", false, err)
		}
	}

	// 更新 qa_get_answer 最大重试次数（必须 ≥1）
	if req.MaxRetries != nil {
		if *req.MaxRetries < 1 {
			return nil, xError.NewError(ctx, xError.BusinessError, "max_retries 必须至少为 1", false, nil)
		}
		if err := l.db.WithContext(ctx).
			Model(&entity.Info{}).
			Where("\"key\" = ?", "qa.get_answer.max_retries").
			Update("value", strconv.Itoa(*req.MaxRetries)).Error; err != nil {
			return nil, xError.NewError(ctx, xError.DatabaseError, "更新max_retries失败", false, err)
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
	var domainInfo entity.Info
	if err := l.db.WithContext(ctx).Where("\"key\" = ?", "runtime.domain").First(&domainInfo).Error; err != nil || domainInfo.Value == "" {
		domainInfo.Value = "http://localhost:3000"
	}
	link := fmt.Sprintf("%s/interact?session=%s", strings.TrimRight(domainInfo.Value, "/"), hash)

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
	retryKey := bConst.CacheQaGetAnswerRetry.Get(sessionID).String()
	if delErr := l.rdb.Del(ctx, retryKey).Err(); delErr != nil {
		l.log.Warn(ctx, fmt.Sprintf("GetAnswer - 重置重试计数器失败（忽略）: %s", delErr.Error()))
	}

	// 格式化回答字符串（需要查询问题元数据）
	return l.formatAnswerString(ctx, answers), nil
}

// IncrementGetAnswerRetry 增加指定会话的 qa_get_answer 重试计数器
// 返回当前计数（递增后的值）。首次调用时自动设置过期时间为 Session TTL。
func (l *QaLogic) IncrementGetAnswerRetry(ctx context.Context, sessionID string, ttlSeconds int) (int, *xError.Error) {
	retryKey := bConst.CacheQaGetAnswerRetry.Get(sessionID).String()
	count, err := l.rdb.Incr(ctx, retryKey).Result()
	if err != nil {
		return 0, xError.NewError(ctx, xError.UnknownError, "增加重试计数器失败", false, err)
	}
	// 首次设置时配置过期时间
	if count == 1 {
		ttl := time.Duration(ttlSeconds) * time.Second
		if ttl <= 0 {
			ttl = 48 * time.Hour
		}
		_ = l.rdb.Expire(ctx, retryKey, ttl).Err()
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
	answerData := question.Answer
	questionType, questionOptions := parseQuestionMetadata(question)
	summary := formatAnswerData(questionID, answerData, questionType, questionOptions)
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

		answerData := question.Answer
		questionType, questionOptions := parseQuestionMetadata(question)
		summary := formatAnswerData(qid, answerData, questionType, questionOptions)
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

	l.queue.RemoveQueue(sessionID)
	retryKey := bConst.CacheQaGetAnswerRetry.Get(sessionID).String()
	_ = l.rdb.Del(ctx, retryKey).Err()
	return nil
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
func toQuestionSummary(q *entity.QaQuestion, supplements []qa.SupplementResponse) qa.QuestionSummaryResponse {
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
		Supplements: supplements,
		CreatedAt:   q.CreatedAt.Format(time.RFC3339),
		AnsweredAt:  formatTimePtr(q.AnsweredAt),
	}
}

// formatAnswerString 将多条回答格式化为人类可读字符串
func (l *QaLogic) formatAnswerString(ctx context.Context, answers []qaQueue.Answer) string {
	var sb strings.Builder
	for i, a := range answers {
		if i > 0 {
			sb.WriteString("\n")
		}

		dataStr := formatAnswerData(a.QuestionID, a.Data, "", nil)

		var marker, content string
		switch {
		case dataStr == `"[SKIPPED]"` || dataStr == "[SKIPPED]":
			marker = "[SKIPPED]"
			content = "用户跳过了此问题"
		case strings.HasPrefix(dataStr, `"`) && strings.Contains(dataStr, "NEED_SUPPLEMENT"):
			cleaned := strings.Trim(dataStr, `"`)
			marker = "[NEED_SUPPLEMENT]"
			content = strings.TrimPrefix(cleaned, "[NEED_SUPPLEMENT] ")
		case strings.HasPrefix(dataStr, "[NEED_SUPPLEMENT]"):
			marker = "[NEED_SUPPLEMENT]"
			content = strings.TrimPrefix(dataStr, "[NEED_SUPPLEMENT] ")
		default:
			marker = "[ANSWER]"
			// 尝试从 DB 查询问题元数据以获取更精确的格式
			parsedQID, err := xSnowflake.ParseSnowflakeID(a.QuestionID)
			if err == nil {
				question, xErr := l.repo.question.GetByID(ctx, parsedQID)
				if xErr == nil {
					questionType, questionOptions := parseQuestionMetadata(question)
					dataStr = formatAnswerData(a.QuestionID, a.Data, questionType, questionOptions)
				}
			}
			content = dataStr
		}

		sb.WriteString(fmt.Sprintf("--- question:%s ---\n%s %s\n", a.QuestionID, marker, content))
	}
	return sb.String()
}

// parseQuestionMetadata 从 QaQuestion 实体提取题型和选项列表
func parseQuestionMetadata(q *entity.QaQuestion) (string, []map[string]interface{}) {
	if q == nil {
		return "", nil
	}
	questionType := q.Type
	var options []map[string]interface{}
	if q.Options != nil {
		_ = json.Unmarshal(q.Options, &options)
	}
	return questionType, options
}

// formatAnswerData 格式化单条回答数据
func formatAnswerData(questionID string, data any, questionType string, options []map[string]interface{}) string {
	if data == nil {
		return ""
	}

	m, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Sprintf("%v", data)
	}

	switch questionType {
	case "select":
		return formatSelectAnswer(m, options)
	case "multi-select":
		return formatMultiSelectAnswer(m, options)
	case "text":
		return formatTextAnswer(m)
	case "boolean":
		return formatBooleanAnswer(m)
	case "code", "image", "file":
		return formatMediaAnswer(m, questionType)
	case "diff", "plan", "options", "review":
		return "已提交"
	case "slider":
		return formatSliderAnswer(m)
	case "rank":
		return formatRankAnswer(m)
	case "rate":
		return formatRateAnswer(m)
	default:
		// 未知题型，尝试通用解析
		return formatGenericAnswer(m)
	}
}

// formatSelectAnswer 格式化单选回答
func formatSelectAnswer(m map[string]interface{}, options []map[string]interface{}) string {
	selected, exists := m["selected"]
	if !exists {
		return fmt.Sprintf("%v", m)
	}

	label, desc := resolveOptionLabel(selected, options)
	result := label
	if desc != "" {
		result = label + "\n" + desc
		result += "\n----DETAIL----"
	}
	return result
}

// formatMultiSelectAnswer 格式化多选回答
func formatMultiSelectAnswer(m map[string]interface{}, options []map[string]interface{}) string {
	selected, exists := m["selected"]
	if !exists {
		return fmt.Sprintf("%v", m)
	}

	var labels []string
	var descs []string

	switch v := selected.(type) {
	case []interface{}:
		for _, item := range v {
			label, desc := resolveOptionLabel(item, options)
			if label != "" {
				labels = append(labels, label)
			}
			if desc != "" {
				descs = append(descs, desc)
			}
		}
	default:
		label, desc := resolveOptionLabel(selected, options)
		if label != "" {
			labels = append(labels, label)
		}
		if desc != "" {
			descs = append(descs, desc)
		}
	}

	result := strings.Join(labels, ", ")
	if len(descs) > 0 {
		result += "\n" + strings.Join(descs, "\n")
		result += "\n----DETAIL----"
	}
	return result
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

// formatMediaAnswer 格式化代码/图片/文件回答
func formatMediaAnswer(m map[string]interface{}, questionType string) string {
	switch questionType {
	case "code":
		if code, ok := m["code"].(string); ok {
			return code
		}
	case "image":
		if images, ok := m["images"].([]interface{}); ok && len(images) > 0 {
			var names []string
			for _, img := range images {
				if imgMap, ok := img.(map[string]interface{}); ok {
					if name, ok := imgMap["name"].(string); ok {
						names = append(names, name)
					}
				}
			}
			if len(names) > 0 {
				return strings.Join(names, ", ")
			}
		}
	case "file":
		if files, ok := m["files"].([]interface{}); ok && len(files) > 0 {
			var names []string
			for _, f := range files {
				if fMap, ok := f.(map[string]interface{}); ok {
					if name, ok := fMap["name"].(string); ok {
						names = append(names, name)
					}
				}
			}
			if len(names) > 0 {
				return strings.Join(names, ", ")
			}
		}
	}
	return fmt.Sprintf("%v", m)
}

// formatSliderAnswer 格式化滑块回答
func formatSliderAnswer(m map[string]interface{}) string {
	if value, ok := m["value"]; ok {
		return fmt.Sprintf("%v", value)
	}
	return fmt.Sprintf("%v", m)
}

// formatRankAnswer 格式化排名回答
func formatRankAnswer(m map[string]interface{}) string {
	if ranked, ok := m["ranked"].([]interface{}); ok {
		var items []string
		for i, item := range ranked {
			if rMap, ok := item.(map[string]interface{}); ok {
				if label, ok := rMap["label"].(string); ok {
					items = append(items, fmt.Sprintf("%d. %s", i+1, label))
				}
			} else {
				items = append(items, fmt.Sprintf("%d. %v", i+1, item))
			}
		}
		if len(items) > 0 {
			return strings.Join(items, ", ")
		}
	}
	return fmt.Sprintf("%v", m)
}

// formatRateAnswer 格式化评分回答
func formatRateAnswer(m map[string]interface{}) string {
	if ratings, ok := m["ratings"].(map[string]interface{}); ok {
		var parts []string
		for k, v := range ratings {
			parts = append(parts, fmt.Sprintf("%s: %v", k, v))
		}
		if len(parts) > 0 {
			return strings.Join(parts, ", ")
		}
	}
	return fmt.Sprintf("%v", m)
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
