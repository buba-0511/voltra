import {
  CartesianGrid,
  Legend,
  Line,
  LineChart as RechartsLineChart,
  ReferenceArea,
  ReferenceLine,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'

export interface TimeSeries {
  key: string
  name: string
  color: string
}

interface HighlightRange {
  x1: number | string
  x2: number | string
  label?: string
}

interface EventMarker {
  x: number | string
  label?: string
}

interface TimeSeriesChartProps<T> {
  data: T[]
  xKey: Extract<keyof T, string>
  series: TimeSeries[]
  height?: number
  formatX?: (x: number | string) => string
  formatY?: (y: number) => string
  hideLegend?: boolean
  highlightRange?: HighlightRange
  eventMarkers?: EventMarker[]
  ariaLabel: string
}

function ChartTooltip({
  active,
  payload,
  label,
  formatX,
  formatY,
  series,
}: {
  active?: boolean
  payload?: { dataKey: string | number; value: number }[]
  label?: number | string
  formatX: (x: number | string) => string
  formatY: (y: number) => string
  series: TimeSeries[]
}) {
  if (!active || !payload?.length) return null
  return (
    <div className="rounded-lg border border-surface-border bg-white px-3 py-2 text-xs shadow-lg">
      <div className="mb-1 font-medium text-ink-900">{formatX(label ?? '')}</div>
      {payload.map((p) => {
        const s = series.find((s) => s.key === p.dataKey)
        if (!s) return null
        return (
          <div key={s.key} className="flex items-center gap-1.5 text-ink-500">
            <span className="h-0.5 w-3 shrink-0 rounded-full" style={{ backgroundColor: s.color }} />
            <span className="font-semibold text-ink-900">{formatY(p.value)}</span>
            {series.length > 1 && <span>{s.name}</span>}
          </div>
        )
      })}
    </div>
  )
}

export function TimeSeriesChart<T extends object>({
  data,
  xKey,
  series,
  height = 220,
  formatX = String,
  formatY = String,
  hideLegend = false,
  highlightRange,
  eventMarkers,
  ariaLabel,
}: TimeSeriesChartProps<T>) {
  return (
    <div role="img" aria-label={ariaLabel}>
      <ResponsiveContainer width="100%" height={height}>
        <RechartsLineChart data={data} margin={{ top: 24, right: 8, bottom: 0, left: 0 }}>
          <CartesianGrid stroke="#e4ecec" vertical={false} />
          <XAxis
            dataKey={xKey}
            tickFormatter={formatX}
            axisLine={false}
            tickLine={false}
            tick={{ fontSize: 9, fill: '#7d9b9c' }}
          />
          <YAxis
            domain={[0, (max: number) => max * 1.1]}
            tickFormatter={formatY}
            axisLine={false}
            tickLine={false}
            width={44}
            tick={{ fontSize: 9, fill: '#7d9b9c' }}
          />
          <Tooltip
            content={<ChartTooltip formatX={formatX} formatY={formatY} series={series} />}
            cursor={{ stroke: '#9dc5c1', strokeDasharray: '3 3' }}
          />
          {!hideLegend && series.length > 1 && (
            <Legend
              iconType="plainline"
              wrapperStyle={{ fontSize: 12, color: '#214c50' }}
              formatter={(value) => <span className="text-ink-500">{value}</span>}
            />
          )}
          {highlightRange && (
            <ReferenceArea
              x1={highlightRange.x1}
              x2={highlightRange.x2}
              fill="#ef4444"
              fillOpacity={0.08}
              label={
                highlightRange.label
                  ? { value: highlightRange.label, position: 'insideTop', fontSize: 9, fill: '#ef4444' }
                  : undefined
              }
            />
          )}
          {eventMarkers?.map((m) => (
            <ReferenceLine
              key={String(m.x)}
              x={m.x}
              stroke="#d97706"
              strokeDasharray="3 3"
              label={m.label ? { value: m.label, position: 'top', fontSize: 9, fill: '#d97706' } : undefined}
            />
          ))}
          {series.map((s) => (
            <Line
              key={s.key}
              type="monotone"
              dataKey={s.key}
              name={s.name}
              stroke={s.color}
              strokeWidth={2}
              dot={false}
              activeDot={{ r: 4, strokeWidth: 2, stroke: '#fff' }}
              isAnimationActive={false}
              connectNulls
            />
          ))}
        </RechartsLineChart>
      </ResponsiveContainer>
    </div>
  )
}
