import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { ArrowLeft } from 'lucide-react'
import { Button } from '#/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '#/components/ui/card'
import { useCreateRepoWikiConfig } from '#/hooks/useRepoWiki'
import { ConfigForm } from '#/components/repowiki/config-form'
import { staggerContainer, staggerItem } from '#/lib/motion'
import { PageHeader } from '#/components/page-header'

export const Route = createFileRoute('/console/repowiki/create')({
	component: CreateConfigPage,
})

function CreateConfigPage() {
	const navigate = useNavigate()
	const createMutation = useCreateRepoWikiConfig()

	return (
		<motion.div
			className="space-y-4"
			initial="hidden"
			animate="visible"
			variants={staggerContainer}
		>
			<PageHeader
				title="创建 RepoWiki 配置"
				description="添加新的代码仓库，开始自动生成 Wiki 文档"
				action={
					<Button
						variant="outline"
						onClick={() => navigate({ to: '/console/repowiki' })}
					>
						<ArrowLeft className="mr-2 size-4" />
						返回列表
					</Button>
				}
			/>

			{/* 表单卡片 */}
			<motion.div variants={staggerItem}>
				<Card className="border-border bg-card">
					<CardHeader>
						<CardTitle>仓库信息</CardTitle>
						<CardDescription>
							填写仓库的基本配置信息，支持 HTTPS 和 SSH 协议
						</CardDescription>
					</CardHeader>
					<CardContent>
						<ConfigForm
							onSubmit={(data) =>
								createMutation.mutate(data, {
									onSuccess: () => {
										navigate({ to: '/console/repowiki' })
									},
								})
							}
							isPending={createMutation.isPending}
							onCancel={() => navigate({ to: '/console/repowiki' })}
						/>
					</CardContent>
				</Card>
			</motion.div>
		</motion.div>
	)
}
