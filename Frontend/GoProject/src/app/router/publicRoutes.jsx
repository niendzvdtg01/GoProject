import { Route } from 'react-router-dom'
import { LoginPage } from '../../pages/LoginPage/index.jsx'
import { RegisterPage } from '../../pages/RegisterPage/index.jsx'
import { ROUTES } from '../../shared/constants/routes.js'
import { PublicRoute } from './PublicRoute.jsx'

export function PublicRoutes() {
  return (
    <Route element={<PublicRoute />}>
      <Route path={ROUTES.login} element={<LoginPage />} />
      <Route path={ROUTES.register} element={<RegisterPage />} />
    </Route>
  )
}
