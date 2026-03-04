import { ValueCard } from './widgets/ValueCard'
import { LineChart } from './widgets/LineChart'
import { BooleanIndicator } from './widgets/BooleanIndicator'
import { EventLog } from './widgets/EventLog'

export function AutoDashboard({ deviceId, keys, latest }) {
  if (!keys?.length) {
    return (
      <div className="flex flex-col items-center justify-center py-20 text-center">
        <div className="text-4xl mb-4">📡</div>
        <div className="text-zinc-400 font-medium">No telemetry data yet</div>
        <div className="text-zinc-600 text-sm mt-1">
          Send data from your device to see widgets here
        </div>
      </div>
    )
  }

  const widgets = keys.map((k) => {
    const entry = latest[k.key]
    const value = entry?.value
    const lastUpdated = entry?.time

    if (k.value_type === 'boolean') {
      return { key: k.key, type: 'boolean', component: (
        <BooleanIndicator key={k.key} title={k.display_name || k.key} value={value} lastUpdated={lastUpdated} />
      )}
    }

    if (k.value_type === 'string' || k.value_type === 'json') {
      return { key: k.key, type: 'log', component: (
        <EventLog key={k.key} title={k.display_name || k.key} value={value} lastUpdated={lastUpdated} />
      )}
    }

    // number
    return { key: k.key, type: 'number', component: (
      <ValueCard
        key={k.key}
        title={k.display_name || k.key}
        value={value}
        unit={k.unit}
        valueType={k.value_type}
        lastUpdated={lastUpdated}
      />
    )}
  })

  // Numeric keys with line charts
  const numericKeys = keys.filter((k) => k.value_type === 'number')

  return (
    <div className="space-y-4">
      {/* Value cards grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
        {widgets.filter((w) => w.type === 'boolean' || w.type === 'number').map((w) => w.component)}
      </div>

      {/* Full-width: logs and charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {widgets.filter((w) => w.type === 'log').map((w) => w.component)}
      </div>

      {/* Time series charts for numeric keys */}
      {numericKeys.length > 0 && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
          {numericKeys.map((k) => (
            <LineChart
              key={k.key}
              deviceId={deviceId}
              keyName={k.key}
              unit={k.unit}
              liveValue={latest[k.key]?.value}
            />
          ))}
        </div>
      )}
    </div>
  )
}
