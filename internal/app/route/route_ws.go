package route

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
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

	// 创建 Hub 并注入消息处理器（若 Hub 单例已由其他路由以 nil handler 创建，
	// 此处覆盖 handler，确保 Q&A 业务消息处理器最终生效）
	hub := websocket.GetHub(msgHandler)

	// 装配 Q&A 连接生命周期钩子（原 hub.go 中 Q&A 专属逻辑迁移至此，
	// 通过闭包捕获会话/问题/补充仓库与数据库实例，Hub 自身保持领域中立）
	questionRepo := repository.NewQaQuestionRepo(db)
	supplementRepo := repository.NewQaSupplementRepo(db)
	hooksLog := xLog.WithName(xLog.NamedCONT, "QAHooks")

	// syncOnlineDevices 异步更新数据库中的 OnlineDevices 字段
	syncOnlineDevices := func(sessionID string, count int) {
		sid, err := xSnowflake.ParseSnowflakeID(sessionID)
		if err != nil {
			return
		}
		db.Model(&entity.QaSession{}).Where("id = ?", sid).Update("online_devices", count)
	}

	// pushSessionHistory 向新连接推送该会话的全部问题历史及补充内容
	//
	// 连接建立时调用。按问题状态区分推送方式：
	//   - pending 问题 → question_push（前端视为待回答的活跃问题）
	//   - 已回答/已跳过 → history_question（前端仅作历史展示，不激活交互）
	//
	// 补充内容（supplement_push）仅推送给目标问题仍为 pending 的内容，
	// 已回答/已跳过的问题无需恢复补充面板，避免前端布局异常。
	pushSessionHistory := func(conn *websocket.Connection) {
		sid, err := xSnowflake.ParseSnowflakeID(conn.SessionID())
		if err != nil {
			return
		}

		// 查询该会话全部问题（含已回答/已跳过/待回答）
		questions, xErr := questionRepo.GetBySessionID(context.Background(), sid)
		if xErr != nil {
			hooksLog.Warn(nil, "推送会话历史失败", slog.String("error", xErr.Error()))
			return
		}

		// 按状态区分推送
		pendingIDs := make(map[string]bool, len(questions))
		for _, q := range questions {
			msgType := websocket.MsgHistoryQuestion
			if q.Status == "pending" {
				msgType = websocket.MsgQuestionPush
				pendingIDs[q.ID.String()] = true
			}
			msg := &websocket.Message{
				Type:      msgType,
				SessionID: conn.SessionID(),
				Data:      q,
				Timestamp: time.Now().UnixMilli(),
			}
			_ = conn.SendMessage(msg)
		}

		// 构建 optionID → questionID 映射，用于判断 option 级 supplement 所属问题状态
		optionToQuestion := make(map[string]string)
		for _, q := range questions {
			if q.Options != nil {
				var opts []map[string]interface{}
				if json.Unmarshal(q.Options, &opts) == nil {
					for _, opt := range opts {
						if optID, ok := opt["id"].(string); ok && optID != "" {
							optionToQuestion[optID] = q.ID.String()
						}
					}
				}
			}
		}

		// 推送该会话所有 supplement（仅目标问题仍为 pending 的才推送）
		supplements, sErr := supplementRepo.GetBySessionID(context.Background(), sid)
		if sErr != nil {
			hooksLog.Warn(nil, "查询会话补充内容失败", slog.String("error", sErr.Error()))
			return
		}
		for _, s := range supplements {
			var belongsToPending bool
			if s.TargetType == "question" {
				belongsToPending = pendingIDs[s.TargetID.String()]
			} else if s.TargetType == "option" {
				qid := optionToQuestion[s.TargetID.String()]
				belongsToPending = pendingIDs[qid]
			}
			if !belongsToPending {
				continue
			}
			msg := &websocket.Message{
				Type:      websocket.MsgSupplementPush,
				SessionID: conn.SessionID(),
				Data:      s,
				Timestamp: time.Now().UnixMilli(),
			}
			_ = conn.SendMessage(msg)
		}
	}

	// 注册钩子：仅处理 Q&A 连接（Kind == "qa"），Preview 连接不参与 Q&A 生命周期
	hub.AddRegisterHook(func(conn *websocket.Connection) {
		if conn.Kind != "qa" {
			return
		}

		// 异步同步 OnlineDevices 到数据库
		go syncOnlineDevices(conn.SessionID(), hub.GetOnlineDevices(conn.SessionID()))

		// 向新连接推送该会话的全部问题历史及补充内容
		go pushSessionHistory(conn)
	})

	hub.AddUnregisterHook(func(conn *websocket.Connection) {
		if conn.Kind != "qa" {
			return
		}

		// 异步同步 OnlineDevices 到数据库
		go syncOnlineDevices(conn.SessionID(), hub.GetOnlineDevices(conn.SessionID()))
	})

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
