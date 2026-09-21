import {
  BadgeCheck,
  CircleCheck,
  CircleHelp,
  Lightbulb,
  StickyNote,
  TriangleAlert,
} from 'lucide-react'
import type React from 'react'
import { useMemo } from 'react'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '../ui/tooltip'
import type { LpwAnnotation, LpwAnnotationKind } from './types'
import { isSafeAnnotationPattern } from './safe-pattern'

/** 正文超过该长度时跳过划线标注，限制多项式级回溯的总开销 */
const MAX_ANNOTATED_TEXT_LENGTH = 10000

const KIND_ICON_MAP: Record<
  LpwAnnotationKind,
  React.ComponentType<{ className?: string }>
> = {
  note: StickyNote,
  suggestion: Lightbulb,
  todo: CircleCheck,
  issue: TriangleAlert,
  approved: BadgeCheck,
  question: CircleHelp,
}

const KIND_COLOR_MAP: Record<LpwAnnotationKind, string> = {
  note: 'text-sea-ink-soft bg-surface-muted border-line',
  suggestion: 'text-lagoon-deep bg-lagoon/10 border-lagoon/30',
  todo: 'text-kicker bg-kicker/10 border-kicker/30',
  issue: 'text-destructive bg-destructive/10 border-destructive/30',
  approved: 'text-palm bg-palm/10 border-palm/30',
  question: 'text-purple-600 bg-purple-50 border-purple-200',
}

export interface AnnotationFrameProps {
  annotation?: LpwAnnotation
  children: React.ReactNode
}

export const AnnotationFrame: React.FC<AnnotationFrameProps> = ({
  annotation,
  children,
}) => {
  if (!annotation) {
    return <>{children}</>
  }

  const Icon = KIND_ICON_MAP[annotation.kind]
  const colorClass = KIND_COLOR_MAP[annotation.kind]

  return (
    <div
      data-testid="annotation-frame"
      data-kind={annotation.kind}
      className="group relative my-2"
    >
      <div className="absolute -top-3 right-2 z-20">
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger asChild>
              <button
                type="button"
                aria-label={`批注：${annotation.kind}`}
                className={`flex cursor-pointer items-center gap-1 rounded-full border px-2 py-0.5 font-mono text-[10px] shadow-2xs transition-transform hover:scale-105 ${colorClass}`}
              >
                <Icon className="h-3 w-3" />
                <span className="font-semibold uppercase tracking-wider">
                  {annotation.label || annotation.kind}
                </span>
              </button>
            </TooltipTrigger>
            <TooltipContent
              side="top"
              className="max-w-xs border border-line bg-surface-strong p-3 text-xs text-sea-ink shadow-md"
            >
              <p className="font-sans font-medium leading-relaxed">
                {annotation.message}
              </p>
              {annotation.author && (
                <p className="mt-1 font-mono text-[10px] text-sea-ink-soft">
                  — {annotation.author}
                </p>
              )}
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      </div>

      <div className="relative">{children}</div>
    </div>
  )
}

export interface AnnotatedTextProps {
  field: string
  value: string
  annotation?: LpwAnnotation
}

export const AnnotatedText: React.FC<AnnotatedTextProps> = ({
  field,
  value,
  annotation,
}) => {
  const parts = useMemo(() => {
    if (!annotation?.targets || annotation.targets.length === 0) {
      return null
    }

    const target = annotation.targets.find((t) => t.field === field)
    if (!target || !target.pattern) {
      return null
    }
    // ReDoS 防线：回溯型引擎执行前先过静态语法子集检查，危险模式回退原文渲染
    if (!isSafeAnnotationPattern(target.pattern)) {
      return null
    }
    if (value.length > MAX_ANNOTATED_TEXT_LENGTH) {
      return null
    }

    try {
      // Q-10 修复：对齐 Schema 枚举 [i, u, iu]，透传 u/iu 标志以支持 Unicode 属性与码点转义
      const hasI = target.flags?.includes('i')
      const hasU = target.flags?.includes('u')
      const flags = `g${hasI ? 'i' : ''}${hasU ? 'u' : ''}`
      const regex = new RegExp(target.pattern, flags)
      const segments: Array<{ text: string; isMatch: boolean }> = []
      let lastIndex = 0
      let match: RegExpExecArray | null
      let iterations = 0
      const maxIterations = 1000

      while ((match = regex.exec(value)) !== null) {
        iterations++
        if (iterations > maxIterations) {
          break
        }

        const matchStart = match.index
        const matchText = match[0]

        if (matchText.length === 0) {
          regex.lastIndex++
          continue
        }

        if (matchStart > lastIndex) {
          segments.push({
            text: value.slice(lastIndex, matchStart),
            isMatch: false,
          })
        }

        segments.push({
          text: matchText,
          isMatch: true,
        })

        lastIndex = matchStart + matchText.length
      }

      if (lastIndex < value.length) {
        segments.push({
          text: value.slice(lastIndex),
          isMatch: false,
        })
      }

      return segments.length > 0 ? segments : null
    } catch {
      return null
    }
  }, [field, value, annotation])

  if (!parts) {
    return <>{value}</>
  }

  return (
    <>
      {parts.map((segment, idx) => {
        if (segment.isMatch) {
          return (
            <mark
              key={idx}
              className="bg-transparent text-inherit underline decoration-current decoration-wavy decoration-2 underline-offset-4"
            >
              {segment.text}
            </mark>
          )
        }
        return <span key={idx}>{segment.text}</span>
      })}
    </>
  )
}
