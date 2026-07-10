import { Card, CardContent, CardHeader, CardTitle } from '#/components/ui/card'

interface EnvInfoItem {
  label: string
  value: string
}

interface EnvInfoCardProps {
  items: EnvInfoItem[]
}

export function EnvInfoCard({ items }: EnvInfoCardProps) {
  return (
    <Card className="mt-6 border-dashed">
      <CardHeader>
        <CardTitle className="text-sm font-medium text-muted-foreground">
          环境变量（只读）
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid gap-3 sm:grid-cols-2">
          {items.map((item) => (
            <div key={item.label} className="space-y-1">
              <p className="text-xs font-medium text-muted-foreground">
                {item.label}
              </p>
              <p className="text-sm font-mono break-all">{item.value}</p>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}
