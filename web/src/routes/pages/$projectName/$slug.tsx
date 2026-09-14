import { createFileRoute, Outlet, redirect } from '@tanstack/react-router'

export const Route = createFileRoute('/pages/$projectName/$slug')({
  beforeLoad: ({ location, params }) => {
    const rest = location.pathname.replace(
      `/pages/${params.projectName}/${params.slug}`,
      '',
    )
    if (rest === '' || rest === '/') {
      throw redirect({
        to: '/pages/$projectName/$slug/$',
        params: {
          projectName: params.projectName,
          slug: params.slug,
          _splat: 'index.html',
        },
      })
    }
  },
  component: () => <Outlet />,
})
