import { ROLE_LABELS } from '../constants/roles.js'

export function RoleBadge({ role }) {
  const styles =
    role === 'manager'
      ? 'bg-emerald-50 text-emerald-700'
      : role === 'member'
      ? 'bg-amber-50 text-amber-700'
      : 'bg-slate-100 text-slate-600'

  return (
    <span className={`inline-flex rounded-full px-2.5 py-1 text-xs font-extrabold uppercase ${styles}`}>
      {role ? ROLE_LABELS[role] ?? role : 'No role'}
    </span>
  )
}
