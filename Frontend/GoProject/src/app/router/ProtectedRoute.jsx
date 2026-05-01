import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { useAuthStore } from '../store/authStore.js'
import { ROUTES } from '../../shared/constants/routes.js'

export function ProtectedRoute({ roles }) {
  const location = useLocation()
  const accessToken = useAuthStore((state) => state.accessToken)
  const user = useAuthStore((state) => state.user)

  if (!accessToken || !user) {
    return <Navigate to={ROUTES.login} replace state={{ from: location }} />
  }

  if (roles?.length && !roles.includes(user.role)) {
    return <Navigate to={ROUTES.dashboard} replace />
  }

  return <Outlet />
}
