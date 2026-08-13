import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from '@lumina/components/ui/sheet'
import { useSessionDetail } from '#/hooks/useQaAdmin'
import { SessionDetail } from './session-detail'
import { QuestionCard } from './question-card'
import type { QuestionSummary } from '#/lib/models/response/qa-admin'

interface SessionDetailDrawerProps {
  sessionId: string | null
  onClose: () => void
}

export function SessionDetailDrawer({ sessionId, onClose }: SessionDetailDrawerProps) {
  // 懒加载：仅当抽屉打开（sessionId 非空）时才发起详情请求
  const { data, isLoading } = useSessionDetail(sessionId ?? '')
  const session = data?.data

  return (
    <Sheet open={!!sessionId} onOpenChange={(open) => { if (!open) onClose() }}>
      <SheetContent side="right" className="w-[min(560px,92vw)] gap-0 border-line sm:max-w-[560px]">
        <SheetHeader className="border-b border-line px-6 py-4">
          <SheetTitle>会话详情</SheetTitle>
          <SheetDescription>查看该会话的问题、选项与补充内容</SheetDescription>
        </SheetHeader>

        <div className="flex-1 overflow-y-auto px-6 py-5">
          {isLoading ? (
            <div className="py-12 text-center text-muted-foreground">加载中...</div>
          ) : session ? (
            <div className="space-y-6">
              <SessionDetail session={session} />
              <div>
                <h3 className="mb-3 text-lg font-semibold">
                  问题列表 ({session.questions?.length ?? 0})
                </h3>
                <div className="space-y-3">
                  {session.questions?.map((q: QuestionSummary) => (
                    <QuestionCard key={q.id} question={q} />
                  ))}
                  {(!session.questions || session.questions.length === 0) && (
                    <p className="py-8 text-center text-muted-foreground">暂无问题</p>
                  )}
                </div>
              </div>
            </div>
          ) : (
            <div className="py-12 text-center text-muted-foreground">会话不存在</div>
          )}
        </div>
      </SheetContent>
    </Sheet>
  )
}
