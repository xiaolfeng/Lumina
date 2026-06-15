package websocket

// MessageType 定义 WebSocket 消息类型
type MessageType string

const (
	// Server → Client 消息类型
	MsgQuestionPush    MessageType = "question_push"     // 推送新问题（pending）
	MsgHistoryQuestion MessageType = "history_question"  // 推送历史问题（已回答/已跳过，连接恢复时）
	MsgSupplementPush  MessageType = "supplement_push"   // 推送补充内容
	MsgAnswerSync      MessageType = "answer_sync"       // 回答同步（跨设备）
	MsgHeartbeat       MessageType = "heartbeat"         // 心跳
	MsgSessionEnd      MessageType = "session_end"       // 会话结束通知
	MsgQuestionCancel  MessageType = "question_cancel"  // 问题取消通知（单个或全部）

	MsgAnswerUnhandled MessageType = "answer_unhandled" // 回答未消费通知（Server → Client）

	// Client → Server 消息类型
	MsgAnswerSubmit      MessageType = "answer_submit"       // 提交回答
	MsgRequestSupplement MessageType = "request_supplement"  // 请求补充
	MsgSkip              MessageType = "skip"                // 跳过问题
	MsgHeartbeatAck      MessageType = "heartbeat_ack"       // 心跳响应

	// 多设备协同消息类型
	MsgDeviceJoin    MessageType = "device_join"     // 设备加入通知（Server → Client）
	MsgDeviceLeave   MessageType = "device_leave"    // 设备离开通知（Server → Client）
	MsgSessionLeave  MessageType = "session_leave"   // 主动离开（Client → Server）
)

// Message WebSocket 通信消息
type Message struct {
	Type      MessageType `json:"type"`                  // 消息类型
	SessionID string      `json:"session_id,omitempty"`  // 会话ID
	Data      any         `json:"data,omitempty"`         // 消息数据
	Timestamp int64       `json:"timestamp"`              // 时间戳（毫秒）
}
