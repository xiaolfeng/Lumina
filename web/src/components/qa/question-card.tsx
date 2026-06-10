import { useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '#/components/ui/card'
import { Badge } from '#/components/ui/badge'
import { ChevronDown, ChevronUp } from 'lucide-react'
import type { SupplementItem } from '#/lib/models/response/qa-admin'
import { useQuestionDetail } from '#/hooks/useQaAdmin'

interface QuestionCardProps {
  sessionId: string
  question: { id: string; type: string; title: string; status: string; created_at: string; answered_at: string }
}

export function QuestionCard({ sessionId, question }: QuestionCardProps) {
  const [expanded, setExpanded] = useState(false)
  const { data, isLoading } = useQuestionDetail(sessionId, question.id)

  const statusVariant = question.status === 'answered' ? 'default' : question.status === 'skipped' ? 'outline' : 'secondary'
  const statusLabel = question.status === 'answered' ? '已回答' : question.status === 'skipped' ? '已跳过' : '待回答'

  return (
    <Card className="cursor-pointer" onClick={() => setExpanded(!expanded)}>
      <CardHeader className="flex flex-row items-center justify-between py-3">
        <div className="flex items-center gap-3">
          <Badge variant="outline">{question.type}</Badge>
          <CardTitle className="text-base">{question.title}</CardTitle>
        </div>
        <div className="flex items-center gap-2">
          <Badge variant={statusVariant}>{statusLabel}</Badge>
          {expanded ? <ChevronUp className="h-4 w-4" /> : <ChevronDown className="h-4 w-4" />}
        </div>
      </CardHeader>
      {expanded && (
        <CardContent>
          {isLoading ? (
            <p className="text-sm text-muted-foreground">加载中...</p>
          ) : data?.data ? (
            <div className="space-y-4">
              {question.status === 'answered' && data.data.answer && (
                <div>
                  <h4 className="text-sm font-semibold mb-1">回答</h4>
                  <pre className="text-sm bg-muted p-3 rounded-md overflow-auto max-h-64">
                    {JSON.stringify(data.data.answer, null, 2)}
                  </pre>
                </div>
              )}
              {data.data.description && (
                <div>
                  <h4 className="text-sm font-semibold mb-1">描述</h4>
                  <p className="text-sm text-muted-foreground">{data.data.description}</p>
                </div>
              )}
              {data.data.supplements?.length > 0 && (
                <div>
                  <h4 className="text-sm font-semibold mb-1">补充内容</h4>
                  {data.data.supplements.map((s: SupplementItem) => (
                    <div key={s.id} className="text-sm bg-muted p-3 rounded-md mt-1">
                      <Badge variant="outline" className="mb-1">{s.content_type}</Badge>
                      <div className="prose prose-sm max-w-none" dangerouslySetInnerHTML={{ __html: s.content }} />
                    </div>
                  ))}
                </div>
              )}
              <div className="text-xs text-muted-foreground">
                创建: {new Date(question.created_at).toLocaleString()}
                {question.answered_at && ` | 回答: ${new Date(question.answered_at).toLocaleString()}`}
              </div>
            </div>
          ) : (
            <p className="text-sm text-muted-foreground">暂无详情</p>
          )}
        </CardContent>
      )}
    </Card>
  )
}
