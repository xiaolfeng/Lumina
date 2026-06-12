import { useState, useCallback } from 'react'
import { useQaWebSocket } from './useQaWebSocket'
import type {
  Question,
  Session,
  SupplementItem,
} from '#/components/interact/types'
import { getQuestionDetail } from '#/lib/apis/qa-admin'

// ── Options ──

interface UseQaSessionOptions {
  sessionHash: string | null
  onReject?: () => void
}

// ── Hook ──

export function useQaSession({ sessionHash, onReject }: UseQaSessionOptions) {
  const [questions, setQuestions] = useState<Question[]>([])
  const [activeSupplement, setActiveSupplement] =
    useState<SupplementItem | null>(null)
  const [session, setSession] = useState<Session | null>(null)
  const [isSupplementLoading, setIsSupplementLoading] = useState(false)

  // ── Question Push Handler ──

  const handleQuestionPush = useCallback(
    (data: any) => {
      const hasSupplement = data.supplement || data.Supplement || false
      const question: Question = {
        id: data.id || data.ID,
        sessionId: sessionHash || '',
        content: data.title || data.Title || '',
        description: data.description || data.Description,
        type: data.type || data.Type || 'text',
        options: data.options || data.Options,
        allowOther: data.allow_other,
        groupLabel: data.group_label || '',
        batch: data.batch,
        config: data.config,
        status: 'pending',
        answered: false,
        answer: undefined,
        supplements: [],
        supplement: hasSupplement,
        createdAt: data.created_at || new Date().toISOString(),
      }
      setQuestions((prev) => [...prev, question])
      if (hasSupplement) {
        setIsSupplementLoading(true)
      }
    },
    [sessionHash],
  )

  // ── Supplement Push Handler ──

  const handleSupplementPush = useCallback((data: any) => {
    const supplement: SupplementItem = {
      id: data.id || data.ID,
      target_type: data.target_type || data.TargetType || 'question',
      target_id: data.target_id || data.TargetID || '',
      content_type: data.content_type || data.ContentType || 'markdown',
      content: data.content || data.Content || '',
      created_at: data.created_at || new Date().toISOString(),
      updated_at: data.updated_at || new Date().toISOString(),
    }
    setActiveSupplement(supplement)
    setIsSupplementLoading(false)
    // Also add to relevant question's supplements
    setQuestions((prev) =>
      prev.map((q) => {
        if (
          supplement.target_type === 'question' &&
          q.id === supplement.target_id
        ) {
          return {
            ...q,
            supplements: [...(q.supplements || []), supplement],
          }
        }
        return q
      }),
    )
  }, [])

  // ── Answer Sync Handler ──

  const handleAnswerSync = useCallback((data: any) => {
    setQuestions((prev) =>
      prev.map((q) => {
        if (q.id === data.question_id) {
          return {
            ...q,
            status: data.status,
            answered: data.status === 'answered',
          }
        }
        return q
      }),
    )
  }, [])

  // ── WS Reconnect: restore supplements for active pending question ──

  const handleReconnect = useCallback(async () => {
    // Find the first pending question that had supplement flag
    // Use functional state read to avoid stale closure
    setQuestions((currentQuestions) => {
      const pendingWithSupplement = currentQuestions.find(
        (q) => q.status === 'pending' && q.supplement,
      )
      if (pendingWithSupplement && sessionHash) {
        // Fire-and-forget: fetch supplement details from REST API
        getQuestionDetail(sessionHash, pendingWithSupplement.id)
          .then((res) => {
            const detail = res.data
            if (detail?.supplements?.length) {
              // Restore the latest supplement as active
              const latest =
                detail.supplements[detail.supplements.length - 1]
              setActiveSupplement(latest)
              setIsSupplementLoading(false)
            }
          })
          .catch(() => {
            // Silently ignore — supplement is best-effort on reconnect
          })
      }
      return currentQuestions // no mutation, just reading
    })
  }, [sessionHash])

  // ── WebSocket Connection ──

  const ws = useQaWebSocket(sessionHash, {
    onQuestionPush: handleQuestionPush,
    onSupplementPush: handleSupplementPush,
    onAnswerSync: handleAnswerSync,
    onReject,
    onConnect: handleReconnect,
  })

  // ── Actions ──

  const submitAnswer = useCallback(
    (questionId: string, answer: any) => {
      ws.sendMessage('answer_submit', { question_id: questionId, answer })
      // Optimistic update
      setQuestions((prev) =>
        prev.map((q) =>
          q.id === questionId
            ? { ...q, status: 'answered' as const, answered: true, answer }
            : q,
        ),
      )
      // Clear active supplement panel after submission
      setActiveSupplement(null)
    },
    [ws],
  )

  const skipQuestion = useCallback(
    (questionId: string) => {
      ws.sendMessage('skip', { question_id: questionId })
      setQuestions((prev) =>
        prev.map((q) =>
          q.id === questionId
            ? { ...q, status: 'skipped' as const }
            : q,
        ),
      )
    },
    [ws],
  )

  const requestSupplement = useCallback(
    (targets: string[]) => {
      ws.sendMessage('request_supplement', { targets })
    },
    [ws],
  )

  return {
    questions,
    activeSupplement,
    isSupplementLoading,
    session,
    setSession,
    connectionStatus: ws.status,
    submitAnswer,
    skipQuestion,
    requestSupplement,
    connect: ws.connect,
    disconnect: ws.disconnect,
  }
}
