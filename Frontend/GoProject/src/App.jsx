import { Navigate, Route, Routes } from 'react-router-dom'
import { NotFoundPage } from './pages/NotFoundPage/index.jsx'
import { DashboardPage } from './pages/DashboardPage/index.jsx'
import { TeamPage } from './pages/TeamPage/index.jsx'
import { ProfilePage } from './pages/ProfilePage/index.jsx'
import { LoginPage } from './pages/LoginPage/index.jsx'
import { RegisterPage } from './pages/RegisterPage/index.jsx'
import { DashboardLayout } from './shared/layouts/DashboardLayout.jsx'
import { AuthLayout } from './shared/layouts/AuthLayout.jsx'
import { USER_ROLES } from './shared/constants/roles.js'
import { ROUTES } from './shared/constants/routes.js'
import { ProtectedRoute } from './shared/utils/ProtectedRoute.jsx'
import { PublicRoute } from './shared/utils/PublicRoute.jsx'

function PublicRoutes() {
  return (
    <Route element={<PublicRoute />}>
      <Route path={ROUTES.login} element={<AuthLayout><LoginPage /></AuthLayout>} />
      <Route path={ROUTES.register} element={<AuthLayout><RegisterPage /></AuthLayout>} />
    </Route>
  )
}

function ProtectedRoutes() {
  return (
    <Route element={<ProtectedRoute />}>
      <Route element={<DashboardLayout />}>
        <Route path={ROUTES.dashboard} element={<DashboardPage />} />
        <Route path={ROUTES.profile} element={<ProfilePage />} />
      </Route>

      <Route element={<ProtectedRoute roles={[USER_ROLES.manager]} />}>
        <Route element={<DashboardLayout />}>
          <Route path={ROUTES.teams} element={<TeamPage />} />
          <Route path={ROUTES.teamDetail} element={<TeamPage />} />
        </Route>
      </Route>
    </Route>
  )
}

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<Navigate to={ROUTES.dashboard} replace />} />
      {PublicRoutes()}
      {ProtectedRoutes()}
      <Route path="*" element={<NotFoundPage />} />
    </Routes>
  )
}
