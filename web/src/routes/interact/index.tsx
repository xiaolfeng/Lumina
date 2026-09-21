import { createFileRoute, useNavigate, useSearch } from '@tanstack/react-router'
import { useState, useEffect } from 'react'

import type { Session } from '#/components/interact/types'
import { DetailPanel } from '#/components/interact/detail-panel'
import { groupHistoryDesc } from '#/components/interact/group-history'
import { HistoryCard } from '#/components/interact/history-card'
import { LobbyView } from '#/components/interact/lobby-view'
import { QuestionCard } from '#/components/interact/question-card'
import { MobileSessionDrawer } from '#/components/interact/session-drawer'
import { SessionSidebarCompact } from '#/components/interact/session-sidebar-compact'
import { EmptyState, LoadingState } from '#/components/interact/primitives'
import { ScrollArea } from '@lumina/components/ui/scroll-area'
import {
  Splitter,
  SplitterPanel,
  SplitterHandle,
} from '@lumina/components/ui/splitter'
import { computeSessionProgress } from '#/components/interact/session-progress'
import { useQaSession } from '#/hooks/useQaSession'
import { useSidebarOpen } from '#/hooks/useSidebarOpen'
import { getSessionByHash, getSessionList } from '#/lib/apis/qa-admin'

interface InteractSearch {
  session?: string
}

export const Route = createFileRoute('/interact/')({
  validateSearch: (search: Record<string, unknown>): InteractSearch => {
    return {
      session: typeof search.session === 'string' ? search.session : undefined,
    }
  },
  component: InteractPage,
})

