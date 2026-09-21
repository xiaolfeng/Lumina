import type { SessionProgress } from './session-progress'
import {
  sessionProgressPercent,
  sessionProgressCancelledPercent,
} from './session-progress'

export function SessionProgressBar({
  progress,
}: {
  progress: SessionProgress
}) {
  const answeredPercent = sessionProgressPercent(progress)
  const cancelledPercent = sessionProgressCancelledPercent(progress)
  const label = `问题进度：已答 ${progress.answered}，取消 ${progress.cancelled}，未答 ${progress.remaining}，共 ${progress.total}`

  return (
    <div className="flex min-w-0 items-center gap-2.5" aria-label={label}>
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
          title={`已答 ${progress.answered}`}
        />
        <div
          className="h-full bg-rose-400 dark:bg-rose-500 transition-[width] duration-500 ease-out"
          style={{ width: `${cancelledPercent}%` }}
          title={`已取消 ${progress.cancelled}`}
        />
      </div>
      <div className="flex items-center gap-1.5 text-[11px] tabular-nums text-sea-ink-soft">
        <span className="font-semibold text-sea-ink sm:hidden">
          {progress.answered}
          {progress.cancelled > 0 ? `+${progress.cancelled}` : ''}/
          {progress.total}
        </span>
        <span className="hidden sm:inline">
          已答{' '}
          <span className="font-semibold text-sea-ink">
            {progress.answered}
          </span>
        </span>
        <span className="hidden text-line sm:inline">·</span>
        <span className="hidden sm:inline">
          取消{' '}
          <span className="font-semibold text-rose-500">
            {progress.cancelled}
          </span>
        </span>
        <span className="hidden text-line sm:inline">·</span>
        <span className="hidden sm:inline">
          未答{' '}
          <span className="font-semibold text-lagoon">
            {progress.remaining}
          </span>
        </span>
        <span className="hidden text-line sm:inline">·</span>
        <span className="hidden sm:inline">
          总计{' '}
          <span className="font-semibold text-sea-ink">{progress.total}</span>
        </span>
      </div>
    </div>
  )
}
