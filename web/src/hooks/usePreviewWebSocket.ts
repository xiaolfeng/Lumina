import { useCallback, useEffect, useRef, useState } from 'react'

// ── WebSocket Message Types ──

export type WsMessageType = 'preview_sync' | 'heartbeat' | 'heartbeat_ack'

export interface WsMessage {
  type: WsMessageType
  session_id?: string
  data?: any
  timestamp: number
}

export type ConnectionStatus = 'idle' | 'connecting' | 'connected' | 'disconnected' | 'rejected'

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
  const scheduleReconnectRef = useRef<() => void>(() => {})
  optionsRef.current = options

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

  // ── Connect ──

  const connect = useCallback(() => {
    if (!sessionHash) return

    // Close existing connection
    if (wsRef.current) {
      wsRef.current.close()
    }

    everConnectedRef.current = false
    rejectedRef.current = false
    setStatus('connecting')

    // Build WebSocket URL (upgrade http → ws)
    // /api/v1/preview/ws 为公开端点（hash 鉴权），无需携带 token
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    const deviceId = localStorage.getItem('preview_device_id') || generateDeviceId()
    localStorage.setItem('preview_device_id', deviceId)

    const wsUrl = `${protocol}//${host}/api/v1/preview/ws?session=${sessionHash}&device_id=${deviceId}`

    const ws = new WebSocket(wsUrl)
    wsRef.current = ws

    // ── Open ──

    ws.onopen = () => {
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
      setStatus('disconnected')
      optionsRef.current.onStatusChange?.('disconnected')
      clearTimers()
      wsRef.current = null

      if (!everConnectedRef.current) {
        rejectedRef.current = true
        setStatus('rejected')
        optionsRef.current.onReject?.()
        return
      }

      scheduleReconnectRef.current()
    }

    // ── Error ──

    ws.onerror = () => {
      // onclose will fire after onerror
    }
  }, [sessionHash, disconnect, clearTimers])

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
  return 'dev_' + Math.random().toString(36).substring(2, 10)
}
