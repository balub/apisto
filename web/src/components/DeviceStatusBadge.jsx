export function DeviceStatusBadge({ isOnline, lastSeen }) {
  return (
    <div className="flex items-center gap-2">
      <div
        className={`w-2 h-2 rounded-full ${
          isOnline ? 'bg-green-400 shadow-[0_0_6px_#4ade80]' : 'bg-zinc-600'
        }`}
      />
      <span className={`text-xs ${isOnline ? 'text-green-400' : 'text-zinc-500'}`}>
        {isOnline ? 'online' : lastSeen ? `offline · ${timeAgo(lastSeen)}` : 'offline'}
      </span>
    </div>
  )
}

function timeAgo(ts) {
  const secs = Math.floor((Date.now() - new Date(ts)) / 1000)
  if (secs < 60) return `${secs}s ago`
  if (secs < 3600) return `${Math.floor(secs / 60)}m ago`
  if (secs < 86400) return `${Math.floor(secs / 3600)}h ago`
  return `${Math.floor(secs / 86400)}d ago`
}
