import { createFileRoute, Outlet } from '@tanstack/react-router'

export const Route = createFileRoute('/console/qa')({
	component: QaLayout,
})

function QaLayout() {
	return <Outlet />
}
