package route

import (
	"time"

	xCtxUtil "github.com/bamboo-services/bamboo-base-go/major/utility/context"
	"github.com/gin-gonic/gin"
	"github.com/xiaolfeng/Lumina/internal/app/middleware"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"github.com/xiaolfeng/Lumina/internal/logic"
	"github.com/xiaolfeng/Lumina/internal/repository"
	"github.com/xiaolfeng/Lumina/internal/service"
	"github.com/xiaolfeng/Lumina/internal/websocket"
)

func (r *route) wsRouter(route gin.IRouter) {
	db := xCtxUtil.MustGetDB(r.context)
	rdb := xCtxUtil.MustGetRDB(r.context)

	// 创建 Session 仓库（用于 WebSocket 连接时 Hash→ID 解析）
	sessionRepo := repository.NewQaSessionRepo(db, rdb)

	// 创建业务消息处理器（注入媒体回答处理服务）
	mediaSvc := service.NewMediaAnswerService()
	msgHandler := websocket.CreateMessageHandler(db, mediaSvc)

	// 创建 Hub 并注入消息处理器、Session 仓库和数据库实例
	hub := websocket.GetHub(msgHandler, sessionRepo, db)

	// Hub 主循环不在此启动：RouteRegistrar 传入的 ctx 未经 Runner.WithCancel 包裹，
	// 若在此 `go hub.Run(r.context)` 会导致 ctx.Done() 永不触发、WebSocket 无法优雅关闭。
	// 改由 startup.NewWebSocketRunner 通过 xMain.Runner 的 goroutineFunc 接收可取消 ctx 后启动。

	// 设置 QaLogic 回调，使其推送问题时通过 WebSocket 广播到在线设备
	logic.OnQuestionPushed = func(sessionID string, question *entity.QaQuestion) {
		msg := &websocket.Message{
			Type:      websocket.MsgQuestionPush,
			SessionID: sessionID,
			Data:      question,
			Timestamp: time.Now().UnixMilli(),
		}
		hub.BroadcastToSession(sessionID, msg)
	}

	logic.OnSupplementPushed = func(sessionID string, supplement *entity.QaSupplement) {
		msg := &websocket.Message{
			Type:      websocket.MsgSupplementPush,
			SessionID: sessionID,
			Data:      supplement,
			Timestamp: time.Now().UnixMilli(),
		}
		hub.BroadcastToSession(sessionID, msg)
	}

	// 问题取消回调：question 为 nil 时表示全部取消，非 nil 时表示单个问题取消
	logic.OnQuestionCancelled = func(sessionID string, question *entity.QaQuestion) {
		msg := &websocket.Message{
			Type:      websocket.MsgQuestionCancel,
			SessionID: sessionID,
			Data: map[string]interface{}{
				"question_id": func() string {
					if question != nil {
						return question.ID.String()
					}
					return ""
				}(),
				"cancel_all": question == nil,
			},
			Timestamp: time.Now().UnixMilli(),
		}
		hub.BroadcastToSession(sessionID, msg)
	}

	logic.OnSessionArchived = func(sessionID string) {
		msg := &websocket.Message{
			Type:      websocket.MsgSessionEnd,
			SessionID: sessionID,
			Data: map[string]interface{}{
				"reason": "archived",
			},
			Timestamp: time.Now().UnixMilli(),
		}
		hub.BroadcastToSession(sessionID, msg)
	}

	// Q&A WebSocket 端点（需认证）
	wsGroup := route.Group("/qa")
	wsGroup.Use(middleware.Auth(r.context))
	wsGroup.GET("/ws", websocket.WSHandler(hub, sessionRepo))
}
