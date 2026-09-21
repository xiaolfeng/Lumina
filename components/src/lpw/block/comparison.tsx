import { Badge } from '../../ui/badge'
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
      className="my-8 overflow-x-auto border-t-2 border-b-2 border-sea-ink bg-surface/30 text-xs shadow-2xs font-sans"
    >
      {title && (
        <div className="border-b border-line bg-surface/60 px-5 py-3 font-serif font-semibold text-base text-sea-ink flex items-center justify-between">
          <span>{title}</span>
          <span className="font-mono text-[10px] text-sea-ink-soft uppercase tracking-widest">
            DIMENSIONAL MATRIX
          </span>
        </div>
      )}

      <table className="w-full border-collapse text-left">
        <thead>
          <tr className="border-b border-sea-ink bg-surface/50">
            <th className="p-3.5 font-mono text-[11px] font-bold uppercase tracking-wider text-sea-ink w-1/4">
              维度
            </th>
            {plans.map((plan, idx) => (
              <th
                key={idx}
                className={`p-3.5 font-mono text-[11px] font-bold uppercase tracking-wider text-sea-ink ${
                  plan.recommended
                    ? 'bg-lagoon/5 border-x border-lagoon/20 text-lagoon-deep'
                    : ''
                }`}
              >
                <div className="flex items-center gap-1.5">
                  <span>{plan.name}</span>
                  {plan.recommended && (
                    <Badge
                      variant="default"
                      className="text-[9.5px] font-mono uppercase bg-lagoon text-foam px-1.5 py-0 rounded-none tracking-normal"
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
                className="p-8 text-center text-xs font-serif italic text-sea-ink-soft/70"
              >
                暂无对比维度
              </td>
            </tr>
          ) : (
            rows.map((row, rIdx) => (
              <tr key={rIdx} className="hover:bg-surface/60 transition-colors">
                <td className="p-3.5 font-semibold text-sea-ink">
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
                      className={`p-3.5 ${
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
