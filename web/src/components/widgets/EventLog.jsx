import { useState, useEffect, useRef } from 'react'

export function EventLog({ title, value, lastUpdated }) {
  const [events, setEvents] = useState([])
  const listRef = useRef(null)

  useEffect(() => {
    if (value == null) return
    setEvents((prev) => {
      const next = [
        { value: String(value), time: lastUpdated || new Date().toISOString() },
        ...prev,
      ]
      return next.slice(0, 50)
    })
  }, [value, lastUpdated])

  return (
    <div className="card col-span-2">
      <div className="text-xs text-zinc-500 uppercase tracking-wider mb-3">{title} — event log</div>
      <div ref={listRef} className="space-y-1 max-h-36 overflow-y-auto">
        {events.length === 0 && (
          <div className="text-zinc-600 text-xs">No events yet</div>
        )}
        {events.map((e, i) => (
          <div key={i} className="flex items-start gap-3 text-xs">
            <span className="mono text-zinc-600 shrink-0">
              {new Date(e.time).toLocaleTimeString()}
            </span>
            <span className="mono text-zinc-300">{e.value}</span>
          </div>
        ))}
      </div>
    </div>
  )
}
