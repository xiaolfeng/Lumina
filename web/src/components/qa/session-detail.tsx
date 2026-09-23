import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@lumina/components/ui/card'
import { Badge } from '@lumina/components/ui/badge'
import { formatDateTime } from '#/lib/format-date'
import type { SessionDetailResponse } from '#/lib/models/response/qa-admin'

interface SessionDetailProps {
  session: SessionDetailResponse
}

export function SessionDetail({ session }: SessionDetailProps) {
  const statusVariant =
    session.status === 'active'
      ? 'default'
      : session.status === 'expired'
        ? 'outline'
        : 'destructive'
  const statusLabel =
    session.status === 'active'
      ? '活跃'
      : session.status === 'expired'
        ? '已过期'
        : '已删除'
  const typeLabel = session.type === 'permanent' ? '永久' : '临时'

  return (
    <Card>
      <CardHeader>
        <CardTitle className="display-title">{session.title}</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
          <div>
            <span className="text-muted-foreground">Agent</span>
            <p className="font-medium">{session.agent}</p>
          </div>
          <div>
            <span className="text-muted-foreground">关联项目</span>
            <p className="font-medium">{session.project_name || '—'}</p>
          </div>
          <div>
            <span className="text-muted-foreground">类型</span>
            <p>
              <Badge variant="secondary">{typeLabel}</Badge>
            </p>
          </div>
          <div>
            <span className="text-muted-foreground">状态</span>
            <p>
              <Badge variant={statusVariant}>{statusLabel}</Badge>
            </p>
          </div>
          <div>
            <span className="text-muted-foreground">在线设备</span>
            <p className="font-medium tabular-nums">{session.online_devices}</p>
          </div>
          <div>
            <span className="text-muted-foreground">创建时间</span>
            <p className="font-medium">{formatDateTime(session.created_at)}</p>
          </div>
          <div>
            <span className="text-muted-foreground">过期时间</span>
            <p className="font-medium">
              {session.expires_at
                ? formatDateTime(session.expires_at)
                : '永久有效'}
            </p>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
