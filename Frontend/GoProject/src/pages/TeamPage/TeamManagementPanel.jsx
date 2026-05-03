import { Button } from '../../shared/components/Button.jsx'
import { Card } from '../../shared/components/Card.jsx'
import { EmptyState } from '../../shared/components/EmptyState.jsx'

export function TeamManagementPanel() {
  return (
    <div className="grid gap-4">
      <Card className="p-5">
        <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
          <div>
            <h2 className="text-lg font-extrabold text-slate-950">Team management</h2>
            <p className="mt-1 text-sm text-slate-600">
              API contract is prepared for /teams. Backend route can be plugged in without changing page composition.
            </p>
          </div>
          <Button type="button" disabled>
            Create team
          </Button>
        </div>
      </Card>
      <EmptyState
        title="No teams endpoint yet"
        description="Team model exists in backend, but handler/routes are not implemented yet."
      />
    </div>
  )
}
