package route

import (
	"time"

	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xCtxUtil "github.com/bamboo-services/bamboo-base-go/major/utility/context"
	"github.com/gin-gonic/gin"
	"github.com/xiaolfeng/Lumina/internal/app/middleware"
	"github.com/xiaolfeng/Lumina/internal/handler"
	"github.com/xiaolfeng/Lumina/internal/logic"
	"github.com/xiaolfeng/Lumina/internal/repository"
	"github.com/xiaolfeng/Lumina/internal/websocket"
)

// previewRouter 注册预览模块公开读取路由与管理路由。
func (r *route) previewRouter(route gin.IRouter) {
	previewHandler := handler.NewHandler[handler.PreviewHandler](r.context, "PreviewHandler")

	// 构建 Preview 模块依赖（会话仓储 + 业务逻辑 + WebSocket Hub 单例）
	db := xCtxUtil.MustGetDB(r.context)
	previewSessionRepo := repository.NewPreviewSessionRepo(db)
	previewLogic := logic.NewPreviewLogic(r.context)
	hub := websocket.GetHub(nil)

	// 注册连接快照推送钩子：预览连接建立后异步推送完整会话详情（session + files）
	hub.AddRegisterHook(func(conn *websocket.Connection) {
		// 仅处理 Preview 类型连接，Q&A 连接的快照推送由 wsRouter 的钩子负责
		if conn.Kind != "preview" {
			return
		}

		sid, err := xSnowflake.ParseSnowflakeID(conn.SessionID())
		if err != nil {
			return
		}

		detail, xErr := previewLogic.GetSessionDetailByID(r.context, sid)
		if xErr != nil {
			return
		}

		_ = conn.SendMessage(&websocket.Message{
			Type:      websocket.MsgPreviewSync,
			SessionID: conn.SessionID(),
			Data:      detail,
			Timestamp: time.Now().UnixMilli(),
		})
	})

	// 设置 PreviewLogic 回调，使文件/会话变更时通过 WebSocket 广播 preview_sync 到在线设备
	logic.OnPreviewChanged = func(sessionID string, eventType string) {
		var data any = map[string]interface{}{
			"event_type": eventType,
		}

		// 会话仍存在时附带最新详情（session + files）；会话删除后仅携带事件标记
		if sid, err := xSnowflake.ParseSnowflakeID(sessionID); err == nil {
			if detail, xErr := previewLogic.GetSessionDetailByID(r.context, sid); xErr == nil {
				data = detail
			}
		}

		hub.BroadcastToSession(sessionID, &websocket.Message{
			Type:      websocket.MsgPreviewSync,
			SessionID: sessionID,
			Data:      data,
			Timestamp: time.Now().UnixMilli(),
		})
	}

	// 公开读取路由（hash 鉴权，无需登录）
	publicGroup := route.Group("/preview")
	publicGroup.GET("/sessions/:hash", previewHandler.GetSession)
	publicGroup.GET("/sessions/:hash/files/:filename", previewHandler.GetFile)
	publicGroup.GET("/files/:id", previewHandler.GetFileByID)
	publicGroup.GET("/ws", websocket.PreviewWSHandler(hub, previewSessionRepo))

	// 管理路由（Bearer Token 认证）
	adminGroup := route.Group("/preview")
	adminGroup.Use(middleware.Auth(r.context))
	adminGroup.POST("/sessions", previewHandler.CreateSession)
	adminGroup.GET("/sessions", previewHandler.ListSessions)
	adminGroup.DELETE("/sessions/:id", previewHandler.DeleteSession)
	adminGroup.DELETE("/files/:id", previewHandler.DeleteFile)
}
