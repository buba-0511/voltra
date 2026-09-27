import { useMemo, useRef, useState } from 'react'

export interface ChartSeries {
  name: string
  color: string
  points: { x: number; y: number }[]
}

interface LineChartProps {
  series: ChartSeries[]
  height?: number
  formatX?: (x: number) => string
  formatY?: (y: number) => string
  ariaLabel: string
}

const PADDING = { top: 12, right: 12, bottom: 24, left: 44 }

export function LineChart({
  series,
  height = 220,
  formatX = String,
  formatY = String,
  ariaLabel,
}: LineChartProps) {
  const svgRef = useRef<SVGSVGElement>(null)
  const [hoverIdx, setHoverIdx] = useState<number | null>(null)
  const width = 600 // viewBox width; scales to container via CSS

  const allPoints = series.flatMap((s) => s.points)
  const xs = allPoints.map((p) => p.x)
  const ys = allPoints.map((p) => p.y)
  const xMin = Math.min(...xs)
  const xMax = Math.max(...xs)
  const yMin = 0 // consumption/voltage/etc. never meaningfully negative here
  const yMax = Math.max(...ys) * 1.08 || 1

  const plotW = width - PADDING.left - PADDING.right
  const plotH = height - PADDING.top - PADDING.bottom

  const scaleX = (x: number) =>
    PADDING.left + ((x - xMin) / (xMax - xMin || 1)) * plotW
  const scaleY = (y: number) =>
    PADDING.top + plotH - ((y - yMin) / (yMax - yMin || 1)) * plotH

  const linePath = (points: { x: number; y: number }[]) =>
    points.map((p, i) => `${i === 0 ? 'M' : 'L'}${scaleX(p.x)},${scaleY(p.y)}`).join(' ')

  const yTicks = useMemo(() => {
    const step = yMax / 4
    return Array.from({ length: 5 }, (_, i) => step * i)
  }, [yMax])

  const xTicks = useMemo(() => {
    const base = series[0]?.points ?? []
    if (base.length <= 1) return base
    const count = Math.min(5, base.length)
    return Array.from({ length: count }, (_, i) =>
      base[Math.round((i / (count - 1)) * (base.length - 1))],
    )
  }, [series])

  function handleMove(e: React.PointerEvent<SVGSVGElement>) {
    const svg = svgRef.current
    if (!svg) return
    const rect = svg.getBoundingClientRect()
    const px = ((e.clientX - rect.left) / rect.width) * width
    const base = series[0]?.points ?? []
    if (base.length === 0) return
    let closest = 0
    let closestDist = Infinity
    base.forEach((p, i) => {
      const d = Math.abs(scaleX(p.x) - px)
      if (d < closestDist) {
        closestDist = d
        closest = i
      }
    })
    setHoverIdx(closest)
  }

  const hoverX = hoverIdx !== null ? series[0]?.points[hoverIdx]?.x : null

  return (
    <div className="relative">
      <svg
        ref={svgRef}
        viewBox={`0 0 ${width} ${height}`}
        className="w-full touch-none"
        role="img"
        aria-label={ariaLabel}
        onPointerMove={handleMove}
        onPointerLeave={() => setHoverIdx(null)}
      >
        {/* gridlines */}
        {yTicks.map((t) => (
          <line
            key={t}
            x1={PADDING.left}
            x2={width - PADDING.right}
            y1={scaleY(t)}
            y2={scaleY(t)}
            stroke="#e4ecec"
            strokeWidth={1}
          />
        ))}

        {/* y-axis labels */}
        {yTicks.map((t) => (
          <text
            key={t}
            x={PADDING.left - 8}
            y={scaleY(t)}
            textAnchor="end"
            dominantBaseline="middle"
            className="fill-ink-400 text-[9px]"
          >
            {formatY(t)}
          </text>
        ))}

        {/* x-axis labels */}
        {xTicks.map((p, i) => (
          <text
            key={i}
            x={scaleX(p.x)}
            y={height - 6}
            textAnchor="middle"
            className="fill-ink-400 text-[9px]"
          >
            {formatX(p.x)}
          </text>
        ))}

        {/* crosshair */}
        {hoverX !== null && (
          <line
            x1={scaleX(hoverX)}
            x2={scaleX(hoverX)}
            y1={PADDING.top}
            y2={PADDING.top + plotH}
            stroke="#9dc5c1"
            strokeWidth={1}
            strokeDasharray="3 3"
          />
        )}

        {/* series lines */}
        {series.map((s) => (
          <path
            key={s.name}
            d={linePath(s.points)}
            fill="none"
            stroke={s.color}
            strokeWidth={2}
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        ))}

        {/* hover dots */}
        {hoverIdx !== null &&
          series.map((s) => {
            const p = s.points[hoverIdx]
            if (!p) return null
            return (
              <circle
                key={s.name}
                cx={scaleX(p.x)}
                cy={scaleY(p.y)}
                r={4}
                fill={s.color}
                stroke="#fff"
                strokeWidth={2}
              />
            )
          })}
      </svg>

      {hoverIdx !== null && hoverX !== null && (
        <div
          className="pointer-events-none absolute top-0 rounded-lg border border-surface-border bg-white px-3 py-2 text-xs shadow-lg"
          style={{
            left: `${(scaleX(hoverX) / width) * 100}%`,
            transform: 'translate(-50%, -110%)',
          }}
        >
          <div className="mb-1 font-medium text-ink-900">{formatX(hoverX)}</div>
          {series.map((s) => {
            const p = s.points[hoverIdx]
            if (!p) return null
            return (
              <div key={s.name} className="flex items-center gap-1.5 text-ink-500">
                <span
                  className="h-0.5 w-3 shrink-0 rounded-full"
                  style={{ backgroundColor: s.color }}
                />
                <span className="font-semibold text-ink-900">{formatY(p.y)}</span>
                {series.length > 1 && <span>{s.name}</span>}
              </div>
            )
          })}
        </div>
      )}

      {series.length > 1 && (
        <div className="mt-2 flex items-center gap-4 text-xs text-ink-500">
          {series.map((s) => (
            <span key={s.name} className="flex items-center gap-1.5">
              <span
                className="h-0.5 w-3 rounded-full"
                style={{ backgroundColor: s.color }}
              />
              {s.name}
            </span>
          ))}
        </div>
      )}
    </div>
  )
}
