import { useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '#/components/ui/card'
import { Button } from '#/components/ui/button'
import { Input } from '#/components/ui/input'
import { Label } from '#/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '#/components/ui/dialog'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '#/components/ui/alert-dialog'
import { Fingerprint, Plus, Trash2, Monitor } from 'lucide-react'
import { useBiometricCredentials, useRegisterBiometric, useDeleteBiometric } from '#/hooks/useBiometric'
import type { BiometricCredentialItem } from '#/lib/models/response/user'
import { toast } from 'sonner'

export function BiometricTab() {
  const { data: credData, isLoading } = useBiometricCredentials()
  const registerBiometric = useRegisterBiometric()
  const deleteBiometric = useDeleteBiometric()
  const [registerOpen, setRegisterOpen] = useState(false)
  const [deviceName, setDeviceName] = useState('')
  const [deleteId, setDeleteId] = useState<string | null>(null)

  const credentials = credData?.data?.items || []

  function handleRegister() {
    if (!deviceName.trim()) {
      toast.error('请输入设备名称')
      return
    }
    registerBiometric.mutate(deviceName, {
      onSuccess: () => {
        toast.success('生物特征注册成功')
        setRegisterOpen(false)
        setDeviceName('')
      },
      onError: (err) => toast.error('注册失败：' + err.message),
    })
  }

  function handleDelete() {
    if (!deleteId) return
    deleteBiometric.mutate(deleteId, {
      onSuccess: () => {
        toast.success('凭证已删除')
        setDeleteId(null)
      },
      onError: (err) => toast.error('删除失败：' + err.message),
    })
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2">
            <Fingerprint className="size-4" />
            生物特征凭证
          </CardTitle>
          <Dialog open={registerOpen} onOpenChange={setRegisterOpen}>
            <DialogTrigger asChild>
              <Button size="sm" variant="outline">
                <Plus className="size-4" />
                注册新设备
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>注册生物特征凭证</DialogTitle>
                <DialogDescription>
                  输入设备名称并使用浏览器的生物特征认证完成注册。
                </DialogDescription>
              </DialogHeader>
              <div className="space-y-2">
                <Label htmlFor="device-name">设备名称</Label>
                <Input
                  id="device-name"
                  placeholder="如：MacBook Pro Touch ID"
                  value={deviceName}
                  onChange={(e) => setDeviceName(e.target.value)}
                />
              </div>
              <DialogFooter>
                <Button variant="outline" onClick={() => setRegisterOpen(false)}>
                  取消
                </Button>
                <Button onClick={handleRegister} disabled={registerBiometric.isPending}>
                  {registerBiometric.isPending ? '注册中…' : '开始注册'}
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </div>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <p className="text-sm text-sea-ink-soft">加载中…</p>
        ) : credentials.length === 0 ? (
          <p className="text-sm text-sea-ink-soft">尚未注册任何生物特征凭证。</p>
        ) : (
          <div className="space-y-3">
            {credentials.map((cred: BiometricCredentialItem) => (
              <div
                key={cred.id}
                className="flex items-center justify-between rounded-lg border border-chip-line p-3"
              >
                <div className="flex items-center gap-3">
                  <div className="flex size-9 items-center justify-center rounded-lg bg-lagoon/10 text-lagoon">
                    <Monitor className="size-4" />
                  </div>
                  <div>
                    <p className="text-sm font-medium text-sea-ink">{cred.device_name}</p>
                    <p className="text-xs text-sea-ink-soft">
                      {cred.last_used_at
                        ? `最后使用：${new Date(cred.last_used_at * 1000).toLocaleString()}`
                        : '从未使用'}
                    </p>
                  </div>
                </div>
                <Button
                  size="icon"
                  variant="ghost"
                  onClick={() => setDeleteId(cred.id)}
                >
                  <Trash2 className="size-4 text-red-500" />
                </Button>
              </div>
            ))}
          </div>
        )}
      </CardContent>

      <AlertDialog open={!!deleteId} onOpenChange={(open) => !open && setDeleteId(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认删除凭证？</AlertDialogTitle>
            <AlertDialogDescription>
              删除后，该设备将无法使用生物特征登录。此操作不可撤销。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={handleDelete}>确认删除</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </Card>
  )
}
