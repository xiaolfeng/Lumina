import { Calculator, Sigma } from 'lucide-react'
import { Badge } from '../../ui/badge'
import React, { useMemo } from 'react'
import type { LpwBlockSlotProps, LpwScorecardProps } from '../types'

export const ScorecardBlock: React.FC<LpwBlockSlotProps<LpwScorecardProps>> = ({
  props,
}) => {
  const { title, criteria, plans } = props

  // 计算每个 plan 的加权总分
  const weightedTotals = useMemo(() => {
    return plans.map((plan) => {
      let sum = 0
      let validWeight = 0
      for (let i = 0; i < criteria.length; i++) {
        const score = plan.scores[i]
        if (typeof score === 'number') {
          const weight = criteria[i].weight
          sum += score * weight
          validWeight += weight
        }
      }
      if (validWeight === 0) return '—'
      return (sum / validWeight).toFixed(1)
    })
  }, [criteria, plans])

  return (
    <div
      data-testid="scorecard-block"
      className="relative min-w-0 max-w-full overflow-hidden border border-line bg-surface/30 text-xs font-sans"
    >
      <div
        aria-hidden="true"
        className="pointer-events-none absolute inset-y-0 right-0 w-8 bg-gradient-to-l from-surface via-surface/40 to-transparent sm:hidden z-10"
      />
      {title && (
        <div className="border-b border-line bg-surface/60 px-5 py-3 font-serif font-semibold text-base text-sea-ink flex items-center justify-between">
          <span>{title}</span>
          <Calculator className="h-4 w-4 text-sea-ink-soft" aria-hidden="true" />
        </div>
      )}

      <div className="overflow-x-auto">
        <table className="w-full border-collapse text-left">
          <thead>
            <tr className="border-b border-sea-ink bg-surface/50">
              <th className="p-3.5 text-xs font-semibold text-sea-ink w-1/3">
                评估准则 (权重)
              </th>
              {plans.map((plan, pIdx) => (
                <th
                  key={pIdx}
                  className={`p-3.5 text-xs font-semibold text-sea-ink ${
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
            {criteria.map((crit, cIdx) => (
              <tr key={cIdx} className="hover:bg-surface/60 transition-colors">
                <td className="p-3.5 text-sea-ink">
                  <span className="font-medium">{crit.name}</span>
                  <span className="ml-1.5 font-mono text-[11px] text-sea-ink-soft/80">
                    ({crit.weight}%)
                  </span>
                </td>
                {plans.map((plan, pIdx) => {
                  const score: number | undefined = plan.scores[cIdx]
                  const maxScore = 5
                  const scorePercent =
                    typeof score === 'number'
                      ? Math.min(100, (score / maxScore) * 100)
                      : 0

                  return (
                    <td
                      key={pIdx}
                      className={`p-3.5 font-mono ${
                        plan.recommended
                          ? 'bg-lagoon/5 border-x border-lagoon/20'
                          : ''
                      }`}
                    >
                      <div className="flex items-center gap-2">
                        <span
                          className={`text-xs font-semibold w-5 ${plan.recommended ? 'text-lagoon-deep' : 'text-sea-ink'}`}
                        >
                          {/* eslint-disable-next-line @typescript-eslint/no-unnecessary-condition */}
                          {score ?? '—'}
                        </span>
                        {typeof score === 'number' && (
                          <div className="flex-1 max-w-[60px] h-1 bg-line/60 rounded-none overflow-hidden">
                            <div
                              className={`h-full ${plan.recommended ? 'bg-lagoon' : 'bg-sea-ink-soft'}`}
                              style={{ width: `${scorePercent}%` }}
                            />
                          </div>
                        )}
                      </div>
                    </td>
                  )
                })}
              </tr>
            ))}

            {/* 末行：加权总分 */}
            <tr className="border-t-[1.5px] border-sea-ink bg-surface/60 font-semibold">
              <td className="p-3.5 font-mono text-xs uppercase tracking-wider text-sea-ink">
                <span className="inline-flex items-center gap-1">
                  <Sigma className="h-3.5 w-3.5" />
                  <span>加权总分</span>
                </span>
              </td>
              {plans.map((plan, pIdx) => (
                <td
                  key={pIdx}
                  data-testid={`scorecard-total-${pIdx}`}
                  className={`p-3.5 font-mono text-sm text-sea-ink ${
                    plan.recommended
                      ? 'bg-lagoon/5 border-x border-lagoon/20 text-lagoon-deep font-bold text-base'
                      : ''
                  }`}
                >
                  {weightedTotals[pIdx]}
                </td>
              ))}
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  )
}
