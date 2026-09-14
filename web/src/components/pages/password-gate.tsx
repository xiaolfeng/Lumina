import { useState } from 'react'
import { Lock } from 'lucide-react'
import { Button } from '@lumina/components/ui/button'
import { Input } from '@lumina/components/ui/input'
import { useUnlockPage } from '#/hooks/usePages'

export function PagesPasswordGate({
  projectName,
  slug,
}: {
  projectName: string
  slug: string
}) {
  const [password, setPassword] = useState('')
  const unlock = useUnlockPage()

  return (
    <div className="flex min-h-screen items-center justify-center bg-sand px-4">
      <form
        className="w-full max-w-sm space-y-4 border border-line bg-foam p-6"
        onSubmit={(event) => {
          event.preventDefault()
          unlock.mutate({ projectName, slug, password })
        }}
      >
        <div className="flex justify-center">
          <div className="grid size-14 place-items-center bg-lagoon/10 text-lagoon">
            <Lock className="size-6" />
          </div>
        </div>
        <div className="text-center">
          <h1 className="display-title text-lg font-semibold text-sea-ink">
            该页面受密码保护
          </h1>
          <p className="mt-1 text-sm text-sea-ink-soft">
            输入访问密码后继续浏览
          </p>
        </div>
        <Input
          type="password"
          value={password}
          onChange={(event) => setPassword(event.target.value)}
          placeholder="访问密码"
        />
        <Button type="submit" className="w-full" disabled={unlock.isPending}>
          解锁
        </Button>
      </form>
    </div>
  )
}
