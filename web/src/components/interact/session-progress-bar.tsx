import type { SessionProgress } from './session-progress'
import {
  sessionProgressCancelledPercent,
  sessionProgressPercent,
  sessionProgressSkippedPercent,
} from './session-progress'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@lumina/components/ui/tooltip'

/** 详情浮层明细行：色点与进度条多色段一一对应 */
const DETAIL_ROWS = [
  { key: 'answered', label: '已回答', dot: 'bg-lagoon' },
  { key: 'remaining', label: '未回答', dot: 'border border-white/30' },
  { key: 'skipped', label: '已跳过', dot: 'bg-amber-400' },
  { key: 'cancelled', label: '已取消', dot: 'bg-rose-500' },
] as const

export function SessionProgressBar({
  progress,
}: {
  progress: SessionProgress
}) {
  const answeredPercent = sessionProgressPercent(progress)
  const skippedPercent = sessionProgressSkippedPercent(progress)
  const cancelledPercent = sessionProgressCancelledPercent(progress)
  const label = `问题进度：已答 ${progress.answered}，跳过 ${progress.skipped}，取消 ${progress.cancelled}，未答 ${progress.remaining}，共 ${progress.total}`

  return (
    <div className="flex min-w-0 items-center gap-2.5">
      {/* 多色段进度条：已答 lagoon / 已跳过琥珀 / 已取消红，未答保留底色空槽 */}
      <div
        className="flex h-1 w-16 overflow-hidden bg-line sm:w-28"
        role="progressbar"
        aria-valuemin={0}
        aria-valuemax={progress.total}
        aria-valuenow={progress.answered}
        aria-valuetext={label}
      >
        <div
          className="h-full bg-lagoon transition-[width] duration-500 ease-out"
          style={{ width: `${answeredPercent}%` }}
        />
        <div
          className="h-full bg-amber-400 transition-[width] duration-500 ease-out"
          style={{ width: `${skippedPercent}%` }}
        />
        <div
          className="h-full bg-rose-500 transition-[width] duration-500 ease-out"
          style={{ width: `${cancelledPercent}%` }}
        />
      </div>

      {/* 紧凑读数「已答 / 总计」：Hover（桌面）或点击聚焦（移动端）弹出明细 Tooltip */}
      <TooltipProvider delayDuration={150}>
        <Tooltip>
          <TooltipTrigger asChild>
            <button
              type="button"
              className="flex cursor-pointer items-center gap-1 px-0.5 text-[11px] tabular-nums text-sea-ink-soft transition-colors hover:text-sea-ink focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-lagoon"
              aria-label={label}
            >
              <span className="hidden sm:inline">已答</span>
              <span className="font-semibold text-sea-ink">
                {progress.answered}
              </span>
              <span className="text-line">/</span>
              <span>{progress.total}</span>
            </button>
          </TooltipTrigger>
          <TooltipContent
            side="bottom"
            align="end"
            className="w-38 px-3 py-2.5"
          >
            <p className="text-[10px] font-semibold uppercase tracking-[0.12em] text-background/60">
              问题进度
            </p>
            <ul className="mt-1.5 space-y-1">
              {DETAIL_ROWS.map((row) => (
                <li
                  key={row.key}
                  className="flex items-center justify-between gap-4 text-xs"
                >
                  <span className="flex items-center gap-1.5">
                    <span className={`size-1.5 ${row.dot}`} aria-hidden />
                    {row.label}
                  </span>
                  <span className="font-semibold tabular-nums">
                    {progress[row.key]}
                  </span>
                </li>
              ))}
            </ul>
            <div className="mt-1.5 flex items-center justify-between gap-4 border-t border-background/15 pt-1.5 text-xs">
              <span>总计</span>
              <span className="font-semibold tabular-nums">
                {progress.total}
              </span>
            </div>
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    </div>
  )
}
