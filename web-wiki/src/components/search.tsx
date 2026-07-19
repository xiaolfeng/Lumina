import { useState, useEffect, useCallback, useRef, Suspense, lazy } from 'react'
import { Link } from '@tanstack/react-router'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@lumina/components/ui/dialog'
import { SearchIcon, FileText } from 'lucide-react'
import type { Orama } from '@orama/orama'
import { wikiReaderApi } from '#/lib/api-client'
import type { PageNode } from '#/lib/source'

// ── Types ──

interface SearchResult {
  id: string
  title: string
  description?: string
  path: string
}

interface SearchIndexEntry {
  title: string
  description: string
  path: string
  content: string
}

type AnyOrama = Orama<any>

// ── Index Cache ──

const indexCache = new Map<string, AnyOrama>()
const indexBuilding = new Map<string, boolean>()

// ── Concurrency Limiter (hand-written semaphore) ──

async function withConcurrencyLimit<T>(
  items: T[],
  limit: number,
  fn: (item: T) => Promise<unknown>,
): Promise<void> {
  let index = 0
  const workers = Array.from({ length: limit }, async () => {
    while (index < items.length) {
      const i = index++
      if (i >= items.length) break
      await fn(items[i])
    }
  })
  await Promise.all(workers)
}

// ── Index Builder ──

async function buildSearchIndex(wikiId: string, leaves: PageNode[]): Promise<AnyOrama> {
  const cached = indexCache.get(wikiId)
  if (cached) return cached

  // Dynamically import Orama to avoid bundling in main chunk
  const { create, insert } = await import('@orama/orama')
  const { createTokenizer } = await import('@orama/tokenizers/mandarin')

  const db = await create({
    schema: {
      title: 'string',
      description: 'string',
      path: 'string',
      content: 'string',
    } as const,
    components: {
      tokenizer: await createTokenizer(),
    },
  })

  const entries: SearchIndexEntry[] = []

  await withConcurrencyLimit(leaves, 6, async (leaf) => {
    try {
      const page = await wikiReaderApi.getPage(wikiId, leaf.path)
      entries.push({
        title: page.title || leaf.title || leaf.path,
        description: page.description || leaf.description || '',
        path: leaf.path,
        content: page.content || '',
      })
    } catch {
      // Fallback to manifest data if page fetch fails
      entries.push({
        title: leaf.title || leaf.path,
        description: leaf.description || '',
        path: leaf.path,
        content: '',
      })
    }
  })

  for (const entry of entries) {
    await insert(db, entry)
  }

  indexCache.set(wikiId, db)
  return db
}

// ── Search Hook ──

function useSearchIndex(wikiId: string, leaves: PageNode[]) {
  const [index, setIndex] = useState<AnyOrama | null>(null)
  const [loading, setLoading] = useState(false)
  const builtRef = useRef(false)

  useEffect(() => {
    if (!wikiId || leaves.length === 0 || builtRef.current) return
    builtRef.current = true

    const cached = indexCache.get(wikiId)
    if (cached) {
      setIndex(cached)
      return
    }

    if (indexBuilding.get(wikiId)) return
    indexBuilding.set(wikiId, true)

    setLoading(true)

    const doBuild = async () => {
      try {
        const db = await buildSearchIndex(wikiId, leaves)
        setIndex(db)
      } finally {
        setLoading(false)
        indexBuilding.set(wikiId, false)
      }
    }

    // Non-blocking: use requestIdleCallback or setTimeout(0) fallback
    if (typeof window !== 'undefined' && 'requestIdleCallback' in window) {
      const id = window.requestIdleCallback(() => {
        doBuild()
      })
      return () => {
        window.cancelIdleCallback(id)
      }
    } else {
      const timer = setTimeout(() => {
        doBuild()
      }, 0)
      return () => clearTimeout(timer)
    }
  }, [wikiId, leaves])

  return { index, loading }
}

// ── Search Dialog Component ──

interface SearchDialogProps {
  wikiId: string
  leaves: PageNode[]
  open: boolean
  onOpenChange: (open: boolean) => void
}

