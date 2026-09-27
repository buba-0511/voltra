const styles: Record<string, string> = {
  HIGH: 'bg-red-50 text-red-600 border-red-100',
  MEDIUM: 'bg-amber-50 text-amber-700 border-amber-100',
  LOW: 'bg-slate-100 text-slate-600 border-slate-200',
  OK: 'bg-emerald-50 text-emerald-600 border-emerald-100',
  ALERT: 'bg-amber-50 text-amber-700 border-amber-100',
  CRITICAL: 'bg-red-50 text-red-600 border-red-100',
  REAL_ANOMALY: 'bg-red-50 text-red-600 border-red-100',
  EXPLAINABLE_ANOMALY: 'bg-amber-50 text-amber-700 border-amber-100',
  FALSE_POSITIVE: 'bg-slate-100 text-slate-600 border-slate-200',
  DATA_QUALITY: 'bg-violet-50 text-violet-600 border-violet-100',
}

const typeLabels: Record<string, string> = {
  REAL_ANOMALY: 'Anomalía real',
  EXPLAINABLE_ANOMALY: 'Explicable',
  FALSE_POSITIVE: 'Falso positivo',
  DATA_QUALITY: 'Calidad de datos',
}

export function Badge({ children }: { children: string }) {
  const label = typeLabels[children] ?? children
  const style = styles[children.toUpperCase()] ?? 'bg-slate-50 text-slate-600 border-slate-200'
  return (
    <span
      className={`inline-flex items-center rounded-md border px-2 py-1 text-[10px] font-semibold ${style}`}
    >
      {label}
    </span>
  )
}
