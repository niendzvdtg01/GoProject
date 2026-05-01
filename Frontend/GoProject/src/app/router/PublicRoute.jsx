import { Navigate, Outlet } from 'react-router-dom'
import { useAuthStore } from '../store/authStore.js'
import { ROUTES } from '../../shared/constants/routes.js'

export function PublicRoute() {
  const accessToken = useAuthStore((state) => state.accessToken)
  const user = useAuthStore((state) => state.user)

  if (accessToken && user) {
    return <Navigate to={ROUTES.dashboard} replace />
  }

  return <Outlet />
}
