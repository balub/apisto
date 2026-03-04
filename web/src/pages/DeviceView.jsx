import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { DeviceStatusBadge } from '../components/DeviceStatusBadge'
import { AutoDashboard } from '../components/AutoDashboard'
import { ControlWidget } from '../components/widgets/ControlWidget'
import { Sidebar } from '../components/Sidebar'
import { useDeviceData } from '../hooks/useDeviceData'
import { api } from '../lib/api'

export function DeviceView() {
  const { projectId, deviceId } = useParams()
  const navigate = useNavigate()
  const { device, keys, latest, loading, error } = useDeviceData(deviceId)
  const [devices, setDevices] = useState([])
  const [project, setProject] = useState(null)
  const [tab, setTab] = useState('dashboard')
  const [commands, setCommands] = useState([])
  const [showAddControl, setShowAddControl] = useState(false)
  const [controls, setControls] = useState([])
  const [controlForm, setControlForm] = useState({ command: '', label: '', type: 'button', payload: 'trigger' })

  useEffect(() => {
    api.getProject(projectId).then(setProject).catch(() => {})
    api.listDevices(projectId).then(setDevices).catch(() => {})
  }, [projectId])

  useEffect(() => {
    if (tab === 'commands') {
      api.listCommands(deviceId, 50).then(setCommands).catch(() => {})
    }
  }, [tab, deviceId])

  if (loading) {
    return (
      <div className="flex h-screen">
        <div className="flex-1 flex items-center justify-center">
          <div className="text-zinc-500">Loading device...</div>
        </div>
      </div>
    )
  }

  if (error || !device) {
    return (
      <div className="flex h-screen">
        <div className="flex-1 flex items-center justify-center">
          <div className="text-red-400">Device not found</div>
        </div>
      </div>
    )
  }

  return (
    <div className="flex h-screen overflow-hidden">
      <Sidebar project={project} devices={devices} />
      <main className="flex-1 overflow-y-auto">
        {/* Device header */}
        <div className="border-b border-zinc-800 px-6 py-4">
          <div className="flex items-start justify-between">
            <div>
              <h1 className="text-xl font-bold text-white">{device.name}</h1>
              <div className="flex items-center gap-4 mt-1">
                <DeviceStatusBadge isOnline={device.is_online} lastSeen={device.last_seen_at} />
                {device.ip_address && (
                  <span className="mono text-xs text-zinc-600">{device.ip_address}</span>
                )}
                {device.firmware_version && (
                  <span className="mono text-xs text-zinc-600">fw {device.firmware_version}</span>
                )}
              </div>
            </div>
            <div className="flex gap-2">
              <button
                onClick={() => api.createShare(deviceId).then(s => {
                  const url = `${location.origin}/public/${s.share_token}`
                  navigator.clipboard.writeText(url)
                  alert('Share URL copied!')
                })}
                className="btn-ghost text-xs"
              >
                Share
              </button>
            </div>
          </div>

          {/* Tabs */}
          <div className="flex gap-1 mt-4">
            {['dashboard', 'controls', 'commands', 'raw'].map((t) => (
              <button
                key={t}
                onClick={() => setTab(t)}
                className={`px-3 py-1.5 text-xs rounded-md capitalize transition-colors ${
                  tab === t
                    ? 'bg-accent/10 text-accent border border-accent/20'
                    : 'text-zinc-500 hover:text-white'
                }`}
              >
                {t}
              </button>
            ))}
          </div>
        </div>

        <div className="p-6">
          {tab === 'dashboard' && (
            <AutoDashboard deviceId={deviceId} keys={keys} latest={latest} />
          )}

          {tab === 'controls' && (
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <h2 className="text-sm font-medium text-zinc-300">Control Widgets</h2>
                <button onClick={() => setShowAddControl(true)} className="btn-ghost text-xs">+ Add Control</button>
              </div>
              {controls.length === 0 && (
                <div className="text-zinc-600 text-sm py-8 text-center">
                  No controls yet. Add a control to send commands to your device.
                </div>
              )}
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                {controls.map((ctrl, i) => (
                  <ControlWidget key={i} deviceId={deviceId} {...ctrl} />
                ))}
              </div>

              {showAddControl && (
                <div className="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-50 p-4">
                  <div className="card w-full max-w-md">
                    <div className="flex justify-between mb-4">
                      <h3 className="font-semibold text-white">Add Control Widget</h3>
                      <button onClick={() => setShowAddControl(false)} className="text-zinc-500 hover:text-white">×</button>
                    </div>
                    <div className="space-y-3">
                      <div>
                        <label className="label">Command Name</label>
                        <input className="input" placeholder="relay" value={controlForm.command} onChange={e => setControlForm(f => ({...f, command: e.target.value}))} />
                      </div>
                      <div>
                        <label className="label">Label</label>
                        <input className="input" placeholder="Toggle Relay" value={controlForm.label} onChange={e => setControlForm(f => ({...f, label: e.target.value}))} />
                      </div>
                      <div>
                        <label className="label">Type</label>
                        <select className="input" value={controlForm.type} onChange={e => setControlForm(f => ({...f, type: e.target.value}))}>
                          <option value="button">Button</option>
                          <option value="toggle">Toggle</option>
                          <option value="slider">Slider</option>
                          <option value="text">Text Input</option>
                        </select>
                      </div>
                      {controlForm.type === 'button' && (
                        <div>
                          <label className="label">Payload</label>
                          <input className="input" placeholder="trigger" value={controlForm.payload} onChange={e => setControlForm(f => ({...f, payload: e.target.value}))} />
                        </div>
                      )}
                      <div className="flex gap-3 pt-2">
                        <button onClick={() => setShowAddControl(false)} className="btn-ghost flex-1">Cancel</button>
                        <button onClick={() => { setControls(c => [...c, {...controlForm}]); setShowAddControl(false) }} className="btn-primary flex-1">Add</button>
                      </div>
                    </div>
                  </div>
                </div>
              )}
            </div>
          )}

          {tab === 'commands' && (
            <div>
              <h2 className="text-sm font-medium text-zinc-300 mb-4">Command History</h2>
              {commands.length === 0 && (
                <div className="text-zinc-600 text-sm py-8 text-center">No commands sent yet</div>
              )}
              <div className="space-y-2">
                {commands.map((cmd) => (
                  <div key={cmd.id} className="card flex items-center gap-4">
                    <div className="flex-1">
                      <span className="mono text-sm text-white">{cmd.command}</span>
                      {cmd.payload && <span className="mono text-xs text-zinc-500 ml-2">"{cmd.payload}"</span>}
                    </div>
                    <span className={`text-xs px-2 py-0.5 rounded-full ${
                      cmd.status === 'acknowledged' ? 'bg-green-900/30 text-green-400' :
                      cmd.status === 'sent' ? 'bg-accent/10 text-accent' :
                      cmd.status === 'failed' ? 'bg-red-900/30 text-red-400' :
                      'bg-zinc-800 text-zinc-400'
                    }`}>
                      {cmd.status}
                    </span>
                    <span className="text-xs text-zinc-600 mono">
                      {new Date(cmd.created_at).toLocaleTimeString()}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}

          {tab === 'raw' && (
            <div>
              <h2 className="text-sm font-medium text-zinc-300 mb-4">Latest Values</h2>
              <div className="space-y-2">
                {Object.entries(latest).map(([key, entry]) => (
                  <div key={key} className="card flex items-center gap-4">
                    <span className="mono text-zinc-400 text-sm w-40 truncate">{key}</span>
                    <span className="mono text-white text-sm flex-1">
                      {JSON.stringify(entry.value)}
                    </span>
                    <span className="text-xs text-zinc-600 mono px-2 py-0.5 rounded bg-zinc-800">{entry.value_type}</span>
                    <span className="text-xs text-zinc-600">
                      {entry.time ? new Date(entry.time).toLocaleTimeString() : ''}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      </main>
    </div>
  )
}
