import { useState } from 'react'
import { api } from '../../lib/api'

export function ControlWidget({ deviceId, command, type = 'button', label, payload = 'trigger' }) {
  const [status, setStatus] = useState(null)
  const [loading, setLoading] = useState(false)
  const [sliderVal, setSliderVal] = useState(50)
  const [textVal, setTextVal] = useState('')
  const [toggled, setToggled] = useState(false)

  async function send(value) {
    setLoading(true)
    try {
      await api.sendCommand(deviceId, { command, payload: String(value) })
      setStatus('sent')
      setTimeout(() => setStatus(null), 2000)
    } catch {
      setStatus('error')
      setTimeout(() => setStatus(null), 2000)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="card flex flex-col gap-3">
      <div className="flex items-center justify-between">
        <span className="text-xs text-zinc-500 uppercase tracking-wider">{label || command}</span>
        {status && (
          <span className={`text-[10px] mono ${status === 'sent' ? 'text-green-400' : 'text-red-400'}`}>
            {status}
          </span>
        )}
      </div>

      {type === 'button' && (
        <button
          onClick={() => send(payload)}
          disabled={loading}
          className="btn-primary w-full"
        >
          {loading ? '...' : label || command}
        </button>
      )}

      {type === 'toggle' && (
        <button
          onClick={() => {
            const next = !toggled
            setToggled(next)
            send(next ? 'on' : 'off')
          }}
          disabled={loading}
          className={`w-full py-2 rounded-lg font-medium text-sm transition-all ${
            toggled
              ? 'bg-green-500/20 border border-green-500 text-green-400'
              : 'bg-zinc-800 border border-zinc-700 text-zinc-400'
          }`}
        >
          {toggled ? 'ON' : 'OFF'}
        </button>
      )}

      {type === 'slider' && (
        <div className="flex flex-col gap-2">
          <input
            type="range"
            min="0"
            max="255"
            value={sliderVal}
            onChange={(e) => setSliderVal(Number(e.target.value))}
            onMouseUp={() => send(sliderVal)}
            className="w-full accent-accent"
          />
          <div className="flex justify-between text-[10px] text-zinc-600">
            <span>0</span>
            <span className="mono text-accent">{sliderVal}</span>
            <span>255</span>
          </div>
        </div>
      )}

      {type === 'text' && (
        <div className="flex gap-2">
          <input
            value={textVal}
            onChange={(e) => setTextVal(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && send(textVal)}
            placeholder="Enter value..."
            className="input flex-1"
          />
          <button onClick={() => send(textVal)} disabled={loading} className="btn-primary">
            Send
          </button>
        </div>
      )}
    </div>
  )
}
