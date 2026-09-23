import { createFileRoute, Outlet } from '@tanstack/react-router'

export const Route = createFileRoute('/console/qa')({
  staticData: { crumb: '问答管理' },
  component: QaLayout,
})

function QaLayout() {
  return <Outlet />
}
