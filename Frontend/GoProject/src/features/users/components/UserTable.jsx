import { EmptyState } from '../../../shared/components/EmptyState.jsx'
import { RoleBadge } from '../../../shared/components/RoleBadge.jsx'
import { Table } from '../../../shared/components/Table.jsx'
import { formatDate } from '../../../shared/utils/formatDate.js'

const columns = ['Name', 'Email', 'Role', 'Created']

export function UserTable({ users }) {
  if (!users.length) {
    return (
      <EmptyState
        title="No users found"
        description="Register a manager or member account to see it in this directory."
      />
    )
  }

  return (
    <Table
      columns={columns}
      data={users}
      renderRow={(user) => (
        <tr key={user.userId || user.email} className="border-b border-slate-200 last:border-b-0">
          <td className="px-5 py-4">
            <strong className="block text-slate-950">{user.username}</strong>
            <small className="text-slate-500">{user.userId}</small>
          </td>
          <td className="px-5 py-4 text-slate-700">{user.email}</td>
          <td className="px-5 py-4">
            <RoleBadge role={user.role} />
          </td>
          <td className="px-5 py-4 text-slate-700">{formatDate(user.createdAt)}</td>
        </tr>
      )}
    />
  )
}
