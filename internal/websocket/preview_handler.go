package websocket

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xiaolfeng/Lumina/internal/repository"
)

// maxPreviewConnPerSession 单个预览会话的最大并发连接数
//
// 防止同一会话被大量 WebSocket 连接占用资源（如异常客户端反复建连、
// 连接泄漏导致 Hub 会话映射无限增长），达到上限后新连接直接拒绝。
const maxPreviewConnPerSession = 50

// PreviewWSHandler 创建预览 WebSocket 升级 Gin 处理器
//
// 升级流程：
//  1. 从 query 中提取 session（访问哈希，必填）和 device_id（可选，不提供则自动生成）
//  2. 通过访问哈希解析预览会话（PreviewSessionRepo.GetByHash）
//  3. 校验会话状态为 active（已删除会话返回 410）
//  4. 校验会话在线连接数未达上限（超出返回 429）
//  5. 将 HTTP 连接升级为 WebSocket
//  6. 创建 Connection 并注册到 Hub（Kind 标记为 "preview"）
//  7. 启动 ReadPump 和 WritePump goroutine
//
// 连接建立后由 Hub 的注册钩子（上层注入）异步推送会话快照与文件列表，
// 文件变更经 preview_sync 消息实时同步到所有在线设备。
func PreviewWSHandler(hub *Hub, sessionRepo *repository.PreviewSessionRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 提取 session hash（必填）
		sessionHash := c.Query("session")
		if sessionHash == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "session 参数必填"})
			c.Abort()
			return
		}

		// 2. 通过访问哈希解析预览会话
		session, xErr := sessionRepo.GetByHash(c.Request.Context(), sessionHash)
		if xErr != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "预览会话不存在或已过期"})
			c.Abort()
			return
		}

		// 3. 校验会话状态 —— 已删除会话拒绝连接，前端会触发 onReject 展示会话不存在
		if session.Status != "active" {
			c.JSON(http.StatusGone, gin.H{"error": "会话已删除", "status": session.Status})
			c.Abort()
			return
		}

		sessionID := session.ID.String() // snowflake ID 字符串，Hub 内部仍使用雪花 ID

		// 4. 提取或生成 device_id
		deviceID := c.Query("device_id")
		if deviceID == "" {
			deviceID = fmt.Sprintf("device_%s", uuid.New().String()[:8])
		}

		// 4.5 校验会话在线连接数未达上限 —— 达到上限时拒绝新连接，避免资源被单会话耗尽
		if hub.GetOnlineDevices(sessionID) >= maxPreviewConnPerSession {
			wsLog.Warn(nil, "预览会话连接数已达上限，拒绝连接",
				slog.String("sessionID", sessionID),
				slog.String("deviceID", deviceID),
				slog.Int("limit", maxPreviewConnPerSession),
			)
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "会话连接数已达上限", "limit": maxPreviewConnPerSession})
			c.Abort()
			return
		}

		// 5. 升级 HTTP 连接为 WebSocket
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			wsLog.Error(nil, "预览 WebSocket 升级失败", slog.String("error", err.Error()))
			return
		}

		wsLog.Info(nil, "预览 WebSocket 连接建立",
			slog.String("sessionHash", sessionHash),
			slog.String("sessionID", sessionID),
			slog.String("deviceID", deviceID),
		)

		// 6. 创建连接封装并注册到 Hub（标记 Kind 为 preview，供业务钩子按类型分发）
		wsConn := NewConnection(conn, sessionID, deviceID, hub)
		wsConn.Kind = "preview"
		wsConn.sessionHash = sessionHash
		hub.Register(wsConn)

		// 7. 启动读写泵
		go wsConn.WritePump()
		go wsConn.ReadPump()
	}
}
