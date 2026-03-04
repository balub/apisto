export function GaugeCard({ title, value, min = 0, max = 100, unit, lastUpdated }) {
  const pct = Math.min(100, Math.max(0, ((value - min) / (max - min)) * 100))
  const angle = -135 + (pct / 100) * 270

  const getColor = () => {
    if (pct < 33) return '#22d3ee'
    if (pct < 66) return '#facc15'
    return '#f87171'
  }

  return (
    <div className="card flex flex-col items-center gap-2 min-h-[140px]">
      <div className="text-xs text-zinc-500 uppercase tracking-wider self-start">{title}</div>
      <div className="relative w-24 h-16 overflow-hidden">
        <svg viewBox="0 0 100 60" className="w-full">
          {/* Background arc */}
          <path
            d="M 10 55 A 40 40 0 0 1 90 55"
            fill="none"
            stroke="#2a2a2a"
            strokeWidth="8"
            strokeLinecap="round"
          />
          {/* Value arc */}
          <path
            d="M 10 55 A 40 40 0 0 1 90 55"
            fill="none"
            stroke={getColor()}
            strokeWidth="8"
            strokeLinecap="round"
            strokeDasharray={`${(pct / 100) * 125.6} 125.6`}
          />
        </svg>
      </div>
      <div className="flex items-end gap-1">
        <span className="mono text-2xl font-bold" style={{ color: getColor() }}>
          {typeof value === 'number' ? value.toFixed(1) : value}
        </span>
        {unit && <span className="text-zinc-400 text-sm mb-0.5">{unit}</span>}
      </div>
      {lastUpdated && (
        <span className="text-[10px] text-zinc-600">{new Date(lastUpdated).toLocaleTimeString()}</span>
      )}
    </div>
  )
}
