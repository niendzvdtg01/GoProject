import { Route } from 'react-router-dom'
import { DashboardPage } from '../../pages/DashboardPage/index.jsx'
import { TeamPage } from '../../pages/TeamPage/index.jsx'
import { AssetPage } from '../../pages/AssetPage/index.jsx'
import { ProfilePage } from '../../pages/ProfilePage/index.jsx'
import { DashboardLayout } from '../../shared/layouts/DashboardLayout.jsx'
import { USER_ROLES } from '../../shared/constants/roles.js'
import { ROUTES } from '../../shared/constants/routes.js'
import { ProtectedRoute } from './ProtectedRoute.jsx'

export function ProtectedRoutes() {
  return (
    <Route element={<ProtectedRoute />}>
      <Route element={<DashboardLayout />}>
        <Route path={ROUTES.dashboard} element={<DashboardPage />} />
        <Route path={ROUTES.assets} element={<AssetPage />} />
        <Route path={ROUTES.assetDetail} element={<AssetPage />} />
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
