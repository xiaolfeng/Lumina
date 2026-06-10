import { createFileRoute } from '@tanstack/react-router'
import { useSessionDetail } from '#/hooks/useQaAdmin'
import { SessionDetail } from '#/components/qa/session-detail'
import { QuestionCard } from '#/components/qa/question-card'
import type { QuestionSummary } from '#/lib/models/response/qa-admin'

export const Route = createFileRoute('/console/qa/$sessionId')({
  component: SessionDetailPage,
})

function SessionDetailPage() {
  const { sessionId } = Route.useParams()
  const { data, isLoading } = useSessionDetail(sessionId)

  if (isLoading) {
    return <div className="text-center py-12 text-muted-foreground">加载中...</div>
  }

  const session = data?.data
  if (!session) {
    return <div className="text-center py-12 text-muted-foreground">会话不存在</div>
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">会话详情</h1>
        <p className="text-muted-foreground">{session.title}</p>
      </div>

      <SessionDetail session={session} />

      <div>
        <h2 className="text-lg font-semibold mb-4">问题列表 ({session.questions?.length ?? 0})</h2>
        <div className="space-y-3">
          {session.questions?.map((q: QuestionSummary) => (
            <QuestionCard key={q.id} sessionId={sessionId} question={q} />
          ))}
          {(!session.questions || session.questions.length === 0) && (
            <p className="text-center py-8 text-muted-foreground">暂无问题</p>
          )}
        </div>
      </div>
    </div>
  )
}
