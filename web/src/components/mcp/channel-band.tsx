import type { ReactNode } from 'react'

interface ChannelBandProps {
  index: number
  alt?: boolean
  children: ReactNode
}

export function ChannelBand({
  index,
  alt = false,
  children,
}: ChannelBandProps) {
  return (
    <div
      className={`grid grid-cols-[48px_1fr] border-b border-line sm:grid-cols-[72px_1fr] ${
        alt ? 'bg-chip-bg' : 'bg-foam'
      }`}
    >
      <div className="pt-7 pl-2 font-mono text-[12px] font-bold tracking-[0.08em] text-lagoon-deep sm:pl-4 sm:text-[13px]">
        {String(index).padStart(2, '0')}
      </div>
      <section className="min-w-0 px-3 py-6 sm:px-8 sm:py-7">
        {children}
      </section>
    </div>
  )
}

export function ChannelKicker({ children }: { children: ReactNode }) {
  return (
    <p className="text-[11px] font-bold uppercase tracking-[0.2em] text-lagoon-deep">
      {children}
    </p>
  )
}
