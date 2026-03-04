import { BrowserRouter, Routes, Route, NavLink, Navigate } from 'react-router-dom'
import { Projects } from './pages/Projects'
import { Dashboard } from './pages/Dashboard'
import { DeviceView } from './pages/DeviceView'
import { QuickStart } from './pages/QuickStart'
import { Settings } from './pages/Settings'

function Nav() {
  return (
    <header className="fixed top-0 left-0 right-0 z-40 border-b border-zinc-800 bg-[#0a0a0a]/90 backdrop-blur-sm">
      <div className="flex items-center gap-6 px-6 h-12">
        <NavLink to="/" className="flex items-center gap-2 shrink-0">
          <span className="text-accent font-bold text-lg">⬡</span>
          <span className="font-semibold text-white">Apisto</span>
        </NavLink>
        <nav className="flex items-center gap-1 flex-1">
          <NavLink
            to="/"
            end
            className={({ isActive }) =>
              `text-sm px-3 py-1.5 rounded-md transition-colors ${
                isActive ? 'text-white bg-zinc-800' : 'text-zinc-500 hover:text-white'
              }`
            }
          >
            Projects
          </NavLink>
          <NavLink
            to="/quickstart"
            className={({ isActive }) =>
              `text-sm px-3 py-1.5 rounded-md transition-colors ${
                isActive ? 'text-white bg-zinc-800' : 'text-zinc-500 hover:text-white'
              }`
            }
          >
            Quick Start
          </NavLink>
          <NavLink
            to="/settings"
            className={({ isActive }) =>
              `text-sm px-3 py-1.5 rounded-md transition-colors ${
                isActive ? 'text-white bg-zinc-800' : 'text-zinc-500 hover:text-white'
              }`
            }
          >
            Settings
          </NavLink>
        </nav>
      </div>
    </header>
  )
}

export default function App() {
  return (
    <BrowserRouter>
      <Nav />
      <div className="pt-12">
        <Routes>
          <Route path="/" element={<Projects />} />
          <Route path="/projects/:projectId" element={<Dashboard />} />
          <Route path="/projects/:projectId/devices/:deviceId" element={<DeviceView />} />
          <Route path="/quickstart" element={<QuickStart />} />
          <Route path="/settings" element={<Settings />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </div>
    </BrowserRouter>
  )
}
