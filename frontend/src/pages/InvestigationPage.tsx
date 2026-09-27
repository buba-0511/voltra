import { AlertTriangle, ArrowLeft, Check, CircleHelp } from 'lucide-react'
import { useMemo, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { Badge } from '../components/Badge'
import { Skeleton } from '../components/Skeleton'
import { TimeSeriesChart } from '../components/TimeSeriesChart'
import {
  useGetAnomalyQuery,
  useGetMeterDailyQuery,
  useUpdateAnomalyStatusMutation,
} from '../api/apiSlice'

function dayKey(iso: string) {
  return iso.slice(0, 10)
}

const eventTypeLabels: Record<string, string> = {
  OPERATIONAL_CHANGE: 'Cambio operativo',
  SCHEDULED_OUTAGE: 'Corte programado',
  DATA_QUALITY: 'Calidad de datos',
  UNKNOWN: 'Evento sin clasificar',
}

function eventTypeLabel(type: string) {
  return eventTypeLabels[type] ?? type
}

function evidenceRelatedEventDay(anomaly: { evidence: { related_event?: { timestamp: string } } }) {
  return anomaly.evidence.related_event ? dayKey(anomaly.evidence.related_event.timestamp) : null
}

function formatDateTime(iso: string) {
  return new Date(iso).toLocaleString('es-ES', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function InvestigationPage() {
  const { id } = useParams<{ id: string }>()
  const anomalyId = Number(id)
  const { data: anomaly, isLoading } = useGetAnomalyQuery(anomalyId, { skip: !anomalyId })
  const { data: daily } = useGetMeterDailyQuery(anomaly?.meter_id ?? '', { skip: !anomaly })
  const [updateStatus, { isLoading: updating }] = useUpdateAnomalyStatusMutation()
  const [statusError, setStatusError] = useState<string | null>(null)

  async function handleStatusChange(id: number, status: string) {
    setStatusError(null)
    try {
      await updateStatus({ id, status }).unwrap()
    } catch (err) {
      const message =
        typeof err === 'object' && err && 'error' in err
          ? String((err as { error: unknown }).error)
          : 'No se pudo actualizar el estado. Probá de nuevo.'
      setStatusError(message)
    }
  }

  const chartData = useMemo(() => {
    if (!daily || !anomaly) return null
    const windowDays = Math.max(
      1,
      Math.round(
        (new Date(anomaly.window_end).getTime() - new Date(anomaly.window_start).getTime()) /
          86_400_000,
      ) || 1,
    )
    const avgBaselinePerDay = anomaly.baseline_kwh / windowDays

    const consumption = daily.map((d) => ({
      day: d.day,
      real: d.consumption_kwh,
      baseline: avgBaselinePerDay,
    }))
    const electrical = daily.map((d) => ({ day: d.day, voltaje: d.avg_voltage_v, pf: d.avg_power_factor }))

    const days = daily.map((d) => d.day)
    const eventDay = evidenceRelatedEventDay(anomaly)
    return {
      consumption,
      electrical,
      highlightRange: {
        x1: anomaly.window_start.slice(0, 10),
        x2: anomaly.window_end.slice(0, 10),
      },
      eventMarkers: eventDay && days.includes(eventDay) ? [{ x: eventDay, label: 'Evento' }] : [],
    }
  }, [daily, anomaly])

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="flex items-start justify-between">
          <div>
            <Skeleton className="h-4 w-24" />
            <Skeleton className="mt-2 h-7 w-56" />
            <Skeleton className="mt-2 h-4 w-40" />
          </div>
          <Skeleton className="h-5 w-16 rounded-md" />
        </div>
        <div className="grid gap-5 xl:grid-cols-[1.45fr_.75fr]">
          <div className="space-y-5">
            <Skeleton className="h-40 w-full rounded-xl" />
            <Skeleton className="h-48 w-full rounded-xl" />
            <Skeleton className="h-55 w-full rounded-xl" />
          </div>
          <div className="space-y-5">
            <Skeleton className="h-40 w-full rounded-xl" />
            <Skeleton className="h-32 w-full rounded-xl" />
            <Skeleton className="h-32 w-full rounded-xl" />
          </div>
        </div>
      </div>
    )
  }
  if (!anomaly) return <div className="text-sm text-ink-400">Anomalía no encontrada.</div>

  const evidence = anomaly.evidence

  const variableRows = [
    {
      label: 'Consumo',
      before: `${Math.round(anomaly.baseline_kwh)} kWh (baseline)`,
      after: `${Math.round(anomaly.actual_kwh)} kWh`,
      delta: `${anomaly.variation_pct > 0 ? '+' : ''}${anomaly.variation_pct.toFixed(1)}%`,
      deltaTone: 'text-red-500',
    },
    {
      label: 'Voltaje',
      before: `${evidence.voltage_actual_start_v.toFixed(1)} V`,
      after: `${evidence.voltage_actual_end_v.toFixed(1)} V`,
      delta: `${(evidence.voltage_actual_end_v - evidence.voltage_actual_start_v).toFixed(1)} V`,
      deltaTone: 'text-amber-600',
    },
    {
      label: 'Factor de potencia',
      before: evidence.power_factor_actual_start.toFixed(2),
      after: evidence.power_factor_actual_end.toFixed(2),
      delta: (evidence.power_factor_actual_end - evidence.power_factor_actual_start).toFixed(2),
      deltaTone: 'text-amber-600',
    },
  ]

  const evidenceItems: string[] = []
  if (evidence.peak_zscore > 0) {
    evidenceItems.push(`Desviación estadística fuerte (z=${evidence.peak_zscore.toFixed(1)})`)
  }
  if (evidence.onset === 'STEP') evidenceItems.push('El cambio fue abrupto, no gradual')
  if (evidence.onset === 'GRADUAL') evidenceItems.push('El cambio fue gradual, en varias horas')
  evidenceItems.push(evidence.ongoing ? 'El cambio sigue activo' : 'El cambio ya se resolvió')
  if (evidence.data_quality_flag_count > 0) {
    evidenceItems.push(
      `${evidence.data_quality_flag_count} de ${evidence.data_quality_sample_size} lecturas con voltaje/PF fuera de rango`,
    )
  }
  evidenceItems.push(
    evidence.related_event
      ? `Evento operativo registrado: "${evidence.related_event.description}"`
      : 'Sin evento operativo que explique el cambio',
  )

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-2">
        <div>
          <Link
            to="/anomalies"
            className="flex items-center gap-1.5 text-xs font-medium text-ink-400 hover:text-ink-700"
          >
            <ArrowLeft size={13} /> Anomalías IA
          </Link>
          <h1 className="mt-2 text-2xl font-semibold tracking-tight text-ink-900">
            Investigación · {anomaly.meter_id}
          </h1>
          <p className="mt-1 text-sm text-ink-400">
            {anomaly.window_start.slice(0, 10)} a {anomaly.window_end.slice(0, 10)}
          </p>
        </div>
        <Badge>{anomaly.severity}</Badge>
      </div>

      <div className="grid gap-5 xl:grid-cols-[1.45fr_.75fr]">
        <div className="space-y-5">
          <section className="rounded-xl border border-surface-border bg-white p-6">
            <div className="flex items-center justify-between">
              <div>
                <div className="flex items-center gap-2 text-xs font-semibold text-red-500">
                  <AlertTriangle size={14} /> Veredicto de la IA
                </div>
                <h2 className="mt-3 text-lg font-semibold text-ink-900">
                  <Badge>{anomaly.type}</Badge>
                </h2>
              </div>
              <div className="relative grid h-16 w-16 shrink-0 place-items-center">
                <svg viewBox="0 0 64 64" className="absolute inset-0 h-16 w-16 -rotate-90">
                  <circle cx={32} cy={32} r={26} fill="none" stroke="#e4ecec" strokeWidth={5} />
                  <circle
                    cx={32}
                    cy={32}
                    r={26}
                    fill="none"
                    stroke="#18c8b3"
                    strokeWidth={5}
                    strokeLinecap="round"
                    strokeDasharray={2 * Math.PI * 26}
                    strokeDashoffset={2 * Math.PI * 26 * (1 - anomaly.confidence)}
                  />
                </svg>
                <span className="text-lg font-semibold text-ink-900">
                  {Math.round(anomaly.confidence * 100)}%
                </span>
                <span className="absolute -bottom-5 text-[9px] font-normal text-ink-400">
                  confianza
                </span>
              </div>
            </div>
            <p className="mt-6 text-sm leading-6 text-ink-600">{anomaly.reason}</p>
          </section>

          <section className="rounded-xl border border-surface-border bg-white p-6">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <div>
                <h2 className="text-sm font-semibold text-ink-900">Variables que cambiaron</h2>
                <p className="mt-1 text-xs text-ink-400">Comparación antes vs después del cambio</p>
              </div>
              <span className="rounded-md bg-red-50 px-2 py-1 text-[10px] font-semibold text-red-600">
                {formatDateTime(anomaly.window_start)}
              </span>
            </div>

            {/* Table on wider screens, stacked cards on phones - a 4-column
                table doesn't fit a phone width without hiding a column
                behind an unsignposted horizontal scroll. */}
            <div className="mt-5 hidden overflow-x-auto sm:block">
              <table className="w-full min-w-115 text-left">
                <thead className="text-[10px] text-ink-400">
                  <tr>
                    <th className="pb-3">Variable</th>
                    <th className="pb-3">Antes</th>
                    <th className="pb-3">Después</th>
                    <th className="pb-3">Delta</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-surface-border/60 text-xs">
                  {variableRows.map((row) => (
                    <tr key={row.label}>
                      <td className="py-3 font-medium text-ink-900">{row.label}</td>
                      <td>{row.before}</td>
                      <td>{row.after}</td>
                      <td className={`font-semibold ${row.deltaTone}`}>{row.delta}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            <div className="mt-5 space-y-3 sm:hidden">
              {variableRows.map((row) => (
                <div key={row.label} className="rounded-lg border border-surface-border p-3">
                  <div className="text-xs font-semibold text-ink-900">{row.label}</div>
                  <div className="mt-2 flex items-center justify-between gap-2 text-xs text-ink-600">
                    <span>{row.before}</span>
                    <span className="text-ink-300">→</span>
                    <span>{row.after}</span>
                  </div>
                  <div className={`mt-1.5 text-right text-xs font-semibold ${row.deltaTone}`}>
                    {row.delta}
                  </div>
                </div>
              ))}
            </div>
          </section>

          <section className="rounded-xl border border-surface-border bg-white p-6">
            <div>
              <h2 className="text-sm font-semibold text-ink-900">Comparación contra baseline</h2>
              <p className="mt-1 text-xs text-ink-400">
                Consumo diario · período completo · la banda roja marca cuándo ocurrió el cambio
              </p>
            </div>
            <div className="mt-5">
              {chartData ? (
                <TimeSeriesChart
                  ariaLabel={`Consumo real vs baseline de ${anomaly.meter_id}`}
                  data={chartData.consumption}
                  xKey="day"
                  series={[
                    { key: 'real', name: 'Real', color: '#18c8b3' },
                    { key: 'baseline', name: 'Baseline (promedio)', color: '#9dc5c1' },
                  ]}
                  formatX={(x) => String(x).slice(5)}
                  formatY={(y) => `${Math.round(Number(y))} kWh`}
                  highlightRange={chartData.highlightRange}
                  eventMarkers={chartData.eventMarkers}
                />
              ) : (
                <Skeleton className="h-55 w-full rounded-lg" />
              )}
            </div>
          </section>

          <section className="rounded-xl border border-surface-border bg-white p-6">
            <div>
              <h2 className="text-sm font-semibold text-ink-900">Voltaje y factor de potencia</h2>
              <p className="mt-1 text-xs text-ink-400">Promedio diario · misma ventana</p>
            </div>
            {chartData ? (
              <div className="mt-5 grid gap-5 sm:grid-cols-2">
                <div>
                  <p className="text-[10px] font-medium text-ink-400">Voltaje (V)</p>
                  <TimeSeriesChart
                    ariaLabel={`Voltaje diario de ${anomaly.meter_id}`}
                    data={chartData.electrical}
                    xKey="day"
                    series={[{ key: 'voltaje', name: 'Voltaje', color: '#18c8b3' }]}
                    formatX={(x) => String(x).slice(5)}
                    formatY={(y) => `${Number(y).toFixed(0)} V`}
                    highlightRange={chartData.highlightRange}
                    height={140}
                  />
                </div>
                <div>
                  <p className="text-[10px] font-medium text-ink-400">Factor de potencia</p>
                  <TimeSeriesChart
                    ariaLabel={`Factor de potencia diario de ${anomaly.meter_id}`}
                    data={chartData.electrical}
                    xKey="day"
                    series={[{ key: 'pf', name: 'Factor de potencia', color: '#0d8577' }]}
                    formatX={(x) => String(x).slice(5)}
                    formatY={(y) => Number(y).toFixed(2)}
                    highlightRange={chartData.highlightRange}
                    height={140}
                  />
                </div>
              </div>
            ) : (
              <div className="mt-5 grid gap-5 sm:grid-cols-2">
                <Skeleton className="h-35 w-full rounded-lg" />
                <Skeleton className="h-35 w-full rounded-lg" />
              </div>
            )}
          </section>
        </div>

        <aside className="space-y-5">
          <section className="rounded-xl border border-surface-border bg-white p-5">
            <h2 className="text-sm font-semibold text-ink-900">Evidencia</h2>
            <div className="mt-4 space-y-3">
              {evidenceItems.map((item) => (
                <div key={item} className="flex items-start gap-2.5 text-xs text-ink-600">
                  <div className="-mt-px grid h-4 w-4 shrink-0 place-items-center rounded-full bg-brand-50 text-brand-700">
                    <Check size={10} />
                  </div>
                  {item}
                </div>
              ))}
            </div>
          </section>

          <section className="rounded-xl border border-amber-100 bg-amber-50/50 p-5">
            <div className="flex items-center gap-2 text-xs font-semibold text-amber-700">
              <CircleHelp size={14} /> Eventos relacionados
            </div>
            {evidence.related_event ? (
              <div className="mt-4 rounded-lg border border-amber-100 bg-white/70 p-3">
                <div className="text-xs font-semibold text-ink-900">
                  {eventTypeLabel(evidence.related_event.type)}
                </div>
                <div className="mt-1 text-[10px] text-ink-500">
                  {evidence.related_event.description}
                </div>
              </div>
            ) : (
              <p className="mt-3 text-[11px] leading-4 text-amber-800">
                Ningún evento operativo explica este cambio.
              </p>
            )}
          </section>

          <section className="rounded-xl border border-surface-border bg-white p-5">
            <h2 className="text-sm font-semibold text-ink-900">Acción recomendada</h2>
            <p className="mt-3 text-xs leading-5 text-ink-500">{anomaly.recommended_action}</p>
            <button
              disabled={updating}
              onClick={() => handleStatusChange(anomaly.id, 'IN_REVIEW')}
              className="mt-4 w-full rounded-lg bg-ink-900 py-2.5 text-xs font-semibold text-white disabled:opacity-50"
            >
              {anomaly.status === 'IN_REVIEW' ? 'En revisión' : 'Marcar en revisión'}
            </button>
            <button
              disabled={updating}
              onClick={() => handleStatusChange(anomaly.id, 'DISMISSED')}
              className="mt-2 w-full rounded-lg border border-surface-border py-2.5 text-xs font-semibold text-ink-700 disabled:opacity-50"
            >
              {anomaly.status === 'DISMISSED' ? 'Descartada' : 'Descartar anomalía'}
            </button>
            {statusError && (
              <p className="mt-2 text-[11px] leading-4 text-red-600">{statusError}</p>
            )}
          </section>
        </aside>
      </div>
    </div>
  )
}
