import { useCallback, useEffect, useRef, useState } from 'react'

// ── WebSocket Message Types ──

export type WsMessageType =
  | 'question_push'
  | 'history_question'
  | 'supplement_push'
  | 'answer_sync'
  | 'heartbeat'
  | 'heartbeat_ack'
  | 'session_end'
  | 'answer_submit'
  | 'request_supplement'
  | 'skip'
  | 'answer_unhandled'

export interface AnswerUnhandledData {
  question_id: string
  answer: any
  message: string
}

export interface WsMessage {
  type: WsMessageType
  session_id?: string
  data?: any
  timestamp: number
}

export type ConnectionStatus = 'idle' | 'connecting' | 'connected' | 'disconnected' | 'rejected'

// ── Hook Options ──

interface UseQaWebSocketOptions {
  /** Callback for when a question is pushed (pending) */
  onQuestionPush?: (data: any) => void
  /** Callback for when a history question is pushed (answered/skipped) */
  onHistoryQuestion?: (data: any) => void
  /** Callback for when a supplement is pushed */
  onSupplementPush?: (data: any) => void
  /** Callback for answer sync from other devices */
  onAnswerSync?: (data: any) => void
  /** Callback for session end */
  onSessionEnd?: () => void
  /** Callback for when the server rejects an answer (unhandled) */
  onAnswerUnhandled?: (data: AnswerUnhandledData) => void
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

export function useQaWebSocket(
  sessionHash: string | null,
  options: UseQaWebSocketOptions = {},
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

  // ── Schedule Reconnect ──

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
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    const deviceId = localStorage.getItem('qa_device_id') || generateDeviceId()
    localStorage.setItem('qa_device_id', deviceId)

    const token = getAccessToken()
    const wsUrl = `${protocol}//${host}/api/v1/qa/ws?session=${sessionHash}&device_id=${deviceId}&token=${token}`

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
          case 'question_push':
            optionsRef.current.onQuestionPush?.(msg.data)
            break
          case 'history_question':
            optionsRef.current.onHistoryQuestion?.(msg.data)
            break
          case 'supplement_push':
            optionsRef.current.onSupplementPush?.(msg.data)
            break
          case 'answer_sync':
            optionsRef.current.onAnswerSync?.(msg.data)
            break
          case 'answer_unhandled':
            optionsRef.current.onAnswerUnhandled?.(msg.data)
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

  // ── Beforeunload: send session_leave on page close (skip refresh) ──

  useEffect(() => {
    function handleBeforeUnload() {
      // Skip on refresh — navigation type 'reload' means F5/Cmd+R
      const navEntries = performance.getEntriesByType('navigation')
      if (
        navEntries.length > 0 &&
        (navEntries[0] as PerformanceNavigationTiming).type === 'reload'
      ) {
        return
      }

      // Best-effort: try to send session_leave via WS before tab closes
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        wsRef.current.send(
          JSON.stringify({
            type: 'session_leave',
            timestamp: Date.now(),
          }),
        )
      }
    }

    window.addEventListener('beforeunload', handleBeforeUnload)
    return () => window.removeEventListener('beforeunload', handleBeforeUnload)
  }, [])

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
