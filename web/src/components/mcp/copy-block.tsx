import { useState } from 'react'
import { Check, Copy, Terminal } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@lumina/components/ui/button'

interface CopyBlockProps {
  code: string
  filename?: string
  className?: string
}

export function CopyBlock({ code, filename, className }: CopyBlockProps) {
  const [copied, setCopied] = useState(false)

  const handleCopy = async () => {
    await navigator.clipboard.writeText(code)
    setCopied(true)
    toast.success('已复制到剪贴板')
    window.setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className={`overflow-hidden bg-sea-ink ${className ?? ''}`}>
      <div className="flex items-center justify-between gap-3 border-b border-sand/10 px-4 py-2">
        <div className="flex min-w-0 items-center gap-2">
          <Terminal className="size-3.5 shrink-0 text-lagoon" aria-hidden />
          <span className="truncate text-[11px] font-semibold uppercase tracking-wider text-lagoon">
            {filename ?? '配置'}
          </span>
        </div>
        <Button
          type="button"
          variant="ghost"
          size="sm"
          className="h-7 shrink-0 px-2 text-sand/70 hover:bg-transparent hover:text-lagoon"
          onClick={handleCopy}
        >
          {copied ? (
            <Check className="size-3.5" />
          ) : (
            <Copy className="size-3.5" />
          )}
          {copied ? '已复制' : '复制'}
        </Button>
      </div>
      <pre className="max-h-80 overflow-auto px-4 py-3">
        <code className="font-mono text-[13px] leading-relaxed whitespace-pre text-sand">
          {code}
        </code>
      </pre>
    </div>
  )
}
