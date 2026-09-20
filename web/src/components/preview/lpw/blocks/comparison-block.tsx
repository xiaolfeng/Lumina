import { Badge } from '@lumina/components/ui/badge'
import type React from 'react'
import type {
  LpwBlockSlotProps,
  LpwComparisonCell,
  LpwComparisonProps,
} from '../types'

const VERDICT_CLASSES: Record<
  NonNullable<LpwComparisonCell['verdict']>,
  string
> = {
  good: 'text-kicker font-semibold',
  warn: 'text-palm font-semibold',
  bad: 'text-destructive font-semibold',
}

export const ComparisonBlock: React.FC<
  LpwBlockSlotProps<LpwComparisonProps>
> = ({ props }) => {
  const { title, plans, rows } = props

  return (
    <div
      data-testid="comparison-block"
      className="my-4 overflow-x-auto border border-line bg-surface text-xs"
    >
      {title && (
        <div className="border-b border-line bg-surface-muted/30 px-4 py-2 font-semibold text-sm text-sea-ink">
          {title}
        </div>
      )}

      <table className="w-full border-collapse text-left">
        <thead>
          <tr className="border-b border-line bg-surface-muted/20">
            <th className="p-3 font-semibold text-sea-ink w-1/4">维度</th>
            {plans.map((plan, idx) => (
              <th
                key={idx}
                className={`p-3 font-semibold text-sea-ink ${
                  plan.recommended
                    ? 'bg-lagoon/5 border-x border-lagoon/20'
                    : ''
                }`}
              >
                <div className="flex items-center gap-1.5">
                  <span>{plan.name}</span>
                  {plan.recommended && (
                    <Badge
                      variant="default"
                      className="text-[10px] bg-lagoon text-foam"
                    >
                      推荐
                    </Badge>
                  )}
                </div>
              </th>
            ))}
          </tr>
        </thead>
        <tbody className="divide-y divide-line/60">
          {rows.length === 0 ? (
            <tr>
              <td
                colSpan={plans.length + 1}
                className="p-8 text-center text-xs text-sea-ink-soft/60"
              >
                暂无对比维度
              </td>
            </tr>
          ) : (
            rows.map((row, rIdx) => (
              <tr key={rIdx} className="hover:bg-surface-muted/10">
                <td className="p-3 font-semibold text-sea-ink">
                  {row.dimension}
                </td>
                {row.values.map((cell, cIdx) => {
                  const plan = plans[cIdx]
                  const isRec = Boolean(plan.recommended)
                  const verdictClass = cell.verdict
                    ? VERDICT_CLASSES[cell.verdict]
                    : 'text-sea-ink'

                  return (
                    <td
                      key={cIdx}
                      className={`p-3 ${
                        isRec ? 'bg-lagoon/5 border-x border-lagoon/20' : ''
                      } ${verdictClass}`}
                    >
                      {cell.text}
                    </td>
                  )
                })}
              </tr>
            ))
          )}
        </tbody>
      </table>
    </div>
  )
}
