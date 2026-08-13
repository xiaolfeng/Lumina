import { Card, CardContent, CardHeader, CardTitle } from '@lumina/components/ui/card'
import { Badge } from '@lumina/components/ui/badge'
import type { QuestionSummary, SupplementItem } from '#/lib/models/response/qa-admin'
import { Markdown, ShadowHtml } from '#/components/interact/primitives'
import { formatAnswer, type AnswerOption } from '#/lib/format-answer'

interface QuestionCardProps {
  question: QuestionSummary
}

const statusVariantMap: Record<QuestionSummary['status'], 'default' | 'outline' | 'destructive' | 'secondary'> = {
  answered: 'default',
  skipped: 'outline',
  cancelled: 'destructive',
  pending: 'secondary',
}

const statusLabelMap: Record<QuestionSummary['status'], string> = {
  answered: '已回答',
  skipped: '已跳过',
  cancelled: '已取消',
  pending: '待回答',
}

export function QuestionCard({ question }: QuestionCardProps) {
  const options = (question.options ?? []) as AnswerOption[]
  const questionSupplements = (question.supplements ?? []).filter(
    (s) => s.target_type === 'question',
  )

  // 选项级补充按 target_id（选项 ID）分组，渲染选项时对齐展示
  const optionSupplements = new Map<string, SupplementItem[]>()
  for (const s of question.supplements ?? []) {
    if (s.target_type !== 'option') continue
    const list = optionSupplements.get(s.target_id) ?? []
    list.push(s)
    optionSupplements.set(s.target_id, list)
  }

  const variant = statusVariantMap[question.status] ?? 'secondary'
  const statusLabel = statusLabelMap[question.status] ?? question.status

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between py-3">
        <div className="flex items-center gap-3">
          <Badge variant="outline">{question.type}</Badge>
          <CardTitle className="text-base">{question.title}</CardTitle>
        </div>
        <Badge variant={variant}>{statusLabel}</Badge>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* 回答（按题型格式化） */}
        {question.status === 'answered' && question.answer != null && (
          <div>
            <h4 className="text-sm font-semibold mb-1">回答</h4>
            <p className="text-sm text-sea-ink">{formatAnswer(question.answer, options)}</p>
          </div>
        )}

        {/* 选项列表（含选项级补充） */}
        {options.length > 0 && (
          <div>
            <h4 className="text-sm font-semibold mb-2">选项</h4>
            <ul className="space-y-2">
              {options.map((opt) => {
                const supplements = optionSupplements.get(opt.id) ?? []
                return (
                  <li key={opt.id} className="text-sm">
                    <span className="font-medium text-sea-ink">{opt.label}</span>
                    {opt.description && (
                      <span className="text-sea-ink-soft"> — {opt.description}</span>
                    )}
                    {supplements.map((s) => (
                      <div key={s.id} className="mt-1.5 rounded-md bg-muted p-2.5">
                        {s.content_type === 'html' ? (
                          <ShadowHtml content={s.content} />
                        ) : (
                          <Markdown>{s.content}</Markdown>
                        )}
                      </div>
                    ))}
                  </li>
                )
              })}
            </ul>
          </div>
        )}

        {/* 问题级补充 */}
        {questionSupplements.length > 0 && (
          <div>
            <h4 className="text-sm font-semibold mb-1">补充内容</h4>
            {questionSupplements.map((s) => (
              <div key={s.id} className="mt-1.5 rounded-md bg-muted p-3">
                <Badge variant="outline" className="mb-1.5">
                  {s.content_type}
                </Badge>
                {s.content_type === 'html' ? (
                  <ShadowHtml content={s.content} />
                ) : (
                  <Markdown>{s.content}</Markdown>
                )}
              </div>
            ))}
          </div>
        )}

        <div className="text-xs text-muted-foreground">
          创建: {new Date(question.created_at).toLocaleString()}
          {question.answered_at && ` | 回答: ${new Date(question.answered_at).toLocaleString()}`}
        </div>
      </CardContent>
    </Card>
  )
}
