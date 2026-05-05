import { NavLink, Outlet } from 'react-router-dom'
import { useAuthStore } from '../../stores/authStore.js'
import { ROUTES } from '../constants/routes.js'
import { Button } from '../components/Button.jsx'
import { useLogout } from '../hooks/useLogout.js'

const navItems = [
  { label: 'Dashboard', to: ROUTES.dashboard },
  { label: 'Teams', to: ROUTES.teams },
  { label: 'Profile', to: ROUTES.profile },
]

export function DashboardLayout() {
  const user = useAuthStore((state) => state.user)
  const logout = useLogout()

  return (
    <div className="grid min-h-screen bg-slate-100 lg:grid-cols-[280px_1fr]">
      <aside className="flex flex-col gap-7 bg-slate-950 p-5 text-white">
        <div className="flex items-center gap-3">
          <span className="flex h-10 w-10 items-center justify-center rounded-md bg-sky-100 text-sm font-black text-sky-800">
            GP
          </span>
          <div>
            <strong className="block">GoProject</strong>
            <small className="text-slate-400">Admin console</small>
          </div>
        </div>

        <nav className="grid gap-1">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              className={({ isActive }) =>
                `rounded-md px-3 py-2 text-sm font-bold ${
                  isActive ? 'bg-white/10 text-white' : 'text-slate-300 hover:bg-white/5'
                }`
              }
            >
              {item.label}
            </NavLink>
          ))}
        </nav>

        <div className="mt-auto grid gap-3 border-t border-white/10 pt-5">
          <div>
            <strong className="block truncate">{user?.username}</strong>
            <small className="block truncate text-slate-400">{user?.email}</small>
          </div>
          <Button variant="secondary" type="button" onClick={logout}>
            Logout
          </Button>
        </div>
      </aside>

      <main className="min-w-0 p-5 md:p-8">
        <Outlet />
      </main>
    </div>
  )
}