function InteractPage() {
  const [selectedSessionId, setSelectedSessionId] = useState<string>('')
  const [sessionHash, setSessionHash] = useState<string>('')
  const [sessions, setSessions] = useState<Session[]>([])
  const [isLoading, setIsLoading] = useState(false)

  const { setProgress } = useSidebarOpen()

  const search = useSearch({ from: '/interact/' })
  const navigate = useNavigate()
  const hashParam = search.session

  const isConnected = !!sessionHash

  // URL 有 hash → 加载会话详情
  useEffect(() => {
    if (!hashParam) return

    setSessionHash(hashParam)
    setIsLoading(true)
    ;(async () => {
      try {
        const res = await getSessionByHash(hashParam)
        const sessionData = res.data?.items?.[0]
        if (sessionData) {
          setSelectedSessionId(sessionData.id)
        }
      } catch (e) {
        console.error('Failed to load session info:', e)
      } finally {
        setIsLoading(false)
      }
    })()
  }, [hashParam])

  useEffect(() => {
    sessionStorage.setItem('qa_page_active', '1')
    return () => {
      sessionStorage.removeItem('qa_page_active')
    }
  }, [])

  useEffect(() => {
    ;(async () => {
      try {
        const res = await getSessionList({
          status: 'active',
          type: 'permanent',
          size: 50,
        })
        const apiSessions = res.data?.items || []

        const mapped: Session[] = apiSessions.map((item) => ({
          id: item.id,
          hash: item.hash,
          title: item.title,
          agent: item.agent,
          type: item.type,
          status: item.status,
          onlineDevices: item.online_devices,
          owner: '',
          questions: [],
          updatedAt: item.updated_at || item.created_at,
          expiresAt: item.expires_at || '',
        }))

        setSessions(mapped)
      } catch (e) {
        console.error('Failed to load sessions:', e)
      }
    })()
  }, [])

  const {
    questions,
    activeSupplement,
    session: currentSession,
    submitAnswer,
    skipQuestion,
    requestSupplement,
    viewOptionDetail,
    backToQuestionDetail,
    isSupplementLoading,
    dismissSupplementLoading,
  } = useQaSession({
    sessionHash: sessionHash || null,
    onReject: () => {
      setSessionHash('')
      setSelectedSessionId('')
      navigate({ to: '/interact/thank', replace: true })
    },
  })

  useEffect(() => {
    if (currentSession) {
      setSessions((prev) => {
        const idx = prev.findIndex((s) => s.id === currentSession.id)
        if (idx >= 0) {
          const next = [...prev]
          next[idx] = currentSession
          return next
        }
        return [...prev, currentSession]
      })
    }
  }, [currentSession])

  const activeQuestion = questions.find((q) => q.status === 'pending')
  // 已回答 / 已取消进入历史区，按事件时间 DESC（最新问答在最上方）。
  const groupedHistory = groupHistoryDesc(questions)

  // 详情面板内容：底层=当前待答问题级 supplement，上层=选项级 supplement
  const questionSupp = activeQuestion?.supplements?.find(
    (s) => s.target_type === 'question',
  )
  const isViewingOption = activeSupplement?.target_type === 'option'
  const activeDetailSupp = isViewingOption
    ? activeSupplement
    : activeSupplement?.target_type === 'question'
      ? activeSupplement
      : questionSupp

  // 只要存在活跃选项 supplement 或当前待答题目自带问题级 supplement，即展开右侧详情
  const hasDetailContent = activeDetailSupp != null

  // 当前激活的选项 ID（用于 OptionDetailLabel 的 isActive 状态）
  const activeOptionId = isViewingOption
    ? activeSupplement?.target_id
    : undefined

  const questionContent = isViewingOption
    ? (questionSupp?.content ?? '')
    : (activeDetailSupp?.content ?? '')
  const questionContentType = isViewingOption
    ? (questionSupp?.content_type ?? 'markdown')
    : (activeDetailSupp?.content_type ?? 'markdown')
  const optionContent = isViewingOption ? (activeSupplement?.content ?? '') : ''
  const optionContentType = isViewingOption
    ? (activeSupplement?.content_type ?? 'markdown')
    : 'markdown'

  const current = sessions.find((s) => s.id === selectedSessionId)

  // 进度以 WebSocket 推送的 questions 为准；列表接口的 session.questions 恒为空。
  useEffect(() => {
    if (!sessionHash) {
      setProgress(null)
      return
    }
    setProgress(computeSessionProgress(questions))
  }, [sessionHash, questions, setProgress])

  if (!isConnected) {
    return <LobbyView sessions={sessions} />
  }

  const questionPanelNode = (
    <aside className="relative flex h-full min-h-0 w-full flex-col overflow-hidden p-4">
      <ScrollArea className="min-h-0 flex-1 pt-2" hideScrollbar>
        <div className="space-y-4">
          {isLoading ? (
            <LoadingState text="正在准备会话…" />
          ) : activeQuestion ? (
            <QuestionCard
              question={activeQuestion}
              agent={current?.agent}
              onSubmit={(answer) => submitAnswer(activeQuestion.id, answer)}
              onSkip={() => skipQuestion(activeQuestion.id)}
              onRequestSupplement={(payload) => requestSupplement(payload)}
              onViewOptionDetail={(optId) => viewOptionDetail(optId)}
              isSupplementLoading={isSupplementLoading}
              onDismissSupplementLoading={dismissSupplementLoading}
              activeOptionId={activeOptionId}
            />
          ) : (
            <EmptyState
              text={
                selectedSessionId ? '等待问题推送…' : '请在右侧选择一个会话'
              }
            />
          )}

          <HistoryCard groupedHistory={groupedHistory} />
        </div>
      </ScrollArea>
    </aside>
  )

  return (
    <div className="flex min-h-0 flex-1 overflow-hidden">
      {/* 第一/二列：问答区 + 详情面板 */}
      <div className="flex min-w-0 flex-1 overflow-hidden">
        {hasDetailContent ? (
          <Splitter
            direction="horizontal"
            className="h-full w-full min-w-0 flex-1"
          >
            <SplitterPanel
              defaultSize={45}
              minSize={30}
              maxSize={50}
              className="relative flex h-full min-h-0 flex-col overflow-hidden"
            >
              {questionPanelNode}
            </SplitterPanel>

            <SplitterHandle />

            <SplitterPanel
              minSize={30}
              className="relative flex h-full min-h-0 flex-col overflow-hidden"
            >
              <DetailPanel
                visible={hasDetailContent}
                activeOption={
                  activeDetailSupp
                    ? { label: activeDetailSupp.content_type }
                    : null
                }
                isMotionDemo={false}
                questionContent={questionContent}
                questionContentType={questionContentType}
                optionContent={optionContent}
                optionContentType={optionContentType}
                optionId={activeDetailSupp?.id ?? ''}
                onBack={() => backToQuestionDetail()}
              />
            </SplitterPanel>
          </Splitter>
        ) : (
          <div className="relative mx-auto flex min-h-0 w-full max-w-3xl flex-1 flex-col overflow-hidden">
            {questionPanelNode}
          </div>
        )}
      </div>

      {/* 第三列：会话列表（≥xl · 真 flex 子列，宽度过渡） */}
      <SessionSidebarCompact
        sessions={sessions}
        selectedId={selectedSessionId}
        onSelect={setSelectedSessionId}
      />

      <div className="xl:hidden">
        <MobileSessionDrawer
          sessions={sessions}
          selectedId={selectedSessionId}
          onSelect={setSelectedSessionId}
        />
      </div>
    </div>
  )
}
