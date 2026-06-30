package logic

import (
	"context"
	"fmt"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	"github.com/xiaolfeng/Lumina/internal/service"
)

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
