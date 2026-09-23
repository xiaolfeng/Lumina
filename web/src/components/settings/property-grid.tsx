import type { ReactNode } from 'react'
import { Button } from '@lumina/components/ui/button'

interface PropertyRowProps {
  title: string
  propKey: string
  description?: string
  children: ReactNode
}

export function PropertyRow({
  title,
  propKey,
  description,
  children,
}: PropertyRowProps) {
  return (
    <div className="grid grid-cols-1 md:grid-cols-[320px_1fr] border-b border-line last:border-b-0 hover:bg-sand/40 transition-colors">
      <div className="p-4 px-5 border-b md:border-b-0 md:border-r border-line bg-sand/60 flex flex-col gap-1">
        <span className="text-xs font-semibold text-sea-ink">{title}</span>
        <span className="font-mono text-[10px] text-lagoon-deep select-all">
          {propKey}
        </span>
        {description ? (
          <p className="text-[11.5px] text-sea-ink-soft leading-relaxed mt-0.5">
            {description}
          </p>
        ) : null}
      </div>
      <div className="p-4 px-6 flex items-center">{children}</div>
    </div>
  )
}

interface PropertyGridProps {
  children: ReactNode
  className?: string
}

export function PropertyGrid({ children, className = '' }: PropertyGridProps) {
  return (
    <div className={`border border-line bg-foam ${className}`}>{children}</div>
  )
}

interface SettingsCommitDockProps {
  dirtyCount: number
  isSubmitting?: boolean
  onReset: () => void
  disabled?: boolean
}

export function SettingsCommitDock({
  dirtyCount,
  isSubmitting,
  onReset,
  disabled,
}: SettingsCommitDockProps) {
  const isDirty = dirtyCount > 0

  return (
    <div className="sticky bottom-0 -mx-4 -mb-4 mt-6 border-t border-line bg-surface-strong/95 p-3.5 px-6 backdrop-blur-md flex items-center justify-between z-30">
      <div className="flex items-center gap-3">
        {isDirty ? (
          <span className="font-mono text-xs font-bold text-lagoon-deep bg-[#fff3e6] border border-lagoon/40 px-2.5 py-0.5">
            已修改 {dirtyCount} 项待保存
          </span>
        ) : (
          <span className="text-xs text-sea-ink-soft">
            所有配置已与服务端保持同步
          </span>
        )}
      </div>
      <div className="flex items-center gap-2">
        {isDirty && (
          <Button
            type="button"
            variant="outline"
            onClick={onReset}
            disabled={isSubmitting}
            className="rounded-none border-line text-xs hover:bg-chip-bg text-sea-ink"
          >
            放弃修改
          </Button>
        )}
        <Button
          type="submit"
          disabled={!isDirty || isSubmitting || disabled}
          className="rounded-none bg-lagoon text-foam hover:bg-lagoon-deep text-xs font-semibold"
        >
          {isSubmitting ? '保存中…' : '保存变更'}
        </Button>
      </div>
    </div>
  )
}
