import { AlertTriangle, ArrowUpRight, Gauge, Sparkles } from 'lucide-react'
import { Link } from 'react-router-dom'
import { Badge } from '../components/Badge'
import { StatTile } from '../components/StatTile'
import { useDashboardSummaryQuery, useListAnomaliesQuery } from '../api/apiSlice'

function formatKwh(n: number) {
  return `${n.toLocaleString('es-ES', { maximumFractionDigits: 0 })} kWh`
}

function formatDateTime(iso: string) {
  return new Date(iso).toLocaleString('es-ES', {
    day: '2-digit',
    month: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function DashboardPage() {
  const { data: summary, isLoading: summaryLoading } = useDashboardSummaryQuery()
  const { data: anomalies, isLoading: anomaliesLoading } = useListAnomaliesQuery()

  const topAnomaly = anomalies?.[0]
  const avgConfidence = anomalies?.length
    ? Math.round((anomalies.reduce((sum, a) => sum + a.confidence, 0) / anomalies.length) * 100)
    : null

  if (summaryLoading) {
    return <div className="text-sm text-ink-400">Cargando…</div>
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight text-ink-900">Dashboard</h1>
        <p className="mt-1 text-sm text-ink-400">
          Estado energético de tu operación.
        </p>
      </div>

      <div className="grid grid-cols-2 gap-3 lg:grid-cols-3 xl:grid-cols-6">
        <StatTile label="Medidores" value={summary?.meters_count ?? '—'} sub="Activos" />
        <StatTile
          label="Consumo total"
          value={formatKwh(summary?.total_consumption_kwh ?? 0)}
          sub="Período completo"
        />
        <StatTile
          label="Anomalías IA"
          value={summary?.anomalies_detected ?? 0}
          sub="Detectadas"
          icon={<Sparkles size={13} className="text-brand-600" />}
        />
        <StatTile
          label="Alta prioridad"
          value={summary?.high_priority ?? 0}
          sub="Requieren atención"
          subTone={summary?.high_priority ? 'danger' : 'default'}
        />
        <StatTile
          label="Confianza IA"
          value={avgConfidence !== null ? `${avgConfidence}%` : '—'}
          sub="Promedio"
        />
        <StatTile
          label="Último análisis"
          value={summary?.last_analysis ? formatDateTime(summary.last_analysis.started_at) : '—'}
          sub={summary?.last_analysis ? summary.last_analysis.status : 'Sin correr'}
        />
      </div>

      {topAnomaly && (
        <div className="rounded-xl bg-ink-900 p-6 text-white shadow-sm">
          <div className="flex items-start justify-between">
            <div>
              <div className="flex items-center gap-2 text-xs text-ink-300">
                <span className="h-2 w-2 rounded-full bg-red-400" />
                REQUIERE ATENCIÓN
              </div>
              <h2 className="mt-3 text-xl font-semibold">
                {topAnomaly.meter_id}{' '}
                <span className="font-normal text-ink-300">
                  · {topAnomaly.variation_pct > 0 ? '+' : ''}
                  {topAnomaly.variation_pct.toFixed(1)}% vs baseline
                </span>
              </h2>
              <p className="mt-2 max-w-2xl text-sm leading-6 text-ink-200">{topAnomaly.reason}</p>
            </div>
            <div className="rounded-lg bg-ink-500 p-2 text-brand-400">
              <AlertTriangle size={18} />
            </div>
          </div>
          <div className="mt-6 flex flex-wrap items-center gap-2">
            <Badge>{topAnomaly.severity}</Badge>
            <span className="rounded-md bg-white/10 px-2 py-1 text-[10px] font-semibold text-white">
              {Math.round(topAnomaly.confidence * 100)}% confianza
            </span>
            <Link
              to="/anomalies"
              className="ml-2 flex items-center gap-1 text-xs font-semibold text-brand-400 hover:text-white"
            >
              Investigar <ArrowUpRight size={14} />
            </Link>
          </div>
        </div>
      )}

      <div className="rounded-xl border border-surface-border bg-white">
        <div className="flex items-center justify-between border-b border-surface-border/60 px-5 py-4">
          <div>
            <h2 className="text-sm font-semibold text-ink-900">Anomalías detectadas</h2>
            <p className="mt-1 text-xs text-ink-400">Priorizadas por severidad y confianza</p>
          </div>
          <Link to="/anomalies" className="text-xs font-semibold text-brand-600 hover:text-brand-700">
            Ver todas <ArrowUpRight className="ml-1 inline" size={13} />
          </Link>
        </div>
        {anomaliesLoading ? (
          <p className="px-5 py-6 text-sm text-ink-400">Cargando…</p>
        ) : anomalies && anomalies.length > 0 ? (
          <div className="divide-y divide-surface-border/60">
            {anomalies.map((a) => (
              <div key={a.id} className="flex items-center gap-3 px-5 py-3.5">
                <div
                  className={`grid h-8 w-8 shrink-0 place-items-center rounded-lg ${
                    a.severity === 'HIGH' ? 'bg-red-50 text-red-500' : 'bg-amber-50 text-amber-600'
                  }`}
                >
                  <AlertTriangle size={15} />
                </div>
                <div className="min-w-0 flex-1">
                  <div className="text-xs font-semibold text-ink-900">
                    {a.meter_id} <span className="font-normal text-ink-400">· {a.type}</span>
                  </div>
                  <div className="mt-1 truncate text-[11px] text-ink-400">{a.reason}</div>
                </div>
                <Badge>{a.severity}</Badge>
                <span className="w-10 text-right text-xs font-semibold tabular-nums text-ink-600">
                  {Math.round(a.confidence * 100)}%
                </span>
              </div>
            ))}
          </div>
        ) : (
          <div className="flex flex-col items-center gap-2 px-5 py-10 text-center">
            <Gauge className="text-ink-300" />
            <p className="text-sm text-ink-500">Todavía no corriste un análisis.</p>
            <Link to="/anomalies" className="text-xs font-semibold text-brand-600">
              Ir a Anomalías IA <ArrowUpRight className="ml-1 inline" size={12} />
            </Link>
          </div>
        )}
      </div>
    </div>
  )
}
