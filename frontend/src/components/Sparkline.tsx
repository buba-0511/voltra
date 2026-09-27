interface SparklineProps {
  values: number[]
  color?: string
}

export function Sparkline({ values, color = '#18c8b3' }: SparklineProps) {
  if (values.length < 2) return null
  const max = Math.max(...values)
  const min = Math.min(...values)
  const points = values
    .map((v, i) => {
      const x = (i / (values.length - 1)) * 100
      const y = 24 - ((v - min) / (max - min || 1)) * 20
      return `${x},${y}`
    })
    .join(' ')

  return (
    <svg viewBox="0 0 100 24" className="h-6 w-16" aria-hidden="true">
      <polyline fill="none" stroke={color} strokeWidth={2} strokeLinecap="round" points={points} />
    </svg>
  )
}
