package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// writeWait 写超时时间
	writeWait = 10 * time.Second
	// pongWait 等待客户端心跳响应超时时间
	pongWait = 60 * time.Second
	// heartbeatPeriod 心跳发送间隔（必须小于 pongWait）
	heartbeatPeriod = 5 * time.Second
	// sendBufSize 发送通道缓冲区大小
	sendBufSize = 256
)

// Connection WebSocket 连接封装
type Connection struct {
	conn        *websocket.Conn // 底层 WebSocket 连接
	hub         *Hub            // 所属 Hub 引用
	sessionID   string          // QA 会话 ID（雪花 ID 字符串）
	sessionHash string          // 会话 Hash 标识，用于跨设备通知和主动离开判断
	deviceID    string          // 设备唯一标识
	send        chan []byte     // 待发送消息缓冲通道
	mu          sync.Mutex      // 写锁，防止并发写入
	lastPing    time.Time       // 最后一次收到心跳响应的时间
	isAlive     bool            // 连接存活标记
	isVoluntary bool            // 是否为主动离开（true=用户主动关闭/发送 session_leave）
}

// NewConnection 创建连接封装实例
func NewConnection(conn *websocket.Conn, sessionID, deviceID string, hub *Hub) *Connection {
	return &Connection{
		conn:      conn,
		hub:       hub,
		sessionID: sessionID,
		deviceID:  deviceID,
		send:      make(chan []byte, sendBufSize),
		lastPing:  time.Now(),
		isAlive:   true,
	}
}

// ReadPump 持续读取 WebSocket 消息并分发给 Hub 处理
//
// 在独立 goroutine 中运行，当连接关闭或读取异常时自动退出。
// 读取到 MsgHeartbeatAck 时更新 lastPing 时间戳以维持心跳检测。
func (c *Connection) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))

	for {
		msgType, data, err := c.conn.ReadMessage()
		if err != nil {
			// 连接异常或客户端主动关闭
			return
		}

		// 重置读取截止时间
		c.conn.SetReadDeadline(time.Now().Add(pongWait))

		// 更新存活心跳时间
		c.mu.Lock()
		c.lastPing = time.Now()
		c.mu.Unlock()

		// 分发消息给 Hub 处理
		c.hub.handleMessage(c, msgType, data)
	}
}

// WritePump 持续从发送通道写入消息到 WebSocket 连接
//
// 在独立 goroutine 中运行，包含定时心跳发送逻辑。
// 当发送通道关闭或写入异常时自动退出。
func (c *Connection) WritePump() {
	heartbeatTicker := time.NewTicker(heartbeatPeriod)
	defer func() {
		heartbeatTicker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// 通道已关闭，发送关闭帧
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-heartbeatTicker.C:
			// 定时发送心跳消息
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			heartbeat := &Message{
				Type:      MsgHeartbeat,
				Timestamp: time.Now().UnixMilli(),
			}
			if err := c.writeJSON(heartbeat); err != nil {
				return
			}
		}
	}
}

// writeJSON 将对象序列化为 JSON 并写入 WebSocket 连接（线程安全）
func (c *Connection) writeJSON(v any) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	return c.conn.WriteJSON(v)
}

// Close 安全关闭连接
func (c *Connection) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isAlive {
		c.isAlive = false
		close(c.send)
	}
}

// SessionID 返回连接所属的会话 ID
func (c *Connection) SessionID() string {
	return c.sessionID
}

// DeviceID 返回连接的设备标识
func (c *Connection) DeviceID() string {
	return c.deviceID
}

// SendMessage 将消息序列化后推送到发送通道
func (c *Connection) SendMessage(msg *Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isAlive {
		return nil
	}

	select {
	case c.send <- data:
	default:
		// 发送通道已满，丢弃消息
	}

	return nil
}
