import { NavLink } from 'react-router-dom'
import { DeviceStatusBadge } from './DeviceStatusBadge'

export function Sidebar({ project, devices }) {
  return (
    <aside className="w-64 shrink-0 bg-[#0d0d0d] border-r border-zinc-800 flex flex-col">
      <div className="px-4 py-4 border-b border-zinc-800">
        <div className="text-xs text-zinc-500 uppercase tracking-wider mb-1">Project</div>
        <div className="font-semibold text-white truncate">{project?.name}</div>
        <div className="text-xs text-zinc-500 mt-0.5">{devices?.length ?? 0} devices</div>
      </div>

      <nav className="flex-1 overflow-y-auto py-2">
        {!devices?.length && (
          <div className="px-4 py-8 text-center text-xs text-zinc-600">No devices yet</div>
        )}
        {devices?.map((device) => (
          <NavLink
            key={device.id}
            to={`/projects/${project.id}/devices/${device.id}`}
            className={({ isActive }) =>
              `flex flex-col px-4 py-3 transition-colors hover:bg-zinc-900 ${
                isActive ? 'bg-zinc-900 border-r-2 border-accent' : ''
              }`
            }
          >
            <span className="text-sm font-medium text-white truncate">{device.name}</span>
            <DeviceStatusBadge isOnline={device.is_online} lastSeen={device.last_seen_at} />
          </NavLink>
        ))}
      </nav>
    </aside>
  )
}
