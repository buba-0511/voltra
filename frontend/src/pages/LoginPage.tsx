import { AlertCircle, Zap } from 'lucide-react'
import { useState, type FormEvent } from 'react'
import { Navigate, useNavigate } from 'react-router-dom'
import { useLoginMutation } from '../api/apiSlice'
import { getToken } from '../api/tokenStorage'
import { credentialsSet } from '../store/authSlice'
import { useAppDispatch } from '../store/hooks'

const highlights = [
  'Detecta anomalías reales, no solo picos de consumo',
  'Explica cada hallazgo con evidencia, no una caja negra',
  'Prioriza qué medidor revisar primero',
]

export function LoginPage() {
  const dispatch = useAppDispatch()
  const navigate = useNavigate()
  const [login, { isLoading }] = useLoginMutation()
  const [email, setEmail] = useState('admin@energy-platform.local')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)

  if (getToken()) {
    return <Navigate to="/dashboard" replace />
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    try {
      const result = await login({ email, password }).unwrap()
      dispatch(credentialsSet(result))
      navigate('/dashboard', { replace: true })
    } catch (err) {
      const message =
        typeof err === 'object' && err && 'error' in err
          ? String((err as { error: unknown }).error)
          : 'No se pudo iniciar sesión'
      setError(message)
    }
  }

  return (
    <div className="flex min-h-screen bg-surface">
      {/* Branding panel - hidden on mobile, the form alone is enough there */}
      <div className="relative hidden w-[62%] flex-col justify-between overflow-hidden bg-ink-900 px-16 py-14 text-white xl:flex">
        <div className="flex items-center gap-3">
          <div className="grid h-10 w-10 place-items-center rounded-xl bg-brand-500 text-ink-900">
            <Zap size={22} strokeWidth={2.8} />
          </div>
          <div>
            <span className="text-xl font-bold tracking-tight">
              Voltra<span className="text-brand-400">.</span>
            </span>
            <p className="text-[11px] text-ink-300">AI Energy Management Platform</p>
          </div>
        </div>

        <div className="max-w-2xl">
          <h1 className="text-6xl leading-[1.08] font-semibold text-balance">
            De datos crudos a{' '}
            <span className="text-brand-400">la acción que hay que tomar.</span>
          </h1>
          <p className="mt-6 max-w-xl text-lg leading-8 text-ink-200">
            Voltra analiza el consumo de cada medidor, detecta lo que se sale
            de lo normal y te dice{' '}
            <span className="font-medium text-white">qué revisar primero</span>{' '}
            — con la evidencia que respalda cada decisión.
          </p>

          <ul className="mt-10 space-y-5">
            {highlights.map((h) => (
              <li key={h} className="flex items-start gap-3 text-base text-ink-100">
                <span className="mt-2 h-1.5 w-1.5 shrink-0 rounded-full bg-brand-400" />
                {h}
              </li>
            ))}
          </ul>
        </div>

        <div className="flex max-w-md flex-wrap gap-2 border-t border-white/10 pt-6">
          {['Detección estadística', 'Explicación con evidencia', 'Priorización automática'].map(
            (tag) => (
              <span
                key={tag}
                className="rounded-full bg-white/5 px-3 py-1.5 text-xs font-medium text-ink-200 ring-1 ring-white/10"
              >
                {tag}
              </span>
            ),
          )}
        </div>
      </div>

      {/* Form panel */}
      <div className="flex flex-1 items-center justify-center px-4 py-12">
        <div className="w-full max-w-76">
          <div className="mb-8 flex items-center gap-2.5 xl:hidden">
            <div className="grid h-8 w-8 place-items-center rounded-lg bg-brand-500 text-ink-900">
              <Zap size={18} strokeWidth={2.8} />
            </div>
            <span className="text-base font-bold tracking-tight text-ink-900">
              Voltra<span className="text-brand-500">.</span>
            </span>
          </div>

          <h2 className="text-lg font-semibold tracking-tight text-ink-900">
            Bienvenido de nuevo
          </h2>
          <p className="mt-1 text-sm text-ink-400">Iniciá sesión para continuar.</p>

          <form onSubmit={handleSubmit} className="mt-6 space-y-3.5">
            <div>
              <label
                className="block text-sm font-medium text-ink-700"
                htmlFor="email"
              >
                Email
              </label>
              <input
                id="email"
                type="email"
                required
                autoComplete="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="mt-1.5 w-full rounded-lg border border-surface-border bg-white px-3 py-2 text-sm text-ink-900 outline-none transition-colors placeholder:text-ink-300 focus:border-brand-500 focus:ring-2 focus:ring-brand-100"
              />
            </div>
            <div>
              <label
                className="block text-sm font-medium text-ink-700"
                htmlFor="password"
              >
                Contraseña
              </label>
              <input
                id="password"
                type="password"
                required
                autoComplete="current-password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="mt-1.5 w-full rounded-lg border border-surface-border bg-white px-3 py-2 text-sm text-ink-900 outline-none transition-colors placeholder:text-ink-300 focus:border-brand-500 focus:ring-2 focus:ring-brand-100"
              />
            </div>

            {error && (
              <p className="flex items-center gap-2 rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">
                <AlertCircle size={15} className="shrink-0" />
                {error}
              </p>
            )}

            <button
              type="submit"
              disabled={isLoading}
              className="w-full rounded-lg bg-ink-700 py-2.5 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-ink-800 disabled:opacity-50"
            >
              {isLoading ? 'Ingresando…' : 'Ingresar'}
            </button>
          </form>
        </div>
      </div>
    </div>
  )
}
