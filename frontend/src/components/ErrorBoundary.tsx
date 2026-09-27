import { AlertTriangle } from 'lucide-react'
import { Component, type ErrorInfo, type ReactNode } from 'react'

interface Props {
  children: ReactNode
}

interface State {
  error: Error | null
}

// React error boundaries can only be class components - there's no hook
// equivalent. Without this, any single component throwing during render
// (a bad API response shape, a stray undefined.toFixed(), etc.) unmounts
// the entire app to a blank page instead of a contained, recoverable
// error - we hit exactly that failure mode earlier in this project.
export class ErrorBoundary extends Component<Props, State> {
  state: State = { error: null }

  static getDerivedStateFromError(error: Error): State {
    return { error }
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('Unhandled error in component tree:', error, info.componentStack)
  }

  render() {
    if (this.state.error) {
      return (
        <div className="flex min-h-screen flex-col items-center justify-center gap-4 bg-surface px-4 text-center">
          <div className="grid h-12 w-12 place-items-center rounded-full bg-red-50 text-red-500">
            <AlertTriangle size={22} />
          </div>
          <div>
            <h1 className="text-lg font-semibold text-ink-900">Algo salió mal</h1>
            <p className="mt-1 max-w-sm text-sm text-ink-400">
              Ocurrió un error inesperado en esta pantalla. Podés intentar recargar la página.
            </p>
          </div>
          <button
            onClick={() => window.location.reload()}
            className="rounded-lg bg-ink-900 px-4 py-2.5 text-xs font-semibold text-white hover:bg-ink-800"
          >
            Recargar
          </button>
        </div>
      )
    }

    return this.props.children
  }
}
