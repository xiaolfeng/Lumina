import { useCallback, useEffect, useRef, useState } from 'react'

// ── WebSocket Message Types ──

export type WsMessageType =
  | 'question_push'
  | 'supplement_push'
  | 'answer_sync'
  | 'heartbeat'
  | 'heartbeat_ack'
  | 'session_end'
  | 'answer_submit'
  | 'request_supplement'
  | 'skip'

export interface WsMessage {
  type: WsMessageType
  session_id?: string
  data?: any
  timestamp: number
}

export type ConnectionStatus = 'idle' | 'connecting' | 'connected' | 'disconnected'

// ── Hook Options ──

interface UseQaWebSocketOptions {
  /** Callback for when a question is pushed */
  onQuestionPush?: (data: any) => void
  /** Callback for when a supplement is pushed */
  onSupplementPush?: (data: any) => void
  /** Callback for answer sync from other devices */
  onAnswerSync?: (data: any) => void
  /** Callback for session end */
  onSessionEnd?: () => void
  /** Callback for connection status changes */
  onStatusChange?: (status: ConnectionStatus) => void
}

// ── Constants ──

const HEARTBEAT_INTERVAL = 5000 // 5s
const RECONNECT_DELAYS = [1000, 2000, 4000, 8000, 16000] // exponential backoff

// ── Hook ──

export function useQaWebSocket(
  sessionId: string | null,
  options: UseQaWebSocketOptions = {},
) {
  const [status, setStatus] = useState<ConnectionStatus>('idle')
  const wsRef = useRef<WebSocket | null>(null)
  const heartbeatTimerRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const reconnectAttemptRef = useRef(0)
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const optionsRef = useRef(options)
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

  // ── Schedule Reconnect ──

  const connect = useCallback(() => {
    if (!sessionId) return

    // Close existing connection
    if (wsRef.current) {
      wsRef.current.close()
    }

    setStatus('connecting')

    // Build WebSocket URL (upgrade http → ws)
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    const deviceId = localStorage.getItem('qa_device_id') || generateDeviceId()
    localStorage.setItem('qa_device_id', deviceId)

    const token = getAccessToken()
    const wsUrl = `${protocol}//${host}/api/v1/qa/ws?session_id=${sessionId}&device_id=${deviceId}&token=${token}`

    const ws = new WebSocket(wsUrl)
    wsRef.current = ws

    // ── Open ──

    ws.onopen = () => {
      setStatus('connected')
      reconnectAttemptRef.current = 0
      optionsRef.current.onStatusChange?.('connected')

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
          case 'question_push':
            optionsRef.current.onQuestionPush?.(msg.data)
            break
          case 'supplement_push':
            optionsRef.current.onSupplementPush?.(msg.data)
            break
          case 'answer_sync':
            optionsRef.current.onAnswerSync?.(msg.data)
            break
          case 'session_end':
            optionsRef.current.onSessionEnd?.()
            disconnect()
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
      scheduleReconnect()
    }

    // ── Error ──

    ws.onerror = () => {
      // onclose will fire after onerror
    }
  }, [sessionId, disconnect, clearTimers])

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

  // ── Lifecycle ──

  useEffect(() => {
    if (sessionId) {
      connect()
    }
    return () => {
      disconnect()
    }
  }, [sessionId, connect, disconnect])

  // ── Send Message ──

  const sendMessage = useCallback(
    (type: WsMessageType, data?: any) => {
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        wsRef.current.send(
          JSON.stringify({
            type,
            session_id: sessionId,
            data,
            timestamp: Date.now(),
          }),
        )
      }
    },
    [sessionId],
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

function getAccessToken(): string {
  const match = document.cookie.match(/(?:^|;\s*)access_token=([^;]*)/)
  return match ? decodeURIComponent(match[1]) : ''
}
