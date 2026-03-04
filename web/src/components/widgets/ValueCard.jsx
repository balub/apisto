import { LineChart as ReLineChart, Line, ResponsiveContainer } from 'recharts'

export function ValueCard({ title, value, unit, valueType, history = [], lastUpdated }) {
  const displayValue = valueType === 'number'
    ? (typeof value === 'number' ? value.toFixed(2).replace(/\.?0+$/, '') : value)
    : String(value ?? '—')

  const sparkData = history.slice(-20).map((v, i) => ({ i, v }))

  return (
    <div className="card flex flex-col gap-2 min-h-[100px]">
      <div className="flex items-start justify-between">
        <span className="text-xs text-zinc-500 uppercase tracking-wider">{title}</span>
        {lastUpdated && (
          <span className="text-[10px] text-zinc-600">
            {new Date(lastUpdated).toLocaleTimeString()}
          </span>
        )}
      </div>
      <div className="flex items-end gap-1">
        <span className="mono text-3xl font-bold text-white tabular-nums">{displayValue}</span>
        {unit && <span className="text-zinc-400 text-sm mb-1">{unit}</span>}
      </div>
      {sparkData.length > 2 && (
        <div className="h-10 -mx-1">
          <ResponsiveContainer width="100%" height="100%">
            <ReLineChart data={sparkData}>
              <Line
                type="monotone"
                dataKey="v"
                stroke="#06b6d4"
                strokeWidth={1.5}
                dot={false}
                isAnimationActive={false}
              />
            </ReLineChart>
          </ResponsiveContainer>
        </div>
      )}
    </div>
  )
}
