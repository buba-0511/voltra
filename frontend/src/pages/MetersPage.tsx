import { ChevronDown, ChevronUp, ChevronsUpDown, Search } from 'lucide-react'
import { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Badge } from '../components/Badge'
import { Skeleton } from '../components/Skeleton'
import { useListAnomaliesQuery, useListMetersQuery } from '../api/apiSlice'
import type { Anomaly, Severity } from '../api/types'
import { meterStatus } from '../utils/meterStatus'

type Filter = 'Todos' | 'Normales' | 'Alertas' | 'Críticas'
type SortKey = 'consumo' | 'variacion' | 'severidad'
type SortDir = 'asc' | 'desc'

const severityRank: Record<Severity, number> = { HIGH: 3, MEDIUM: 2, LOW: 1 }

function severityFilter(severity: Severity | undefined, filter: Filter) {
  if (filter === 'Todos') return true
  if (filter === 'Normales') return !severity
  if (filter === 'Alertas') return severity === 'MEDIUM'
  return severity === 'HIGH'
}

function SortHeader({
  label,
  sortKey,
  sort,
  onToggle,
}: {
  label: string
  sortKey: SortKey
  sort: { key: SortKey; dir: SortDir } | null
  onToggle: (key: SortKey) => void
}) {
  const active = sort?.key === sortKey
  return (
    <button onClick={() => onToggle(sortKey)} className="flex items-center gap-1 hover:text-ink-700">
      {label}
      {active ? (
        sort?.dir === 'asc' ? (
          <ChevronUp size={12} />
        ) : (
          <ChevronDown size={12} />
        )
      ) : (
        <ChevronsUpDown size={12} className="text-ink-300" />
      )}
    </button>
  )
}

