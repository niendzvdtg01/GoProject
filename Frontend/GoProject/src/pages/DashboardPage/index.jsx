import { Card } from '../../shared/components/Card.jsx'
import { EmptyState } from '../../shared/components/EmptyState.jsx'
import { LoadingSkeleton } from '../../shared/components/LoadingSkeleton.jsx'
import { UserSummaryCards } from './UserSummaryCards.jsx'
import { UserTable } from './UserTable.jsx'
import { useUsers } from '../../shared/hooks/useUsers.js'
import { useAuthStore } from '../../stores/authStore.js'
import { isManager } from '../../shared/types/user.js'
import { DashboardStats } from '../../shared/components/DashboardStats.jsx'

export function DashboardPage() {
  const user = useAuthStore((state) => state.user)
  const usersQuery = useUsers()
  const users = usersQuery.data ?? []

  return (
    <div className="grid gap-6">
      <header>
        <span className="text-xs font-extrabold uppercase text-sky-700">Dashboard</span>
        <h1 className="mt-2 text-3xl font-black text-slate-950">Workspace overview</h1>
        <p className="mt-2 max-w-3xl text-slate-600">
          Server state is loaded through TanStack Query; auth and UI state stay in Zustand. Team management and user directory are the current focus.
        </p>
      </header>

      <DashboardStats users={users} />

      {isManager(user) ? (
        <>
          <UserSummaryCards users={users} />
          <Card className="overflow-hidden">
            <div className="border-b border-slate-200 px-5 py-4">
              <h2 className="text-lg font-extrabold text-slate-950">User directory</h2>
            </div>
            <div className="p-5">
              {usersQuery.isLoading ? <LoadingSkeleton rows={5} /> : <UserTable users={users} />}
              {usersQuery.isError ? (
                <p className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm font-semibold text-red-700">
                  {usersQuery.error.message}
                </p>
              ) : null}
            </div>
          </Card>
        </>
      ) : (
        <EmptyState
          title="Member dashboard"
          description="Member accounts can view assigned teams and shared assets after those backend APIs are available."
        />
      )}
    </div>
  )
}
