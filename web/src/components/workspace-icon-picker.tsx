import { useMemo, useState } from 'react'
import { ChevronsUpDown } from 'lucide-react'
import { Button } from '@lumina/components/ui/button'
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@lumina/components/ui/command'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@lumina/components/ui/popover'
import { cn } from '#/lib/utils'
import { WorkspaceIcon } from './workspace-icon'
import {
  isEmojiValue,
  parseWorkspaceIcon,
  searchLucideIconNames,
  workspaceIconDisplayLabel,
  workspaceIconPickerItemLabel,
} from './workspace-icon-utils'

interface WorkspaceIconPickerProps {
  value: string
  onChange: (value: string) => void
  disabled?: boolean
  id?: string
}

export function WorkspaceIconPicker({
  value,
  onChange,
  disabled = false,
  id,
}: WorkspaceIconPickerProps) {
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState('')

  const emojiCandidate = useMemo(() => {
    const trimmed = query.trim()
    if (!trimmed) return null
    const parsed = parseWorkspaceIcon(trimmed)
    return parsed.kind === 'emoji' || isEmojiValue(trimmed)
      ? parsed.kind === 'emoji'
        ? parsed.value
        : trimmed
      : null
  }, [query])

  const lucideNames = useMemo(
    () => (emojiCandidate ? [] : searchLucideIconNames(query)),
    [emojiCandidate, query],
  )

  const displayLabel = workspaceIconDisplayLabel(value)

  function select(next: string) {
    onChange(next)
    setOpen(false)
    setQuery('')
  }

  return (
    <Popover
      open={open}
      onOpenChange={(next) => {
        if (disabled) return
        setOpen(next)
        if (!next) setQuery('')
      }}
    >
      <PopoverTrigger asChild>
        <Button
          id={id}
          type="button"
          variant="outline"
          role="combobox"
          aria-expanded={open}
          aria-label={`空间图标：${displayLabel}`}
          disabled={disabled}
          className="w-full justify-between"
        >
          <span className="flex min-w-0 items-center gap-2">
            <WorkspaceIcon name={value} label="空间图标预览" />
            <span className="truncate">{displayLabel}</span>
          </span>
          <ChevronsUpDown className="ml-2 size-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-72 p-0" align="start">
        <Command shouldFilter={false}>
          <CommandInput
            value={query}
            onValueChange={setQuery}
            placeholder="搜索图标名称或输入表情"
          />
          <CommandList>
            <CommandEmpty>没有匹配的图标</CommandEmpty>
            {emojiCandidate ? (
              <CommandGroup heading="表情">
                <CommandItem
                  value={emojiCandidate}
                  onSelect={() => select(emojiCandidate)}
                >
                  <WorkspaceIcon name={emojiCandidate} label="表情预览" />
                  使用该表情
                </CommandItem>
              </CommandGroup>
            ) : null}
            {lucideNames.length > 0 ? (
              <CommandGroup heading={query.trim() ? '搜索结果' : '常用图标'}>
                {lucideNames.map((name) => (
                  <CommandItem
                    key={name}
                    value={`${name} ${workspaceIconPickerItemLabel(name)}`}
                    onSelect={() => select(name)}
                  >
                    <WorkspaceIcon name={name} label="图标预览" />
                    <span>{workspaceIconPickerItemLabel(name)}</span>
                    <span
                      className={cn(
                        'ml-auto size-2 rounded-full bg-lagoon',
                        value === name ? 'opacity-100' : 'opacity-0',
                      )}
                      aria-hidden
                    />
                  </CommandItem>
                ))}
              </CommandGroup>
            ) : null}
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}
