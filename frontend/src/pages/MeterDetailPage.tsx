import { ArrowLeft, Gauge, Activity, Zap as ZapIcon } from 'lucide-react'
import { useMemo } from 'react'
import { Link, useParams } from 'react-router-dom'
import { Badge } from '../components/Badge'
import { LineChart } from '../components/LineChart'
import { StatTile } from '../components/StatTile'
import {
  useGetMeterQuery,
  useGetMeterReadingsQuery,
  useListAnomaliesQuery,
} from '../api/apiSlice'

function dayKey(iso: string) {
  return iso.slice(0, 10)
}

export function MeterDetailPage() {
  const { meterId = '' } = useParams<{ meterId: string }>()
  const { data: meter, isLoading: meterLoading } = useGetMeterQuery(meterId)
  const { data: readings, isLoading: readingsLoading } = useGetMeterReadingsQuery(meterId)
  const { data: anomalies } = useListAnomaliesQuery()

  const anomaly = anomalies?.find((a) => a.meter_id === meterId)

  const dailyConsumption = useMemo(() => {
    if (!readings) return []
    const byDay = new Map<string, number>()
    for (const r of readings) {
      const key = dayKey(r.timestamp)
      byDay.set(key, (byDay.get(key) ?? 0) + r.consumption_kwh)
    }
    return Array.from(byDay.entries())
      .sort(([a], [b]) => a.localeCompare(b))
      .map(([day, total], i) => ({ x: i, day, y: total }))
  }, [readings])

  if (meterLoading || readingsLoading) {
    return <div className="text-sm text-ink-400">Cargando…</div>
  }

  if (!meter) {
    return <div className="text-sm text-ink-400">Medidor no encontrado.</div>
  }

  return (
    <div className="space-y-6">
      <div>
        <Link
          to="/meters"
          className="flex items-center gap-1.5 text-xs font-medium text-ink-400 hover:text-ink-700"
        >
          <ArrowLeft size={13} /> Medidores
        </Link>
        <div className="mt-2 flex items-start justify-between">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight text-ink-900">{meter.meter_id}</h1>
            <p className="mt-1 text-sm text-ink-400">{meter.location}</p>
          </div>
          {anomaly && <Badge>{anomaly.severity}</Badge>}
        </div>
      </div>

      <div className="grid grid-cols-2 gap-3 md:grid-cols-4">
        <StatTile
          label="Consumo total"
          value={`${meter.consumption_kwh.toLocaleString('es-ES', { maximumFractionDigits: 0 })} kWh`}
          icon={<ZapIcon size={13} className="text-brand-600" />}
        />
        <StatTile
          label="Voltaje promedio"
          value={`${meter.avg_voltage_v.toFixed(1)} V`}
          icon={<Activity size={13} className="text-brand-600" />}
        />
        <StatTile label="Corriente promedio" value={`${meter.avg_current_a.toFixed(1)} A`} />
        <StatTile
          label="Factor de potencia"
          value={meter.avg_power_factor.toFixed(2)}
          icon={<Gauge size={13} className="text-brand-600" />}
        />
      </div>

      {anomaly && (
        <div className="rounded-xl border border-amber-100 bg-amber-50/50 p-5">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2 text-xs font-semibold text-amber-800">
              <Badge>{anomaly.type}</Badge>
              {Math.round(anomaly.confidence * 100)}% confianza
            </div>
            <Link to="/anomalies" className="text-xs font-semibold text-brand-700 hover:underline">
              Ver investigación
            </Link>
          </div>
          <p className="mt-3 text-sm leading-6 text-ink-700">{anomaly.reason}</p>
          <p className="mt-3 text-xs font-medium text-ink-600">
            Acción recomendada: {anomaly.recommended_action}
          </p>
        </div>
      )}

      <div className="rounded-xl border border-surface-border bg-white p-5">
        <div>
          <h2 className="text-sm font-semibold text-ink-900">Consumo diario</h2>
          <p className="mt-1 text-xs text-ink-400">Últimos 14 días · {readings?.length ?? 0} lecturas</p>
        </div>
        <div className="mt-4">
          {dailyConsumption.length > 0 ? (
            <LineChart
              ariaLabel={`Consumo diario de ${meter.meter_id}`}
              series={[{ name: 'Consumo', color: '#18c8b3', points: dailyConsumption }]}
              formatX={(x) => {
                const point = dailyConsumption[x]
                return point ? point.day.slice(5) : ''
              }}
              formatY={(y) => `${Math.round(y)} kWh`}
            />
          ) : (
            <p className="text-sm text-ink-400">Sin lecturas para este medidor.</p>
          )}
        </div>
      </div>
    </div>
  )
}
