import { Card } from '../shared/components/Card.jsx'

export function DashboardStats({ users }) {
  const items = [
    { label: 'Team count', value: 'Pending API' },
    { label: 'Member count', value: users.filter((user) => user.role === 'member').length },
    { label: 'Recent assets', value: 'Pending API' },
    { label: 'Permissions', value: '3 levels' },
  ]

  return (
    <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
      {items.map((item) => (
        <Card key={item.label} className="p-5">
          <span className="text-sm font-bold text-slate-500">{item.label}</span>
          <strong className="mt-2 block text-2xl font-black text-slate-950">{item.value}</strong>
        </Card>
      ))}
    </div>
  )
}
