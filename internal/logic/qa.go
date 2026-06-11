package logic

import (
	"context"
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
	sessions, total, xErr := l.repo.session.List(ctx, page, size, req.Status, req.Type)
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

	// 映射问题摘要
	questionSummaries := make([]qa.QuestionSummaryResponse, 0, len(questions))
	for _, q := range questions {
		questionSummaries = append(questionSummaries, toQuestionSummary(q))
	}

	return &qa.SessionDetailResponse{
		ID:            session.ID.String(),
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
				TargetID:    strconv.FormatInt(supplement.TargetID, 10),
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

	// 清理队列
	l.queue.RemoveQueue(id)

	return nil
}

// GetQaConfig 获取Q&A配置
func (l *QaLogic) GetQaConfig(ctx context.Context) (*qa.QaConfigResponse, *xError.Error) {
	l.log.Info(ctx, "GetQaConfig - 获取Q&A配置")

	// 读取 Session TTL
	var ttlInfo entity.Info
	if err := l.db.WithContext(ctx).Where("`key` = ?", "qa.session.ttl").First(&ttlInfo).Error; err != nil {
		l.log.Warn(ctx, fmt.Sprintf("读取qa.session.ttl失败: %s", err.Error()))
		ttlInfo.Value = "604800" // 默认7天
	}
	ttl, _ := strconv.Atoi(ttlInfo.Value)
	if ttl <= 0 {
		ttl = 604800
	}

	// 读取运行时域名
	var domainInfo entity.Info
	if err := l.db.WithContext(ctx).Where("`key` = ?", "runtime.domain").First(&domainInfo).Error; err != nil {
		l.log.Warn(ctx, fmt.Sprintf("读取runtime.domain失败: %s", err.Error()))
		domainInfo.Value = "http://localhost:3000"
	}

	return &qa.QaConfigResponse{
		SessionTTL:    ttl,
		RuntimeDomain: domainInfo.Value,
	}, nil
}

// UpdateQaConfig 更新Q&A配置
func (l *QaLogic) UpdateQaConfig(ctx context.Context, req *qa.UpdateQaConfigRequest) (*qa.QaConfigResponse, *xError.Error) {
	l.log.Info(ctx, "UpdateQaConfig - 更新Q&A配置")

	// 更新 Session TTL
	if req.SessionTTL != nil {
		if err := l.db.WithContext(ctx).
			Model(&entity.Info{}).
			Where("`key` = ?", "qa.session.ttl").
			Update("value", strconv.Itoa(*req.SessionTTL)).Error; err != nil {
			return nil, xError.NewError(ctx, xError.DatabaseError, "更新Session TTL失败", false, err)
		}
	}

	// 更新运行时域名
	if req.RuntimeDomain != nil {
		if err := l.db.WithContext(ctx).
			Model(&entity.Info{}).
			Where("`key` = ?", "runtime.domain").
			Update("value", *req.RuntimeDomain).Error; err != nil {
			return nil, xError.NewError(ctx, xError.DatabaseError, "更新运行时域名失败", false, err)
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
	var prjID int64
	if projectID != "" {
		var parseErr error
		prjID, parseErr = strconv.ParseInt(projectID, 10, 64)
		if parseErr != nil {
			return "", "", xError.NewError(ctx, xError.BusinessError, "无效的项目ID格式", false, nil)
		}
	}

	// 生成雪花ID
	id := xSnowflake.GenerateID(bConst.GeneQaSession)

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
			ExpiresAt:  nil,
		}
		if xErr := l.repo.session.Create(ctx, entity); xErr != nil {
			return "", "", xErr
		}
	}

	// 生成浏览器链接
	link := fmt.Sprintf("http://localhost:3000/interact?session=%s", id.String())

	return id.String(), link, nil
}

// PushQuestion 推送问题到会话（MCP工具）
func (l *QaLogic) PushQuestion(ctx context.Context, sessionID, qType, title, description string, options, config, batch any, groupLabel string) (string, map[string]string, *xError.Error) {
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
		SessionID:   parsedSID.Int64(),
		Type:        qType,
		Title:       title,
		Description: description,
		Options:     toJSON(options),
		Config:      toJSON(config),
		Batch:       toJSON(batch),
		GroupLabel:  groupLabel,
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
func (l *QaLogic) PushSupplement(ctx context.Context, sessionID, targetType string, targetID int64, contentType, content string) *xError.Error {
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
		SessionID:   parsedSID.Int64(),
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

	// 格式化回答字符串
	return formatAnswerString(answers), nil
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
		return "问题尚未回答，请使用 get_answer 阻塞等待回答", nil
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
	summary := formatAnswerData(questionID, answerData)
	return fmt.Sprintf("--- question:%s ---\n[ANSWERED] %s", questionID, summary), nil
}

// RegetAnswers 批量重新获取回答（MCP工具）
//
// 遍历所有 questionIDs，对每个执行：
// - 已回答 → "[ANSWERED] {question_id}: {答案摘要}"
// - 未回答 → "[PENDING] {question_id}: 请使用 qa_get_answer 阻塞等待用户回答。"
// - 无效ID → "[ERROR] {question_id}: 无效的问题ID格式"
// - 不存在 → "[ERROR] {question_id}: 问题不存在"
func (l *QaLogic) RegetAnswers(ctx context.Context, sessionID string, questionIDs []string) (string, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("RegetAnswers - 批量重新获取回答 [session=%s, count=%d]", sessionID, len(questionIDs)))

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("批量查询结果（%d 个问题）:\n\n", len(questionIDs)))

	for _, qid := range questionIDs {
		parsedQID, err := xSnowflake.ParseSnowflakeID(qid)
		if err != nil {
			sb.WriteString(fmt.Sprintf("[ERROR] %s: 无效的问题ID格式\n", qid))
			continue
		}

		question, xErr := l.repo.question.GetByID(ctx, parsedQID)
		if xErr != nil {
			sb.WriteString(fmt.Sprintf("[ERROR] %s: 问题不存在\n", qid))
			continue
		}

		if question.Status == "answered" {
			answerData := question.Answer
			summary := formatAnswerData(qid, answerData)
			sb.WriteString(fmt.Sprintf("[ANSWERED] %s: %s\n", qid, summary))
		} else {
			sb.WriteString(fmt.Sprintf("[PENDING] %s: 请使用 qa_get_answer 阻塞等待用户回答。\n", qid))
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
	sb.WriteString(fmt.Sprintf("关联项目: %d\n", session.ProjectID))
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

// ─── Helper Methods ─────────────────────────────────────────────────────

// toSessionResponse 将会话实体映射为响应 DTO
func toSessionResponse(session *entity.QaSession) qa.SessionResponse {
	return qa.SessionResponse{
		ID:            session.ID.String(),
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
func toQuestionSummary(q *entity.QaQuestion) qa.QuestionSummaryResponse {
	return qa.QuestionSummaryResponse{
		ID:         q.ID.String(),
		Type:       q.Type,
		Title:      q.Title,
		Status:     q.Status,
		CreatedAt:  q.CreatedAt.Format(time.RFC3339),
		AnsweredAt: formatTimePtr(q.AnsweredAt),
	}
}

// formatAnswerString 将多条回答格式化为人类可读字符串
func formatAnswerString(answers []qaQueue.Answer) string {
	var sb strings.Builder
	for i, a := range answers {
		if i > 0 {
			sb.WriteString("\n")
		}
		summary := formatAnswerData(a.QuestionID, a.Data)
		sb.WriteString(fmt.Sprintf("--- question:%s ---\n[ANSWERED] %s\n", a.QuestionID, summary))
	}
	return sb.String()
}

// formatAnswerData 格式化单条回答数据
func formatAnswerData(questionID string, data any) string {
	if data == nil {
		return ""
	}

	// 尝试解析为 map 以判断回答类型
	m, ok := data.(map[string]interface{})
	if !ok {
		// 非对象类型，JSON 序列化
		bytes, err := json.Marshal(data)
		if err != nil {
			return fmt.Sprintf("%v", data)
		}
		return string(bytes)
	}

	// 选择题型：有 selected 字段
	if selected, exists := m["selected"]; exists {
		if selMap, ok := selected.(map[string]interface{}); ok {
			label, _ := selMap["label"].(string)
			id, _ := selMap["id"]
			return fmt.Sprintf("选择: %s (id: %v)", label, id)
		}
	}

	// 文本题型：有 text 字段
	if text, exists := m["text"]; exists {
		if textStr, ok := text.(string); ok {
			return textStr
		}
	}

	// 布尔题型：有 choice 字段
	if choice, exists := m["choice"]; exists {
		if choiceStr, ok := choice.(string); ok {
			if choiceStr == "yes" {
				return "是"
			}
			return "否"
		}
		if choiceBool, ok := choice.(bool); ok {
			if choiceBool {
				return "是"
			}
			return "否"
		}
	}

	// 默认：JSON 序列化
	bytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Sprintf("%v", data)
	}
	return string(bytes)
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
