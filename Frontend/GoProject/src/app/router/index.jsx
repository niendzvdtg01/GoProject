import { Navigate, Route, Routes } from 'react-router-dom'
import { NotFoundPage } from '../../pages/NotFoundPage/index.jsx'
import { ROUTES } from '../../shared/constants/routes.js'
import { ProtectedRoutes } from './protectedRoutes.jsx'
import { PublicRoutes } from './publicRoutes.jsx'

export function AppRouter() {
  return (
    <Routes>
      <Route path="/" element={<Navigate to={ROUTES.dashboard} replace />} />
      {PublicRoutes()}
      {ProtectedRoutes()}
      <Route path="*" element={<NotFoundPage />} />
    </Routes>
  )
}
