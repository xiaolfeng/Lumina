import { useState, useCallback, useEffect, useRef } from 'react'
import { toast } from 'sonner'
import { useQaWebSocket } from './useQaWebSocket'
import type {
  Question,
  Session,
  SupplementItem,
} from '#/components/interact/types'

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

  // questions 的同步镜像 —— WS handler 里读取最新状态判断目标问题是否 pending，
  // 避免依赖 setState updater 的异步时序（updater 不保证在 setState 返回前同步执行）。
  const questionsRef = useRef<Question[]>([])

  const updateQuestions = useCallback(
    (updater: (prev: Question[]) => Question[]) => {
      setQuestions((prev) => {
        const next = updater(prev)
        questionsRef.current = next
        return next
      })
    },
    [],
  )

  // 切换会话时清空上一个会话的残留状态，避免跨会话数据污染
  // （questions/activeSupplement/isSupplementLoading/session 均为本地 state，
  //  WebSocket 重连后由后端重新推送，不会丢失）
  useEffect(() => {
    questionsRef.current = []
    setQuestions([])
    setActiveSupplement(null)
    setIsSupplementLoading(false)
    setSession(null)
  }, [sessionHash])

  const startSupplementLoading = useCallback(() => {
    setIsSupplementLoading(true)
  }, [])

  const stopSupplementLoading = useCallback(() => {
    setIsSupplementLoading(false)
  }, [])

  // 用户手动忽略补充加载（恢复可点击，不等待 Agent 推送）
  const dismissSupplementLoading = useCallback(() => {
    setIsSupplementLoading(false)
  }, [])

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
      updateQuestions((prev) => [...prev, question])
      if (hasSupplement) {
        startSupplementLoading()
      }
    },
    [sessionHash, startSupplementLoading, updateQuestions],
  )

  // ── History Question Handler ──
  // 连接恢复时后端推送已回答/已跳过的历史问题，仅作展示，不激活交互

  const handleHistoryQuestion = useCallback(
    (data: any) => {
      const qid = data.id || data.ID
      const status = (data.status ||
        data.Status ||
        'answered') as Question['status']
      updateQuestions((prev) => {
        // 去重：已存在则跳过
        if (prev.some((q) => q.id === qid)) return prev
        const question: Question = {
          id: qid,
          sessionId: sessionHash || '',
          content: data.title || data.Title || '',
          description: data.description || data.Description,
          type: data.type || data.Type || 'text',
          options: data.options || data.Options,
          allowOther: data.allow_other,
          groupLabel: data.group_label || '',
          batch: data.batch,
          config: data.config,
          status,
          answered: status === 'answered',
          answer: data.answer || data.Answer,
          supplements: [],
          supplement: data.supplement || data.Supplement || false,
          media: data.media || data.Media,
          createdAt: data.created_at || new Date().toISOString(),
          answeredAt: data.answered_at || data.AnsweredAt,
        }
        return [...prev, question]
      })
    },
    [sessionHash, updateQuestions],
  )

  // ── Supplement Push Handler ──

  const handleSupplementPush = useCallback(
    (data: any) => {
      const supplement: SupplementItem = {
        id: data.id || data.ID,
        target_type: data.target_type || data.TargetType || 'question',
        target_id: data.target_id || data.TargetID || '',
        content_type: data.content_type || data.ContentType || 'markdown',
        content: data.content || data.Content || '',
        created_at: data.created_at || new Date().toISOString(),
        updated_at: data.updated_at || new Date().toISOString(),
      }
      // 将 supplement 附加到对应 question 的 supplements 数组（数据层，历史卡片展示用）
      // - 问题级（target_type=question）：直接按 target_id 匹配 question.id
      // - 选项级（target_type=option）：按 target_id 匹配 question.options[].id
      updateQuestions((prev) =>
        prev.map((q) => {
          const belongsToThis =
            (supplement.target_type === 'question' &&
              q.id === supplement.target_id) ||
            (supplement.target_type === 'option' &&
              (q.options ?? []).some((o) => o.id === supplement.target_id))
          if (!belongsToThis) return q
          // 覆写：按 target_type + target_id 替换同目标的旧 supplement（后端 CreateOrUpdate 会生成新 ID）
          const filtered = (q.supplements ?? []).filter(
            (s) =>
              !(
                s.target_type === supplement.target_type &&
                s.target_id === supplement.target_id
              ),
          )
          return {
            ...q,
            supplements: [...filtered, supplement],
          }
        }),
      )
      // 仅当存在 pending 问题（用户正在回答）时才打开详情面板。
      // 重连时后端推送的历史 supplement（目标问题已 answered/skipped）只作数据挂载，
      // 不激活面板，避免无 pending 问题时详情列空白占位、第一列无法居中。
      const hasPending = questionsRef.current.some(
        (q) => q.status === 'pending',
      )
      if (hasPending) {
        setActiveSupplement(supplement)
      }
      stopSupplementLoading()
    },
    [updateQuestions, stopSupplementLoading],
  )

  // ── Answer Sync Handler ──

  const handleAnswerSync = useCallback(
    (data: any) => {
      updateQuestions((prev) =>
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
    },
    [updateQuestions],
  )

  // ── Answer Unhandled Handler ──

  const handleAnswerUnhandled = useCallback(
    (data: any) => {
      const promptText = `请处理以下回答：\n会话ID: ${sessionHash}\n问题ID: ${data.question_id}\n回答内容: ${JSON.stringify(data.answer, null, 2)}\n\n提示：使用 qa_reget_answer 工具获取最新回答。`

      toast.info('等待 AI Agent 处理问题', {
        description:
          '您的回答已保存。AI Agent 可能正在处理其他任务或已停止获取回答，点击下方按钮可复制提示词手动提醒 Agent 继续处理。',
        duration: 12000,
        action: {
          label: '复制提示词',
          onClick: () => navigator.clipboard.writeText(promptText),
        },
      })
    },
    [sessionHash],
  )

  // ── Question Cancel Handler ──
  // Agent 通过 MCP 取消问题时后端推送。cancel_all=true 时批量标记所有 pending 为 cancelled；
  // 否则按 question_id 定位单个问题标记。已 answered/skipped 的问题不受影响（后端仅取消 pending）。

  const handleQuestionCancel = useCallback(
    (data: any) => {
      const cancelAll = data?.cancel_all === true
      const targetId = data?.question_id || ''
      updateQuestions((prev) => {
        if (cancelAll || !targetId) {
          // 全部取消：仅 pending → cancelled
          return prev.map((q) =>
            q.status === 'pending' ? { ...q, status: 'cancelled' as const } : q,
          )
        }
        // 单个取消：按 question_id 定位（仅 pending 生效）
        return prev.map((q) =>
          q.id === targetId && q.status === 'pending'
            ? { ...q, status: 'cancelled' as const }
            : q,
        )
      })
      // 若当前激活问题被取消，清空详情面板
      setActiveSupplement(null)
      // 提示用户
      toast.info(cancelAll ? '所有待回答问题已被取消' : '问题已被取消', {
        description: 'AI Agent 已取消该问题',
        duration: 4000,
      })
    },
    [updateQuestions],
  )

  // ── WebSocket Connection ──
  // 连接建立时后端自动推送 pending 问题（question_push）及其 supplement（supplement_push），
  // 前端通过 handleQuestionPush / handleSupplementPush 自动恢复，无需 REST 历史加载。

  const ws = useQaWebSocket(sessionHash, {
    onQuestionPush: handleQuestionPush,
    onHistoryQuestion: handleHistoryQuestion,
    onSupplementPush: handleSupplementPush,
    onAnswerSync: handleAnswerSync,
    onAnswerUnhandled: handleAnswerUnhandled,
    onQuestionCancel: handleQuestionCancel,
    onSessionEnd: () => {
      setActiveSupplement(null)
      onReject?.()
    },
    onReject,
  })

  // ── Actions ──

  const submitAnswer = useCallback(
    (questionId: string, answer: any) => {
      ws.sendMessage('answer_submit', { question_id: questionId, answer })
      // Optimistic update
      updateQuestions((prev) =>
        prev.map((q) =>
          q.id === questionId
            ? { ...q, status: 'answered' as const, answered: true, answer }
            : q,
        ),
      )
      // Clear active supplement panel after submission
      setActiveSupplement(null)
    },
    [ws, updateQuestions],
  )

  const skipQuestion = useCallback(
    (questionId: string) => {
      ws.sendMessage('skip', { question_id: questionId })
      updateQuestions((prev) =>
        prev.map((q) =>
          q.id === questionId ? { ...q, status: 'skipped' as const } : q,
        ),
      )
      // 跳过后清除详情面板，避免残留上一个问题的补充内容
      setActiveSupplement(null)
    },
    [ws, updateQuestions],
  )

  const requestSupplement = useCallback(
    (payload: {
      questionId: string
      note?: string
      withOptions?: boolean
      optionIds?: string[]
    }) => {
      startSupplementLoading()
      ws.sendMessage('request_supplement', {
        question_id: payload.questionId,
        note: payload.note ?? '',
        with_options: payload.withOptions ?? false,
        option_ids: payload.withOptions ? (payload.optionIds ?? []) : [],
      })
    },
    [ws, startSupplementLoading],
  )

  // 在右侧 DetailPanel 展示某选项的补充内容（从 question.supplements 查找并设为 active）
  const viewOptionDetail = useCallback(
    (optId: string) => {
      const found = questions
        .flatMap((q) => q.supplements ?? [])
        .find((s) => s.target_type === 'option' && s.target_id === optId)
      if (found) {
        setActiveSupplement(found)
      }
    },
    [questions],
  )

  // 从选项详情返回到问题级补充内容（查找当前 pending 问题的问题级 supplement）
  const backToQuestionDetail = useCallback(() => {
    const pending = questions.find((q) => q.status === 'pending')
    if (!pending) return
    const questionSupp = (pending.supplements ?? []).find(
      (s) => s.target_type === 'question',
    )
    if (questionSupp) {
      setActiveSupplement(questionSupp)
    } else {
      setActiveSupplement(null)
    }
  }, [questions])

  return {
    questions,
    activeSupplement,
    isSupplementLoading,
    dismissSupplementLoading,
    session,
    setSession,
    connectionStatus: ws.status,
    submitAnswer,
    skipQuestion,
    requestSupplement,
    viewOptionDetail,
    backToQuestionDetail,
    connect: ws.connect,
    disconnect: ws.disconnect,
  }
}
