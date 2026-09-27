import { useMemo, useState } from 'react'
import { useGetAllMetersDailyQuery, useListAnomaliesQuery, useListMetersQuery } from '../api/apiSlice'
import { Skeleton } from './Skeleton'
import { TimeSeriesChart } from './TimeSeriesChart'

type View = 'total' | 'meters'

const NORMAL_COLOR = '#9dc5c1'
const HIGH_SEVERITY_COLOR = '#ef4444'

export function AllMetersTimeline() {
  const { data: meters } = useListMetersQuery()
  const { data: anomalies } = useListAnomaliesQuery()
  const { data: dailyRows } = useGetAllMetersDailyQuery()
  const [view, setView] = useState<View>('total')

  const highSeverityMeters = useMemo(
    () => new Set((anomalies ?? []).filter((a) => a.severity === 'HIGH').map((a) => a.meter_id)),
    [anomalies],
  )

  const allDays = useMemo(() => {
    if (!dailyRows) return []
    return Array.from(new Set(dailyRows.map((r) => r.day))).sort()
  }, [dailyRows])

  const totalRows = useMemo(() => {
    if (!dailyRows) return []
    const dayTotals = new Map<string, number>()
    for (const r of dailyRows) {
      dayTotals.set(r.day, (dayTotals.get(r.day) ?? 0) + r.consumption_kwh)
    }
    return allDays.map((day) => ({ day, total: dayTotals.get(day) ?? 0 }))
  }, [dailyRows, allDays])

  const perMeterRows = useMemo(() => {
    if (!dailyRows || !meters) return []
    const lookup = new Map<string, Map<string, number>>()
    for (const r of dailyRows) {
      if (!lookup.has(r.meter_id)) lookup.set(r.meter_id, new Map())
      lookup.get(r.meter_id)!.set(r.day, r.consumption_kwh)
    }
    return allDays.map((day) => {
      const row: Record<string, number | string | null> = { day }
      for (const m of meters) {
        row[m.meter_id] = lookup.get(m.meter_id)?.get(day) ?? null
      }
      return row
    })
  }, [dailyRows, meters, allDays])

  const perMeterSeries = useMemo(
    () =>
      (meters ?? []).map((m) => ({
        key: m.meter_id,
        name: m.meter_id,
        color: highSeverityMeters.has(m.meter_id) ? HIGH_SEVERITY_COLOR : NORMAL_COLOR,
      })),
    [meters, highSeverityMeters],
  )

  const flaggedMeters = (meters ?? []).filter((m) => highSeverityMeters.has(m.meter_id))

  return (
    <div className="space-y-5">
      <div className="rounded-xl border border-surface-border bg-white p-5">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-sm font-semibold text-ink-900">Consumo</h2>
            <p className="mt-1 text-xs text-ink-400">
              {view === 'total'
                ? 'Los 12 medidores combinados · últimos 14 días'
                : 'Mismo período, los 12 medidores en un solo gráfico'}
            </p>
          </div>
          <div className="flex rounded-lg border border-surface-border bg-surface p-1">
            {(['total', 'meters'] as View[]).map((v) => (
              <button
                key={v}
                onClick={() => setView(v)}
                className={`rounded-md px-3 py-1.5 text-[11px] font-medium transition-colors ${
                  view === v ? 'bg-ink-900 text-white' : 'text-ink-500 hover:text-ink-900'
                }`}
              >
                {v === 'total' ? 'Total' : 'Por medidor'}
              </button>
            ))}
          </div>
        </div>

        <div className="mt-4">
          {view === 'total' ? (
            totalRows.length > 0 ? (
              <TimeSeriesChart
                ariaLabel="Consumo total combinado de los 12 medidores"
                data={totalRows}
                xKey="day"
                series={[{ key: 'total', name: 'Consumo total', color: '#18c8b3' }]}
                formatX={(x) => String(x).slice(5)}
                formatY={(y) => `${Math.round(Number(y) / 1000)}k kWh`}
              />
            ) : (
              <Skeleton className="h-55 w-full rounded-lg" />
            )
          ) : perMeterRows.length > 0 ? (
            <>
              <TimeSeriesChart
                ariaLabel="Consumo diario de los 12 medidores superpuestos"
                data={perMeterRows}
                xKey="day"
                series={perMeterSeries}
                hideLegend
                formatX={(x) => String(x).slice(5)}
                formatY={(y) => `${Math.round(Number(y))} kWh`}
              />
              <div className="mt-3 flex flex-wrap items-center gap-4 text-xs text-ink-500">
                <span className="flex items-center gap-1.5">
                  <span
                    className="h-0.5 w-3 rounded-full"
                    style={{ backgroundColor: NORMAL_COLOR }}
                  />
                  Medidores normales
                </span>
                {flaggedMeters.map((m) => (
                  <span key={m.meter_id} className="flex items-center gap-1.5">
                    <span
                      className="h-0.5 w-3 rounded-full"
                      style={{ backgroundColor: HIGH_SEVERITY_COLOR }}
                    />
                    {m.meter_id}
                  </span>
                ))}
              </div>
            </>
          ) : (
            <>
              <Skeleton className="h-55 w-full rounded-lg" />
              <Skeleton className="mt-3 h-4 w-48 rounded" />
            </>
          )}
        </div>
      </div>
    </div>
  )
}
