import { useCallback, useEffect, useState  } from 'react'
import type {FormEvent} from 'react';
import { createFileRoute } from '@tanstack/react-router'
import { CheckCircle2, ShieldCheck, XCircle } from 'lucide-react'
import { toast } from 'sonner'

import { Button } from '@lumina/components/ui/button'
import { Input } from '@lumina/components/ui/input'
import { Label } from '@lumina/components/ui/label'

import { useAuth, useLogin } from '#/hooks/useAuth'
import {
  getOAuthConsentDetail,
  postOAuthConsent
  
} from '#/lib/apis/oauth'
import type {OAuthConsentDetail} from '#/lib/apis/oauth';

interface OAuthSearch {
  authorize_id?: string
}

export const Route = createFileRoute('/_public/oauth')({
  validateSearch: (search: Record<string, unknown>): OAuthSearch => {
    return {
      authorize_id:
        typeof search.authorize_id === 'string'
          ? search.authorize_id
          : undefined,
    }
  },
  component: OAuthAuthorizePage,
})

function OAuthAuthorizePage() {
  const { authorize_id: authorizeId } = useOAuthSearch()
  const { isAuthenticated } = useAuth()
  const login = useLogin()

  const [detail, setDetail] = useState<OAuthConsentDetail | null>(null)
  const [loadError, setLoadError] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)
  const [account, setAccount] = useState('')
  const [password, setPassword] = useState('')

  const loadDetail = useCallback(async () => {
    if (!authorizeId) return
    try {
      const response = await getOAuthConsentDetail(authorizeId)
      setDetail(response.data ?? null)
      setLoadError(null)
    } catch {
      setLoadError('授权请求不存在或已过期，请回到客户端重新发起连接。')
    }
  }, [authorizeId])

  useEffect(() => {
    if (isAuthenticated) void loadDetail()
  }, [isAuthenticated, loadDetail])

  function submitLogin(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!account.trim() || !password) {
      toast.error('请输入账号与密码')
      return
    }
    login.mutate(
      { account: account.trim(), password },
      {
        onSuccess: () => void loadDetail(),
        onError: () => toast.error('登录失败，请检查账号与密码'),
      },
    )
  }

  async function decide(approve: boolean) {
    if (!authorizeId || busy) return
    setBusy(true)
    try {
      const response = await postOAuthConsent(authorizeId, approve)
      const redirect = response.data?.redirect
      if (!redirect) throw new Error('missing redirect')
      window.location.assign(redirect)
    } catch {
      toast.error('授权处理失败，请重试')
      setBusy(false)
    }
  }

  return (
    <div className="mx-auto flex min-h-[70vh] w-full max-w-lg flex-col justify-center px-4 py-12">
      <div className="mb-6 flex items-center gap-2 text-sea-ink">
        <ShieldCheck className="size-5 text-lagoon-deep" aria-hidden />
        <span className="text-sm font-semibold tracking-wide">
          Lumina · MCP 授权
        </span>
      </div>

      <div className="border border-line bg-sand p-6 sm:p-8">
        {!authorizeId ? (
          <Notice text="缺少授权请求参数。请回到客户端重新添加 MCP 地址并发起连接。" />
        ) : loadError ? (
          <Notice text={loadError} />
        ) : !isAuthenticated || !detail ? (
          <form onSubmit={submitLogin} className="space-y-4">
            <div>
              <h1 className="display-title text-[22px] font-medium text-sea-ink">
                登录 Lumina
              </h1>
              <p className="mt-2 text-[13px] leading-relaxed text-sea-ink-soft">
                客户端正在请求 MCP 访问授权，请先用控制台账号登录以继续。
              </p>
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="oauth-account">账号</Label>
              <Input
                id="oauth-account"
                value={account}
                onChange={(event) => setAccount(event.target.value)}
                placeholder="用户名或邮箱"
                autoComplete="username"
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="oauth-password">密码</Label>
              <Input
                id="oauth-password"
                type="password"
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                placeholder="登录密码"
                autoComplete="current-password"
              />
            </div>
            <Button type="submit" className="w-full" disabled={login.isPending}>
              {login.isPending ? '登录中…' : '登录并继续授权'}
            </Button>
          </form>
        ) : (
          <div className="space-y-5">
            <div>
              <h1 className="display-title text-[22px] font-medium text-sea-ink">
                确认授权
              </h1>
              <p className="mt-2 text-[13px] leading-relaxed text-sea-ink-soft">
                <span className="font-semibold text-sea-ink">
                  {detail.client_name || 'MCP 客户端'}
                </span>{' '}
                请求访问你的 Lumina MCP 工具（作用域：
                <code className="mx-0.5">{detail.scope}</code>
                ）。授权后客户端将获得访问令牌并自动续期。
              </p>
            </div>
            <div className="grid gap-2 sm:grid-cols-2">
              <Button onClick={() => void decide(true)} disabled={busy}>
                <CheckCircle2 className="mr-1 size-4" aria-hidden />
                同意授权
              </Button>
              <Button
                variant="outline"
                onClick={() => void decide(false)}
                disabled={busy}
              >
                <XCircle className="mr-1 size-4" aria-hidden />
                拒绝
              </Button>
            </div>
            <p className="text-xs leading-relaxed text-sea-ink-soft">
              点击后页面将跳回客户端回调地址。拒绝会通知客户端授权被取消。
            </p>
          </div>
        )}
      </div>
    </div>
  )
}

function useOAuthSearch(): OAuthSearch {
  const search = Route.useSearch()
  return search
}

function Notice({ text }: { text: string }) {
  return (
    <div className="flex items-start gap-2 text-[13px] leading-relaxed text-sea-ink-soft">
      <XCircle
        className="mt-0.5 size-4 shrink-0 text-lagoon-deep"
        aria-hidden
      />
      <span>{text}</span>
    </div>
  )
}