export function MetersPage() {
  const navigate = useNavigate()
  const [filter, setFilter] = useState<Filter>('Todos')
  const [search, setSearch] = useState('')
  const [sort, setSort] = useState<{ key: SortKey; dir: SortDir } | null>(null)
  const { data: meters, isLoading: metersLoading } = useListMetersQuery()
  const { data: anomalies } = useListAnomaliesQuery()

  const anomalyByMeter = useMemo(() => {
    const map = new Map<string, Anomaly>()
    for (const a of anomalies ?? []) {
      const existing = map.get(a.meter_id)
      if (!existing || a.confidence > existing.confidence) map.set(a.meter_id, a)
    }
    return map
  }, [anomalies])

  const rows = (meters ?? []).map((m) => ({ meter: m, anomaly: anomalyByMeter.get(m.meter_id) }))

  const filtered = rows
    .filter((r) => severityFilter(r.anomaly?.severity, filter))
    .filter((r) => r.meter.meter_id.toLowerCase().includes(search.trim().toLowerCase()))

  function sortValue(r: (typeof filtered)[number], key: SortKey) {
    if (key === 'consumo') return r.meter.consumption_kwh
    if (key === 'variacion') return r.meter.variation_pct
    return r.anomaly ? severityRank[r.anomaly.severity] : 0
  }

  const sorted = sort
    ? [...filtered].sort(
        (a, b) => (sortValue(a, sort.key) - sortValue(b, sort.key)) * (sort.dir === 'asc' ? 1 : -1),
      )
    : filtered

  function toggleSort(key: SortKey) {
    setSort((prev) => {
      if (!prev || prev.key !== key) return { key, dir: 'desc' }
      if (prev.dir === 'desc') return { key, dir: 'asc' }
      return null
    })
  }

  return (
    <div className="flex h-full flex-col gap-6">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight text-ink-900">Medidores</h1>
          <p className="mt-1 text-sm text-ink-400">
            {meters?.length ?? 0} medidores conectados a tu operación.
          </p>
        </div>
        <div className="flex flex-wrap items-center gap-3">
          <div className="relative">
            <Search
              size={13}
              className="pointer-events-none absolute top-1/2 left-2.5 -translate-y-1/2 text-ink-300"
            />
            <input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Buscar medidor…"
              className="rounded-lg border border-surface-border bg-white py-1.5 pr-3 pl-7 text-xs text-ink-900 placeholder:text-ink-300 focus:border-brand-400 focus:outline-none"
            />
          </div>
          <div className="flex rounded-lg border border-surface-border bg-white p-1">
            {(['Todos', 'Normales', 'Alertas', 'Críticas'] as Filter[]).map((f) => (
              <button
                key={f}
                onClick={() => setFilter(f)}
                className={`rounded-md px-3 py-1.5 text-[11px] font-medium transition-colors ${
                  filter === f ? 'bg-ink-900 text-white' : 'text-ink-500 hover:text-ink-900'
                }`}
              >
                {f}
              </button>
            ))}
          </div>
        </div>
      </div>

      <div className="min-h-0 flex-1 overflow-hidden rounded-xl border border-surface-border bg-white">
        <div className="h-full overflow-auto">
          <table className="w-full min-w-180 text-left">
            <thead className="sticky top-0 z-10">
              <tr className="border-b border-surface-border bg-surface text-[10px] font-semibold text-ink-400">
                <th className="px-5 py-3">Medidor</th>
                <th className="px-4 py-3">
                  <SortHeader label="Consumo" sortKey="consumo" sort={sort} onToggle={toggleSort} />
                </th>
                <th className="px-4 py-3">
                  <SortHeader label="Variación" sortKey="variacion" sort={sort} onToggle={toggleSort} />
                </th>
                <th className="px-4 py-3">Estado</th>
                <th className="px-4 py-3">
                  <SortHeader label="Anomalía" sortKey="severidad" sort={sort} onToggle={toggleSort} />
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-surface-border/60">
              {metersLoading ? (
                Array.from({ length: 6 }).map((_, i) => (
                  <tr key={i}>
                    <td className="px-5 py-4">
                      <Skeleton className="h-3 w-16" />
                      <Skeleton className="mt-2 h-2 w-24" />
                    </td>
                    <td className="px-4 py-4">
                      <Skeleton className="h-3 w-14" />
                    </td>
                    <td className="px-4 py-4">
                      <Skeleton className="h-3 w-12" />
                    </td>
                    <td className="px-4 py-4">
                      <Skeleton className="h-5 w-16 rounded-md" />
                    </td>
                    <td className="px-4 py-4">
                      <Skeleton className="h-5 w-16 rounded-md" />
                    </td>
                  </tr>
                ))
              ) : sorted.length === 0 ? (
                <tr>
                  <td colSpan={5} className="px-5 py-8 text-center text-sm text-ink-400">
                    No hay medidores que coincidan.
                  </td>
                </tr>
              ) : (
                sorted.map(({ meter, anomaly }) => (
                  <tr
                    key={meter.meter_id}
                    onClick={() => navigate(`/meters/${meter.meter_id}`)}
                    className="cursor-pointer hover:bg-surface"
                  >
                    <td className="px-5 py-4">
                      <span className="text-xs font-semibold text-ink-900">{meter.meter_id}</span>
                      <div className="mt-1 text-[10px] text-ink-400">{meter.location}</div>
                    </td>
                    <td className="px-4 py-4 text-xs font-semibold tabular-nums text-ink-900">
                      {meter.consumption_kwh.toLocaleString('es-ES', { maximumFractionDigits: 0 })}{' '}
                      <span className="font-normal text-ink-400">kWh</span>
                    </td>
                    <td
                      className={`px-4 py-4 text-xs font-semibold tabular-nums ${
                        Math.abs(meter.variation_pct) > 20 ? 'text-red-500' : 'text-ink-600'
                      }`}
                    >
                      {meter.variation_pct > 0 ? '+' : ''}
                      {meter.variation_pct.toFixed(1)}%
                    </td>
                    <td className="px-4 py-4">
                      <Badge>{meterStatus(anomaly?.severity)}</Badge>
                    </td>
                    <td className="px-4 py-4">
                      {anomaly ? (
                        <Badge>{anomaly.severity}</Badge>
                      ) : (
                        <span className="text-xs text-ink-300">—</span>
                      )}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
