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
import { useDeleteApikey } from '#/hooks/useApikey'
import type { ApikeyItem } from '#/lib/models/response/apikey'

interface DeleteDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  item: ApikeyItem | null
}

export function DeleteDialog({ open, onOpenChange, item }: DeleteDialogProps) {
  const deleteMutation = useDeleteApikey()

  const handleDelete = () => {
    if (!item) return
    deleteMutation.mutate(item.id, {
      onSuccess: () => onOpenChange(false),
    })
  }

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>确定要删除此令牌吗？</AlertDialogTitle>
          <AlertDialogDescription>
            此操作不可撤销。删除后，使用该令牌的所有 API
            请求将立即失效。
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>取消</AlertDialogCancel>
          <AlertDialogAction
            onClick={handleDelete}
            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
          >
            {deleteMutation.isPending ? '删除中...' : '确认删除'}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
