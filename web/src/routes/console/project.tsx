import { createFileRoute, Outlet } from '@tanstack/react-router'

export const Route = createFileRoute('/console/project')({
	staticData: { crumb: '项目管理' },
	component: ProjectLayout,
})

function ProjectLayout() {
	return <Outlet />
}
