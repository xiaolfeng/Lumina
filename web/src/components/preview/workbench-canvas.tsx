import { useEffect, useRef, useState } from 'react'
import { PreviewFileViewer } from '#/components/preview/file-viewer'
import { previewKindFromFilename } from '#/lib/preview-file'

export type WorkbenchDevice = 'desktop' | 'tablet' | 'mobile'

export const DEVICE_PRESETS: {
  id: WorkbenchDevice
  label: string
  width: number
}[] = [
  { id: 'desktop', label: '桌面', width: 0 },
  { id: 'tablet', label: '平板', width: 768 },
  { id: 'mobile', label: '手机', width: 375 },
]

const QUICK_WIDTHS = [375, 414, 768, 1024]

const MIN_DEVICE_WIDTH = 320
const MAX_DEVICE_WIDTH = 1280
const VIEWPORT_SIDE_GAP = 32

/**
 * 设备框宽度钳制：下限固定 320，上限随视口动态收缩为 min(1280, 视口宽 - 32)，
 * 窄视口下设备框不允许超出可视区域；拖拽 / 滑杆 / 快捷宽度共用本函数
 */
export function clampDeviceWidth(width: number, viewportWidth: number): number {
  const upperBound = Math.min(
    MAX_DEVICE_WIDTH,
    Math.max(MIN_DEVICE_WIDTH, viewportWidth - VIEWPORT_SIDE_GAP),
  )
  return Math.min(upperBound, Math.max(MIN_DEVICE_WIDTH, Math.round(width)))
}

export function WorkbenchCanvas({
  src,
  filename,
  sourceMode,
  device,
}: {
  src: string
  filename: string
  sourceMode: boolean
  device: WorkbenchDevice
}) {
  const [viewportWidth, setViewportWidth] = useState(() =>
    typeof window === 'undefined' ? MAX_DEVICE_WIDTH : window.innerWidth,
  )
  const [width, setWidth] = useState(() => clampDeviceWidth(375, viewportWidth))
  const dragging = useRef(false)
  const wrapperRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const preset = DEVICE_PRESETS.find((item) => item.id === device)
    if (preset?.width) {
      setWidth((current) =>
        clampDeviceWidth(
          preset.width,
          typeof window === 'undefined' ? current : window.innerWidth,
        ),
      )
    }
  }, [device])

  // 视口变化时同步动态上限，并把已超限的当前宽度收拢
  useEffect(() => {
    const onResize = () => {
      setViewportWidth(window.innerWidth)
      setWidth((current) => clampDeviceWidth(current, window.innerWidth))
    }
    window.addEventListener('resize', onResize)
    return () => window.removeEventListener('resize', onResize)
  }, [])

  useEffect(() => {
    const onMove = (event: PointerEvent) => {
      if (!dragging.current || !wrapperRef.current) return
      const rect = wrapperRef.current.getBoundingClientRect()
      setWidth(clampDeviceWidth(event.clientX - rect.left, viewportWidth))
    }
    const onUp = () => {
      dragging.current = false
    }
    window.addEventListener('pointermove', onMove)
    window.addEventListener('pointerup', onUp)
    return () => {
      window.removeEventListener('pointermove', onMove)
      window.removeEventListener('pointerup', onUp)
    }
  }, [viewportWidth])

  const kind = sourceMode ? 'code' : previewKindFromFilename(filename)
  const isDevice = !sourceMode && device !== 'desktop'
  const frameWidth = sourceMode || device === 'desktop' ? '100%' : `${width}px`
  // 动态上限：min(1280, 视口宽 - 32)，滑杆最大值再与 1100 取小
  const widthUpperBound = clampDeviceWidth(MAX_DEVICE_WIDTH, viewportWidth)

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      <div
        className={`flex min-h-0 flex-1 flex-col ${
          isDevice ? 'bg-[#e9e3d8] px-4 pt-9 pb-3 dark:bg-[#0d0b09]' : 'bg-sand'
        }`}
      >
        <div
          ref={wrapperRef}
          className={`relative mx-auto flex min-h-0 flex-1 flex-col ${
            isDevice
              ? 'border border-line-strong shadow-[0_12px_36px_rgba(0,0,0,0.14)]'
              : ''
          }`}
          style={{
            width: frameWidth,
            maxWidth: '100%',
            height: isDevice ? 'calc(100% - 38px)' : '100%',
          }}
        >
          <div className="min-h-0 flex-1 overflow-y-auto overflow-x-hidden">
            {src ? (
              <PreviewFileViewer kind={kind} src={src} filename={filename} />
            ) : null}
          </div>
          {isDevice ? (
            <button
              type="button"
              aria-label="拖拽调整宽度"
              className="absolute top-1/2 right-[-13px] h-16 w-3 -translate-y-1/2 cursor-ew-resize bg-lagoon/80"
              onPointerDown={() => {
                dragging.current = true
              }}
            />
          ) : null}
        </div>
        {isDevice ? (
          <div className="device-bottom-tip mt-2 flex flex-wrap items-center justify-center gap-3 text-[11px] text-sea-ink-soft">
            <span>
              ↔ 宽度: {Math.round(width)}px
              {width <= 430 ? ' (手机)' : width <= 900 ? ' (平板)' : ''}
            </span>
            <input
              type="range"
              min={320}
              max={Math.min(1100, widthUpperBound)}
              value={width}
              onChange={(event) =>
                setWidth(
                  clampDeviceWidth(Number(event.target.value), viewportWidth),
                )
              }
            />
            {QUICK_WIDTHS.map((value) => (
              <button
                key={value}
                type="button"
                className="border border-line px-1.5 py-0.5"
                onClick={() => setWidth(clampDeviceWidth(value, viewportWidth))}
              >
                {value}
              </button>
            ))}
          </div>
        ) : null}
      </div>
    </div>
  )
}
