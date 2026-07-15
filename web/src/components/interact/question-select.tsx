import { Pencil } from 'lucide-react'
import { useState } from 'react'

import { Input } from '@lumina/components/ui/input'
import { Label } from '@lumina/components/ui/label'
import { RadioGroup, RadioGroupItem } from '@lumina/components/ui/radio-group'
import { SupplementLoadingBanner } from './supplement-loading-banner'

import { OptionDetailLabel } from './option-detail-label'
import { QuestionShell } from './question-shell'
import type { QuestionComponentProps } from './question-shell'

export type {
  QuestionComponentProps,
  SupplementRequestArgs,
} from './question-shell'

export function QuestionSelect({
  question,
  onSubmit,
  onSkip,
  onRequestSupplement,
  isSupplementLoading = false,
  onDismissSupplementLoading,
  onViewOptionDetail,
  activeOptionId,
}: QuestionComponentProps) {
  const [selected, setSelected] = useState<string>('')
  const [otherText, setOtherText] = useState('')
  const [showOtherInput, setShowOtherInput] = useState(false)

  const isOther = selected === '__other__'
  const options = question.options ?? []
  const hasSupplements = (question.supplements?.length ?? 0) > 0

  const hasOptionSupplement = (optId: string): boolean => {
    return (
      question.supplements?.some(
        (s) => s.target_type === 'option' && s.target_id === optId,
      ) ?? false
    )
  }

  const handleSubmit = () => {
    if (!selected) return
    if (isOther) {
      onSubmit({ selected: '__other__', other: otherText })
    } else {
      onSubmit({ selected })
    }
  }

  const handleRadioChange = (value: string) => {
    setSelected(value)
    setShowOtherInput(value === '__other__')
    if (value !== '__other__' && hasOptionSupplement(value)) {
      onViewOptionDetail?.(value)
    }
  }

  return (
    <QuestionShell
      question={question}
      isSupplementLoading={isSupplementLoading}
      onSkip={onSkip}
      onRequestSupplement={onRequestSupplement}
      supplement={{
        kind: 'options',
        optionCount: options.length,
        optionIds: options.map((o) => o.id),
      }}
      supplementButtonLabel={hasSupplements ? '重新获取详情' : '请求补充信息'}
      submitDisabled={!selected || (isOther && !otherText.trim())}
      onSubmit={handleSubmit}
    >
      {isSupplementLoading && (
        <SupplementLoadingBanner
          onDismiss={() => onDismissSupplementLoading?.()}
        />
      )}
      <RadioGroup
        value={selected}
        onValueChange={handleRadioChange}
        className="space-y-2"
      >
        {options.map((opt) => (
          <Label
            key={opt.id}
            htmlFor={`select-${question.id}-${opt.id}`}
            className="flex cursor-pointer items-start gap-3 rounded-lg border border-line bg-foam px-3 py-2.5 transition-colors duration-150 has-[data-state=checked]:border-lagoon/30 has-[data-state=checked]:bg-lagoon/10 hover:border-lagoon/30"
          >
            <RadioGroupItem
              value={opt.id}
              id={`select-${question.id}-${opt.id}`}
              className="mt-0.5"
            />
            <div className="min-w-0 flex-1">
              <div className="flex items-start gap-2">
                <p className="min-w-0 flex-1 text-sm font-medium leading-snug">
                  {opt.label}
                </p>
                {hasOptionSupplement(opt.id) && (
                  <OptionDetailLabel
                    optId={opt.id}
                    onClick={() => onViewOptionDetail?.(opt.id)}
                    isActive={activeOptionId === opt.id}
                  />
                )}
              </div>
              {opt.description && (
                <p className="mt-0.5 text-xs leading-relaxed text-sea-ink-soft">
                  {opt.description}
                </p>
              )}
            </div>
          </Label>
        ))}

        <Label
          htmlFor={`select-${question.id}-other`}
          className="flex cursor-pointer items-start gap-3 rounded-lg border border-line bg-foam px-3 py-2.5 transition-colors duration-150 has-[data-state=checked]:border-lagoon/30 has-[data-state=checked]:bg-lagoon/10 hover:border-lagoon/30"
        >
          <RadioGroupItem
            value="__other__"
            id={`select-${question.id}-other`}
            className="mt-0.5"
          />
          <Pencil className="size-3.5 shrink-0 text-lagoon-deep" />
          <span className="text-sm font-medium">其他</span>
        </Label>
      </RadioGroup>

      {isOther && showOtherInput && (
        <Input
          placeholder="输入自定义选项..."
          value={otherText}
          onChange={(e) => setOtherText(e.target.value)}
          className="rounded-lg border-line bg-foam"
        />
      )}
    </QuestionShell>
  )
}
