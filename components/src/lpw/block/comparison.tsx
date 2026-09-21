import { Award, Layers } from 'lucide-react'
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
      className="relative my-8 border-t-2 border-b-2 border-sea-ink bg-surface/30 text-xs shadow-2xs font-sans"
    >
      <div
        aria-hidden="true"
        className="pointer-events-none absolute inset-y-0 right-0 w-8 bg-gradient-to-l from-surface via-surface/40 to-transparent sm:hidden z-10"
      />
      {title && (
        <div className="border-b border-line bg-surface/60 px-5 py-3 font-serif font-semibold text-base text-sea-ink flex items-center justify-between">
          <span>{title}</span>
          <span className="inline-flex items-center gap-1 font-mono text-[10px] text-sea-ink-soft uppercase tracking-widest">
            <Layers className="h-3 w-3" />
            <span>DIMENSIONAL MATRIX</span>
          </span>
        </div>
      )}

      <div className="overflow-x-auto">
        <table className="w-full min-w-[520px] border-collapse text-left">
          <thead>
            <tr className="border-b border-sea-ink bg-surface/50">
              <th
                scope="col"
                className="p-3.5 font-mono text-[11px] font-bold uppercase tracking-wider text-sea-ink min-w-[100px] sm:w-1/4"
              >
                维度
              </th>
              {plans.map((plan, idx) => {
                const isRec = Boolean(plan.recommended)
                return (
                  <th
                    key={idx}
                    scope="col"
                    className={`p-3.5 font-mono text-[11px] font-bold uppercase tracking-wider text-sea-ink ${
                      isRec
                        ? 'bg-lagoon/5 border-x border-lagoon/20 text-lagoon-deep'
                        : ''
                    }`}
                  >
                    <div className="flex items-center gap-1.5">
                      <span>{plan.name}</span>
                      {isRec && (
                        <Badge
                          variant="default"
                          className="inline-flex items-center gap-0.5 text-[9.5px] font-mono uppercase bg-lagoon text-foam px-1.5 py-0 rounded-none tracking-normal"
                        >
                          <Award className="h-3 w-3" />
                          <span>推荐</span>
                        </Badge>
                      )}
                    </div>
                  </th>
                )
              })}
            </tr>
          </thead>
          <tbody className="divide-y divide-line/60">
            {plans.length === 0 ? (
              <tr>
                <td
                  colSpan={1}
                  className="p-8 text-center text-xs font-serif italic text-sea-ink-soft/70"
                >
                  暂无方案对比数据
                </td>
              </tr>
            ) : rows.length === 0 ? (
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
                  <th
                    scope="row"
                    className="p-3.5 font-semibold text-sea-ink text-left"
                  >
                    {row.dimension}
                  </th>
                  {row.values.slice(0, plans.length).map((cell, cIdx) => {
                    const plan = plans[cIdx] as (typeof plans)[number] | undefined
                    const isRec = Boolean(plan?.recommended)
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
                        {cell.verdict && (
                          <span className="sr-only">
                            ({cell.verdict === 'good' ? '优势' : cell.verdict === 'warn' ? '注意' : '劣势'})
                          </span>
                        )}
                      </td>
                    )
                  })}
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}
