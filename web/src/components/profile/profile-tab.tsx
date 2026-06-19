import { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '#/components/ui/card'
import { Button } from '#/components/ui/button'
import { Input } from '#/components/ui/input'
import { Label } from '#/components/ui/label'
import { useAuth } from '#/hooks/useAuth'
import { useUpdateProfile } from '#/hooks/useProfile'
import { toast } from 'sonner'

export function ProfileTab() {
  const { currentUser } = useAuth()
  const updateProfile = useUpdateProfile()
  const [username, setUsername] = useState('')
  const [email, setEmail] = useState('')

  useEffect(() => {
    if (currentUser.data?.data) {
      setUsername(currentUser.data.data.username)
      setEmail(currentUser.data.data.email)
    }
  }, [currentUser.data])

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    updateProfile.mutate(
      { username, email },
      {
        onSuccess: () => toast.success('个人资料更新成功'),
        onError: (err) => toast.error('更新失败：' + err.message),
      },
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>个人资料</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="username">用户名</Label>
            <Input
              id="username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
              minLength={3}
              maxLength={32}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="email">邮箱</Label>
            <Input
              id="email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />
          </div>
          <Button type="submit" disabled={updateProfile.isPending}>
            {updateProfile.isPending ? '保存中…' : '保存'}
          </Button>
        </form>
      </CardContent>
    </Card>
  )
}
