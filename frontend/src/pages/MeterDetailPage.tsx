import {
  Activity,
  ArrowLeft,
  Gauge,
  Percent,
  TrendingUp,
  Waves,
  Zap as ZapIcon,
} from 'lucide-react'
import { Link, useParams } from 'react-router-dom'
import { Badge } from '../components/Badge'
import { Skeleton } from '../components/Skeleton'
import { TimeSeriesChart } from '../components/TimeSeriesChart'
import { StatTile } from '../components/StatTile'
import { useGetMeterDailyQuery, useGetMeterQuery, useListAnomaliesQuery } from '../api/apiSlice'
import { meterStatus } from '../utils/meterStatus'

export function MeterDetailPage() {
  const { meterId = '' } = useParams<{ meterId: string }>()
  const { data: meter, isLoading: meterLoading } = useGetMeterQuery(meterId)
  const { data: daily, isLoading: dailyLoading } = useGetMeterDailyQuery(meterId)
  const { data: anomalies } = useListAnomaliesQuery()

  const anomaly = anomalies?.find((a) => a.meter_id === meterId)

  if (meterLoading || dailyLoading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-4 w-24" />
        <div>
          <Skeleton className="h-7 w-32" />
          <Skeleton className="mt-2 h-4 w-48" />
        </div>
        <div className="grid grid-cols-2 gap-3 md:grid-cols-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className="rounded-xl border border-surface-border bg-white p-4">
              <Skeleton className="h-3 w-20" />
              <Skeleton className="mt-3 h-6 w-16" />
            </div>
          ))}
        </div>
        <Skeleton className="h-64 w-full rounded-xl" />
      </div>
    )
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
          label="Consumo actual"
          value={`${meter.consumption_kwh.toLocaleString('es-ES', { maximumFractionDigits: 0 })} kWh`}
          icon={<ZapIcon size={13} className="text-brand-600" />}
        />
        <StatTile
          label="Baseline"
          value={`${Math.round(meter.baseline_kwh).toLocaleString('es-ES')} kWh`}
          sub="Esperado para el período"
          icon={<TrendingUp size={13} className="text-brand-600" />}
        />
        <StatTile
          label="Variación"
          value={`${meter.variation_pct > 0 ? '+' : ''}${meter.variation_pct.toFixed(1)}%`}
          subTone={Math.abs(meter.variation_pct) > 20 ? 'danger' : 'default'}
          icon={<Percent size={13} className="text-brand-600" />}
        />
        <StatTile
          label="Estado"
          value={<Badge>{meterStatus(anomaly?.severity)}</Badge>}
          icon={<Gauge size={13} className="text-brand-600" />}
        />
      </div>

      <div className="grid grid-cols-2 gap-3 md:grid-cols-3">
        <StatTile
          label="Voltaje promedio"
          value={`${meter.avg_voltage_v.toFixed(1)} V`}
          icon={<Activity size={13} className="text-brand-600" />}
        />
        <StatTile
          label="Corriente promedio"
          value={`${meter.avg_current_a.toFixed(1)} A`}
          icon={<Waves size={13} className="text-brand-600" />}
        />
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
            <Link
              to={`/anomalies/${anomaly.id}`}
              className="text-xs font-semibold text-brand-700 hover:underline"
            >
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
          <p className="mt-1 text-xs text-ink-400">Últimos {daily?.length ?? 0} días</p>
        </div>
        <div className="mt-4">
          {daily && daily.length > 0 ? (
            <TimeSeriesChart
              ariaLabel={`Consumo diario de ${meter.meter_id}`}
              data={daily}
              xKey="day"
              series={[{ key: 'consumption_kwh', name: 'Consumo', color: '#18c8b3' }]}
              formatX={(x) => String(x).slice(5)}
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
