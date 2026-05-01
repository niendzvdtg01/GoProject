import { Card } from '../../../shared/components/Card.jsx'

const rows = [
  { role: 'Viewer', permission: 'Read' },
  { role: 'Editor', permission: 'Read + Update' },
  { role: 'Owner', permission: 'Full Access' },
]

export function PermissionMatrix() {
  return (
    <Card className="p-5">
      <h2 className="text-lg font-extrabold text-slate-950">Permission overview</h2>
      <div className="mt-4 grid gap-2">
        {rows.map((row) => (
          <div
            key={row.role}
            className="grid grid-cols-[120px_1fr] rounded-md border border-slate-200 bg-slate-50 px-4 py-3 text-sm"
          >
            <strong className="text-slate-950">{row.role}</strong>
            <span className="text-slate-700">{row.permission}</span>
          </div>
        ))}
      </div>
    </Card>
  )
}
