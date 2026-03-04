import { useState, useEffect } from 'react'

export function Settings() {
  const [info] = useState({
    version: 'v1.0.0',
    mqttPort: 1883,
    wsPort: 9001,
    httpPort: location.port || 8080,
  })

  return (
    <div className="min-h-screen p-8 max-w-2xl mx-auto">
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-white">Settings</h1>
        <p className="text-zinc-500 mt-1">Server configuration and information</p>
      </div>

      <div className="space-y-4">
        <div className="card">
          <h2 className="font-semibold text-white mb-4">Server Info</h2>
          <div className="space-y-3">
            {[
              ['Version', info.version],
              ['HTTP Port', info.httpPort],
              ['MQTT Port', info.mqttPort],
              ['MQTT WebSocket Port', info.wsPort],
            ].map(([label, value]) => (
              <div key={label} className="flex items-center justify-between text-sm">
                <span className="text-zinc-500">{label}</span>
                <span className="mono text-zinc-300">{value}</span>
              </div>
            ))}
          </div>
        </div>

        <div className="card">
          <h2 className="font-semibold text-white mb-4">Configuration</h2>
          <p className="text-zinc-500 text-sm mb-4">
            These settings are configured via environment variables. Edit your <code className="mono text-accent">.env</code> file and restart the server.
          </p>
          <div className="space-y-2">
            {[
              ['APISTO_RETENTION_DAYS', '30', 'Days of telemetry to retain'],
              ['APISTO_HEARTBEAT_TIMEOUT', '60', 'Seconds before device marked offline'],
              ['APISTO_LOG_LEVEL', 'info', 'Log verbosity'],
              ['APISTO_CORS_ORIGINS', '*', 'Allowed CORS origins'],
            ].map(([key, defaultVal, desc]) => (
              <div key={key} className="rounded-lg bg-[#1a1a1a] p-3">
                <div className="flex items-start justify-between mb-1">
                  <code className="mono text-xs text-accent">{key}</code>
                  <code className="mono text-xs text-zinc-500">default: {defaultVal}</code>
                </div>
                <div className="text-xs text-zinc-600">{desc}</div>
              </div>
            ))}
          </div>
        </div>

        <div className="card">
          <h2 className="font-semibold text-white mb-2">MQTT Topics</h2>
          <div className="space-y-2 text-xs mono text-zinc-400">
            <div><span className="text-accent">apisto/</span><span className="text-yellow-500">{'{token}'}</span><span className="text-accent">/telemetry</span> <span className="text-zinc-600">— device → server</span></div>
            <div><span className="text-accent">apisto/</span><span className="text-yellow-500">{'{token}'}</span><span className="text-accent">/status</span> <span className="text-zinc-600">— heartbeat</span></div>
            <div><span className="text-accent">apisto/</span><span className="text-yellow-500">{'{token}'}</span><span className="text-accent">/commands</span> <span className="text-zinc-600">— server → device</span></div>
            <div><span className="text-accent">apisto/</span><span className="text-yellow-500">{'{token}'}</span><span className="text-accent">/commands/ack</span> <span className="text-zinc-600">— device acknowledgment</span></div>
          </div>
        </div>
      </div>
    </div>
  )
}
