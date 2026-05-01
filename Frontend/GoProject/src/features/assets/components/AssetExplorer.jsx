import { Button } from '../../../shared/components/Button.jsx'
import { Card } from '../../../shared/components/Card.jsx'

const folders = [
  { id: 'notes', name: 'Notes', kind: 'folder' },
  { id: 'shared-files', name: 'Shared Files', kind: 'folder' },
  { id: 'permissions', name: 'Permission drafts', kind: 'note' },
]

export function AssetExplorer() {
  return (
    <div className="grid gap-4 lg:grid-cols-[280px_1fr]">
      <Card className="p-4">
        <div className="mb-4 flex items-center justify-between">
          <h2 className="font-extrabold text-slate-950">Folder Tree</h2>
          <Button type="button" variant="secondary" disabled>
            New
          </Button>
        </div>
        <div className="grid gap-2">
          {folders.map((item) => (
            <button
              key={item.id}
              type="button"
              className="rounded-md border border-slate-200 bg-white px-3 py-2 text-left text-sm font-bold text-slate-700 hover:bg-slate-50"
            >
              {item.name}
            </button>
          ))}
        </div>
      </Card>

      <Card className="p-5">
        <h2 className="text-lg font-extrabold text-slate-950">Asset actions</h2>
        <div className="mt-4 grid gap-3 md:grid-cols-3">
          <Action title="Create folder" />
          <Action title="Create note" />
          <Action title="Share asset" />
        </div>
      </Card>
    </div>
  )
}

function Action({ title }) {
  return (
    <div className="rounded-lg border border-slate-200 bg-slate-50 p-4">
      <strong className="block text-slate-950">{title}</strong>
      <p className="mt-1 text-sm text-slate-600">Ready for assets microservice integration.</p>
    </div>
  )
}
