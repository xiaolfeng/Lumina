import {
  ChevronDown,
  ChevronRight,
  ChevronsDownUp,
  ChevronsUpDown,
} from 'lucide-react'
import { useMemo, useState } from 'react'

import { Button } from '@lumina/components/ui/button'
import { Textarea } from '@lumina/components/ui/textarea'

import { DecisionButtons } from './decision-buttons'
import { Markdown, proseQuestion } from './primitives'
import { QuestionShell } from './question-shell'
import type { QuestionComponentProps } from './question-shell'

export interface ReviewSection {
  id: string
  title: string
  content: string
}

type ReviewDecision = 'approve' | 'revise'

function parseReviewConfig(rawConfig: any): Record<string, any> {
  let cfg = rawConfig
  while (typeof cfg === 'string') {
    const trimmed = cfg.trim()
    if (trimmed.startsWith('{') && trimmed.endsWith('}')) {
      try {
        cfg = JSON.parse(trimmed)
      } catch {
        break
      }
    } else {
      break
    }
  }
  return typeof cfg === 'object' && cfg !== null ? cfg : {}
}

export function QuestionReview({
  question,
  onSubmit,
  onSkip,
  onRequestSupplement,
  isSupplementLoading = false,
}: QuestionComponentProps) {
  const [decision, setDecision] = useState<ReviewDecision | null>(null)
  const [feedback, setFeedback] = useState('')
  const [annotations, setAnnotations] = useState<Record<string, string>>({})

  const config = useMemo(
    () => parseReviewConfig(question.config),
    [question.config],
  )
  const sections = useMemo(() => {
    if (Array.isArray(config.sections)) {
      return config.sections as ReviewSection[]
    }
    return []
  }, [config.sections])

  const rawContext =
    typeof config.context === 'string' ? config.context : undefined
  const context = rawContext ?? question.description ?? ''
  const customContent =
    typeof config.content === 'string'
      ? config.content
      : typeof config.markdown === 'string'
        ? config.markdown
        : ''
  const showCustomContent =
    sections.length === 0 &&
    Boolean(customContent) &&
    customContent.trim() !== question.content.trim()

  const [expandedSections, setExpandedSections] = useState<Set<string>>(
    () => new Set(sections.map((s) => s.id)),
  )

  const isRevising = decision === 'revise'

  const toggleSection = (id: string) => {
    setExpandedSections((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  const expandAll = () => {
    setExpandedSections(new Set(sections.map((s) => s.id)))
  }

  const collapseAll = () => {
    setExpandedSections(new Set())
  }

  const allExpanded =
    sections.length > 0 && expandedSections.size === sections.length

  const annotationsList = useMemo(() => {
    return Object.entries(annotations)
      .filter(([, v]) => v.trim())
      .map(([sectionId, content]) => ({ sectionId, content }))
  }, [annotations])

  const hasFeedback = Boolean(feedback.trim() || annotationsList.length > 0)

  const handleSubmit = () => {
    if (!decision) return

    const summaryParts: string[] = []
    if (feedback.trim()) {
      summaryParts.push(feedback.trim())
    }
    if (annotationsList.length > 0) {
      const sectionFeedback = annotationsList
        .map(({ sectionId, content }) => {
          const sec = sections.find((s) => s.id === sectionId)
          return `[${sec?.title ?? sectionId}] ${content}`
        })
        .join('\n')
      if (!feedback.trim()) {
        summaryParts.push(sectionFeedback)
      }
    }

    onSubmit({
      decision,
      ...(summaryParts.length > 0
        ? { feedback: summaryParts.join('\n\n') }
        : {}),
      ...(annotationsList.length > 0 ? { annotations: annotationsList } : {}),
    })
  }

  return (
    <QuestionShell
      question={question}
      isSupplementLoading={isSupplementLoading}
      onSkip={onSkip}
      onRequestSupplement={onRequestSupplement}
      showDescription={false}
      submitDisabled={!decision || (isRevising && !hasFeedback)}
      onSubmit={handleSubmit}
    >
      {context && (
        <div className="rounded-lg border border-amber-200 bg-amber-50/80 px-3 py-2 dark:border-amber-800/40 dark:bg-amber-900/15">
          <p className="text-[11px] font-semibold text-amber-700 dark:text-amber-400">
            上下文
          </p>
          <div className="prose prose-sm mt-1 max-w-none break-words [overflow-wrap:anywhere] [&_p]:mb-0 [&_p]:text-xs [&_p]:leading-relaxed [&_p]:text-amber-600 dark:[&_p]:text-amber-300 [&_code]:rounded [&_code]:bg-lagoon/8 [&_code]:px-1.5 [&_code]:py-0.5 [&_code]:text-xs [&_code]:text-lagoon-deep [&_code]:break-all [&_pre]:rounded [&_pre]:border [&_pre]:border-line [&_pre]:bg-white [&_pre]:p-2 [&_pre]:text-xs [&_pre]:leading-relaxed [&_pre]:font-mono [&_pre]:text-gray-800 [&_pre]:overflow-x-auto [&_pre]:max-w-full [&_pre_code]:bg-transparent [&_pre_code]:p-0 [&_pre_code]:text-inherit">
            <Markdown>{context}</Markdown>
          </div>
        </div>
      )}

      {sections.length > 0 ? (
        <div className="space-y-3">
          <div className="flex items-center justify-between px-1">
            <span className="text-xs font-medium text-sea-ink-soft">
              共 {sections.length} 个审阅段落
            </span>
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={allExpanded ? collapseAll : expandAll}
              disabled={isSupplementLoading}
              className="h-7 gap-1 px-2 text-xs text-sea-ink-soft hover:text-sea-ink"
            >
              {allExpanded ? (
                <>
                  <ChevronsDownUp className="size-3.5" />
                  全部收起
                </>
              ) : (
                <>
                  <ChevronsUpDown className="size-3.5" />
                  全部展开
                </>
              )}
            </Button>
          </div>

          <div className="space-y-2">
            {sections.map((section, idx) => {
              const isOpen = expandedSections.has(section.id)
              const annotationText = annotations[section.id] as
                string | undefined
              const hasSectionAnnotation = Boolean(
                annotationText && annotationText.trim(),
              )

              return (
                <div
                  key={section.id}
                  className="overflow-hidden rounded-lg border border-line bg-foam transition-colors"
                >
                  <button
                    type="button"
                    onClick={() => toggleSection(section.id)}
                    disabled={isSupplementLoading}
                    className="flex w-full items-center justify-between gap-2 px-3 py-2.5 text-left transition-colors hover:bg-line/30 disabled:cursor-not-allowed disabled:opacity-50"
                  >
                    <div className="flex min-w-0 items-center gap-2">
                      {isOpen ? (
                        <ChevronDown className="size-4 shrink-0 text-sea-ink-soft" />
                      ) : (
                        <ChevronRight className="size-4 shrink-0 text-sea-ink-soft" />
                      )}
                      <span className="truncate text-sm font-medium text-sea-ink">
                        {section.title || `段落 ${idx + 1}`}
                      </span>
                    </div>

                    {hasSectionAnnotation && (
                      <span className="shrink-0 rounded bg-amber-500/10 px-1.5 py-0.5 text-[10px] font-medium text-amber-600 dark:text-amber-400">
                        已批注
                      </span>
                    )}
                  </button>

                  {isOpen && (
                    <div className="border-t border-line/50 px-3.5 pb-3.5">
                      <div className={`mt-2.5 ${proseQuestion}`}>
                        <Markdown>{section.content}</Markdown>
                      </div>

                      {isRevising && (
                        <div className="mt-3 rounded-md border border-line/60 bg-white/70 p-2.5 dark:bg-black/20">
                          <p className="text-[11px] font-medium text-sea-ink-soft">
                            针对本段的修改意见：
                          </p>
                          <Textarea
                            placeholder={`对「${section.title || section.id}」的修改批注...`}
                            value={annotations[section.id] ?? ''}
                            onChange={(e) =>
                              setAnnotations((prev) => ({
                                ...prev,
                                [section.id]: e.target.value,
                              }))
                            }
                            disabled={isSupplementLoading}
                            className="mt-1.5 min-h-[56px] resize-y rounded border-line bg-foam text-xs disabled:opacity-50"
                          />
                        </div>
                      )}
                    </div>
                  )}
                </div>
              )
            })}
          </div>
        </div>
      ) : showCustomContent ? (
        <div className="max-h-[420px] overflow-auto rounded-lg border border-line bg-foam p-4">
          <div className={proseQuestion}>
            <Markdown>{customContent}</Markdown>
          </div>
        </div>
      ) : null}

      <DecisionButtons
        variant="two"
        value={decision}
        onChange={(v) => setDecision(v as ReviewDecision)}
        disabled={isSupplementLoading}
      />

      {isRevising && (
        <Textarea
          placeholder={
            sections.length > 0
              ? '总体修改意见（可选，也可直接在上述各段落填写批注）...'
              : '请描述需要修改的内容...'
          }
          value={feedback}
          onChange={(e) => setFeedback(e.target.value)}
          disabled={isSupplementLoading}
          className="min-h-[80px] resize-y rounded-lg border-line bg-foam text-sm disabled:opacity-50"
        />
      )}
    </QuestionShell>
  )
}
