package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/xiaolfeng/Lumina/internal/qa"
	"github.com/xiaolfeng/Lumina/internal/repository"
	"gorm.io/gorm"
)

// upgrader WebSocket 升级器配置
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 允许所有来源（认证由 Auth 中间件保障）
	CheckOrigin: func(r *http.Request) bool { return true },
}

// wsLog WebSocket handler 专用日志
var wsLog = xLog.WithName(xLog.NamedCONT, "WSHandler")

// WSHandler 创建 WebSocket 升级 Gin 处理器
//
// 升级流程：
//  1. 从 query 中提取 session（Hash 标识，必填）和 device_id（可选，不提供则自动生成）
//  2. 通过 Hash 解析为雪花 ID（调用 QaSessionRepo.GetByHash）
//  3. 将 HTTP 连接升级为 WebSocket
//  4. 创建 Connection 并注册到 Hub
//  5. 启动 ReadPump 和 WritePump goroutine
func WSHandler(hub *Hub, sessionRepo *repository.QaSessionRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 提取 session hash（必填）
		sessionHash := c.Query("session")
		if sessionHash == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "session 参数必填"})
			c.Abort()
			return
		}

		// 2. 通过 hash 解析为雪花 ID
		session, xErr := sessionRepo.GetByHash(c.Request.Context(), sessionHash)
		if xErr != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "会话不存在或已过期"})
			c.Abort()
			return
		}
		sessionID := session.ID.String() // snowflake ID 字符串，Hub 内部仍使用雪花 ID

		// 3. 提取或生成 device_id
		deviceID := c.Query("device_id")
		if deviceID == "" {
			deviceID = fmt.Sprintf("device_%s", uuid.New().String()[:8])
		}

		// 4. 升级 HTTP 连接为 WebSocket
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			wsLog.Error(nil, "WebSocket 升级失败", slog.String("error", err.Error()))
			return
		}

		wsLog.Info(nil, "WebSocket 连接建立",
			slog.String("sessionHash", sessionHash),
			slog.String("sessionID", sessionID),
			slog.String("deviceID", deviceID),
		)

		// 5. 创建连接封装并注册到 Hub
		wsConn := NewConnection(conn, sessionID, deviceID, hub)
		wsConn.sessionHash = sessionHash
		hub.Register(wsConn)

		// 6. 启动读写泵
		go wsConn.WritePump()
		go wsConn.ReadPump()
	}
}

// CreateMessageHandler 创建业务消息处理器，将 WebSocket 消息与数据库和回答队列连接
//
// 该函数创建一个 MessageHandler 回调，处理 answer_submit、request_supplement、skip 三种业务消息。
// 调用方在创建 Hub 时传入此回调即可启用业务处理逻辑。
func CreateMessageHandler(db *gorm.DB) MessageHandler {
	log := xLog.WithName(xLog.NamedLOGC, "WSMsgHandler")
	questionRepo := repository.NewQaQuestionRepo(db)
	queue := qa.GetAnswerQueue()

	return func(ctx context.Context, conn *Connection, msg *Message) {
		log.Info(ctx, "收到消息", slog.String("type", string(msg.Type)), slog.String("session", conn.sessionID))

		switch msg.Type {
		case MsgAnswerSubmit:
			handleAnswerSubmit(ctx, conn, msg, questionRepo, queue, log)
		case MsgRequestSupplement:
			handleRequestSupplement(ctx, conn, msg, questionRepo, queue, log)
		case MsgSkip:
			handleSkip(ctx, conn, msg, questionRepo, queue, log)
		case MsgSessionLeave:
			conn.isVoluntary = true
			conn.Close()
		case MsgHeartbeatAck:
			// 心跳响应已在 Connection.ReadPump 中更新 lastPing，无需额外处理
		default:
			log.Warn(ctx, "未知消息类型", slog.String("type", string(msg.Type)))
		}
	}
}

