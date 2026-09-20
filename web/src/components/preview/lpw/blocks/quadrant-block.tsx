import type React from 'react'
import type {
  LpwBlockSlotProps,
  LpwQuadrantItem,
  LpwQuadrantProps,
} from '../types'

export const QuadrantBlock: React.FC<LpwBlockSlotProps<LpwQuadrantProps>> = ({
  props,
}) => {
  const { title, items } = props
  const xLow = props.xLow ?? '低'
  const xHigh = props.xHigh ?? '高'
  const yLow = props.yLow ?? '低'
  const yHigh = props.yHigh ?? '高'

  // 四个象限数据分组
  const qLeftTop = items.filter((it) => it.x === 'low' && it.y === 'high')
  const qRightTop = items.filter((it) => it.x === 'high' && it.y === 'high')
  const qLeftBottom = items.filter((it) => it.x === 'low' && it.y === 'low')
  const qRightBottom = items.filter((it) => it.x === 'high' && it.y === 'low')

  const renderCellContent = (
    cellItems: LpwQuadrantItem[],
    tag: string,
    testId: string,
  ) => {
    return (
      <div
        data-testid={testId}
        className="flex min-h-24 flex-col bg-surface p-3 text-xs"
      >
        <div className="mb-2 font-mono text-[10px] text-sea-ink-soft/60">
          {tag}
        </div>
        {cellItems.length === 0 ? (
          <div className="flex flex-1 items-center justify-center text-sm text-sea-ink-soft/40">
            —
          </div>
        ) : (
          <div className="space-y-1.5">
            {cellItems.map((item, idx) => (
              <div key={idx} className="leading-snug">
                <span className="font-semibold text-sea-ink">{item.label}</span>
                {item.note && (
                  <span className="ml-1 text-[11px] text-sea-ink-soft/70">
                    ({item.note})
                  </span>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    )
  }

  return (
    <div
      data-testid="quadrant-block"
      className="my-4 border border-line bg-surface p-4 text-xs"
    >
      {title && (
        <div className="mb-3 font-semibold text-sm text-sea-ink">{title}</div>
      )}

      <div className="flex items-stretch gap-2">
        {/* Y 轴名 */}
        {props.yLabel && (
          <div
            style={{ writingMode: 'vertical-rl' }}
            className="flex select-none items-center justify-center text-center font-semibold text-[11px] text-sea-ink-soft/80"
          >
            {props.yLabel}
          </div>
        )}

        {/* 2x2 网格 */}
        <div className="flex-1">
          <div className="grid grid-cols-2 gap-px border border-line bg-line">
            {/* 左上: xLow · yHigh */}
            {renderCellContent(
              qLeftTop,
              `${xLow} · ${yHigh}`,
              'quadrant-left-top',
            )}
            {/* 右上: xHigh · yHigh */}
            {renderCellContent(
              qRightTop,
              `${xHigh} · ${yHigh}`,
              'quadrant-right-top',
            )}
            {/* 左下: xLow · yLow */}
            {renderCellContent(
              qLeftBottom,
              `${xLow} · ${yLow}`,
              'quadrant-left-bottom',
            )}
            {/* 右下: xHigh · yLow */}
            {renderCellContent(
              qRightBottom,
              `${xHigh} · ${yLow}`,
              'quadrant-right-bottom',
            )}
          </div>

          {/* X 轴名 */}
          {props.xLabel && (
            <div className="mt-2 select-none text-center font-semibold text-[11px] text-sea-ink-soft/80">
              {props.xLabel}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
