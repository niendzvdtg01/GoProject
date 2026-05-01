import { AssetExplorer } from '../../features/assets/components/AssetExplorer.jsx'

export function AssetPage() {
  return (
    <div className="grid gap-6">
      <header>
        <span className="text-xs font-extrabold uppercase text-sky-700">Assets</span>
        <h1 className="mt-2 text-3xl font-black text-slate-950">Folder and note assets</h1>
        <p className="mt-2 max-w-3xl text-slate-600">
          Asset feature is structured for folders, notes, sharing and permission controls.
        </p>
      </header>
      <AssetExplorer />
    </div>
  )
}
