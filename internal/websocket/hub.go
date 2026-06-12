package websocket

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"github.com/xiaolfeng/Lumina/internal/repository"
	"gorm.io/gorm"
)

// MessageHandler 消息处理回调函数
//
// 当 Hub 收到客户端消息时调用此回调，由上层业务逻辑决定如何处理。
type MessageHandler func(ctx context.Context, conn *Connection, msg *Message)

// heartbeatCheckInterval 心跳检测间隔
const heartbeatCheckInterval = 5 * time.Second

// heartbeatTimeout 心跳超时阈值（3 倍心跳周期）
const heartbeatTimeout = 15 * time.Second

// Hub WebSocket 连接管理器
//
// 管理所有活跃的 WebSocket 连接，按 sessionID → deviceID 二级索引组织。
// 通过 register/unregister 通道实现并发安全的连接生命周期管理。
// 新连接注册时自动推送该会话下所有待回答问题（用于前端重连恢复）。
type Hub struct {
	sessions     map[string]map[string]*Connection // sessionID → deviceID → Connection
	register     chan *Connection                   // 注册通道
	unregister   chan *Connection                   // 注销通道
	mu           sync.RWMutex                       // sessions 读写锁
	log          *xLog.LogNamedLogger               // 日志记录器
	handler      MessageHandler                     // 消息处理回调
	sessionRepo  *repository.QaSessionRepo          // QA 会话仓库（用于临时会话硬删除）
	questionRepo *repository.QaQuestionRepo          // QA 问题仓库（用于连接时推送 pending 问题）
	db           *gorm.DB                           // 数据库实例（用于异步更新 OnlineDevices）
}

// 全局单例 Hub 实例
var (
	globalHub *Hub
	hubOnce   sync.Once
)

// GetHub 获取或创建全局 Hub 单例
//
// 首次调用时使用传入的参数创建 Hub 实例，后续调用忽略参数。
// 当 handler 为 nil 时，收到的客户端消息仅记录日志不做处理。
func GetHub(handler MessageHandler, sessionRepo *repository.QaSessionRepo, db *gorm.DB) *Hub {
	hubOnce.Do(func() {
		globalHub = NewHub(handler, sessionRepo, db)
	})
	return globalHub
}

// NewHub 创建 Hub 实例
func NewHub(handler MessageHandler, sessionRepo *repository.QaSessionRepo, db *gorm.DB) *Hub {
	return &Hub{
		sessions:     make(map[string]map[string]*Connection),
		register:     make(chan *Connection),
		unregister:   make(chan *Connection),
		log:          xLog.WithName(xLog.NamedCONT, "WebSocketHub"),
		handler:      handler,
		sessionRepo:  sessionRepo,
		questionRepo: repository.NewQaQuestionRepo(db),
		db:           db,
	}
}

// Run 启动 Hub 主循环
//
// 监听注册/注销事件，定期执行心跳检测。
// 应在独立 goroutine 中调用，通常在应用启动时启动。
func (h *Hub) Run(ctx context.Context) {
	heartbeatTicker := time.NewTicker(heartbeatCheckInterval)
	defer heartbeatTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			h.shutdownAll()
			return

		case conn := <-h.register:
			h.registerConn(conn)

		case conn := <-h.unregister:
			h.unregisterConn(conn)

		case <-heartbeatTicker.C:
			h.checkHeartbeats()
		}
	}
}

// Register 将连接注册到 Hub
func (h *Hub) Register(conn *Connection) {
	h.register <- conn
}

// Unregister 从 Hub 注销连接
func (h *Hub) Unregister(conn *Connection) {
	h.unregister <- conn
}

// BroadcastToSession 向指定会话的所有在线设备广播消息
func (h *Hub) BroadcastToSession(sessionID string, msg *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	devices, ok := h.sessions[sessionID]
	if !ok {
		return
	}

	for _, conn := range devices {
		_ = conn.SendMessage(msg)
	}
}

// SendToDevice 向指定会话的特定设备发送消息
func (h *Hub) SendToDevice(sessionID, deviceID string, msg *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	devices, ok := h.sessions[sessionID]
	if !ok {
		return
	}

	if conn, exists := devices[deviceID]; exists {
		_ = conn.SendMessage(msg)
	}
}

// GetOnlineDevices 获取指定会话的在线设备数量
func (h *Hub) GetOnlineDevices(sessionID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	devices, ok := h.sessions[sessionID]
	if !ok {
		return 0
	}

	count := 0
	for _, conn := range devices {
		conn.mu.Lock()
		if conn.isAlive {
			count++
		}
		conn.mu.Unlock()
	}

	return count
}

// handleMessage 处理收到的客户端消息
func (h *Hub) handleMessage(conn *Connection, msgType int, data []byte) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		h.log.Info(nil, "消息解析失败", slog.String("error", err.Error()), slog.String("sessionID", conn.sessionID), slog.String("deviceID", conn.deviceID))
		return
	}

	// 心跳响应仅更新时间戳，不触发业务回调
	if msg.Type == MsgHeartbeatAck {
		return
	}

	// 填充 sessionID（优先使用消息中携带的，否则使用连接绑定的）
	if msg.SessionID == "" {
		msg.SessionID = conn.sessionID
	}

	// 调用业务回调
	if h.handler != nil {
		h.handler(context.Background(), conn, &msg)
	}
}

