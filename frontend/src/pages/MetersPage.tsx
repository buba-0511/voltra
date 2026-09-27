import { useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { Badge } from '../components/Badge'
import { useListAnomaliesQuery, useListMetersQuery } from '../api/apiSlice'
import type { Anomaly, Severity } from '../api/types'

type Filter = 'Todos' | 'Normales' | 'Alertas' | 'Críticas'

function severityFilter(severity: Severity | undefined, filter: Filter) {
  if (filter === 'Todos') return true
  if (filter === 'Normales') return !severity
  if (filter === 'Alertas') return severity === 'MEDIUM'
  return severity === 'HIGH'
}

// The meter table's "Estado" column uses OK/Alert/Critical, separate from
// "Anomalía" which shows the raw severity (Medium/High) - two different
// vocabularies for two different questions (brief, section 6).
function meterStatus(severity: Severity | undefined) {
  if (severity === 'HIGH') return 'CRITICAL'
  if (severity === 'MEDIUM' || severity === 'LOW') return 'ALERT'
  return 'OK'
}

export function MetersPage() {
  const [filter, setFilter] = useState<Filter>('Todos')
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
  const filtered = rows.filter((r) => severityFilter(r.anomaly?.severity, filter))

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight text-ink-900">Medidores</h1>
          <p className="mt-1 text-sm text-ink-400">
            {meters?.length ?? 0} medidores conectados a tu operación.
          </p>
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

      <div className="overflow-hidden rounded-xl border border-surface-border bg-white">
        <div className="overflow-x-auto">
          <table className="w-full min-w-180 text-left">
            <thead>
              <tr className="border-b border-surface-border bg-surface text-[10px] font-semibold tracking-wide text-ink-400 uppercase">
                <th className="px-5 py-3">Medidor</th>
                <th className="px-4 py-3">Consumo</th>
                <th className="px-4 py-3">Variación</th>
                <th className="px-4 py-3">Estado</th>
                <th className="px-4 py-3">Anomalía</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-surface-border/60">
              {metersLoading ? (
                <tr>
                  <td colSpan={5} className="px-5 py-8 text-center text-sm text-ink-400">
                    Cargando…
                  </td>
                </tr>
              ) : filtered.length === 0 ? (
                <tr>
                  <td colSpan={5} className="px-5 py-8 text-center text-sm text-ink-400">
                    No hay medidores en este filtro.
                  </td>
                </tr>
              ) : (
                filtered.map(({ meter, anomaly }) => (
                  <tr key={meter.meter_id} className="hover:bg-surface">
                    <td className="px-5 py-4">
                      <Link
                        to={`/meters/${meter.meter_id}`}
                        className="text-xs font-semibold text-ink-900 hover:text-brand-600"
                      >
                        {meter.meter_id}
                      </Link>
                      <div className="mt-1 text-[10px] text-ink-400">{meter.location}</div>
                    </td>
                    <td className="px-4 py-4 text-xs font-semibold tabular-nums text-ink-900">
                      {meter.consumption_kwh.toLocaleString('es-ES', { maximumFractionDigits: 0 })}{' '}
                      <span className="font-normal text-ink-400">kWh</span>
                    </td>
                    <td
                      className={`px-4 py-4 text-xs font-semibold tabular-nums ${
                        anomaly && Math.abs(anomaly.variation_pct) > 20 ? 'text-red-500' : 'text-ink-600'
                      }`}
                    >
                      {anomaly
                        ? `${anomaly.variation_pct > 0 ? '+' : ''}${anomaly.variation_pct.toFixed(1)}%`
                        : '—'}
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
