import { useAuthStore } from '../../app/store/authStore.js'
import { Card } from '../../shared/components/Card.jsx'
import { RoleBadge } from '../../shared/components/RoleBadge.jsx'

export function ProfilePage() {
  const user = useAuthStore((state) => state.user)

  return (
    <div className="grid gap-6">
      <header>
        <span className="text-xs font-extrabold uppercase text-sky-700">Profile</span>
        <h1 className="mt-2 text-3xl font-black text-slate-950">Current user</h1>
      </header>

      <Card className="max-w-2xl p-5">
        <dl className="grid gap-4">
          <Item label="Username" value={user?.username} />
          <Item label="Email" value={user?.email} />
          <div className="grid gap-1">
            <dt className="text-sm font-bold text-slate-500">Role</dt>
            <dd>
              <RoleBadge role={user?.role} />
            </dd>
          </div>
        </dl>
      </Card>
    </div>
  )
}

function Item({ label, value }) {
  return (
    <div className="grid gap-1">
      <dt className="text-sm font-bold text-slate-500">{label}</dt>
      <dd className="font-semibold text-slate-950">{value}</dd>
    </div>
  )
}
