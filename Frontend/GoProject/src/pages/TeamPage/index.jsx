import { useParams } from 'react-router-dom'
import { TeamManagementPanel } from './TeamManagementPanel.jsx'
import TeamWorkspace from './TeamWorkspace.jsx'

export function TeamPage() {
  const { teamName } = useParams()
  if (teamName) return <TeamWorkspace />
  return <TeamManagementPanel />
}
