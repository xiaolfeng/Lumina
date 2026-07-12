import { createFileRoute, Outlet } from '@tanstack/react-router'

export const Route = createFileRoute('/console/project')({
	component: ProjectLayout,
})

function ProjectLayout() {
	return <Outlet />
}
