import { ExternalLink } from 'lucide-react'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@lumina/components/ui/sheet'
import { Button } from '@lumina/components/ui/button'
import { useSessionDetail } from '#/hooks/useQaAdmin'
import { SessionDetail } from './session-detail'
import { QuestionCard } from './question-card'
import type { QuestionSummary } from '#/lib/models/response/qa-admin'

interface SessionDetailDrawerProps {
  sessionId: string | null
  onClose: () => void
}

export function SessionDetailDrawer({
  sessionId,
  onClose,
}: SessionDetailDrawerProps) {
  // 懒加载：仅当抽屉打开（sessionId 非空）时才发起详情请求
  const { data, isLoading } = useSessionDetail(sessionId ?? '')
  const session = data?.data

  return (
    <Sheet
      open={!!sessionId}
      onOpenChange={(open) => {
        if (!open) onClose()
      }}
    >
      <SheetContent
        side="right"
        className="w-full rounded-none border-l border-line bg-foam p-0 sm:max-w-xl flex flex-col"
      >
        <SheetHeader className="border-b border-line bg-surface-strong px-6 py-5 text-left">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <span className="size-2 rotate-45 bg-lagoon" />
              <SheetTitle className="text-base font-semibold text-sea-ink">
                问答会话全景
              </SheetTitle>
            </div>
            <span className="font-mono text-[10px] text-lagoon-deep bg-sand px-2 py-0.5 border border-line">
              SESSION
            </span>
          </div>
          <SheetDescription className="text-xs text-sea-ink-soft">
            查看该会话的关联项目、在线设备、问题流与回答时间线
          </SheetDescription>
        </SheetHeader>

        <div className="flex-1 overflow-y-auto px-6 py-5 space-y-5">
          {isLoading ? (
            <div className="py-16 text-center text-xs text-sea-ink-soft">
              加载会话详细数据中…
            </div>
          ) : session ? (
            <div className="space-y-6">
              <SessionDetail session={session} />
              <div>
                <div className="flex items-center justify-between mb-3">
                  <h3 className="text-xs font-semibold uppercase tracking-wider text-lagoon-deep">
                    交互问题流 ({session.questions.length})
                  </h3>
                  <span className="text-[11px] text-sea-ink-soft">
                    按推送顺序排序
                  </span>
                </div>
                <div className="space-y-3">
                  {session.questions.map((q: QuestionSummary) => (
                    <QuestionCard key={q.id} question={q} />
                  ))}
                  {session.questions.length === 0 && (
                    <div className="border border-dashed border-line bg-sand/30 py-8 text-center text-xs text-sea-ink-soft">
                      该会话目前尚未产生任何问题
                    </div>
                  )}
                </div>
              </div>
            </div>
          ) : (
            <div className="py-16 text-center text-xs text-sea-ink-soft">
              会话不存在或已被归档清理
            </div>
          )}
        </div>

        {/* 底部快捷操作 */}
        <div className="border-t border-line bg-surface-strong px-6 py-4 flex items-center justify-between gap-3">
          <Button
            type="button"
            variant="outline"
            className="rounded-none border-line text-sea-ink hover:bg-chip-bg flex-1"
            onClick={onClose}
          >
            关闭详情
          </Button>
          <Button
            type="button"
            className="rounded-none bg-lagoon text-foam hover:bg-lagoon-deep flex-1"
            onClick={() => window.open('/interact', '_blank')}
          >
            <span>进入问答交互大厅</span>
            <ExternalLink className="ml-1.5 size-3.5" />
          </Button>
        </div>
      </SheetContent>
    </Sheet>
  )
}