// registerConn 注册连接到 sessions 映射
func (h *Hub) registerConn(conn *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.sessions[conn.sessionID]; !ok {
		h.sessions[conn.sessionID] = make(map[string]*Connection)
	}

	deviceCount := len(h.sessions[conn.sessionID])
	h.sessions[conn.sessionID][conn.deviceID] = conn

	h.log.Info(nil, "设备上线",
		slog.String("sessionID", conn.sessionID),
		slog.String("deviceID", conn.deviceID),
		slog.Int("online", deviceCount+1),
	)

	// 向同会话的其他设备广播 device_join
	if deviceCount > 0 {
		joinMsg := &Message{
			Type:      MsgDeviceJoin,
			SessionID: conn.sessionID,
			Data: map[string]interface{}{
				"device_id": conn.deviceID,
				"hash":      conn.sessionHash,
			},
			Timestamp: time.Now().UnixMilli(),
		}
		for devID, existingConn := range h.sessions[conn.sessionID] {
			if devID != conn.deviceID {
				_ = existingConn.SendMessage(joinMsg)
			}
		}
	}

	// 异步同步 OnlineDevices 到数据库
	go h.syncOnlineDevices(conn.sessionID, deviceCount+1)

	// 向新连接推送该会话下所有待回答问题（支持前端重连恢复）
	go h.pushPendingQuestions(conn)
}

// unregisterConn 从 sessions 映射中移除连接
func (h *Hub) unregisterConn(conn *Connection) {
	h.mu.Lock()
	devices, ok := h.sessions[conn.sessionID]
	if !ok {
		h.mu.Unlock()
		return
	}

	if _, exists := devices[conn.deviceID]; exists {
		delete(devices, conn.deviceID)
		conn.Close()

		remainingCount := len(devices)

		h.log.Info(nil, "设备离线",
			slog.String("sessionID", conn.sessionID),
			slog.String("deviceID", conn.deviceID),
			slog.Bool("voluntary", conn.isVoluntary),
			slog.Int("remaining", remainingCount),
		)

		// 会话无在线设备时清理映射
		if remainingCount == 0 {
			delete(h.sessions, conn.sessionID)
		}
		h.mu.Unlock()

		// 向剩余设备广播 device_leave
		if remainingCount > 0 {
			leaveMsg := &Message{
				Type:      MsgDeviceLeave,
				SessionID: conn.sessionID,
				Data: map[string]interface{}{
					"device_id": conn.deviceID,
				},
				Timestamp: time.Now().UnixMilli(),
			}
			h.mu.RLock()
			for _, remainingConn := range devices {
				_ = remainingConn.SendMessage(leaveMsg)
			}
			h.mu.RUnlock()
		}

		// 异步同步 OnlineDevices 到数据库
		go h.syncOnlineDevices(conn.sessionID, remainingCount)

		// 最后一个设备主动离开的临时会话 → 触发硬删除
		if remainingCount == 0 && conn.isVoluntary {
			go h.checkAndDeleteTemporarySession(conn.sessionID)
		}
	} else {
		h.mu.Unlock()
	}
}

// checkHeartbeats 检测所有连接的心跳状态，超时则标记死亡并注销
func (h *Hub) checkHeartbeats() {
	h.mu.RLock()
	defer h.mu.RUnlock()

	now := time.Now()
	for _, devices := range h.sessions {
		for _, conn := range devices {
			conn.mu.Lock()
			if now.Sub(conn.lastPing) > heartbeatTimeout {
				conn.isAlive = false
				conn.mu.Unlock()

				// 异步注销避免死锁（Unregister 会写 unregister 通道）
				go h.Unregister(conn)
				continue
			}
			conn.mu.Unlock()
		}
	}
}

// shutdownAll 关闭所有活跃连接
func (h *Hub) shutdownAll() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for sessionID, devices := range h.sessions {
		for deviceID, conn := range devices {
			conn.Close()
			h.log.Info(nil, "关闭连接", slog.String("sessionID", sessionID), slog.String("deviceID", deviceID))
		}
		delete(h.sessions, sessionID)
	}
}

// syncOnlineDevices 异步更新数据库中的 OnlineDevices 字段
func (h *Hub) syncOnlineDevices(sessionID string, count int) {
	sid, err := xSnowflake.ParseSnowflakeID(sessionID)
	if err != nil {
		return
	}
	h.db.Model(&entity.QaSession{}).Where("id = ?", sid).Update("online_devices", count)
}

// checkAndDeleteTemporarySession 检查并删除临时会话（最后一个设备主动离开时触发硬删除）
func (h *Hub) checkAndDeleteTemporarySession(sessionID string) {
	sid, err := xSnowflake.ParseSnowflakeID(sessionID)
	if err != nil {
		return
	}
	session, xErr := h.sessionRepo.GetByID(context.Background(), sid)
	if xErr != nil {
		return
	}
	if session.Type == "temporary" {
		h.sessionRepo.Delete(context.Background(), sid)
	}
}

// pushPendingQuestions 向新连接推送该会话下所有待回答问题
func (h *Hub) pushPendingQuestions(conn *Connection) {
	sid, err := xSnowflake.ParseSnowflakeID(conn.sessionID)
	if err != nil {
		return
	}

	questions, xErr := h.questionRepo.GetPendingBySessionID(context.Background(), sid)
	if xErr != nil {
		h.log.Warn(nil, "推送待回答问题失败", slog.String("error", xErr.Error()))
		return
	}

	for _, q := range questions {
		msg := &Message{
			Type:      MsgQuestionPush,
			SessionID: conn.sessionID,
			Data:      q,
			Timestamp: time.Now().UnixMilli(),
		}
		_ = conn.SendMessage(msg)
	}
}