function SearchDialog({ wikiId, leaves, open, onOpenChange }: SearchDialogProps) {
  const { index } = useSearchIndex(wikiId, leaves)
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<SearchResult[]>([])
  const [selectedIndex, setSelectedIndex] = useState(0)
  const inputRef = useRef<HTMLInputElement>(null)
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  // Focus input when dialog opens
  useEffect(() => {
    if (open) {
      setTimeout(() => inputRef.current?.focus(), 50)
      setQuery('')
      setResults([])
      setSelectedIndex(0)
    }
  }, [open])

  // Debounced search
  useEffect(() => {
    if (debounceRef.current) {
      clearTimeout(debounceRef.current)
    }

    if (!query.trim() || !index) {
      setResults([])
      setSelectedIndex(0)
      return
    }

    debounceRef.current = setTimeout(async () => {
      const { search: oramaSearch } = await import('@orama/orama')
      const searchResults = await oramaSearch(index, {
        term: query,
        limit: 10,
        boost: {
          title: 2,
          description: 1.5,
        },
      })

      const mapped: SearchResult[] = searchResults.hits.map((hit: any) => ({
        id: String(hit.id),
        title: String(hit.document.title || ''),
        description: String(hit.document.description || ''),
        path: String(hit.document.path || ''),
      }))

      setResults(mapped)
      setSelectedIndex(0)
    }, 200)

    return () => {
      if (debounceRef.current) {
        clearTimeout(debounceRef.current)
      }
    }
  }, [query, index])

  // Keyboard navigation
  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === 'ArrowDown') {
        e.preventDefault()
        setSelectedIndex((prev) => (prev + 1) % Math.max(results.length, 1))
      } else if (e.key === 'ArrowUp') {
        e.preventDefault()
        setSelectedIndex((prev) =>
          prev <= 0 ? Math.max(results.length - 1, 0) : prev - 1,
        )
      } else if (e.key === 'Enter' && results[selectedIndex]) {
        e.preventDefault()
        onOpenChange(false)
      }
    },
    [results, selectedIndex, onOpenChange],
  )

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        className="overflow-hidden p-0 sm:max-w-lg"
        showCloseButton={false}
      >
        <DialogHeader className="sr-only">
          <DialogTitle>搜索 Wiki</DialogTitle>
        </DialogHeader>
        <div className="flex flex-col">
          {/* Search Input */}
          <div className="flex items-center gap-2 border-b px-3 py-3">
            <SearchIcon className="h-4 w-4 shrink-0 opacity-50" />
            <input
              ref={inputRef}
              type="text"
              placeholder="搜索页面..."
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              onKeyDown={handleKeyDown}
              className="flex h-10 w-full rounded-md bg-transparent text-sm outline-none placeholder:text-muted-foreground"
            />
            <kbd className="hidden rounded border bg-muted px-1.5 py-0.5 text-xs font-mono text-muted-foreground sm:inline-block">
              ESC
            </kbd>
          </div>

          {/* Results */}
          <div className="max-h-[300px] overflow-y-auto">
            {results.length === 0 && query.trim() && index && (
              <div className="py-6 text-center text-sm text-muted-foreground">
                未找到结果
              </div>
            )}
            {results.length === 0 && !query.trim() && (
              <div className="py-6 text-center text-sm text-muted-foreground">
                输入关键词搜索页面...
              </div>
            )}
            {results.map((result, idx) => (
              <Link
                key={result.id}
                to="/wiki/$wikiId/$"
                params={{ wikiId, _splat: result.path }}
                onClick={() => onOpenChange(false)}
                className={`flex items-start gap-3 px-4 py-3 text-sm transition-colors hover:bg-accent ${
                  idx === selectedIndex ? 'bg-accent' : ''
                }`}
              >
                <FileText className="mt-0.5 h-4 w-4 shrink-0 text-muted-foreground" />
                <div className="flex flex-col gap-0.5">
                  <span className="font-medium">{result.title}</span>
                  {result.description && (
                    <span className="text-xs text-muted-foreground line-clamp-1">
                      {result.description}
                    </span>
                  )}
                  <span className="text-xs text-muted-foreground/60">
                    {result.path}
                  </span>
                </div>
              </Link>
            ))}
          </div>

          {/* Footer */}
          <div className="flex items-center justify-between border-t px-4 py-2 text-xs text-muted-foreground">
            <div className="flex items-center gap-2">
              <span>
                {index ? `${results.length} 个结果` : '索引构建中...'}
              </span>
            </div>
            <div className="flex items-center gap-1">
              <kbd className="rounded border bg-muted px-1.5 py-0.5 text-xs font-mono">
                ↑
              </kbd>
              <kbd className="rounded border bg-muted px-1.5 py-0.5 text-xs font-mono">
                ↓
              </kbd>
              <span>导航</span>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}

// ── Lazy-loaded Search Component ──

const LazySearchDialog = lazy(() =>
  Promise.resolve({ default: SearchDialog }),
)

// ── Search Trigger Component ──

interface WikiSearchProps {
  wikiId: string
  leaves: PageNode[]
}

export function WikiSearch({ wikiId, leaves }: WikiSearchProps) {
  const [open, setOpen] = useState(false)

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault()
        setOpen(true)
      }
    }
    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [])

  return (
    <>
      <button
        onClick={() => setOpen(true)}
        className="flex items-center gap-2 rounded-md border bg-background px-3 py-1.5 text-sm text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
        type="button"
      >
        <SearchIcon className="h-4 w-4" />
        <span className="hidden sm:inline">搜索</span>
        <kbd className="ml-2 hidden rounded border bg-muted px-1.5 py-0.5 text-xs font-mono text-muted-foreground lg:inline-block">
          ⌘K
        </kbd>
      </button>
      <Suspense fallback={null}>
        <LazySearchDialog
          wikiId={wikiId}
          leaves={leaves}
          open={open}
          onOpenChange={setOpen}
        />
      </Suspense>
    </>
  )
}

export { SearchDialog, LazySearchDialog }
export type { SearchResult, SearchIndexEntry }
