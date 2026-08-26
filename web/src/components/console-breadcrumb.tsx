import { Fragment } from 'react'
import { Link, useMatches, useParams } from '@tanstack/react-router'
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@lumina/components/ui/breadcrumb'
import { useProjectNameMap } from '#/hooks/useProject'

interface Crumb {
  label: string
  to: string
}

export function ConsoleBreadcrumb() {
  const matches = useMatches()
  const { projectId } = useParams({ strict: false })
  const { names } = useProjectNameMap({ enabled: Boolean(projectId) })

  const crumbs: Crumb[] = []
  for (const match of matches) {
    const label = match.staticData.crumb
    if (!label) continue
    crumbs.push({ label, to: match.pathname })
  }

  if (projectId) {
    const wikiIdx = crumbs.findIndex((crumb) => crumb.label === 'Wiki')
    if (wikiIdx >= 0) {
      crumbs.splice(wikiIdx, 0, {
        label: names[projectId] || '项目',
        to: `/console/project/${projectId}/repowiki`,
      })
    }
  }

  if (crumbs.length === 0) return null

  return (
    <Breadcrumb>
      <BreadcrumbList>
        {crumbs.map((crumb, index) => {
          const isLast = index === crumbs.length - 1
          return (
            <Fragment key={`${crumb.to}-${crumb.label}`}>
              {index > 0 && <BreadcrumbSeparator />}
              <BreadcrumbItem>
                {isLast ? (
                  <BreadcrumbPage>{crumb.label}</BreadcrumbPage>
                ) : (
                  <BreadcrumbLink
                    asChild
                    className="text-sea-ink-soft transition-colors hover:text-lagoon"
                  >
                    <Link to={crumb.to as never}>{crumb.label}</Link>
                  </BreadcrumbLink>
                )}
              </BreadcrumbItem>
            </Fragment>
          )
        })}
      </BreadcrumbList>
    </Breadcrumb>
  )
}
