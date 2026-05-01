import { ROLE_LABELS, USER_ROLES } from '../constants/roles.js'

export function RoleBadge({ role }) {
  const styles =
    role === USER_ROLES.manager
      ? 'bg-emerald-50 text-emerald-700'
      : 'bg-amber-50 text-amber-700'

  return (
    <span className={`inline-flex rounded-full px-2.5 py-1 text-xs font-extrabold uppercase ${styles}`}>
      {ROLE_LABELS[role] ?? role}
    </span>
  )
}
