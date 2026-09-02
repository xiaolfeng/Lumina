import type { SessionProgress } from './session-progress'
import { sessionProgressPercent } from './session-progress'

export function SessionProgressBar({
  progress,
}: {
  progress: SessionProgress
}) {
  const percent = sessionProgressPercent(progress)
  const label = `问题进度：已答 ${progress.answered}，未答 ${progress.remaining}，共 ${progress.total}`

  return (
    <div className="flex min-w-0 items-center gap-2.5" aria-label={label}>
      <div
        className="h-1 w-16 overflow-hidden bg-line sm:w-24"
        role="progressbar"
        aria-valuemin={0}
        aria-valuemax={progress.total}
        aria-valuenow={progress.answered}
        aria-valuetext={label}
      >
        <div
          className="h-full bg-lagoon transition-[width] duration-500 ease-out"
          style={{ width: `${percent}%` }}
        />
      </div>
      <div className="flex items-center gap-1.5 text-[11px] tabular-nums text-sea-ink-soft">
        <span className="font-semibold text-sea-ink sm:hidden">
          {progress.answered}/{progress.total}
        </span>
        <span className="hidden sm:inline">
          已答{' '}
          <span className="font-semibold text-sea-ink">
            {progress.answered}
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
