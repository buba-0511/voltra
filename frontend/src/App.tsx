import { Provider } from 'react-redux'
import { Navigate, Route, BrowserRouter, Routes } from 'react-router-dom'
import { ProtectedRoute } from './auth/ProtectedRoute'
import { ErrorBoundary } from './components/ErrorBoundary'
import { AppLayout } from './layout/AppLayout'
import { AnomaliesPage } from './pages/AnomaliesPage'
import { DashboardPage } from './pages/DashboardPage'
import { LoginPage } from './pages/LoginPage'
import { InvestigationPage } from './pages/InvestigationPage'
import { MeterDetailPage } from './pages/MeterDetailPage'
import { MetersPage } from './pages/MetersPage'
import { store } from './store/store'

function App() {
  return (
    <ErrorBoundary>
      <Provider store={store}>
        <BrowserRouter>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route element={<ProtectedRoute />}>
              <Route element={<AppLayout />}>
                <Route path="/dashboard" element={<DashboardPage />} />
                <Route path="/meters" element={<MetersPage />} />
                <Route path="/meters/:meterId" element={<MeterDetailPage />} />
                <Route path="/anomalies" element={<AnomaliesPage />} />
                <Route path="/anomalies/:id" element={<InvestigationPage />} />
              </Route>
            </Route>
            <Route path="/" element={<Navigate to="/dashboard" replace />} />
            <Route path="*" element={<Navigate to="/dashboard" replace />} />
          </Routes>
        </BrowserRouter>
      </Provider>
    </ErrorBoundary>
  )
}

export default App
