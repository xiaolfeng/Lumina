import { useCallback, useEffect, useMemo, useState } from 'react'
import { ListTree, RefreshCw } from 'lucide-react'
import { Markdown, proseArticle, TableOfContents } from '../markdown'
import { parseFrontmatter } from './frontmatter'
import { MdxHeader } from './mdx-header'

export interface PreviewMdxViewerProps {
  src: string
  filename: string
  variant?: 'page' | 'inline'
}

export function PreviewMdxViewer({
  src,
  filename,
  variant = 'page',
}: PreviewMdxViewerProps) {
  const [source, setSource] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [fetchIndex, setFetchIndex] = useState(0)
  const [mobileTocOpen, setMobileTocOpen] = useState(false)

  const reload = useCallback(() => {
    setFetchIndex((i) => i + 1)
  }, [])

  useEffect(() => {
    const ac = new AbortController()
    setLoading(true)
    setError(null)

    void (async () => {
      try {
        const rawUrl = src.includes('?') ? `${src}&raw=1` : `${src}?raw=1`
        const res = await fetch(rawUrl, { signal: ac.signal })
        if (!res.ok) {
          throw new Error(`HTTP ${res.status}`)
        }
        const text = await res.text()
        if (!ac.signal.aborted) {
          setSource(text)
          setLoading(false)
        }
      } catch (err) {
        if (ac.signal.aborted) return
        setError(err instanceof Error ? err.message : '读取 MDX 文件失败')
        setLoading(false)
      }
    })()

    return () => ac.abort()
  }, [src, fetchIndex])

  const { frontmatter, cleanBody, fallbackTitle } = useMemo(() => {
    if (source === null) {
      return { frontmatter: null, cleanBody: '', fallbackTitle: filename }
    }
    const { frontmatter: fm, body } = parseFrontmatter(source)

    // 尝试从文件名推导备选标题（去除扩展名与路径）
    const baseName = filename.split(/[/\\]/).pop() ?? filename
    const nameWithoutExt = baseName.replace(/\.[^/.]+$/, '')

    // 检查正文开头是否存在重复的一级标题
    const trimmedBody = body.trimStart()
    const firstH1Match = trimmedBody.match(/^#\s+(.+)$/m)
    let bodyTitle: string | null = null
    let processedBody = body

    if (firstH1Match) {
      bodyTitle = firstH1Match[1].trim()
      // 若 frontmatter 明确提供了 title 且与正文首个 # 相同，或正文首行即为 # 标题，
      // 剥离该行以避免双重主标题显示
      if (
        fm?.title &&
        fm.title.trim().toLowerCase() === bodyTitle.toLowerCase()
      ) {
        processedBody = trimmedBody.replace(/^#\s+.+$/m, '').trimStart()
      }
    }

    const title = fm?.title || bodyTitle || nameWithoutExt

    return {
      frontmatter: fm,
      cleanBody: processedBody,
      fallbackTitle: title,
    }
  }, [source, filename])

  if (loading) {
    return (
      <div className="flex h-full min-h-80 w-full items-center justify-center bg-surface p-8">
        <p className="text-xs text-sea-ink-soft/60">加载 MDX 文档中…</p>
      </div>
    )
  }

  if (error || source === null) {
    return (
      <div className="flex h-full min-h-80 w-full flex-col items-center justify-center gap-3 bg-surface p-8 text-xs text-red-600">
        <p>加载失败: {error || '无法读取文件内容'}</p>
        <button
          type="button"
          onClick={reload}
          className="flex items-center gap-1.5 cursor-pointer border border-line bg-surface px-3 py-1.5 text-xs text-sea-ink hover:bg-surface-muted transition-colors shadow-xs"
        >
          <RefreshCw className="size-3.5" />
          <span>重试</span>
        </button>
      </div>
    )
  }

  if (variant === 'inline') {
    return (
      <div className="min-h-80 w-full overflow-y-auto border border-line bg-surface p-5 sm:p-6 text-sea-ink">
        <MdxHeader
          frontmatter={frontmatter}
          fallbackTitle={fallbackTitle}
        />
        <article className={`${proseArticle} max-w-full`}>
          <Markdown>{cleanBody}</Markdown>
        </article>
      </div>
    )
  }

  return (
    <div className="relative flex h-full w-full overflow-y-auto bg-surface">
      <div className="mx-auto flex w-full max-w-6xl justify-center gap-8 px-4 py-8 sm:px-8 lg:px-12">
        <div className="min-w-0 flex-1 max-w-3xl">
          <MdxHeader
            frontmatter={frontmatter}
            fallbackTitle={fallbackTitle}
          />
          <article className={`${proseArticle} max-w-none`}>
            <Markdown>{cleanBody}</Markdown>
          </article>
        </div>

        {/* 桌面端右侧固定目录 */}
        <aside className="hidden w-56 shrink-0 xl:block">
          <div className="sticky top-8 border-l border-line/60 pl-4">
            <div className="mb-2 flex items-center gap-1.5 text-[11px] font-semibold tracking-wider text-lagoon-deep uppercase">
              <ListTree className="size-3.5" />
              <span>目录大纲</span>
            </div>
            <TableOfContents content={cleanBody} />
          </div>
        </aside>

        {/* 移动端目录悬浮开关 */}
        <div className="fixed bottom-6 right-6 xl:hidden z-30">
          <button
            type="button"
            onClick={() => setMobileTocOpen((prev) => !prev)}
            className="flex size-10 items-center justify-center rounded-none border border-line bg-surface text-sea-ink shadow-md hover:bg-surface-muted cursor-pointer"
            aria-label="切换目录大纲"
          >
            <ListTree className="size-5" />
          </button>
          {mobileTocOpen && (
            <div className="absolute bottom-12 right-0 w-64 max-h-80 overflow-y-auto border border-line bg-surface p-4 shadow-xl text-xs">
              <p className="font-semibold text-lagoon-deep mb-2">目录大纲</p>
              <TableOfContents content={cleanBody} />
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

export function PreviewMdxInlineViewer(props: {
  src: string
  filename: string
}) {
  return <PreviewMdxViewer {...props} variant="inline" />
}
