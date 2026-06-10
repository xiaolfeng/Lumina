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
import type { SessionItem } from '#/lib/models/response/qa-admin'

interface DeleteSessionDialogProps {
	open: boolean
	onOpenChange: (open: boolean) => void
	session: SessionItem | null
	onConfirm: () => void
	isPending?: boolean
}

export function DeleteSessionDialog({
	open,
	onOpenChange,
	session,
	onConfirm,
	isPending,
}: DeleteSessionDialogProps) {
	return (
		<AlertDialog open={open} onOpenChange={onOpenChange}>
			<AlertDialogContent>
				<AlertDialogHeader>
					<AlertDialogTitle>删除会话</AlertDialogTitle>
					<AlertDialogDescription>
						确定要删除会话「{session?.title}」吗？删除后所有问答数据将不可恢复。
					</AlertDialogDescription>
				</AlertDialogHeader>
				<AlertDialogFooter>
					<AlertDialogCancel disabled={isPending}>取消</AlertDialogCancel>
					<AlertDialogAction
						onClick={onConfirm}
						disabled={isPending}
						className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
					>
						{isPending ? '删除中...' : '确认删除'}
					</AlertDialogAction>
				</AlertDialogFooter>
			</AlertDialogContent>
		</AlertDialog>
	)
}
