import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@lumina/components/ui/card'
import { Badge } from '@lumina/components/ui/badge'
import { RadioGroup, RadioGroupItem } from '@lumina/components/ui/radio-group'
import { Label } from '@lumina/components/ui/label'
import { SkeletonTable } from '#/components/skeleton-table'
import { useRepoWikiVersions, useUpdateSelectedVersion } from '#/hooks/useRepoWiki'
import { Check, Loader2, Clock } from 'lucide-react'

interface VersionSwitcherProps {
	configId: string
	selectedVersionId?: string
}

export function VersionSwitcher({ configId, selectedVersionId }: VersionSwitcherProps) {
	const { data: versionsData, isLoading } = useRepoWikiVersions(configId, 1, 50)
	const updateMutation = useUpdateSelectedVersion()

	const versions = versionsData?.items ?? []
	const completedVersions = versions.filter((v) => v.status === 'completed')

	if (isLoading) return <SkeletonTable rows={3} />

	if (completedVersions.length === 0) {
		return (
			<Card className="border-border bg-card">
				<CardHeader>
					<CardTitle>选中版本</CardTitle>
					<CardDescription>选择一个已完成的 Wiki 版本作为当前展示版本</CardDescription>
				</CardHeader>
				<CardContent>
					<div className="text-center py-8 text-muted-foreground">
						<p className="text-sm">暂无已完成版本，完成分析后可在此切换</p>
					</div>
				</CardContent>
			</Card>
		)
	}

	return (
		<Card className="border-border bg-card">
			<CardHeader>
				<CardTitle>选中版本</CardTitle>
				<CardDescription>选择一个已完成的 Wiki 版本作为当前展示版本</CardDescription>
			</CardHeader>
			<CardContent>
				<RadioGroup
					value={selectedVersionId ?? ''}
					onValueChange={(value) => {
						updateMutation.mutate({
							configId,
							versionId: value,
						})
					}}
					className="space-y-3"
				>
					{completedVersions.map((version) => {
						const isSelected = version.id === selectedVersionId

						return (
							<div
								key={version.id}
								className={`flex items-start gap-3 rounded-lg border p-3 transition-colors ${
									isSelected
										? 'border-lagoon bg-lagoon/5'
										: 'border-border/50 hover:bg-muted/30'
								}`}
							>
								<RadioGroupItem value={version.id} id={`version-${version.id}`} className="mt-0.5" />
								<div className="flex-1 min-w-0 space-y-1">
									<div className="flex items-center gap-2 flex-wrap">
										<Label
											htmlFor={`version-${version.id}`}
											className={`text-sm font-medium cursor-pointer ${
												isSelected ? 'text-lagoon' : ''
											}`}
										>
											<span className="font-mono">#{version.id}</span>
										</Label>
										{isSelected && (
											<Badge variant="default" className="gap-1 bg-lagoon text-white hover:bg-lagoon-deep text-[10px] px-1.5 py-0">
												<Check className="size-3" />
												当前选中
											</Badge>
										)}
										<Badge variant="outline" className="text-[10px] px-1.5 py-0">
											已完成
										</Badge>
									</div>
									<div className="flex items-center gap-3 text-xs text-muted-foreground">
										<code className="bg-muted px-1.5 py-0.5 rounded font-mono">
											{version.commit_hash.slice(0, 7)}
										</code>
										<span className="flex items-center gap-1">
											<Clock className="size-3" />
											创建: {new Date(version.created_at).toLocaleString('zh-CN')}
										</span>
										{version.completed_at && (
											<span className="flex items-center gap-1">
												<Check className="size-3" />
												完成: {new Date(version.completed_at).toLocaleString('zh-CN')}
											</span>
										)}
									</div>
								</div>
								{updateMutation.isPending &&
									updateMutation.variables?.versionId === version.id && (
										<Loader2 className="size-4 animate-spin text-muted-foreground mt-0.5 shrink-0" />
									)}
							</div>
						)
					})}
				</RadioGroup>
			</CardContent>
		</Card>
	)
}
