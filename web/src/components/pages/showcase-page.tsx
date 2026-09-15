import { isAxiosError } from 'axios'
import { PagesPasswordGate } from '#/components/pages/password-gate'
import { ShowcaseShell } from '#/components/pages/showcase-shell'
import { usePageAuthCheck, usePageMeta, usePageVersions } from '#/hooks/usePages'

// client.ts 拦截器会把带 error_message 的错误重包成裸 Error（HTTP status 丢失），
// 仅透传 AxiosError 时可辨认 404；拿不到 status 时一律按可重试处理，避免误报「页面不存在」
function isHttpNotFound(error: unknown): boolean {
  return isAxiosError(error) && error.response?.status === 404
}

function ShowcaseNotFound() {
  return (
    <div className="flex min-h-screen items-center justify-center text-sm text-red-500">
      页面不存在
    </div>
  )
}

function ShowcaseErrorState({
  message,
  onRetry,
}: {
  message: string
  onRetry: () => void
}) {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-3 text-sm text-red-500">
      <p>{message}</p>
      <button
        type="button"
        className="text-sea-ink-soft underline underline-offset-4"
        onClick={onRetry}
      >
        重试
      </button>
    </div>
  )
}

export function PagesShowcasePage({
  projectName,
  slug,
  filepath,
}: {
  projectName: string
  slug: string
  /** URL 携带的文件路径；为空时由版本入口文件兜底 */
  filepath: string
}) {
  const auth = usePageAuthCheck(projectName, slug)
  const authData = auth.data?.data
  const unlocked = authData
    ? !authData.password_required || Boolean(authData.authenticated)
    : false
  const meta = usePageMeta(projectName, slug, undefined, unlocked)
  const versions = usePageVersions(meta.data?.data?.page.id)

  if (auth.isError) {
    if (isHttpNotFound(auth.error)) {
      return <ShowcaseNotFound />
    }
    return (
      <ShowcaseErrorState
        message={auth.error.message || '认证信息加载失败，请稍后重试'}
        onRetry={() => void auth.refetch()}
      />
    )
  }
  if (auth.isLoading || meta.isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center text-sm text-sea-ink-soft">
        加载中…
      </div>
    )
  }
  if (auth.data?.data?.password_required && !auth.data.data.authenticated) {
    return <PagesPasswordGate projectName={projectName} slug={slug} />
  }
  if (meta.isError) {
    if (isHttpNotFound(meta.error)) {
      return <ShowcaseNotFound />
    }
    return (
      <ShowcaseErrorState
        message={meta.error.message || '页面信息加载失败，请稍后重试'}
        onRetry={() => void meta.refetch()}
      />
    )
  }
  if (!meta.data?.data) {
    return <ShowcaseNotFound />
  }

  return (
    <ShowcaseShell
      projectName={projectName}
      slug={slug}
      filepath={filepath}
      page={meta.data.data.page}
      version={meta.data.data.version}
      files={meta.data.data.files}
      versions={versions.data?.data?.items ?? [meta.data.data.version]}
    />
  )
}
