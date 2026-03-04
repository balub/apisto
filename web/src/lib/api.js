const BASE = '/api/v1'

async function request(method, path, body) {
  const res = await fetch(`${BASE}${path}`, {
    method,
    headers: body ? { 'Content-Type': 'application/json' } : {},
    body: body ? JSON.stringify(body) : undefined,
  })

  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(err.error || 'Request failed')
  }

  if (res.status === 204) return null
  return res.json()
}

export const api = {
  // Projects
  listProjects: () => request('GET', '/projects'),
  getProject: (id) => request('GET', `/projects/${id}`),
  createProject: (data) => request('POST', '/projects', data),
  updateProject: (id, data) => request('PUT', `/projects/${id}`, data),
  deleteProject: (id) => request('DELETE', `/projects/${id}`),

  // Devices
  listDevices: (projectId) => request('GET', `/projects/${projectId}/devices`),
  getDevice: (id) => request('GET', `/devices/${id}`),
  createDevice: (projectId, data) => request('POST', `/projects/${projectId}/devices`, data),
  updateDevice: (id, data) => request('PUT', `/devices/${id}`, data),
  deleteDevice: (id) => request('DELETE', `/devices/${id}`),
  getDeviceKeys: (id) => request('GET', `/devices/${id}/keys`),

  // Telemetry
  getTelemetry: (deviceId, params = {}) => {
    const qs = new URLSearchParams(params).toString()
    return request('GET', `/devices/${deviceId}/telemetry${qs ? '?' + qs : ''}`)
  },
  getLatestTelemetry: (deviceId) => request('GET', `/devices/${deviceId}/telemetry/latest`),
  ingestTelemetry: (token, data) => request('POST', `/devices/${token}/telemetry`, data),

  // Commands
  sendCommand: (deviceId, data) => request('POST', `/devices/${deviceId}/commands`, data),
  listCommands: (deviceId, limit = 20) => request('GET', `/devices/${deviceId}/commands?limit=${limit}`),

  // Shares
  createShare: (deviceId) => request('POST', `/devices/${deviceId}/share`),
  revokeShare: (shareToken) => request('DELETE', `/shares/${shareToken}`),
  getPublicData: (shareToken) => request('GET', `/public/${shareToken}`),
}

export function wsUrl(path) {
  const protocol = location.protocol === 'https:' ? 'wss' : 'ws'
  return `${protocol}://${location.host}/api/v1${path}`
}
