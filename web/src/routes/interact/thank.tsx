import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { ArrowRight, CheckCircle2, MessageSquare } from 'lucide-react'

import { Button } from '#/components/ui/button'
import { staggerContainer, staggerItem, staggerItemLeft } from '#/lib/motion'

export const Route = createFileRoute('/interact/thank')({
	component: ThankPage,
})

function ThankPage() {
	const navigate = useNavigate()

	return (
		<div className="flex min-h-0 flex-1 items-center justify-center p-4">
			<motion.div
				className="w-full max-w-lg rounded-2xl border border-line bg-surface p-8 shadow-[0_1px_3px_rgba(42,36,32,0.04),0_8px_24px_-8px_rgba(42,36,32,0.10)]"
				initial="hidden"
				animate="visible"
				variants={staggerContainer}
			>
				<motion.div
					className="flex flex-col items-center text-center"
					variants={staggerItem}
				>
					<div className="mb-4 flex size-16 items-center justify-center rounded-full bg-emerald-50">
						<CheckCircle2 className="size-8 text-emerald-500" aria-hidden />
					</div>

					<h1 className="text-xl font-bold tracking-tight text-sea-ink">
						感谢您的回答
					</h1>
					<p className="mt-2 text-sm leading-relaxed text-sea-ink-soft">
						本次会话已归档，所有问答记录已保存。
						请回到 AI Agent 继续后续的开发工作。
					</p>
				</motion.div>

				<motion.div
					className="mt-6 space-y-3 rounded-xl border border-line bg-foam p-4"
					variants={staggerItem}
				>
					<div className="flex items-start gap-3">
						<MessageSquare
							className="mt-0.5 size-4 shrink-0 text-lagoon"
							aria-hidden
						/>
						<div className="min-w-0 flex-1">
							<p className="text-sm font-medium text-sea-ink">
								会话已结束
							</p>
							<p className="mt-0.5 text-xs leading-relaxed text-sea-ink-soft">
								AI Agent 已归档此会话，后续问题将需要创建新的会话进行交互。
							</p>
						</div>
					</div>
				</motion.div>

				<motion.div
					className="mt-6 flex flex-col gap-2"
					variants={staggerItemLeft}
				>
					<Button
						onClick={() =>
							navigate({ to: '/interact', replace: true })
						}
						className="w-full rounded-lg bg-gradient-to-b from-lagoon to-lagoon-deep text-white shadow-sm hover:from-lagoon/95 hover:to-lagoon-deep/95"
					>
						返回大厅
						<ArrowRight className="ml-1.5 size-4" aria-hidden />
					</Button>
					<Button
						variant="ghost"
						size="sm"
						onClick={() => navigate({ to: '/console/qa', replace: true })}
						className="w-full text-xs text-sea-ink-soft"
					>
						查看历史会话
					</Button>
				</motion.div>
			</motion.div>
		</div>
	)
}
