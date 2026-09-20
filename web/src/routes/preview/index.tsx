import { createFileRoute, redirect } from '@tanstack/react-router'

export const Route = createFileRoute('/preview/')({
  validateSearch: (search: Record<string, unknown>) => ({
    session: typeof search.session === 'string' ? search.session : undefined,
    file: typeof search.file === 'string' ? search.file : undefined,
  }),
  beforeLoad: ({ search }) => {
    if (search.session) {
      throw redirect({
        to: '/preview/$sessionHash/$',
        params: {
          sessionHash: search.session,
          _splat: search.file || '',
        },
      })
    }
  },
  component: PreviewLegacyPage,
})

function PreviewLegacyPage() {
  return (
    <div className="flex flex-1 items-center justify-center">
      <p className="text-sm text-sea-ink-soft">缺少预览会话路径</p>
    </div>
  )
}
