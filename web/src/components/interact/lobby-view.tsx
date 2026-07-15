import { useEffect, useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import { Bot, Clock, FolderOpen, Loader2, Plus, Users } from 'lucide-react'

import { ScrollArea } from '@lumina/components/ui/scroll-area'
import { createSession } from '#/lib/apis/qa-admin'
import { getProjectList } from '#/lib/apis/project'
import type { ProjectItem } from '#/lib/models/response/project'
import { useSidebarOpen } from '#/hooks/useSidebarOpen'

import { Kicker, PanelCard, LoadingState  } from './primitives'
import type { Session } from './types'

/**
 * 未连接态大厅：项目选择 + 活跃会话列表。
 */
export function LobbyView({ sessions }: { sessions: Session[] }) {
	const navigate = useNavigate()
	const [projects, setProjects] = useState<ProjectItem[]>([])
	const [loadingProjects, setLoadingProjects] = useState(true)
	const [creating, setCreating] = useState(false)

	useEffect(() => {
		;(async () => {
			try {
				const res = await getProjectList({ size: 100 })
				setProjects(res.data?.items || [])
			} catch (e) {
				console.error('Failed to load projects:', e)
			} finally {
				setLoadingProjects(false)
			}
		})()
	}, [])

	async function handleCreate(projectId: string) {
		setCreating(true)
		try {
			const createRes = await createSession({
				project_id: projectId,
				title: `临时问答 ${new Date().toLocaleString('zh-CN')}`,
				agent: 'web',
				type: 'temporary',
			})
			const newSession = createRes.data
			if (newSession?.hash) {
				navigate({
					to: '/interact',
					search: { session: newSession.hash },
					replace: true,
				})
			}
		} catch (e) {
			console.error('Failed to create session:', e)
		} finally {
			setCreating(false)
		}
	}

	return (
		<div className="flex min-h-0 flex-1 items-center justify-center gap-8 p-6">
			<div className="w-full max-w-sm space-y-4">
				<PanelCard
					flushHeader
					header={
						<div className="px-4 py-2.5">
							<Kicker>选择项目</Kicker>
						</div>
					}
				>
					{loadingProjects ? (
						<LoadingState text="加载项目中…" />
					) : projects.length === 0 ? (
						<div className="py-8 text-center">
							<FolderOpen
								className="mx-auto mb-2 size-8 text-sea-ink-soft/30"
								aria-hidden
							/>
							<p className="text-xs text-sea-ink-soft/50">暂无可用项目</p>
						</div>
					) : (
						<div className="divide-y divide-line/30">
							{projects.map((project) => (
								<button
									key={project.id}
									type="button"
									disabled={creating}
									onClick={() => handleCreate(project.id)}
									className="group flex w-full cursor-pointer items-center gap-3 px-4 py-3 text-left transition-colors duration-150 hover:bg-lagoon/5 disabled:cursor-wait"
								>
									<FolderOpen
										className="size-4 shrink-0 text-lagoon/70 group-hover:text-lagoon"
										aria-hidden
									/>
									<div className="min-w-0 flex-1">
										<p className="truncate text-sm font-medium text-sea-ink">
											{project.name}
										</p>
										{project.description && (
											<p className="truncate text-[11px] text-sea-ink-soft/60">
												{project.description}
											</p>
										)}
									</div>
									{creating ? (
										<Loader2
											className="size-4 animate-spin text-sea-ink-soft/40"
											aria-hidden
										/>
									) : (
										<Plus
											className="size-4 text-sea-ink-soft/30 group-hover:text-lagoon"
											aria-hidden
										/>
									)}
								</button>
							))}
						</div>
					)}
				</PanelCard>
			</div>

			<ActiveSessionList sessions={sessions} />
		</div>
	)
}

function ActiveSessionList({ sessions }: { sessions: Session[] }) {
	const navigate = useNavigate()
	const { setOpen } = useSidebarOpen()

	if (sessions.length === 0) return null

	return (
		<div className="w-full max-w-sm">
			<PanelCard
				flushHeader
				header={
					<div className="px-4 py-2.5">
						<Kicker>活跃会话</Kicker>
					</div>
				}
				bodyClassName="p-0"
			>
				<ScrollArea className="max-h-[400px]">
					<div className="space-y-1 p-2">
						{sessions.map((session) => (
							<button
								key={session.id}
								type="button"
								onClick={() => {
									setOpen(false)
									if (session.hash) {
										navigate({
											to: '/interact',
											search: { session: session.hash },
											replace: true,
										})
									}
								}}
								className="group flex w-full cursor-pointer flex-col gap-1 rounded-lg px-3 py-2.5 text-left transition-colors duration-150 hover:bg-lagoon/5"
							>
								<span className="text-sm font-medium leading-tight text-sea-ink">
									{session.title}
								</span>
								<div className="flex items-center gap-2 text-[11px] text-sea-ink-soft">
									<span className="flex items-center gap-0.5">
										<Bot className="size-3" aria-hidden />
										{session.agent}
									</span>
									<span className="flex items-center gap-0.5">
										<Users className="size-3" aria-hidden />
										{session.onlineDevices}
									</span>
									<span className="flex items-center gap-0.5">
										<Clock className="size-2.5" aria-hidden />
										{session.updatedAt}
									</span>
								</div>
							</button>
						))}
					</div>
				</ScrollArea>
			</PanelCard>
		</div>
	)
}
