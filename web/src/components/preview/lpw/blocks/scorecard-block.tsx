import { Badge } from '@lumina/components/ui/badge'
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
      for (let i = 0; i < criteria.length; i++) {
        const weight = criteria[i].weight
        const score = plan.scores[i] ?? 0
        sum += score * weight
      }
      return (sum / 100).toFixed(1)
    })
  }, [criteria, plans])

  return (
    <div
      data-testid="scorecard-block"
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
            <th className="p-3 font-semibold text-sea-ink w-1/3">
              评估准则 (权重)
            </th>
            {plans.map((plan, pIdx) => (
              <th
                key={pIdx}
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
          {criteria.map((crit, cIdx) => (
            <tr key={cIdx} className="hover:bg-surface-muted/10">
              <td className="p-3 text-sea-ink">
                <span className="font-semibold">{crit.name}</span>
                <span className="ml-1 text-[11px] text-sea-ink-soft/70">
                  ({crit.weight}%)
                </span>
              </td>
              {plans.map((plan, pIdx) => (
                <td
                  key={pIdx}
                  className={`p-3 font-mono ${
                    plan.recommended
                      ? 'bg-lagoon/5 border-x border-lagoon/20'
                      : ''
                  }`}
                >
                  {plan.scores[cIdx] ?? '—'}
                </td>
              ))}
            </tr>
          ))}

          {/* 末行：加权总分 */}
          <tr className="border-t-2 border-line bg-surface-muted/30 font-semibold">
            <td className="p-3 text-sea-ink">加权总分</td>
            {plans.map((plan, pIdx) => (
              <td
                key={pIdx}
                data-testid={`scorecard-total-${pIdx}`}
                className={`p-3 font-mono text-sm text-sea-ink ${
                  plan.recommended
                    ? 'bg-lagoon/5 border-x border-lagoon/20 text-lagoon-deep font-bold'
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
  )
}
