import { useEffect } from 'react'
import { Navigate, Outlet } from 'react-router-dom'
import { useMeQuery } from '../api/apiSlice'
import { useAppDispatch, useAppSelector } from '../store/hooks'
import { loggedOut, userLoaded } from '../store/authSlice'

export function ProtectedRoute() {
  const dispatch = useAppDispatch()
  const hasSession = useAppSelector((s) => s.auth.hasSession)
  const { data, error, isLoading } = useMeQuery(undefined, { skip: !hasSession })

  useEffect(() => {
    if (data) dispatch(userLoaded(data))
    if (error) dispatch(loggedOut())
  }, [data, error, dispatch])

  if (!hasSession) {
    return <Navigate to="/login" replace />
  }

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center text-slate-500">
        Cargando…
      </div>
    )
  }

  return <Outlet />
}
