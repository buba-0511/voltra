import { Gauge, LayoutDashboard, LogOut, Sparkles, Zap } from 'lucide-react'
import { useState } from 'react'
import { NavLink, Outlet } from 'react-router-dom'
import { useLogoutMutation } from '../api/apiSlice'
import { loggedOut } from '../store/authSlice'
import { useAppDispatch, useAppSelector } from '../store/hooks'

const navItems = [
  { to: '/dashboard', label: 'Dashboard', icon: LayoutDashboard },
  { to: '/meters', label: 'Medidores', icon: Gauge },
  { to: '/anomalies', label: 'Anomalías IA', icon: Sparkles },
]

function Logo() {
  return (
    <div className="flex items-center gap-2">
      <div className="grid h-7 w-7 place-items-center rounded-md bg-brand-500 text-ink-900">
        <Zap size={15} strokeWidth={2.8} />
      </div>
      <span className="text-sm font-bold tracking-tight text-ink-900">
        Voltra<span className="text-brand-500">.</span>
      </span>
    </div>
  )
}

export function AppLayout() {
  const dispatch = useAppDispatch()
  const user = useAppSelector((s) => s.auth.user)
  const [logoutRequest] = useLogoutMutation()
  const [profileMenuOpen, setProfileMenuOpen] = useState(false)
  const logout = () => {
    // Clear the local state regardless of whether the request itself
    // succeeds - the cookie has a 24h expiry either way, but there's no
    // reason to leave the UI looking logged-in if the network call fails.
    logoutRequest().finally(() => dispatch(loggedOut()))
  }

  return (
    <div className="flex h-screen overflow-hidden bg-surface">
      {/* Sidebar - desktop only. On mobile, nav moves to the fixed
          bottom bar and identity/logout move to the compact top header
          below, since there's no room for a permanent side column. */}
      <aside className="hidden w-56 shrink-0 flex-col border-r border-surface-border bg-white md:flex">
        <div className="px-5 py-5">
          <Logo />
        </div>
        <nav className="flex-1 space-y-1 px-3">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              className={({ isActive }) =>
                `flex items-center gap-2.5 rounded-md px-3 py-2 text-sm font-medium transition-colors ${
                  isActive ? 'bg-ink-700 text-white' : 'text-ink-500 hover:bg-surface'
                }`
              }
            >
              <item.icon size={16} />
              {item.label}
            </NavLink>
          ))}
        </nav>
        <div className="border-t border-surface-border p-3">
          <div className="flex items-center gap-2 px-2 py-1.5">
            <div className="grid h-7 w-7 shrink-0 place-items-center rounded-full bg-ink-100 text-[11px] font-semibold text-ink-700">
              {user?.email?.[0]?.toUpperCase() ?? '?'}
            </div>
            <span className="truncate text-xs text-ink-400">{user?.email}</span>
          </div>
          <button
            onClick={logout}
            className="mt-1 flex w-full items-center gap-2.5 rounded-md px-3 py-2 text-sm font-medium text-ink-500 hover:bg-surface"
          >
            <LogOut size={16} />
            Salir
          </button>
        </div>
      </aside>

      <div className="flex min-h-0 min-w-0 flex-1 flex-col">
        {/* Compact top header - mobile only */}
        <header className="flex shrink-0 items-center justify-between border-b border-surface-border bg-white px-4 py-3 md:hidden">
          <Logo />
          <div className="relative">
            <button
              onClick={() => setProfileMenuOpen((v) => !v)}
              className="grid h-8 w-8 place-items-center rounded-full bg-ink-100 text-xs font-semibold text-ink-700"
            >
              {user?.email?.[0]?.toUpperCase() ?? '?'}
            </button>
            {profileMenuOpen && (
              <>
                <button
                  aria-label="Cerrar menú"
                  onClick={() => setProfileMenuOpen(false)}
                  className="fixed inset-0 z-30 cursor-default"
                />
                <div className="absolute top-full right-0 z-40 mt-2 w-48 rounded-lg border border-surface-border bg-white p-2 shadow-lg">
                  <p className="truncate px-2 py-1.5 text-xs text-ink-400">{user?.email}</p>
                  <button
                    onClick={logout}
                    className="mt-1 flex w-full items-center gap-2 rounded-md px-2 py-2 text-sm font-medium text-ink-700 hover:bg-surface"
                  >
                    <LogOut size={15} />
                    Cerrar sesión
                  </button>
                </div>
              </>
            )}
          </div>
        </header>

        <main className="min-h-0 flex-1 overflow-y-auto pb-16 md:pb-0">
          <div className="mx-auto h-full max-w-6xl px-4 py-8 md:px-6">
            <Outlet />
          </div>
        </main>

        <nav className="fixed inset-x-0 bottom-0 z-20 flex shrink-0 border-t border-surface-border bg-white pb-[env(safe-area-inset-bottom)] md:hidden">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              className={({ isActive }) =>
                `flex flex-1 flex-col items-center gap-1 py-2.5 text-[11px] font-medium ${
                  isActive ? 'text-brand-600' : 'text-ink-400'
                }`
              }
            >
              <item.icon size={18} />
              {item.label}
            </NavLink>
          ))}
        </nav>
      </div>
    </div>
  )
}
