import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../lib/api'

export function Projects() {
  const [projects, setProjects] = useState([])
  const [loading, setLoading] = useState(true)
  const [showModal, setShowModal] = useState(false)
  const [form, setForm] = useState({ name: '', description: '' })
  const [creating, setCreating] = useState(false)
  const navigate = useNavigate()

  useEffect(() => {
    api.listProjects()
      .then(setProjects)
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  async function createProject(e) {
    e.preventDefault()
    if (!form.name.trim()) return
    setCreating(true)
    try {
      const p = await api.createProject(form)
      setProjects((prev) => [p, ...prev])
      setShowModal(false)
      setForm({ name: '', description: '' })
      navigate(`/projects/${p.id}`)
    } catch (err) {
      alert(err.message)
    } finally {
      setCreating(false)
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-zinc-500">Loading...</div>
      </div>
    )
  }

  return (
    <div className="min-h-screen p-8">
      {/* Header */}
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-bold text-white">Projects</h1>
          <p className="text-zinc-500 text-sm mt-1">Manage your IoT projects and devices</p>
        </div>
        <button onClick={() => setShowModal(true)} className="btn-primary">
          + New Project
        </button>
      </div>

      {/* Empty state */}
      {projects.length === 0 && (
        <div className="flex flex-col items-center justify-center py-20 text-center">
          <div className="text-5xl mb-4">🛰️</div>
          <div className="text-xl font-semibold text-white mb-2">No projects yet</div>
          <div className="text-zinc-500 mb-6">Create your first project to get started</div>
          <button onClick={() => setShowModal(true)} className="btn-primary">
            Create Project
          </button>
        </div>
      )}

      {/* Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        {projects.map((p) => (
          <div
            key={p.id}
            onClick={() => navigate(`/projects/${p.id}`)}
            className="card cursor-pointer hover:border-zinc-600 transition-colors group"
          >
            <div className="flex items-start justify-between mb-3">
              <div className="w-8 h-8 rounded-lg bg-accent/10 border border-accent/20 flex items-center justify-center">
                <span className="text-accent text-sm">⬡</span>
              </div>
              <span className="text-xs text-zinc-600 mono">{p.id}</span>
            </div>
            <h3 className="font-semibold text-white group-hover:text-accent transition-colors">
              {p.name}
            </h3>
            {p.description && (
              <p className="text-zinc-500 text-sm mt-1 truncate">{p.description}</p>
            )}
            <div className="flex items-center gap-4 mt-4 pt-3 border-t border-zinc-800">
              <span className="text-xs text-zinc-500">
                {p.device_count} device{p.device_count !== 1 ? 's' : ''}
              </span>
              <span className="text-xs text-zinc-600">
                {new Date(p.created_at).toLocaleDateString()}
              </span>
            </div>
          </div>
        ))}
      </div>

      {/* Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-50 p-4">
          <div className="card w-full max-w-md">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-lg font-semibold text-white">New Project</h2>
              <button
                onClick={() => setShowModal(false)}
                className="text-zinc-500 hover:text-white text-xl"
              >
                ×
              </button>
            </div>
            <form onSubmit={createProject} className="space-y-4">
              <div>
                <label className="label">Project Name *</label>
                <input
                  className="input"
                  placeholder="My IoT Project"
                  value={form.name}
                  onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
                  autoFocus
                />
              </div>
              <div>
                <label className="label">Description</label>
                <input
                  className="input"
                  placeholder="Optional description"
                  value={form.description}
                  onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
                />
              </div>
              <div className="flex gap-3 pt-2">
                <button type="button" onClick={() => setShowModal(false)} className="btn-ghost flex-1">
                  Cancel
                </button>
                <button type="submit" disabled={creating} className="btn-primary flex-1">
                  {creating ? 'Creating...' : 'Create Project'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
