import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Sidebar } from '../components/Sidebar'
import { api } from '../lib/api'

export function Dashboard() {
  const { projectId } = useParams()
  const navigate = useNavigate()
  const [project, setProject] = useState(null)
  const [devices, setDevices] = useState([])
  const [showAddDevice, setShowAddDevice] = useState(false)
  const [deviceForm, setDeviceForm] = useState({ name: '', description: '' })
  const [newDevice, setNewDevice] = useState(null)
  const [creating, setCreating] = useState(false)

  useEffect(() => {
    Promise.all([api.getProject(projectId), api.listDevices(projectId)])
      .then(([p, d]) => {
        setProject(p)
        setDevices(d || [])
        if (d?.length > 0) navigate(`/projects/${projectId}/devices/${d[0].id}`, { replace: true })
      })
      .catch(() => {})
  }, [projectId])

  async function addDevice(e) {
    e.preventDefault()
    setCreating(true)
    try {
      const d = await api.createDevice(projectId, deviceForm)
      setDevices((prev) => [...prev, d])
      setNewDevice(d)
      setDeviceForm({ name: '', description: '' })
    } catch (err) {
      alert(err.message)
    } finally {
      setCreating(false)
    }
  }

  return (
    <div className="flex h-screen overflow-hidden">
      <Sidebar project={project} devices={devices} />
      <main className="flex-1 overflow-y-auto p-8">
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-lg font-semibold text-white">{project?.name}</h2>
          <button onClick={() => setShowAddDevice(true)} className="btn-primary">
            + Add Device
          </button>
        </div>

        {!devices.length && (
          <div className="flex flex-col items-center justify-center py-20 text-center">
            <div className="text-4xl mb-4">📡</div>
            <div className="text-zinc-400 font-medium">No devices yet</div>
            <div className="text-zinc-600 text-sm mt-1 mb-6">Add your first device to get started</div>
            <button onClick={() => setShowAddDevice(true)} className="btn-primary">
              Add Device
            </button>
          </div>
        )}

        {showAddDevice && (
          <div className="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-50 p-4">
            <div className="card w-full max-w-lg">
              <div className="flex items-center justify-between mb-6">
                <h2 className="text-lg font-semibold text-white">Add Device</h2>
                <button onClick={() => { setShowAddDevice(false); setNewDevice(null) }} className="text-zinc-500 hover:text-white text-xl">×</button>
              </div>

              {!newDevice ? (
                <form onSubmit={addDevice} className="space-y-4">
                  <div>
                    <label className="label">Device Name *</label>
                    <input
                      className="input"
                      placeholder="Living Room Sensor"
                      value={deviceForm.name}
                      onChange={(e) => setDeviceForm((f) => ({ ...f, name: e.target.value }))}
                      autoFocus
                    />
                  </div>
                  <div>
                    <label className="label">Description</label>
                    <input
                      className="input"
                      placeholder="Optional"
                      value={deviceForm.description}
                      onChange={(e) => setDeviceForm((f) => ({ ...f, description: e.target.value }))}
                    />
                  </div>
                  <div className="flex gap-3">
                    <button type="button" onClick={() => setShowAddDevice(false)} className="btn-ghost flex-1">Cancel</button>
                    <button type="submit" disabled={creating} className="btn-primary flex-1">{creating ? 'Creating...' : 'Create Device'}</button>
                  </div>
                </form>
              ) : (
                <div className="space-y-4">
                  <div className="bg-green-900/20 border border-green-800 rounded-lg p-3 text-sm text-green-400">
                    Device created! Copy your token — it won't be shown again.
                  </div>
                  <div>
                    <label className="label">Device Token</label>
                    <div className="flex gap-2">
                      <input className="input mono" readOnly value={newDevice.token} />
                      <button onClick={() => navigator.clipboard.writeText(newDevice.token)} className="btn-ghost shrink-0">Copy</button>
                    </div>
                  </div>
                  <div className="text-xs text-zinc-500">
                    Use this token in your device sketch. See the{' '}
                    <a href="/quickstart" className="text-accent hover:underline">Quick Start guide</a>.
                  </div>
                  <button
                    onClick={() => { setShowAddDevice(false); setNewDevice(null); navigate(`/projects/${projectId}/devices/${newDevice.id}`) }}
                    className="btn-primary w-full"
                  >
                    Go to Device Dashboard
                  </button>
                </div>
              )}
            </div>
          </div>
        )}
      </main>
    </div>
  )
}
