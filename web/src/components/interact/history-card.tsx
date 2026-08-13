import { Ban, Check } from 'lucide-react'

import { Kicker, PanelCard } from './primitives'
import type { Question } from './types'
import { formatAnswer } from '#/lib/format-answer'

/** 将 Markdown 内容提取为纯文本摘要，用于历史列表展示 */
function plainTextSummary(md: string, maxLen = 120): string {
  const plain = md
    .replace(/!?\[([^\]]*)\]\([^)]*\)/g, '$1') // 链接/图片 → 文本
    .replace(/```[\s\S]*?```/g, ' ') // 代码块
    .replace(/`([^`]+)`/g, '$1') // 行内代码
    .replace(/[#*~>_+\-=`]/g, '') // 语法标记
    .replace(/\n+/g, ' ') // 换行 → 空格
    .trim()
  return plain.length > maxLen ? plain.slice(0, maxLen) + '…' : plain
}


interface HistoryCardProps {
  answeredQuestions: Question[]
  groupedHistory: Record<string, Question[]>
}

export function HistoryCard({
  groupedHistory,
  answeredQuestions,
}: HistoryCardProps) {
  return (
    <PanelCard
      flushHeader
      header={
        <div className="px-4 py-2.5">
          <Kicker>历史问答</Kicker>
        </div>
      }
      bodyClassName="p-0"
    >
      <div className="space-y-0 divide-y divide-line/30">
        {Object.entries(groupedHistory).map(([group, questions]) => (
          <div key={group} className="px-3 py-2">
            <div className="mb-1.5 flex items-center gap-2">
              <span className="inline-flex items-center gap-1 bg-lagoon/8 px-2 py-0.5 text-[10px] font-semibold text-lagoon-deep">
                {group}
              </span>
              <span className="text-[10px] text-sea-ink-soft">
                {questions.length} 个问答
              </span>
            </div>

            <div className="space-y-1.5">
              {questions.map((q) => {
                const isCancelled = q.status === 'cancelled'
                return (
                  <div
                    key={q.id}
                    className={`flex items-start gap-2 ${isCancelled ? 'opacity-60' : ''}`}
                  >
                    {isCancelled ? (
                      <Ban
                        className="mt-0.5 size-3.5 shrink-0 text-sea-ink-soft/60"
                        aria-hidden
                      />
                    ) : (
                      <Check
                        className="mt-0.5 size-3.5 shrink-0 text-lagoon"
                        aria-hidden
                      />
                    )}
                    <div className="min-w-0 flex-1">
                      <p
                        className="line-clamp-2 text-xs leading-relaxed text-sea-ink-soft"
                        title={q.content}
                      >
                        {plainTextSummary(q.content)}
                      </p>
                      <p className="mt-0.5 text-xs font-medium text-sea-ink">
                        {isCancelled ? (
                          <span className="text-sea-ink-soft/70">已取消</span>
                        ) : (
                          `→ ${formatAnswer(q.answer, q.options)}`
                        )}
                      </p>
                    </div>
                  </div>
                )
              })}
            </div>
          </div>
        ))}

        {answeredQuestions.length === 0 && (
          <div className="px-4 py-6 text-center">
            <p className="text-xs text-sea-ink-soft/50">暂无历史记录</p>
          </div>
        )}
      </div>
    </PanelCard>
  )
}
