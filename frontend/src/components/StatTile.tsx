import type { ReactNode } from 'react'

interface StatTileProps {
  label: string
  value: ReactNode
  sub?: ReactNode
  subTone?: 'default' | 'warning' | 'danger'
  icon?: ReactNode
}

const subToneClass: Record<NonNullable<StatTileProps['subTone']>, string> = {
  default: 'text-ink-400',
  warning: 'text-amber-600',
  danger: 'text-red-500',
}

export function StatTile({ label, value, sub, subTone = 'default', icon }: StatTileProps) {
  return (
    <div className="rounded-xl border border-surface-border bg-white p-4">
      <div className="flex items-center justify-between text-[11px] text-ink-400">
        <span>{label}</span>
        {icon}
      </div>
      {/* Fixed min-heights so a tile's content (e.g. "Último análisis"
          going from "—" to a full date/time string) doesn't change the
          tile's height and, via CSS grid's row stretch, the whole row's
          height along with it. */}
      <div className="mt-3 line-clamp-2 min-h-16 text-2xl font-semibold tracking-tight text-ink-900">
        {value}
      </div>
      <div className={`mt-1 min-h-3.5 text-[11px] ${subToneClass[subTone]}`}>{sub}</div>
    </div>
  )
}
