import {
  HeadContent,
  Scripts,
  createRootRoute,
  redirect,
} from '@tanstack/react-router'
import { TanStackRouterDevtoolsPanel } from '@tanstack/react-router-devtools'
import { TanStackDevtools } from '@tanstack/react-devtools'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

import { getStatus } from '#/lib/apis/auth'
import appCss from '../styles.css?url'

const queryClient = new QueryClient()

export const Route = createRootRoute({
  head: () => ({
    meta: [
      {
        charSet: 'utf-8',
      },
      {
        name: 'viewport',
        content: 'width=device-width, initial-scale=1',
      },
      {
        name: 'theme-color',
        content: '#e7f3ec',
      },
      {
        title: 'Lumina · 微明 — AI 深度代码认知与长期记忆的知识中枢',
      },
      {
        name: 'description',
        content:
          '赋予 AI 深度代码认知与长期记忆的知识中枢。RepoWiki、Memory、Q&A 三合一，通过 MCP 协议向所有 AI Agent 开放。',
      },
    ],
    links: [
      {
        rel: 'stylesheet',
        href: appCss,
      },
    ],
  }),
  beforeLoad: async ({ location }) => {
    let isInitial = true

    try {
      const res = await getStatus()
      isInitial = res.data?.is_initial ?? true
    } catch {
      // Graceful degradation: allow access on API failure
    }

    if (!isInitial && location.pathname !== '/new') {
      throw redirect({ to: '/new' })
    }

    if (isInitial && location.pathname === '/new') {
      throw redirect({ to: '/' })
    }
  },
  shellComponent: RootDocument,
})

function RootDocument({ children }: { children: React.ReactNode }) {
  return (
    <html lang="zh-CN">
      <head>
        <HeadContent />
      </head>
      <body>
        <QueryClientProvider client={queryClient}>
          {children}
          <TanStackDevtools
            config={{
              position: 'bottom-right',
            }}
            plugins={[
              {
                name: 'Tanstack Router',
                render: <TanStackRouterDevtoolsPanel />,
              },
            ]}
          />
        </QueryClientProvider>
        <Scripts />
      </body>
    </html>
  )
}
