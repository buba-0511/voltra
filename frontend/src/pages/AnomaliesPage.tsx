import { Check, Clock, Sparkles, X } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Badge } from '../components/Badge'
import { Skeleton } from '../components/Skeleton'
import { useListAnomaliesQuery, useRunAnalysisMutation } from '../api/apiSlice'
import type { AnalysisRun } from '../api/types'

const STAGES = [
  'Lecturas',
  'Baseline',
  'Detección',
  'Correlación',
  'Eventos',
  'Explicación',
  'Recomendación',
]
const STAGE_INTERVAL_MS = 420

function AnalysisModal({ close }: { close: () => void }) {
  const [runAnalysis] = useRunAnalysisMutation()
  const [stage, setStage] = useState(0)
  const [result, setResult] = useState<AnalysisRun | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    runAnalysis()
      .unwrap()
      .then(setResult)
      .catch(() => setError('El análisis falló. Probá de nuevo.'))
  }, [runAnalysis])

  useEffect(() => {
    if (stage >= STAGES.length) return
    const t = setTimeout(() => setStage((s) => s + 1), STAGE_INTERVAL_MS)
    return () => clearTimeout(t)
  }, [stage])

  // Only "done" once both the staged reveal finished AND the real
  // response came back - never show a result before it's actually real.
  const done = stage >= STAGES.length && result !== null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-ink-900/50 p-4">
      <div className="w-full max-w-lg rounded-2xl bg-white p-6 shadow-2xl">
        <div className="flex items-start justify-between">
          <div>
            <div className="flex items-center gap-2 text-xs font-semibold text-brand-700">
              <Sparkles size={15} /> Análisis IA
            </div>
            <h2 className="mt-2 text-xl font-semibold text-ink-900">
              Analizando tu operación
            </h2>
            <p className="mt-1 text-xs text-ink-400">
              Buscando señales en las lecturas de los últimos 14 días.
            </p>
          </div>
          <button onClick={close} className="text-ink-400 hover:text-ink-700">
            <X size={19} />
          </button>
        </div>

        <div className="mt-7 space-y-3">
          {STAGES.map((s, i) => (
            <div key={s} className="flex items-center gap-3">
              <div
                className={`grid h-7 w-7 shrink-0 place-items-center rounded-full text-[11px] font-semibold ${
                  i < stage
                    ? 'bg-brand-50 text-brand-700'
                    : i === stage
                      ? 'bg-ink-900 text-white'
                      : 'bg-slate-100 text-slate-400'
                }`}
              >
                {i < stage ? <Check size={13} /> : i + 1}
              </div>
              <div className="flex-1">
                <div
                  className={`text-xs font-semibold ${i <= stage ? 'text-ink-900' : 'text-ink-300'}`}
                >
                  {s}
                </div>
              </div>
              {i < stage && <span className="text-[10px] text-emerald-600">Completado</span>}
            </div>
          ))}
        </div>

        {error && <p className="mt-6 text-sm text-red-600">{error}</p>}

        {done && result && (
          <div className="mt-7 rounded-lg bg-brand-50 p-4 text-center text-sm font-semibold text-brand-700">
            {result.anomalies_found} anomalías detectadas
            {result.high_priority > 0 && ` · ${result.high_priority} requieren atención prioritaria`}
          </div>
        )}

        <button
          onClick={close}
          disabled={!done && !error}
          className="mt-3 w-full rounded-lg bg-ink-900 py-3 text-xs font-semibold text-white disabled:opacity-40"
        >
          {done || error ? 'Ver anomalías' : 'Procesando…'}
        </button>
      </div>
    </div>
  )
}

