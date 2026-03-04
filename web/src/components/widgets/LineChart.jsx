import { useState, useEffect } from 'react'
import {
  LineChart as ReLineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts'
import { api } from '../../lib/api'

export function LineChart({ deviceId, keyName, unit, liveValue }) {
  const [data, setData] = useState([])

  useEffect(() => {
    if (!deviceId || !keyName) return
    const from = new Date(Date.now() - 60 * 60 * 1000).toISOString()
    api
      .getTelemetry(deviceId, { key: keyName, from, aggregate: 'avg', interval: '1m', limit: 500 })
      .then((res) => {
        const points = (res?.data || []).map((d) => ({
          t: new Date(d.time).toLocaleTimeString(),
          v: d.value,
        }))
        setData(points)
      })
      .catch(() => {})
  }, [deviceId, keyName])

  // Append live value
  useEffect(() => {
    if (liveValue == null) return
    setData((prev) => {
      const next = [
        ...prev,
        { t: new Date().toLocaleTimeString(), v: liveValue },
      ]
      return next.slice(-200)
    })
  }, [liveValue])

  const CustomTooltip = ({ active, payload, label }) => {
    if (!active || !payload?.length) return null
    return (
      <div className="bg-[#1a1a1a] border border-zinc-700 rounded-lg px-3 py-2 text-xs">
        <div className="text-zinc-400">{label}</div>
        <div className="text-accent font-mono font-bold">
          {payload[0].value?.toFixed(2)} {unit}
        </div>
      </div>
    )
  }

  return (
    <div className="card col-span-2">
      <div className="text-xs text-zinc-500 uppercase tracking-wider mb-3">{keyName} — last hour</div>
      <div className="h-40">
        <ResponsiveContainer width="100%" height="100%">
          <ReLineChart data={data}>
            <CartesianGrid strokeDasharray="3 3" stroke="#222" />
            <XAxis
              dataKey="t"
              tick={{ fontSize: 10, fill: '#52525b' }}
              tickLine={false}
              axisLine={false}
              interval="preserveStartEnd"
            />
            <YAxis
              tick={{ fontSize: 10, fill: '#52525b', fontFamily: 'monospace' }}
              tickLine={false}
              axisLine={false}
              width={40}
            />
            <Tooltip content={<CustomTooltip />} />
            <Line
              type="monotone"
              dataKey="v"
              stroke="#06b6d4"
              strokeWidth={2}
              dot={false}
              isAnimationActive={false}
            />
          </ReLineChart>
        </ResponsiveContainer>
      </div>
    </div>
  )
}
