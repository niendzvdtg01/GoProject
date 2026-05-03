import { Card } from './Card.jsx'

export function DashboardStats({ users }) {
  const items = [
    { label: 'Total users', value: users.length },
    { label: 'Managers', value: users.filter((user) => user.role === 'manager').length },
    { label: 'Members', value: users.filter((user) => user.role === 'member').length },
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
