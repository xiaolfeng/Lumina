import { ExternalLink } from 'lucide-react'
import type React from 'react'
import type { LpwBlockSlotProps, LpwCardsProps } from '../types'

// 协议白名单（S-01）：仅放行 http(s)/mailto 绝对地址与纯相对引用（/、./、../、# 开头
// 或不含冒号的同会话文件名）。浏览器解析 URL 时会剔除 scheme 内的 TAB/LF/CR 等
// 控制字符，黑名单前缀匹配可被穿透，因此含任何控制字符直接整体拒绝。
// eslint-disable-next-line no-control-regex -- 需精确匹配 C0 控制字符与 DEL 以拒绝恶意 scheme
const CONTROL_CHARS = /[\u0000-\u001f\u007f]/

function sanitizeHref(href?: string): string | undefined {
  if (!href) return undefined
  const value = href.trim()
  if (CONTROL_CHARS.test(value)) return undefined
  if (/^(https?:|mailto:)/i.test(value)) return value
  if (/^\/(?!\/)/.test(value) || /^[?#]/.test(value)) return value
  if (/^\.\.?\//.test(value)) return value
  if (!value.includes(':') && !value.startsWith('//')) return value
  return undefined
}

export const CardsBlock: React.FC<LpwBlockSlotProps<LpwCardsProps>> = ({
  props,
}) => {
  return (
    <div className="my-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-3 font-sans">
      {props.items.map((item, idx) => {
        const safeHref = sanitizeHref(item.href)
        const isExternal = Boolean(safeHref && /^https?:\/\//i.test(safeHref))
        const content = (
          <>
            <div className="font-serif font-semibold text-sm text-sea-ink tracking-tight flex items-center justify-between group-hover:text-lagoon transition-colors gap-2">
              <span className="min-w-0 break-words">{item.title}</span>
              {safeHref && (
                <span
                  aria-hidden="true"
                  className="font-mono text-xs text-sea-ink-soft/60 group-hover:translate-x-0.5 transition-transform shrink-0"
                >
                  <ExternalLink className="h-3.5 w-3.5" />
                </span>
              )}
            </div>
            {item.description && (
              <p className="mt-2 text-xs text-sea-ink-soft leading-relaxed font-sans">
                {item.description}
              </p>
            )}
          </>
        )

        if (safeHref) {
          const externalProps = isExternal
            ? { target: '_blank', rel: 'noopener noreferrer' }
            : {}

          return (
            <a
              key={idx}
              href={safeHref}
              {...externalProps}
              className="block border border-line bg-surface p-5 transition-all duration-150 hover:border-sea-ink/60 hover:bg-surface-muted/40 cursor-pointer shadow-2xs group"
            >
              {content}
            </a>
          )
        }

        return (
          <div
            key={idx}
            className="block border border-line bg-surface p-5 shadow-2xs"
          >
            {content}
          </div>
        )
      })}
    </div>
  )
}
