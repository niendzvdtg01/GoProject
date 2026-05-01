import { TeamManagementPanel } from '../../features/teams/components/TeamManagementPanel.jsx'

export function TeamPage() {
  return (
    <div className="grid gap-6">
      <header>
        <span className="text-xs font-extrabold uppercase text-sky-700">Teams</span>
        <h1 className="mt-2 text-3xl font-black text-slate-950">Team management</h1>
        <p className="mt-2 max-w-3xl text-slate-600">
          Manager-only route prepared for teams, members and future role assignment flows.
        </p>
      </header>
      <TeamManagementPanel />
    </div>
  )
}
