import { Card } from '../../shared/components/Card.jsx'

export function UserSummaryCards({ users }) {
  return (
    <div className="grid gap-4 md:grid-cols-3">
      <SummaryCard label="Total users" value={users.length} />
      <SummaryCard label="Active teams" value="—" />
      <SummaryCard label="Role assignment" value="Team-only" />
    </div>
  )
}

function SummaryCard({ label, value }) {
  return (
    <Card className="p-5">
      <span className="text-sm font-bold text-slate-500">{label}</span>
      <strong className="mt-2 block text-3xl font-black text-slate-950">{value}</strong>
    </Card>
  )
}
