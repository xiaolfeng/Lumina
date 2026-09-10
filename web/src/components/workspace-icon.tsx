import type { IconName } from 'lucide-react/dynamic'
import { DynamicIcon } from 'lucide-react/dynamic'
import { LayoutGrid } from 'lucide-react'
import { cn } from '#/lib/utils'
import { parseWorkspaceIcon } from './workspace-icon-utils'

export {
  createHiddenWorkspaceSlug,
  parseWorkspaceIcon,
  workspaceIconDisplayLabel,
} from './workspace-icon-utils'

function LucideFallback() {
  return <LayoutGrid className="size-4" aria-hidden />
}

interface WorkspaceIconProps {
  name?: string | null
  className?: string
  label?: string
}

export function WorkspaceIcon({
  name,
  className,
  label = '空间图标',
}: WorkspaceIconProps) {
  const parsed = parseWorkspaceIcon(name)

  return (
    <span
      role="img"
      aria-label={label}
      className={cn(
        'inline-flex size-4 shrink-0 items-center justify-center',
        className,
      )}
    >
      {parsed.kind === 'emoji' ? (
        <span aria-hidden className="text-base leading-none">
          {parsed.value}
        </span>
      ) : parsed.kind === 'lucide' ? (
        <DynamicIcon
          name={parsed.name as IconName}
          className="size-4"
          aria-hidden
          fallback={LucideFallback}
        />
      ) : (
        <LayoutGrid className="size-4" aria-hidden />
      )}
    </span>
  )
}
