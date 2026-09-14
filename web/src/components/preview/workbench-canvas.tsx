import { useEffect, useRef, useState } from 'react'
import { PreviewFileViewer } from '#/components/preview/file-viewer'
import { previewKindFromFilename } from '#/lib/preview-file'

type Device = 'desktop' | 'tablet' | 'mobile'

const PRESETS: { id: Device; label: string; width: number }[] = [
  { id: 'desktop', label: '桌面 100%', width: 0 },
  { id: 'tablet', label: '平板 768px', width: 768 },
  { id: 'mobile', label: '手机 375px', width: 375 },
]

const QUICK_WIDTHS = [375, 414, 768, 1024]

export function WorkbenchCanvas({
  src,
  filename,
  sourceMode,
}: {
  src: string
  filename: string
  sourceMode: boolean
}) {
  const [device, setDevice] = useState<Device>('desktop')
  const [width, setWidth] = useState(375)
  const dragging = useRef(false)
  const wrapperRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (sourceMode) setDevice('desktop')
  }, [sourceMode])

  useEffect(() => {
    const onMove = (event: PointerEvent) => {
      if (!dragging.current || !wrapperRef.current) return
      const rect = wrapperRef.current.getBoundingClientRect()
      const next = Math.min(1200, Math.max(320, event.clientX - rect.left))
      setWidth(next)
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
  }, [])

  const kind = sourceMode ? 'code' : previewKindFromFilename(filename)
  const isDevice = !sourceMode && device !== 'desktop'
  const frameWidth = sourceMode || device === 'desktop' ? '100%' : `${width}px`

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      {!sourceMode ? (
        <div className="flex shrink-0 items-center gap-2 overflow-x-auto whitespace-nowrap border-b border-line px-3 py-2">
          {PRESETS.map((item) => (
            <button
              key={item.id}
              type="button"
              onClick={() => {
                setDevice(item.id)
                if (item.width) setWidth(item.width)
              }}
              className={`shrink-0 px-2 py-1 text-[11px] font-semibold ${
                device === item.id
                  ? 'bg-lagoon/15 text-lagoon-deep'
                  : 'text-sea-ink-soft hover:bg-line/40'
              }`}
            >
              {item.label}
            </button>
          ))}
        </div>
      ) : null}

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
              max={1100}
              value={width}
              onChange={(event) => setWidth(Number(event.target.value))}
            />
            {QUICK_WIDTHS.map((value) => (
              <button
                key={value}
                type="button"
                className="border border-line px-1.5 py-0.5"
                onClick={() => setWidth(value)}
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
