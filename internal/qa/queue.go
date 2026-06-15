package qa

import (
	"context"
	"log/slog"
	"sync"
	"time"

	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
)

// Answer 用户回答结构
//
// 当用户通过 WebSocket 提交 answer_submit 消息时，回答数据封装为此结构入队。
type Answer struct {
	QuestionID string    `json:"question_id"` // 问题ID
	Data       any       `json:"data"`        // 回答数据（类型因题型而异）
	Timestamp  time.Time `json:"timestamp"`   // 回答时间
}

// sessionQueue 单个会话的 FIFO 回答队列
//
// 每个 Q&A Session 拥有独立的回答队列。Enqueue 追加到尾部，
// Consume 一次性取走全部待消费回答并清空队列。
type sessionQueue struct {
	answers       []Answer      // 待消费的回答列表
	notifier      chan struct{} // 通知通道（队列非空时唤醒阻塞的消费者）
	mu            sync.Mutex    // answers 切片保护锁
	consumerCount int           // 等待中的消费者数量
	consumerMu    sync.Mutex    // 消费者计数锁
}

// AnswerQueue 全局回答队列管理器
//
// 管理所有活跃 Session 的回答队列。WebSocket 收到 answer_submit 时
// 调用 Enqueue 推入回答；MCP get_answer 工具调用 Consume 阻塞等待。
type AnswerQueue struct {
	sessions map[string]*sessionQueue // sessionID → 会话队列
	mu       sync.RWMutex             // sessions map 保护锁
	log      *xLog.LogNamedLogger     // 日志记录器
}

// 全局单例 AnswerQueue 实例
var (
	globalAnswerQueue *AnswerQueue
	answerQueueOnce   sync.Once
)

// GetAnswerQueue 获取或创建全局 AnswerQueue 单例
func GetAnswerQueue() *AnswerQueue {
	answerQueueOnce.Do(func() {
		globalAnswerQueue = NewAnswerQueue()
	})
	return globalAnswerQueue
}

// NewAnswerQueue 创建 AnswerQueue 实例
func NewAnswerQueue() *AnswerQueue {
	return &AnswerQueue{
		sessions: make(map[string]*sessionQueue),
		log:      xLog.WithName(xLog.NamedLOGC, "AnswerQueue"),
	}
}

// Enqueue 将用户回答推入指定会话的队列
//
// 非阻塞操作。WebSocket 收到 answer_submit 消息时调用。
// 如果会话队列不存在会自动创建。
func (aq *AnswerQueue) Enqueue(sessionID string, answer Answer) {
	q := aq.getOrCreateQueue(sessionID)

	q.mu.Lock()
	q.answers = append(q.answers, answer)
	q.mu.Unlock()

	// 非阻塞通知（缓冲为 1，已有待处理信号时无需重复）
	select {
	case q.notifier <- struct{}{}:
	default:
	}

	aq.log.Info(nil, "回答入队",
		slog.String("sessionID", sessionID),
		slog.String("questionID", answer.QuestionID),
		slog.Int("queueSize", aq.queueSize(q)),
	)
}

// Consume 阻塞消费指定会话的全部回答
//
// MCP get_answer 工具调用此方法等待用户回答。
// 如果队列中已有待消费回答，立即返回全部并清空队列。
// 如果队列为空，阻塞等待直到：收到回答 / 超时 / 上下文取消。
//
// timeout 为 0 表示无限等待（仅依赖 context 取消）。
// 超时或无回答时返回 (nil, nil)。
func (aq *AnswerQueue) Consume(ctx context.Context, sessionID string, timeout time.Duration) ([]Answer, error) {
	q := aq.getOrCreateQueue(sessionID)

	q.consumerMu.Lock()
	q.consumerCount++
	q.consumerMu.Unlock()

	defer func() {
		q.consumerMu.Lock()
		q.consumerCount--
		q.consumerMu.Unlock()
	}()

	// 快速路径：队列非空时直接取走
	if answers := aq.tryDrain(q); len(answers) > 0 {
		aq.log.Info(nil, "立即消费回答",
			slog.String("sessionID", sessionID),
			slog.Int("count", len(answers)),
		)
		return answers, nil
	}

	// 慢路径：队列为空，需要等待
	var timeoutCh <-chan time.Time
	if timeout > 0 {
		timeoutCh = time.After(timeout)
	}

	for {
		select {
		case <-q.notifier:
			// 被唤醒，尝试取走回答
			if answers := aq.tryDrain(q); len(answers) > 0 {
				aq.log.Info(nil, "等待后消费回答",
					slog.String("sessionID", sessionID),
					slog.Int("count", len(answers)),
				)
				return answers, nil
			}
			// 虚假唤醒，继续等待
			continue

		case <-timeoutCh:
			aq.log.Info(nil, "消费超时",
				slog.String("sessionID", sessionID),
				slog.Duration("timeout", timeout),
			)
			return nil, nil

		case <-ctx.Done():
			aq.log.Info(nil, "消费取消",
				slog.String("sessionID", sessionID),
				slog.String("error", ctx.Err().Error()),
			)
			return nil, ctx.Err()
		}
	}
}

// RemoveQueue 移除指定会话的队列（会话结束时调用）
func (aq *AnswerQueue) RemoveQueue(sessionID string) {
	aq.mu.Lock()
	defer aq.mu.Unlock()

	if _, ok := aq.sessions[sessionID]; ok {
		delete(aq.sessions, sessionID)
		aq.log.Info(nil, "移除会话队列", slog.String("sessionID", sessionID))
	}
}

// getOrCreateQueue 获取或创建指定会话的队列
func (aq *AnswerQueue) getOrCreateQueue(sessionID string) *sessionQueue {
	// 快速路径：读锁检查
	aq.mu.RLock()
	q, ok := aq.sessions[sessionID]
	aq.mu.RUnlock()
	if ok {
		return q
	}

	// 慢路径：写锁创建
	aq.mu.Lock()
	defer aq.mu.Unlock()

	// 双重检查（并发场景下可能已被其他 goroutine 创建）
	if q, ok = aq.sessions[sessionID]; ok {
		return q
	}

	q = &sessionQueue{
		answers:  make([]Answer, 0),
		notifier: make(chan struct{}, 1),
	}
	aq.sessions[sessionID] = q
	return q
}

// tryDrain 尝试一次性取走队列中的全部回答
//
// 线程安全：取走后创建新的 notifier 通道供后续消费者使用。
func (aq *AnswerQueue) tryDrain(q *sessionQueue) []Answer {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.answers) == 0 {
		return nil
	}

	// 复制并清空
	result := make([]Answer, len(q.answers))
	copy(result, q.answers)
	q.answers = q.answers[:0]

	// 重建 notifier 通道（确保下一次 Consume 能正确阻塞）
	q.notifier = make(chan struct{}, 1)

	return result
}

// queueSize 获取队列中的待消费回答数（调试用）
func (aq *AnswerQueue) queueSize(q *sessionQueue) int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.answers)
}

// GetConsumerCount 返回指定会话的当前等待消费者数量
func (aq *AnswerQueue) GetConsumerCount(sessionID string) int {
	aq.mu.RLock()
	q, exists := aq.sessions[sessionID]
	aq.mu.RUnlock()
	if !exists {
		return 0
	}
	q.consumerMu.Lock()
	defer q.consumerMu.Unlock()
	return q.consumerCount
}

// HasConsumer 检查指定会话是否有消费者在等待
func (aq *AnswerQueue) HasConsumer(sessionID string) bool {
	return aq.GetConsumerCount(sessionID) > 0
}
