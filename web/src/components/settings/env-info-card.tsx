interface EnvInfoItem {
  label: string
  value: string
}

interface EnvInfoCardProps {
  items: EnvInfoItem[]
}

export function EnvInfoCard({ items }: EnvInfoCardProps) {
  return (
    <div className="mt-8 border border-line bg-foam">
      <div className="border-b border-line bg-sand px-5 py-3">
        <span className="text-[11px] font-bold uppercase tracking-wider text-sea-ink-soft">
          底层环境变量与只读参数 (ENVIRONMENT READONLY)
        </span>
      </div>
      <div className="grid gap-0 sm:grid-cols-2">
        {items.map((item, idx) => (
          <div
            key={item.label}
            className={`p-4 px-5 border-b sm:border-b-0 border-line last:border-b-0 flex flex-col gap-1 ${
              idx % 2 === 0 ? 'sm:border-r' : ''
            }`}
          >
            <span className="font-mono text-[10.5px] font-bold uppercase tracking-wider text-lagoon-deep">
              {item.label}
            </span>
            <span className="font-mono text-xs text-sea-ink break-all select-all">
              {item.value}
            </span>
          </div>
        ))}
      </div>
    </div>
  )
}
