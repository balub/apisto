import { useState, useEffect, useCallback } from 'react'
import { api } from '../lib/api'
import { useWebSocket } from './useWebSocket'

export function useDeviceData(deviceId) {
  const [device, setDevice] = useState(null)
  const [keys, setKeys] = useState([])
  const [latest, setLatest] = useState({})
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  const fetchData = useCallback(async () => {
    if (!deviceId) return
    try {
      setLoading(true)
      const [deviceData, keysData, latestData] = await Promise.all([
        api.getDevice(deviceId),
        api.getDeviceKeys(deviceId),
        api.getLatestTelemetry(deviceId),
      ])
      setDevice(deviceData)
      setKeys(keysData || [])
      setLatest(latestData?.data || {})
      setError(null)
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }, [deviceId])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  useWebSocket(deviceId ? `/devices/${deviceId}/ws` : null, (msg) => {
    if (msg.type === 'telemetry') {
      setLatest((prev) => {
        const next = { ...prev }
        Object.entries(msg.data || {}).forEach(([key, value]) => {
          next[key] = {
            value,
            value_type: typeof value === 'number' ? 'number' : typeof value === 'boolean' ? 'boolean' : 'string',
            time: msg.timestamp,
          }
        })
        return next
      })
      // Re-fetch device keys in case new ones appeared
      api.getDeviceKeys(deviceId).then((k) => setKeys(k || []))
    }

    if (msg.type === 'status') {
      setDevice((prev) => prev ? { ...prev, is_online: msg.is_online, last_seen_at: msg.last_seen_at } : prev)
    }
  })

  return { device, keys, latest, loading, error, refetch: fetchData }
}
