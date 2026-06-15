import { CircleDot } from 'lucide-react'

import type { Question } from './types'
import { Kicker, Markdown, PanelCard, proseQuestion } from './primitives'

import { QuestionBoolean } from './question-boolean'
import { QuestionCode } from './question-code'
import { QuestionDiff } from './question-diff'
import { QuestionFile } from './question-file'
import { QuestionImage } from './question-image'
import type {
  QuestionComponentProps,
  SupplementRequestArgs,
} from './question-select'
import { QuestionMultiSelect } from './question-multi-select'
import { QuestionOptions } from './question-options'
import { QuestionPlan } from './question-plan'
import { QuestionRate } from './question-rate'
import { QuestionRank } from './question-rank'
import { QuestionReview } from './question-review'
import { QuestionSelect } from './question-select'
import { QuestionSlider } from './question-slider'
import { QuestionText } from './question-text'

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
      {renderByType(question.type, props)}
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
