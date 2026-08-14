import * as React from "react"

import { cn } from "../lib/utils.ts"

/* ────────────────────────────────────────────────────────────
   Splitter — 微明主题可拖拽分割栏
   组合式 API：<Splitter> 容器 + <SplitterPanel> 面板 + <SplitterHandle> 分割条
   支持 horizontal / vertical 双轴拖拽，Keyboard 微调，minSize 钳制。
   ──────────────────────────────────────────────────────────── */

type Direction = "horizontal" | "vertical"

/* ── 常量 ── */
const MIN_SIZE = 12 // 面板最小占比（%）
const KEY_STEP = 1 // 键盘单次步进（%）
const KEY_STEP_FAST = 5 // Shift + 方向键步进（%）
const HANDLE_SIZE_VAR = "var(--splitter-hit-size, 14px)" // 命中区宽度（CSS 变量驱动）

/* ── 类型 ── */
type SplitterProps = React.ComponentProps<"div"> & {
  direction?: Direction
}

type SplitterPanelProps = React.ComponentProps<"div"> & {
  /** 初始占比（%），默认均分 */
  defaultSize?: number
  /** 最小占比（%），默认 12 */
  minSize?: number
}

type SplitterHandleProps = React.ComponentProps<"div"> & {
  /** 注入：该 handle 前面的 panel 下标 */
  index?: number
  /** 注入：拖拽开始 / 结束回调（联动 body 光标） */
  onDragStart?: () => void
  onDragEnd?: () => void
  /** 注入：比例更新回调 */
  onResize?: (index: number, clientX: number, clientY: number) => void
  /** 注入：键盘微调回调 */
  onKeyResize?: (index: number, delta: number) => void
}

/* ────────────────────────────────────────────────
   Splitter — 容器
   ──────────────────────────────────────────────── */
function Splitter({ direction = "horizontal", className, children, ...props }: SplitterProps) {
  const containerRef = React.useRef<HTMLDivElement>(null)

  // 拆分 children：识别 panel 与 handle（按 JSX 顺序）
  const panels = React.useMemo(
    () =>
      React.Children.toArray(children).filter(
        (child): child is React.ReactElement<SplitterPanelProps> =>
          React.isValidElement(child) && child.type === SplitterPanel
      ),
    [children]
  )

  // 面板占比状态（%）；最后一个面板由 1fr 吸收，不占数值
  const [sizes, setSizes] = React.useState<number[]>(() =>
    panels.slice(0, -1).map((p) => p.props.defaultSize ?? 100 / panels.length)
  )

  // panel 数量变化时同步默认占比（拖拽中不重置；跳过首屏，避免无效渲染）
  const prevPanelCount = React.useRef(panels.length)
  React.useEffect(() => {
    if (prevPanelCount.current === panels.length) return
    prevPanelCount.current = panels.length
    setSizes(panels.slice(0, -1).map((p) => p.props.defaultSize ?? 100 / panels.length))
  }, [panels.length])

  // 收集 minSize（%）
  const minSizes = React.useMemo(
    () => panels.slice(0, -1).map((p) => p.props.minSize ?? MIN_SIZE),
    [panels]
  )

  /* ── 拖拽更新：计算 handle 前方面板的新占比 ── */
  const handleResize = React.useCallback(
    (index: number, clientX: number, clientY: number) => {
      const rect = containerRef.current?.getBoundingClientRect()
      if (!rect || index < 0 || index >= sizes.length) return

      const total =
        direction === "horizontal"
          ? (clientX - rect.left) / rect.width
          : (clientY - rect.top) / rect.height
      const target = total * 100

      setSizes((prev) => {
        // 该 handle 之前所有面板已占的比例
        const before = prev.slice(0, index).reduce((sum, s) => sum + s, 0)
        // 为后续面板保留最小空间
        const reserve = minSizes.slice(index + 1).reduce((sum, s) => sum + s, 0)
        const min = minSizes[index] ?? MIN_SIZE
        const max = Math.max(min, 100 - before - reserve)

        const nextSizes = [...prev]
        nextSizes[index] = Math.min(max, Math.max(min, target - before))
        return nextSizes
      })
    },
    [direction, sizes.length, minSizes]
  )

  const handleKeyResize = React.useCallback(
    (index: number, delta: number) => {
      if (index < 0 || index >= sizes.length) return
      setSizes((prev) => {
        const nextSizes = [...prev]
        const before = prev.slice(0, index).reduce((sum, s) => sum + s, 0)
        const reserve = minSizes.slice(index + 1).reduce((sum, s) => sum + s, 0)
        const min = minSizes[index] ?? MIN_SIZE
        const max = Math.max(min, 100 - before - reserve)
        nextSizes[index] = Math.min(max, Math.max(min, prev[index] + delta))
        return nextSizes
      })
    },
    [sizes.length, minSizes]
  )

  /* ── 组装 Grid 模板：p0 | handle | p1 | handle | … | 1fr ── */
  const template = React.useMemo(() => {
    const tracks: string[] = []
    let panelIndex = 0
    React.Children.forEach(children, (child) => {
      if (!React.isValidElement(child)) return
      if (child.type === SplitterPanel) {
        if (panelIndex < sizes.length) {
          tracks.push(`${sizes[panelIndex]}%`)
        } else {
          tracks.push("1fr") // 最后一个面板吸收剩余
        }
        panelIndex++
      } else if (child.type === SplitterHandle) {
        tracks.push(HANDLE_SIZE_VAR)
      }
    })
    return tracks.join(" ")
  }, [children, sizes])

  // 拖拽中联动 body 光标，避免指针飘出分割条时光标回跳
  const isVertical = direction === "vertical"
  const bodyCursor = isVertical ? "row-resize" : "col-resize"
  const onDragStart = React.useCallback(() => {
    document.body.style.cursor = bodyCursor
  }, [bodyCursor])
  const onDragEnd = React.useCallback(() => {
    document.body.style.cursor = ""
  }, [])

  // 计数：当前遍历到的 handle 下标（避免重复遍历 children）
  let handleSeq = 0

  return (
    <div
      ref={containerRef}
      data-slot="splitter"
      data-direction={direction}
      className={cn(
        "grid min-h-0 min-w-0",
        isVertical
          ? "grid-rows-[var(--splitter-track,1fr)]"
          : "grid-cols-[var(--splitter-track,1fr)]",
        className
      )}
      style={{ "--splitter-track": template } as React.CSSProperties}
      {...props}
    >
      {React.Children.map(children, (child) => {
        if (!React.isValidElement(child)) return child
        // 注入 handle 交互 props（保持用户 className / children）
        if (child.type === SplitterHandle) {
          return React.cloneElement(child, {
            index: handleSeq++,
            direction,
            onDragStart,
            onDragEnd,
            onResize: handleResize,
            onKeyResize: handleKeyResize,
          } as Partial<SplitterHandleProps>)
        }
        return child
      })}
    </div>
  )
}

