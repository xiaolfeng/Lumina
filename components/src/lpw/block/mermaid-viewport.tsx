import { Maximize2, MoveHorizontal, ZoomIn } from 'lucide-react'
import type React from 'react'
import { useCallback, useEffect, useRef, useState } from 'react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '../../ui/dialog'

export interface MermaidViewportProps {
  children: React.ReactNode
}

type ViewportMode = 'fit' | 'actual'

export const MermaidViewport: React.FC<MermaidViewportProps> = ({
  children,
}) => {
  const containerRef = useRef<HTMLDivElement>(null)
  const fullscreenButtonRef = useRef<HTMLButtonElement>(null)
  const [mode, setMode] = useState<ViewportMode>('fit')
  const [fullscreen, setFullscreen] = useState(false)
  const [reducedMotion, setReducedMotion] = useState(false)

  useEffect(() => {
    if (typeof window.matchMedia === 'function') {
      const mq = window.matchMedia('(prefers-reduced-motion: reduce)')
      setReducedMotion(mq.matches)
      const handler = (e: MediaQueryListEvent) => setReducedMotion(e.matches)
      mq.addEventListener('change', handler)
      return () => mq.removeEventListener('change', handler)
    }
  }, [])

  const applySvgStyles = useCallback((targetMode: ViewportMode) => {
    const el = containerRef.current
    if (!el) return
    const svg = el.querySelector<SVGElement>(
      '[data-testid="mermaid-svg-container"] svg',
    )
    if (!svg) return

    // 补齐缺失的 viewBox
    if (!svg.getAttribute('viewBox')) {
      let numW = 0
      let numH = 0

      const rawW = svg.getAttribute('width')
      const rawH = svg.getAttribute('height')

      if (rawW && !rawW.includes('%')) {
        const parsed = parseFloat(rawW)
        if (!isNaN(parsed) && parsed > 0) numW = parsed
      }
      if (rawH && !rawH.includes('%')) {
        const parsed = parseFloat(rawH)
        if (!isNaN(parsed) && parsed > 0) numH = parsed
      }

      // 如果未通过属性获取到非百分比绝对尺寸，优先使用 clientWidth / clientHeight
      if (numW === 0 && svg.clientWidth > 0) {
        numW = svg.clientWidth
      }
      if (numH === 0 && svg.clientHeight > 0) {
        numH = svg.clientHeight
      }

      // 回退使用 getBBox
      if ((numW === 0 || numH === 0) && typeof (svg as unknown as SVGGraphicsElement).getBBox === 'function') {
        try {
          const bbox = (svg as unknown as SVGGraphicsElement).getBBox()
          if (bbox.width > 0 && bbox.height > 0) {
            numW = bbox.width
            numH = bbox.height
          }
        } catch {
          // ignore getBBox error in non-rendered or detached SVG
        }
      }

      if (numW > 0 && numH > 0) {
        svg.setAttribute('viewBox', `0 0 ${numW} ${numH}`)
      }
    }

    if (targetMode === 'fit') {
      svg.style.width = '100%'
      svg.style.height = 'auto'
      svg.style.maxWidth = 'none'
    } else {
      svg.style.width = ''
      svg.style.height = ''
      svg.style.maxWidth = ''
    }
  }, [])

  useEffect(() => {
    applySvgStyles(mode)

    const el = containerRef.current
    if (!el) return

    const observer = new MutationObserver(() => {
      applySvgStyles(mode)
    })

    observer.observe(el, { childList: true, subtree: true })
    return () => observer.disconnect()
  }, [mode, applySvgStyles])

  const handleModeChange = (newMode: ViewportMode) => {
    setMode(newMode)
    applySvgStyles(newMode)
  }

  const handleDialogClose = (open: boolean) => {
    setFullscreen(open)
    if (!open) {
      setTimeout(() => {
        fullscreenButtonRef.current?.focus()
      }, 0)
    }
  }

  const transitionClass = reducedMotion
    ? ''
    : 'transition-colors duration-150'

  return (
    <div
      data-testid="mermaid-viewport"
      className="w-full border border-line/60 bg-surface/30 shadow-2xs font-sans"
    >
      {/* 视口工具栏 */}
      <div className="flex items-center justify-between border-b border-line/40 bg-surface/50 px-3 py-1.5 text-xs text-sea-ink-soft">
        <span className="font-mono text-[11px] tracking-wider uppercase">
          MERMAID DIAGRAM
        </span>
        <div className="flex items-center gap-1">
          <button
            type="button"
            aria-label="适应宽度"
            onClick={() => handleModeChange('fit')}
            className={`flex items-center gap-1 rounded px-2 py-0.5 font-mono text-[11px] cursor-pointer ${
              mode === 'fit'
                ? 'bg-surface font-semibold text-lagoon-deep border border-line/60'
                : 'text-sea-ink-soft hover:text-sea-ink'
            } ${transitionClass}`}
          >
            <MoveHorizontal className="h-3 w-3" />
            <span>适应宽度</span>
          </button>
          <button
            type="button"
            aria-label="原始比例"
            onClick={() => handleModeChange('actual')}
            className={`flex items-center gap-1 rounded px-2 py-0.5 font-mono text-[11px] cursor-pointer ${
              mode === 'actual'
                ? 'bg-surface font-semibold text-lagoon-deep border border-line/60'
                : 'text-sea-ink-soft hover:text-sea-ink'
            } ${transitionClass}`}
          >
            <ZoomIn className="h-3 w-3" />
            <span>原始比例</span>
          </button>
          <button
            ref={fullscreenButtonRef}
            type="button"
            aria-label="全屏查看"
            onClick={() => setFullscreen(true)}
            className={`flex items-center gap-1 rounded px-2 py-0.5 font-mono text-[11px] cursor-pointer text-sea-ink-soft hover:text-sea-ink ${transitionClass}`}
          >
            <Maximize2 className="h-3 w-3" />
            <span>全屏查看</span>
          </button>
        </div>
      </div>

      {/* SVG 容器 */}
      <div
        ref={containerRef}
        data-testid="mermaid-svg-container"
        className={`p-6 w-full ${mode === 'actual' ? 'overflow-x-auto' : ''}`}
      >
        {fullscreen ? (
          <div
            data-testid="mermaid-fullscreen-placeholder"
            className="flex h-32 w-full flex-col items-center justify-center border border-dashed border-line bg-surface/40 p-4 text-xs font-serif italic text-sea-ink-soft/70"
          >
            <span>图表已在全屏窗口中打开</span>
          </div>
        ) : (
          <div className="flex justify-center w-full">{children}</div>
        )}
      </div>

      {/* 全屏 Dialog */}
      <Dialog open={fullscreen} onOpenChange={handleDialogClose}>
        <DialogContent className="max-w-5xl max-h-[90vh] overflow-auto p-6 bg-surface-strong">
          <DialogHeader>
            <DialogTitle className="font-serif text-lg text-sea-ink">
              架构与时序图全屏查看
            </DialogTitle>
          </DialogHeader>
          <div className="mt-4 flex justify-center overflow-auto p-4 bg-surface/50 border border-line/40">
            {fullscreen && children}
          </div>
        </DialogContent>
      </Dialog>
    </div>
  )
}