// handleAnswerSubmit 处理回答提交消息
//
// 流程：解析数据 → 更新 DB → 入队 → 跨设备广播同步
func handleAnswerSubmit(ctx context.Context, conn *Connection, msg *Message, questionRepo *repository.QaQuestionRepo, queue *qa.AnswerQueue, log *xLog.LogNamedLogger) {
	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		log.Warn(ctx, "answer_submit 消息 data 格式无效")
		return
	}

	questionIDStr, _ := data["question_id"].(string)
	answerData := data["answer"]

	// 解析问题雪花 ID
	qID, err := xSnowflake.ParseSnowflakeID(questionIDStr)
	if err != nil {
		log.Warn(ctx, "解析 question_id 失败", slog.String("question_id", questionIDStr))
		return
	}

	// 更新问题状态为已回答，写入回答数据
	if xErr := questionRepo.UpdateAnswer(ctx, qID, "answered", answerData); xErr != nil {
		log.Warn(ctx, "更新问题回答失败", slog.String("error", xErr.Error()))
		return
	}

	// 将回答推入队列，供 MCP get_answer 消费
	queue.Enqueue(conn.sessionID, qa.Answer{
		QuestionID: questionIDStr,
		Data:       answerData,
		Timestamp:  time.Now(),
	})

	// 检查是否有消费者在等待，若无则广播未消费通知
	if !queue.HasConsumer(conn.sessionID) {
		unhandledMsg := &Message{
			Type:      MsgAnswerUnhandled,
			SessionID: conn.sessionID,
			Data: map[string]interface{}{
				"question_id": questionIDStr,
				"answer":      answerData,
				"message":     "回答已提交，但当前没有 AI Agent 在等待处理。请手动复制此提示词提供给 Agent。",
			},
			Timestamp: time.Now().UnixMilli(),
		}
		conn.hub.BroadcastToSession(conn.sessionID, unhandledMsg)
	}

	// 跨设备广播回答同步（通知同一会话的其他设备该问题已作答）
	syncMsg := &Message{
		Type:      MsgAnswerSync,
		SessionID: conn.sessionID,
		Data: map[string]interface{}{
			"question_id": questionIDStr,
			"status":      "answered",
		},
		Timestamp: time.Now().UnixMilli(),
	}
	conn.hub.BroadcastToSession(conn.sessionID, syncMsg)
}

// handleRequestSupplement 处理补充内容请求消息
//
// 将补充请求以 [NEED_SUPPLEMENT] 富令牌格式推入回答队列，
// Agent 端通过 get_answer 消费后可使用 qa_push_supplement 推送补充内容。
// 令牌采用 [FIELD] value 序列格式，缺省字段整行省略。
func handleRequestSupplement(ctx context.Context, conn *Connection, msg *Message, questionRepo *repository.QaQuestionRepo, queue *qa.AnswerQueue, log *xLog.LogNamedLogger) {
	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		log.Warn(ctx, "request_supplement 消息 data 格式无效")
		return
	}

	questionIDStr, _ := data["question_id"].(string)
	note, _ := data["note"].(string)
	withOptions, _ := data["with_options"].(bool)

	// 解析 option_ids（前端勾选开关时传入全部 option id）
	var optionIDs []string
	if rawIDs, ok := data["option_ids"].([]interface{}); ok {
		for _, id := range rawIDs {
			if s, ok := id.(string); ok && s != "" {
				optionIDs = append(optionIDs, s)
			}
		}
	}

	// 构建 [NEED_SUPPLEMENT] 富令牌（缺省字段整行省略）
	var sb strings.Builder
	sb.WriteString("[NEED_SUPPLEMENT] 用户请求补充内容，请使用 qa_push_supplement 推送 Markdown 或 HTML 格式的详细说明，帮助用户更全面地理解问题\n")
	sb.WriteString(fmt.Sprintf("[TARGET] question %s\n", questionIDStr))

	if strings.TrimSpace(note) != "" {
		sb.WriteString(fmt.Sprintf("[USER_NOTE] %s\n", note))
	}

	if withOptions && len(optionIDs) > 0 {
		optionLabelMap := resolveOptionLabels(ctx, questionRepo, questionIDStr, optionIDs)
		sb.WriteString("[WITH_OPTIONS] 用户期望你为问题补充内容后，也使用 qa_push_supplement 为以下每个选项逐一提供详细说明\n")
		sb.WriteString("[OPTION_LIST]\n")
		for _, optID := range optionIDs {
			label := optionLabelMap[optID]
			if label == "" {
				label = optID
			}
			sb.WriteString(fmt.Sprintf("  - %s → %s\n", label, optID))
		}
	}

	queue.Enqueue(conn.sessionID, qa.Answer{
		QuestionID: questionIDStr,
		Data:       sb.String(),
		Timestamp:  time.Now(),
	})
}

