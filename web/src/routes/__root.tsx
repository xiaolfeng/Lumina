import { createRootRoute, Outlet, redirect } from '@tanstack/react-router'
import { TanStackRouterDevtools } from '@tanstack/react-router-devtools'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

import { getStatus } from '#/lib/apis/auth'

const queryClient = new QueryClient()

export const Route = createRootRoute({
  beforeLoad: async ({ location }) => {
    const res = await queryClient.ensureQueryData({
      queryKey: ['auth', 'status'],
      queryFn: getStatus,
      staleTime: 0,
    })

    if (res.data?.is_initial === true && location.pathname !== '/auth/new') {
      throw redirect({ to: '/auth/new' })
    }

    if (res.data?.is_initial === false && location.pathname === '/auth/new') {
      throw redirect({ to: '/auth/login' })
    }
  },
  component: RootComponent,
})

function RootComponent() {
  return (
    <QueryClientProvider client={queryClient}>
      <Outlet />
      {import.meta.env.DEV && <TanStackRouterDevtools position="bottom-right" />}
    </QueryClientProvider>
  )
}
