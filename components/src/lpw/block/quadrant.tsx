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
        className="flex min-h-28 flex-col bg-surface p-4 text-xs"
      >
        <div className="mb-2.5 font-mono text-[10px] uppercase tracking-wider text-sea-ink-soft/70 border-b border-line/40 pb-1">
          {tag}
        </div>
        {cellItems.length === 0 ? (
          <div className="flex flex-1 items-center justify-center font-serif text-sm text-sea-ink-soft/40">
            —
          </div>
        ) : (
          <div className="space-y-2">
            {cellItems.map((item, idx) => (
              <div key={idx} className="leading-snug">
                <span className="font-serif font-semibold text-sea-ink text-[13px]">
                  {item.label}
                </span>
                {item.note && (
                  <span className="ml-1.5 font-mono text-[11px] text-sea-ink-soft/80">
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
      className="my-8 border-t-2 border-b-2 border-sea-ink bg-surface/30 p-6 text-xs shadow-2xs font-sans"
    >
      {title && (
        <div className="mb-4 pb-2 border-b border-line/60 font-serif font-semibold text-base text-sea-ink flex items-center justify-between">
          <span>{title}</span>
          <span className="font-mono text-[10px] text-sea-ink-soft uppercase tracking-widest">
            QUADRANT MATRIX // 战略象限
          </span>
        </div>
      )}

      <div className="flex items-stretch gap-3">
        {/* Y 轴名 */}
        {props.yLabel && (
          <div
            style={{ writingMode: 'vertical-rl' }}
            className="flex select-none items-center justify-center text-center font-mono font-semibold text-xs tracking-widest text-sea-ink-soft/90 px-1 border-r border-line/60"
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
            <div className="mt-2 text-center font-mono font-semibold text-xs tracking-widest text-sea-ink-soft/90">
              {props.xLabel}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
