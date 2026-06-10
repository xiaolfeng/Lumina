package route

import (
	"time"

	xCtxUtil "github.com/bamboo-services/bamboo-base-go/common/utility/context"
	"github.com/gin-gonic/gin"
	"github.com/xiaolfeng/Lumina/internal/app/middleware"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"github.com/xiaolfeng/Lumina/internal/logic"
	"github.com/xiaolfeng/Lumina/internal/websocket"
)

func (r *route) wsRouter(route gin.IRouter) {
	db := xCtxUtil.MustGetDB(r.context)

	// 创建业务消息处理器
	msgHandler := websocket.CreateMessageHandler(db)

	// 创建 Hub 并注入消息处理器
	hub := websocket.GetHub(msgHandler)

	// 启动 Hub 主循环
	go hub.Run(r.context)

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

	// Q&A WebSocket 端点（需认证）
	wsGroup := route.Group("/qa")
	wsGroup.Use(middleware.Auth(r.context))
	wsGroup.GET("/ws", websocket.WSHandler(hub))
}
