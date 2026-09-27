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
      <div className="mt-3 text-2xl font-semibold tracking-tight text-ink-900">{value}</div>
      {sub && <div className={`mt-1 text-[11px] ${subToneClass[subTone]}`}>{sub}</div>}
    </div>
  )
}