/* ────────────────────────────────────────────────
   SplitterPanel — 面板
   ──────────────────────────────────────────────── */
function SplitterPanel({ className, children, ...props }: SplitterPanelProps) {
  return (
    <div
      data-slot="splitter-panel"
      className={cn("min-h-0 min-w-0 overflow-hidden", className)}
      {...props}
    >
      {children}
    </div>
  )
}

/* ────────────────────────────────────────────────
   SplitterHandle — 分割条（Pointer Events + Keyboard）
   ──────────────────────────────────────────────── */
function SplitterHandle({
  index = 0,
  direction = "horizontal",
  onDragStart,
  onDragEnd,
  onResize,
  onKeyResize,
  className,
  ...props
}: SplitterHandleProps & { direction?: Direction }) {
  const isVertical = direction === "vertical"
  const elRef = React.useRef<HTMLDivElement>(null)
  const dragging = React.useRef(false)

  const setDragging = (on: boolean) => {
    dragging.current = on
    const el = elRef.current
    if (!el) return
    if (on) el.setAttribute("data-dragging", "true")
    else el.removeAttribute("data-dragging")
  }

  const handlePointerDown = (e: React.PointerEvent) => {
    if (e.pointerType === "touch" || e.pointerType === "pen") e.preventDefault()
    setDragging(true)
    onDragStart?.()
    // 捕获指针：拖出分割条后仍持续接收事件
    e.currentTarget.setPointerCapture(e.pointerId)
    // 立即同步一次，避免 click 与拖拽起始抖动
    if (onResize) onResize(index, e.clientX, e.clientY)
  }

  const handlePointerMove = (e: React.PointerEvent) => {
    // 门控：仅在按下拖拽中生效（未按下时 pointermove 也会触发）
    if (!dragging.current || !onResize) return
    e.preventDefault()
    onResize(index, e.clientX, e.clientY)
  }

  const endDrag = () => {
    setDragging(false)
    onDragEnd?.()
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (!onKeyResize) return
    let delta: number | null = null
    const step = e.shiftKey ? KEY_STEP_FAST : KEY_STEP
    if (isVertical) {
      if (e.key === "ArrowUp") delta = -step
      else if (e.key === "ArrowDown") delta = step
    } else {
      if (e.key === "ArrowLeft") delta = -step
      else if (e.key === "ArrowRight") delta = step
    }
    if (delta === null) return
    e.preventDefault()
    onKeyResize(index, delta)
  }

  return (
    <div
      ref={elRef}
      data-slot="splitter-handle"
      data-direction={direction}
      role="separator"
      aria-orientation={isVertical ? "horizontal" : "vertical"}
      aria-label="调整面板比例"
      tabIndex={0}
      onPointerDown={handlePointerDown}
      onPointerMove={handlePointerMove}
      onPointerUp={endDrag}
      onPointerCancel={endDrag}
      onKeyDown={handleKeyDown}
      className={cn(
        "relative flex shrink-0 touch-none items-center justify-center outline-none select-none",
        "focus-visible:ring-2 focus-visible:ring-ring/50",
        "before:pointer-events-none before:absolute before:bg-border before:transition-colors before:duration-200",
        "hover:before:bg-ring/60 focus-visible:before:bg-ring/60 active:before:bg-ring data-[dragging=true]:before:bg-ring",
        // 分割条形态：X 轴竖条，Y 轴横条
        isVertical
          ? "h-full w-full cursor-row-resize before:h-px before:w-full before:left-0 before:right-0 before:top-1/2 before:-translate-y-1/2"
          : "w-full cursor-col-resize before:w-px before:h-full before:top-0 before:bottom-0 before:left-1/2 before:-translate-x-1/2",
        // grip 圆点：烛光的圆，只给状态点
        "after:pointer-events-none after:absolute after:size-2 after:rounded-full after:border after:border-ring after:bg-background",
        "after:opacity-0 hover:after:opacity-100 focus-visible:after:opacity-100 active:after:opacity-100 data-[dragging=true]:after:opacity-100",
        className
      )}
      {...props}
    >
      {/* 触控笔/触屏下放大命中区 */}
      <span aria-hidden="true" className="absolute inset-0 -m-1" />
    </div>
  )
}

export { Splitter, SplitterPanel, SplitterHandle }
export type { Direction, SplitterProps, SplitterPanelProps, SplitterHandleProps }
