import { Copy, AlertCircle } from 'lucide-react'
import type React from 'react'
import { useState } from 'react'
import { formatLocation } from './render-location'
import type { LpwRenderLocation } from './render-location'

export type LpwDiagnosticCode =
  | 'UNKNOWN_BLOCK'
  | 'INVALID_LAYOUT_SLOT'
  | 'INVALID_CONTAINER_CHILD'
  | 'DEPTH_LIMIT'
  | 'RENDER_EXCEPTION'
  | 'ANNOTATION_PATTERN'
  | 'RESOURCE_LIMIT'
  | 'UNSUPPORTED_VERSION'

export interface LpwDiagnostic {
  code: LpwDiagnosticCode
  location: LpwRenderLocation
  nodeId: string
  nodeType: string
  componentName?: string
  reason: string
  suggestion?: string
  propsSummary?: unknown
  stack?: string
}

const SENSITIVE_KEY_REGEX =
  /\b(token|secret|password|passwd|authorization|bearer|cookie|key|credential|private|cert|signature|salt|session|hash|jwt|nonce|access)\b/i

/**
 * 递归生成安全的 props 摘要对象
 */
export function summarizeProps(props: unknown): unknown {
  if (props === null || props === undefined) {
    return props
  }

  const seen = new WeakSet<object>()

  function sanitize(value: unknown, depth = 0): unknown {
    if (depth > 8) return '[...]'
    if (typeof value === 'string') {
      if (value.length > 240) {
        return `${value.slice(0, 240)}…（截断）`
      }
      return value
    }
    if (typeof value !== 'object' || value === null) {
      return value
    }

    if (seen.has(value)) {
      return '[Circular]'
    }
    seen.add(value)

    if (Array.isArray(value)) {
      return value.map((item) => sanitize(item, depth + 1))
    }

    const result: Record<string, unknown> = {}
    for (const [k, v] of Object.entries(value)) {
      if (SENSITIVE_KEY_REGEX.test(k)) {
        result[k] = '***'
      } else {
        result[k] = sanitize(v, depth + 1)
      }
    }
    return result
  }

  const sanitized = sanitize(props)
  try {
    const rawJson = JSON.stringify(sanitized, null, 2)
    if (rawJson && new TextEncoder().encode(rawJson).length > 2048) {
      // 整体截断
      return `${rawJson.slice(0, 1900)}…（截断）`
    }
  } catch {
    return '[Unserializable]'
  }

  return sanitized
}

export interface DiagnosticCardProps {
  diagnostic: LpwDiagnostic
  debug?: boolean
}

export const DiagnosticCard: React.FC<DiagnosticCardProps> = ({
  diagnostic,
  debug = false,
}) => {
  const [copied, setCopied] = useState(false)

  const handleCopy = async () => {
    const payload = { ...diagnostic }
    if (!debug) {
      delete payload.stack
    }
    try {
      await navigator.clipboard.writeText(JSON.stringify(payload, null, 2))
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      // 复制降级
    }
  }

  const summaryText =
    typeof diagnostic.propsSummary === 'string'
      ? diagnostic.propsSummary
      : JSON.stringify(diagnostic.propsSummary, null, 2)

  return (
    <div
      data-testid="diagnostic-card"
      className="my-6 border border-destructive/40 bg-destructive/5 p-4 text-xs font-sans text-destructive shadow-2xs"
    >
      <div className="flex flex-wrap items-center justify-between gap-2 border-b border-destructive/20 pb-2.5">
        <div className="flex items-center gap-2">
          <AlertCircle className="h-4 w-4 shrink-0 text-destructive" />
          <span
            data-testid="diagnostic-code"
            className="rounded bg-destructive/15 px-1.5 py-0.5 font-mono font-semibold text-[11px] tracking-wide text-destructive"
          >
            {diagnostic.code}
          </span>
          <span
            data-testid="diagnostic-location"
            className="font-mono text-[11px] text-sea-ink-soft/80"
          >
            {formatLocation(diagnostic.location)}
          </span>
        </div>
        <button
          type="button"
          onClick={handleCopy}
          aria-label="复制诊断信息"
          className="flex cursor-pointer items-center gap-1.5 rounded border border-destructive/30 bg-surface/80 px-2 py-0.5 font-mono text-[11px] text-sea-ink hover:bg-surface"
        >
          <Copy className="h-3 w-3" />
          <span>{copied ? '已复制' : '复制诊断'}</span>
        </button>
      </div>

      <div className="mt-3 space-y-1.5">
        <div className="flex items-center gap-2 font-mono text-[11px] text-sea-ink">
          <span className="font-semibold">
            {diagnostic.nodeType}#{diagnostic.nodeId}
          </span>
          {diagnostic.componentName && (
            <span className="text-sea-ink-soft">
              ({diagnostic.componentName})
            </span>
          )}
        </div>

        <p
          data-testid="diagnostic-reason"
          className="font-sans text-sm font-medium text-sea-ink"
        >
          {diagnostic.reason}
        </p>

        {diagnostic.suggestion && (
          <p
            data-testid="diagnostic-suggestion"
            className="text-[11px] text-sea-ink-soft/90"
          >
            建议：{diagnostic.suggestion}
          </p>
        )}
      </div>

      {diagnostic.propsSummary !== undefined && (
        <details className="mt-3 rounded border border-destructive/20 bg-surface/50 p-2 text-[11px]">
          <summary className="cursor-pointer font-mono text-sea-ink-soft select-none hover:text-sea-ink">
            属性摘要（已过滤敏感信息）
          </summary>
          <pre className="mt-2 max-h-48 overflow-auto font-mono text-[10px] text-sea-ink-soft whitespace-pre-wrap">
            {summaryText}
          </pre>
        </details>
      )}

      {debug && diagnostic.stack && (
        <details className="mt-2 rounded border border-destructive/20 bg-surface/50 p-2 text-[11px]">
          <summary className="cursor-pointer font-mono text-sea-ink-soft select-none hover:text-sea-ink">
            调用栈详情
          </summary>
          <pre className="mt-2 max-h-48 overflow-auto font-mono text-[10px] text-destructive/80 whitespace-pre-wrap">
            {diagnostic.stack}
          </pre>
        </details>
      )}
    </div>
  )
}
