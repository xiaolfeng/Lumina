import { useCallback, useEffect, useRef, useState } from 'react'

import {
  getPreviewSessionDetail,
  isPreviewSessionGoneError,
} from '#/lib/apis/preview'

// ── WebSocket Message Types ──

export type WsMessageType = 'preview_sync' | 'heartbeat' | 'heartbeat_ack'

export interface WsMessage {
  type: WsMessageType
  session_id?: string
  data?: any
  timestamp: number
}

export type ConnectionStatus =
  'idle' | 'connecting' | 'connected' | 'disconnected' | 'rejected'

// ── Hook Options ──

interface UsePreviewWebSocketOptions {
  /** Callback for preview sync (connection snapshot / file change push) */
  onSync?: (data: any) => void
  /** Callback for connection status changes */
  onStatusChange?: (status: ConnectionStatus) => void
  /** Callback for when the server rejects the connection (session invalid/expired) */
  onReject?: () => void
  /** Callback for when the connection is established (initial + reconnect) */
  onConnect?: () => void
}

// ── Constants ──

const HEARTBEAT_INTERVAL = 5000 // 5s
const RECONNECT_DELAYS = [1000, 2000, 4000, 8000, 16000] // exponential backoff

// ── Hook ──

export function usePreviewWebSocket(
  sessionHash: string | null,
  options: UsePreviewWebSocketOptions = {},
) {
  const [status, setStatus] = useState<ConnectionStatus>('idle')
  const wsRef = useRef<WebSocket | null>(null)
  const heartbeatTimerRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const reconnectAttemptRef = useRef(0)
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const optionsRef = useRef(options)
  const everConnectedRef = useRef(false)
  const rejectedRef = useRef(false)
  const sessionHashRef = useRef(sessionHash)
  const scheduleReconnectRef = useRef<() => void>(() => {})
  optionsRef.current = options
  sessionHashRef.current = sessionHash

  // ── Timer Cleanup ──

  const clearTimers = useCallback(() => {
    if (heartbeatTimerRef.current) {
      clearInterval(heartbeatTimerRef.current)
      heartbeatTimerRef.current = null
    }
    if (reconnectTimerRef.current) {
      clearTimeout(reconnectTimerRef.current)
      reconnectTimerRef.current = null
    }
  }, [])

  // ── Disconnect ──

  const disconnect = useCallback(() => {
    clearTimers()
    reconnectAttemptRef.current = 0
    if (wsRef.current) {
      wsRef.current.close()
      wsRef.current = null
    }
    setStatus('disconnected')
  }, [clearTimers])

  // ── REST Probe ──

  // 断线/握手失败后探活确认会话是否确实已被删除：仅 404/410 判 rejected 并撤销
  // 重连；网络错误或其它状态码一律维持指数退避重试（探活异常不得抛出未捕获错误）
  const probeSession = useCallback(
    (hash: string) => {
      void getPreviewSessionDetail(hash).catch((error: unknown) => {
        if (!isPreviewSessionGoneError(error)) return
        // 会话切换后到达的迟到探活结果直接忽略，避免误杀新会话状态
        if (sessionHashRef.current !== hash) return
        if (wsRef.current?.readyState === WebSocket.OPEN) return
        rejectedRef.current = true
        clearTimers()
        setStatus('rejected')
        optionsRef.current.onReject?.()
      })
    },
    [clearTimers],
  )

  // ── Connect ──

  const connect = useCallback(() => {
    // 会话已被探活确认删除后不再自动重连，避免对已删会话的无效循环
    if (!sessionHash || rejectedRef.current) return

    // Close existing connection
    if (wsRef.current) {
      wsRef.current.close()
    }

    setStatus('connecting')

    // Build WebSocket URL (upgrade http → ws)
    // /api/v1/preview/ws 需登录，同源 Cookie 会随 WebSocket 升级自动携带
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    // 每次建连生成全新 device_id，不持久化：localStorage 在同源所有标签页共享，
    // 复用同一 device_id 会让后打开的标签页覆盖 Hub 中旧连接的映射，关闭旧标签页
    // 时误删存活连接（孤儿连接注销问题）
    const deviceId = generateDeviceId()

    const wsUrl = `${protocol}//${host}/api/v1/preview/ws?session=${sessionHash}&device_id=${deviceId}`

    // 握手失败统一处理：先按指数退避排期重连，REST 探活确认会话已被删除后再定格 rejected
    const handleHandshakeFailure = () => {
      scheduleReconnectRef.current()
      probeSession(sessionHash)
    }

    let ws: WebSocket
    try {
      ws = new WebSocket(wsUrl)
    } catch (e) {
      // WS 构造直接抛出（环境禁用 / URL 非法）：视作一次握手失败，走探活 + 重试
      console.error('Failed to construct WebSocket:', e)
      handleHandshakeFailure()
      return
    }
    wsRef.current = ws

    // ── Open ──

    ws.onopen = () => {
      // 代际保护：旧连接的 open/close 晚到时已被新连接替换，忽略不再操作共享状态
      if (wsRef.current !== ws) return
      setStatus('connected')
      everConnectedRef.current = true
      reconnectAttemptRef.current = 0
      optionsRef.current.onStatusChange?.('connected')
      optionsRef.current.onConnect?.()

      // Start heartbeat
      heartbeatTimerRef.current = setInterval(() => {
        if (ws.readyState === WebSocket.OPEN) {
          ws.send(
            JSON.stringify({
              type: 'heartbeat_ack',
              timestamp: Date.now(),
            }),
          )
        }
      }, HEARTBEAT_INTERVAL)
    }

    // ── Message ──

    ws.onmessage = (event) => {
      try {
        const msg: WsMessage = JSON.parse(event.data)
        switch (msg.type) {
          case 'preview_sync':
            optionsRef.current.onSync?.(msg.data)
            break
          case 'heartbeat':
            ws.send(
              JSON.stringify({ type: 'heartbeat_ack', timestamp: Date.now() }),
            )
            break
        }
      } catch (e) {
        console.error('Failed to parse WebSocket message:', e)
      }
    }

    // ── Close ──

    ws.onclose = () => {
      // 代际保护：若当前 wsRef 已指向更新连接（新连接建立、或 disconnect 置空），
      // 本连接为被替换的旧连接，其关闭事件不再改动共享状态/触发重连判断
      if (wsRef.current !== ws) return
      setStatus('disconnected')
      optionsRef.current.onStatusChange?.('disconnected')
      clearTimers()
      wsRef.current = null

      // 断线（含重连握手失败）一律先按指数退避重试，由 REST 探活确认会话确实
      // 已被删除后才定格 rejected：网络瞬断 / 服务重启不再被误判为「会话不存在」
      handleHandshakeFailure()
    }

    // ── Error ──

    ws.onerror = () => {
      // onclose will fire after onerror
    }
  }, [sessionHash, clearTimers, probeSession])

  // ── Reconnect with exponential backoff ──

  const scheduleReconnect = useCallback(() => {
    const delay =
      RECONNECT_DELAYS[
        Math.min(reconnectAttemptRef.current, RECONNECT_DELAYS.length - 1)
      ]
    reconnectAttemptRef.current++
    reconnectTimerRef.current = setTimeout(() => {
      connect()
    }, delay)
  }, [connect])

  scheduleReconnectRef.current = scheduleReconnect

  // ── Lifecycle ──

  useEffect(() => {
    if (sessionHash) {
      // 连接纪元仅在挂载/会话切换时重置：纪元内曾成功建立过连接，则后续任何
      // 一次重连握手失败都不得直接判 rejected（交由 REST 探活确认）
      everConnectedRef.current = false
      rejectedRef.current = false
      connect()
    }
    return () => {
      disconnect()
    }
  }, [sessionHash, connect, disconnect])

  // ── Send Message ──

  const sendMessage = useCallback(
    (type: WsMessageType, data?: any) => {
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        wsRef.current.send(
          JSON.stringify({
            type,
            session_id: sessionHash,
            data,
            timestamp: Date.now(),
          }),
        )
      }
    },
    [sessionHash],
  )

  return {
    status,
    connect,
    disconnect,
    sendMessage,
  }
}

// ── Helpers ──

function generateDeviceId(): string {
  // crypto.randomUUID 需安全上下文（https/localhost），不可用时回退 Math.random
  const uuid =
    typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function'
      ? crypto.randomUUID().slice(0, 8)
      : Math.random().toString(36).substring(2, 10)
  return 'dev_' + uuid
}
