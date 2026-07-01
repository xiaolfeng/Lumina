import { lazy, Suspense } from 'react'
import { CircleDot } from 'lucide-react'

import type { Question } from './types'
import { Kicker, Markdown, PanelCard, proseQuestion } from './primitives'

import type {
  QuestionComponentProps,
  SupplementRequestArgs,
} from './question-select'

/* ── 题型组件按需加载 ──────────────────────────────────────
 * 14 种题型组件通过 React.lazy 拆分为独立 chunk，
 * 避免将 CodeMirror（15 个语言包）、react-diff-viewer 等
 * 重型库全部打入 interact 主 chunk。
 */
const QuestionBoolean = lazy(() =>
  import('./question-boolean').then((m) => ({ default: m.QuestionBoolean })),
)
const QuestionCode = lazy(() =>
  import('./question-code').then((m) => ({ default: m.QuestionCode })),
)
const QuestionDiff = lazy(() =>
  import('./question-diff').then((m) => ({ default: m.QuestionDiff })),
)
const QuestionFile = lazy(() =>
  import('./question-file').then((m) => ({ default: m.QuestionFile })),
)
const QuestionImage = lazy(() =>
  import('./question-image').then((m) => ({ default: m.QuestionImage })),
)
const QuestionMultiSelect = lazy(() =>
  import('./question-multi-select').then((m) => ({
    default: m.QuestionMultiSelect,
  })),
)
const QuestionOptions = lazy(() =>
  import('./question-options').then((m) => ({ default: m.QuestionOptions })),
)
const QuestionPlan = lazy(() =>
  import('./question-plan').then((m) => ({ default: m.QuestionPlan })),
)
const QuestionRate = lazy(() =>
  import('./question-rate').then((m) => ({ default: m.QuestionRate })),
)
const QuestionRank = lazy(() =>
  import('./question-rank').then((m) => ({ default: m.QuestionRank })),
)
const QuestionReview = lazy(() =>
  import('./question-review').then((m) => ({ default: m.QuestionReview })),
)
const QuestionSelect = lazy(() =>
  import('./question-select').then((m) => ({ default: m.QuestionSelect })),
)
const QuestionSlider = lazy(() =>
  import('./question-slider').then((m) => ({ default: m.QuestionSlider })),
)
const QuestionText = lazy(() =>
  import('./question-text').then((m) => ({ default: m.QuestionText })),
)

interface QuestionCardProps {
  question: Question | undefined
  onSubmit: (answer: any) => void
  onSkip: () => void
  onRequestSupplement: (payload: SupplementRequestArgs) => void
  isSupplementLoading?: boolean
  onDismissSupplementLoading?: () => void
  onViewOptionDetail?: (optId: string) => void
  activeOptionId?: string
}

export function QuestionCard({
  question,
  onSubmit,
  onSkip,
  onRequestSupplement,
  isSupplementLoading = false,
  onDismissSupplementLoading,
  onViewOptionDetail,
  activeOptionId,
}: QuestionCardProps) {
  if (!question) {
    return (
      <PanelCard>
        <div className="py-4 text-center">
          <p className="text-sm text-sea-ink-soft">所有问题已回答完毕</p>
        </div>
      </PanelCard>
    )
  }

  const props: QuestionComponentProps = {
    question,
    onSubmit,
    onSkip,
    onRequestSupplement,
    isSupplementLoading,
    onDismissSupplementLoading,
    onViewOptionDetail,
    activeOptionId,
  }

  return (
    <PanelCard
      flushHeader
      header={
        <div className="flex items-center justify-between px-4 py-2.5">
          <Kicker tone="lagoon-deep">当前问题</Kicker>
          <span className="inline-flex items-center gap-1 rounded-full bg-lagoon/10 px-2 py-0.5 text-[10px] font-semibold text-lagoon-deep">
            <CircleDot className="size-2.5" aria-hidden />
            {question.groupLabel}
          </span>
        </div>
      }
    >
      <Suspense
        fallback={
          <div className="flex items-center justify-center py-8">
            <span className="text-xs text-sea-ink-soft">加载题型组件…</span>
          </div>
        }
      >
        {renderByType(question.type, props)}
      </Suspense>
    </PanelCard>
  )
}

function renderByType(type: Question['type'], props: QuestionComponentProps) {
  switch (type) {
    case 'select':
      return <QuestionSelect {...props} />
    case 'multi-select':
      return <QuestionMultiSelect {...props} />
    case 'text':
      return <QuestionText {...props} />
    case 'boolean':
      return <QuestionBoolean {...props} />
    case 'code':
      return <QuestionCode {...props} />
    case 'image':
      return <QuestionImage {...props} />
    case 'file':
      return <QuestionFile {...props} />
    case 'diff':
      return <QuestionDiff {...props} />
    case 'plan':
      return <QuestionPlan {...props} />
    case 'options':
      return <QuestionOptions {...props} />
    case 'review':
      return <QuestionReview {...props} />
    case 'slider':
      return <QuestionSlider {...props} />
    case 'rank':
      return <QuestionRank {...props} />
    case 'rate':
      return <QuestionRate {...props} />
    default:
      return (
        <div className="space-y-3">
          <div className={proseQuestion}>
            <Markdown>{props.question.content}</Markdown>
          </div>
          <div className="rounded-lg border border-dashed border-line bg-foam/50 p-6 text-center">
            <p className="text-xs text-sea-ink-soft">
              暂不支持的问题类型：
              <code className="rounded bg-line px-1.5 py-0.5 font-mono">
                {type}
              </code>
            </p>
            <p className="mt-1 text-[10px] text-sea-ink-soft/60">
              该类型将在后续版本中支持
            </p>
          </div>
        </div>
      )
  }
}
