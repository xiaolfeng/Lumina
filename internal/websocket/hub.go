package websocket

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
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
// 业务连接生命周期逻辑（如 QA 的会话历史推送、在线设备同步）由上层通过
// [Hub.AddRegisterHook] / [Hub.AddUnregisterHook] 注入，Hub 自身保持领域中立，
// 可同时承载 QA、Preview 等多种连接类型。
type Hub struct {
	sessions        map[string]map[string]*Connection // sessionID → deviceID → Connection
	register        chan *Connection                  // 注册通道
	unregister      chan *Connection                  // 注销通道
	mu              sync.RWMutex                      // sessions 读写锁
	log             *xLog.LogNamedLogger              // 日志记录器
	handler         MessageHandler                    // 消息处理回调
	registerHooks   []func(*Connection)               // 连接注册钩子（异步执行）
	unregisterHooks []func(*Connection)               // 连接注销钩子（异步执行）
}

// 全局单例 Hub 实例
var (
	globalHub *Hub
	hubOnce   sync.Once
)

// GetHub 获取或创建全局 Hub 单例
//
// 首次调用时使用传入的 handler 创建 Hub 实例，后续调用复用已有单例。
// 若单例已存在且本次传入的 handler 非 nil，则覆盖单例的 handler，
// 保证无论 wsRouter / previewRouter 的注册顺序如何，Q&A 业务消息处理器最终生效。
func GetHub(handler MessageHandler) *Hub {
	hubOnce.Do(func() {
		globalHub = NewHub(handler)
	})

	// 单例已存在且本次传入 handler 非 nil 时覆盖，确保业务消息处理器最终生效
	if handler != nil {
		globalHub.handler = handler
	}
	return globalHub
}

// GetGlobalHub 返回全局 Hub 单例（不创建）。
//
// 仅在 Hub 已通过 [GetHub] 创建后调用，未创建时返回 nil。
// 供「仅获取」场景使用：如 startup 的 goroutineFunc 需获取已由 route 层
// 在 Register 阶段创建的 Hub 以启动主循环（避免因时序/参数误用创建空 Hub）。
func GetGlobalHub() *Hub {
	return globalHub
}

// NewHub 创建 Hub 实例
func NewHub(handler MessageHandler) *Hub {
	return &Hub{
		sessions:   make(map[string]map[string]*Connection),
		register:   make(chan *Connection),
		unregister: make(chan *Connection),
		log:        xLog.WithName(xLog.NamedCONT, "WebSocketHub"),
		handler:    handler,
	}
}

// AddRegisterHook 注册连接注册钩子
//
// 钩子在连接成功注册到 sessions 映射后异步执行（go fn(conn)）。
// 由上层业务注入连接建立时的生命周期逻辑（如 QA 会话历史推送）。
func (h *Hub) AddRegisterHook(fn func(*Connection)) {
	h.registerHooks = append(h.registerHooks, fn)
}

// AddUnregisterHook 注册连接注销钩子
//
// 钩子在连接从 sessions 映射移除后异步执行（go fn(conn)）。
// 由上层业务注入连接断开时的生命周期逻辑（如 QA 在线设备同步）。
func (h *Hub) AddUnregisterHook(fn func(*Connection)) {
	h.unregisterHooks = append(h.unregisterHooks, fn)
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
		slog.String("kind", conn.Kind),
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

	// 异步执行全部注册钩子
	for _, hook := range h.registerHooks {
		go hook(conn)
	}
}

// unregisterConn 从 sessions 映射中移除连接
func (h *Hub) unregisterConn(conn *Connection) {
	h.mu.Lock()
	devices, ok := h.sessions[conn.sessionID]
	if !ok {
		h.mu.Unlock()
		// 异步执行全部注销钩子
		h.runUnregisterHooks(conn)
		return
	}

	if _, exists := devices[conn.deviceID]; exists {
		delete(devices, conn.deviceID)
		conn.Close()

		remainingCount := len(devices)

		h.log.Info(nil, "设备离线",
			slog.String("sessionID", conn.sessionID),
			slog.String("deviceID", conn.deviceID),
			slog.String("kind", conn.Kind),
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

		// 异步执行全部注销钩子
		h.runUnregisterHooks(conn)
	} else {
		h.mu.Unlock()
		// 异步执行全部注销钩子
		h.runUnregisterHooks(conn)
	}
}

// runUnregisterHooks 异步执行全部注销钩子
func (h *Hub) runUnregisterHooks(conn *Connection) {
	for _, hook := range h.unregisterHooks {
		go hook(conn)
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