// resolveOptionLabels 查询问题实体，按 option_id 收集 label 映射。
// 查询失败或选项缺失时静默降级（返回空 map，调用方用 optID 兜底）。
func resolveOptionLabels(ctx context.Context, questionRepo *repository.QaQuestionRepo, questionIDStr string, optionIDs []string) map[string]string {
	labelMap := make(map[string]string)
	if len(optionIDs) == 0 {
		return labelMap
	}
	parsedQID, err := xSnowflake.ParseSnowflakeID(questionIDStr)
	if err != nil {
		return labelMap
	}
	question, xErr := questionRepo.GetByID(ctx, parsedQID)
	if xErr != nil || len(question.Options) == 0 {
		return labelMap
	}
	var options []map[string]interface{}
	if jsonErr := json.Unmarshal(question.Options, &options); jsonErr != nil {
		return labelMap
	}
	for _, opt := range options {
		id, _ := opt["id"].(string)
		label, _ := opt["label"].(string)
		if id != "" && label != "" {
			labelMap[id] = label
		}
	}
	return labelMap
}

// handleSkip 处理跳过问题消息
//
// 流程：更新 DB 状态 → 入队标记 → 跨设备广播同步
func handleSkip(ctx context.Context, conn *Connection, msg *Message, questionRepo *repository.QaQuestionRepo, queue *qa.AnswerQueue, log *xLog.LogNamedLogger) {
	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		log.Warn(ctx, "skip 消息 data 格式无效")
		return
	}

	questionIDStr, _ := data["question_id"].(string)

	// 解析问题雪花 ID
	qID, err := xSnowflake.ParseSnowflakeID(questionIDStr)
	if err != nil {
		log.Warn(ctx, "解析 question_id 失败", slog.String("question_id", questionIDStr))
		return
	}

	// 更新问题状态为已跳过
	if xErr := questionRepo.UpdateStatus(ctx, qID, "skipped"); xErr != nil {
		log.Warn(ctx, "更新问题状态失败", slog.String("error", xErr.Error()))
		return
	}

	// 推入队列，供 MCP get_answer 消费
	queue.Enqueue(conn.sessionID, qa.Answer{
		QuestionID: questionIDStr,
		Data:       "[SKIPPED]",
		Timestamp:  time.Now(),
	})

	// 检查是否有消费者在等待，若无则广播未消费通知
	if !queue.HasConsumer(conn.sessionID) {
		unhandledMsg := &Message{
			Type:      MsgAnswerUnhandled,
			SessionID: conn.sessionID,
			Data: map[string]interface{}{
				"question_id": questionIDStr,
				"answer":      "[SKIPPED]",
				"message":     "回答已跳过，但当前没有 AI Agent 在等待处理。请手动复制此提示词提供给 Agent。",
			},
			Timestamp: time.Now().UnixMilli(),
		}
		conn.hub.BroadcastToSession(conn.sessionID, unhandledMsg)
	}

	// 跨设备广播跳过同步
	syncMsg := &Message{
		Type:      MsgAnswerSync,
		SessionID: conn.sessionID,
		Data: map[string]interface{}{
			"question_id": questionIDStr,
			"status":      "skipped",
		},
		Timestamp: time.Now().UnixMilli(),
	}
	conn.hub.BroadcastToSession(conn.sessionID, syncMsg)
}
