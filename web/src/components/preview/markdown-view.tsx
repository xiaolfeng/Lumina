import { Markdown, proseArticle } from '@lumina/components/markdown'

export function PreviewMarkdownView({ source }: { source: string }) {
  return (
    <div className="min-h-0 flex-1 overflow-y-auto bg-surface px-6 py-6">
      <article className={`${proseArticle} mx-auto max-w-3xl`}>
        <Markdown>{source}</Markdown>
      </article>
    </div>
  )
}
