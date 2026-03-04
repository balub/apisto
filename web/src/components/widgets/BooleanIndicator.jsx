export function BooleanIndicator({ title, value, lastUpdated }) {
  const isOn = value === true || value === 'true' || value === 1

  return (
    <div className="card flex flex-col gap-3 min-h-[100px]">
      <div className="flex items-start justify-between">
        <span className="text-xs text-zinc-500 uppercase tracking-wider">{title}</span>
        {lastUpdated && (
          <span className="text-[10px] text-zinc-600">
            {new Date(lastUpdated).toLocaleTimeString()}
          </span>
        )}
      </div>
      <div className="flex items-center gap-3">
        <div
          className={`w-4 h-4 rounded-full transition-colors duration-300 ${
            isOn ? 'bg-green-400 shadow-[0_0_8px_#4ade80]' : 'bg-zinc-600'
          }`}
        />
        <span
          className={`mono text-2xl font-bold transition-colors duration-300 ${
            isOn ? 'text-green-400' : 'text-zinc-500'
          }`}
        >
          {isOn ? 'ON' : 'OFF'}
        </span>
      </div>
    </div>
  )
}
