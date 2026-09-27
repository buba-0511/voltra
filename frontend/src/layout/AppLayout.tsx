import { Zap } from 'lucide-react'
import { NavLink, Outlet } from 'react-router-dom'
import { loggedOut } from '../store/authSlice'
import { useAppDispatch, useAppSelector } from '../store/hooks'

const navItems = [
  { to: '/dashboard', label: 'Dashboard' },
  { to: '/meters', label: 'Medidores' },
  { to: '/anomalies', label: 'Anomalías IA' },
]

export function AppLayout() {
  const dispatch = useAppDispatch()
  const user = useAppSelector((s) => s.auth.user)
  const logout = () => dispatch(loggedOut())

  return (
    <div className="min-h-screen bg-surface">
      <header className="border-b border-surface-border bg-white">
        <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-3">
          <div className="flex items-center gap-8">
            <div className="flex items-center gap-2">
              <div className="grid h-7 w-7 place-items-center rounded-md bg-brand-500 text-ink-900">
                <Zap size={15} strokeWidth={2.8} />
              </div>
              <span className="text-sm font-bold tracking-tight text-ink-900">
                Voltra<span className="text-brand-500">.</span>
              </span>
            </div>
            <nav className="flex gap-1">
              {navItems.map((item) => (
                <NavLink
                  key={item.to}
                  to={item.to}
                  className={({ isActive }) =>
                    `rounded-md px-3 py-1.5 text-sm font-medium transition-colors ${
                      isActive
                        ? 'bg-ink-700 text-white'
                        : 'text-ink-500 hover:bg-surface'
                    }`
                  }
                >
                  {item.label}
                </NavLink>
              ))}
            </nav>
          </div>
          <div className="flex items-center gap-3">
            <span className="text-sm text-ink-400">{user?.email}</span>
            <button
              onClick={logout}
              className="rounded-md px-3 py-1.5 text-sm font-medium text-ink-500 hover:bg-surface"
            >
              Salir
            </button>
          </div>
        </div>
      </header>
      <main className="mx-auto max-w-6xl px-6 py-8">
        <Outlet />
      </main>
    </div>
  )
}