export function AnomaliesPage() {
  const navigate = useNavigate()
  const [showModal, setShowModal] = useState(false)
  const { data: anomalies, isLoading } = useListAnomaliesQuery()

  const avgConfidence = anomalies?.length
    ? Math.round((anomalies.reduce((sum, a) => sum + a.confidence, 0) / anomalies.length) * 100)
    : null
  const highPriority = anomalies?.filter((a) => a.severity === 'HIGH').length ?? 0

  return (
    <div className="flex h-full flex-col gap-6">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight text-ink-900">Anomalías IA</h1>
          <p className="mt-1 text-sm text-ink-400">
            {anomalies && anomalies.length > 0
              ? `De ${anomalies.length} señales, ${highPriority} necesitan una acción inmediata.`
              : 'Corré un análisis para detectar anomalías en el período.'}
          </p>
        </div>
        <button
          onClick={() => setShowModal(true)}
          className="flex items-center gap-2 rounded-lg bg-ink-900 px-4 py-2.5 text-xs font-semibold text-white hover:bg-ink-800"
        >
          <Sparkles size={14} className="text-brand-400" /> Run AI Analysis
        </button>
      </div>

      {anomalies && anomalies.length > 0 && (
        <div className="grid gap-4 md:grid-cols-2">
          <div className="rounded-xl border border-surface-border bg-white p-5">
            <div className="text-xs text-ink-400">Confianza promedio</div>
            <div className="mt-3 text-2xl font-semibold text-ink-900">{avgConfidence}%</div>
          </div>
          <div className="rounded-xl border border-surface-border bg-white p-5">
            <div className="flex items-center gap-2 text-xs text-ink-400">
              <Clock size={13} /> Alta prioridad
            </div>
            <div className="mt-3 text-2xl font-semibold text-ink-900">{highPriority}</div>
          </div>
        </div>
      )}

      <div className="min-h-0 flex-1 overflow-hidden rounded-xl border border-surface-border bg-white">
        {isLoading ? (
          <div className="h-full overflow-auto">
            <table className="w-full min-w-180 text-left">
              <thead className="sticky top-0 z-10">
                <tr className="border-b border-surface-border bg-surface text-[10px] font-semibold text-ink-400">
                  <th className="px-5 py-3">Medidor / tipo</th>
                  <th className="px-4 py-3">Severidad</th>
                  <th className="px-4 py-3">Confianza</th>
                  <th className="px-4 py-3">Acción</th>
                  <th className="px-4 py-3">Estado</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-surface-border/60">
                {Array.from({ length: 5 }).map((_, i) => (
                  <tr key={i}>
                    <td className="px-5 py-4">
                      <Skeleton className="h-3 w-16" />
                      <Skeleton className="mt-2 h-2 w-20" />
                    </td>
                    <td className="px-4 py-4">
                      <Skeleton className="h-5 w-16 rounded-md" />
                    </td>
                    <td className="px-4 py-4">
                      <Skeleton className="h-2 w-24" />
                    </td>
                    <td className="px-4 py-4">
                      <Skeleton className="h-3 w-28" />
                    </td>
                    <td className="px-4 py-4">
                      <Skeleton className="h-5 w-16 rounded-md" />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : !anomalies || anomalies.length === 0 ? (
          <div className="flex flex-col items-center gap-3 px-5 py-14 text-center">
            <Sparkles className="text-ink-300" />
            <p className="text-sm text-ink-500">Todavía no corriste un análisis.</p>
            <button
              onClick={() => setShowModal(true)}
              className="text-xs font-semibold text-brand-600 hover:underline"
            >
              Correr el primer análisis
            </button>
          </div>
        ) : (
          <div className="h-full overflow-auto">
            <table className="w-full min-w-180 text-left">
              <thead className="sticky top-0 z-10">
                <tr className="border-b border-surface-border bg-surface text-[10px] font-semibold text-ink-400">
                  <th className="px-5 py-3">Medidor / tipo</th>
                  <th className="px-4 py-3">Severidad</th>
                  <th className="px-4 py-3">Confianza</th>
                  <th className="px-4 py-3">Acción</th>
                  <th className="px-4 py-3">Estado</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-surface-border/60">
                {anomalies.map((a) => (
                  <tr
                    key={a.id}
                    onClick={() => navigate(`/anomalies/${a.id}`)}
                    className={`cursor-pointer hover:bg-surface ${
                      a.status === 'DISMISSED' || a.status === 'RESOLVED' ? 'opacity-50' : ''
                    }`}
                  >
                    <td className="px-5 py-4">
                      <span className="text-xs font-semibold text-ink-900">{a.meter_id}</span>
                      <div className="mt-1 text-[10px] text-ink-400">
                        <Badge>{a.type}</Badge>
                      </div>
                    </td>
                    <td className="px-4 py-4">
                      <Badge>{a.severity}</Badge>
                    </td>
                    <td className="px-4 py-4">
                      <div className="flex items-center gap-2">
                        <div className="h-1.5 w-16 rounded-full bg-slate-100">
                          <div
                            className="h-full rounded-full bg-brand-500"
                            style={{ width: `${Math.round(a.confidence * 100)}%` }}
                          />
                        </div>
                        <span className="text-[11px] font-semibold tabular-nums text-ink-900">
                          {Math.round(a.confidence * 100)}%
                        </span>
                      </div>
                    </td>
                    <td className="px-4 py-4">
                      <span className="text-[11px] font-semibold text-brand-600">
                        {a.recommended_action.slice(0, 40)}
                        {a.recommended_action.length > 40 ? '…' : ''}
                      </span>
                    </td>
                    <td className="px-4 py-4">
                      <Badge>{a.status}</Badge>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {showModal && <AnalysisModal close={() => setShowModal(false)} />}
    </div>
  )
}
