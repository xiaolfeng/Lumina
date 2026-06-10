package websocket

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
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
//  1. 从 query 中提取 session_id（必填）和 device_id（可选，不提供则自动生成）
//  2. 将 HTTP 连接升级为 WebSocket
//  3. 创建 Connection 并注册到 Hub
//  4. 启动 ReadPump 和 WritePump goroutine
func WSHandler(hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 提取 session_id（必填）
		sessionID := c.Query("session_id")
		if sessionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "session_id 参数必填"})
			c.Abort()
			return
		}

		// 2. 提取或生成 device_id
		deviceID := c.Query("device_id")
		if deviceID == "" {
			deviceID = fmt.Sprintf("device_%s", uuid.New().String()[:8])
		}

		// 3. 升级 HTTP 连接为 WebSocket
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			wsLog.Error(nil, "WebSocket 升级失败", slog.String("error", err.Error()))
			return
		}

		wsLog.Info(nil, "WebSocket 连接建立", slog.String("sessionID", sessionID), slog.String("deviceID", deviceID))

		// 4. 创建连接封装并注册到 Hub
		wsConn := NewConnection(conn, sessionID, deviceID, hub)
		hub.Register(wsConn)

		// 5. 启动读写泵
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
			handleRequestSupplement(ctx, conn, msg, queue, log)
		case MsgSkip:
			handleSkip(ctx, conn, msg, questionRepo, queue, log)
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
// 将补充请求以 [NEED_SUPPLEMENT] 标记格式推入回答队列，
// Agent 端通过 get_answer 消费后可使用 qa_push_supplement 推送补充内容。
func handleRequestSupplement(ctx context.Context, conn *Connection, msg *Message, queue *qa.AnswerQueue, log *xLog.LogNamedLogger) {
	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		log.Warn(ctx, "request_supplement 消息 data 格式无效")
		return
	}

	questionIDStr, _ := data["question_id"].(string)

	// 格式化为 Agent 可识别的补充请求标记
	supplementMsg := fmt.Sprintf("[NEED_SUPPLEMENT] 用户请求补充内容\n目标: %v\n提示: 使用 qa_push_supplement 推送补充内容", data)

	queue.Enqueue(conn.sessionID, qa.Answer{
		QuestionID: questionIDStr,
		Data:       supplementMsg,
		Timestamp:  time.Now(),
	})
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
