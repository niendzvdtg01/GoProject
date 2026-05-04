import { api } from './axios.js'

export async function getTeams() {
  const response = await api.get('/teams')
  return response.data.teams ?? response.data ?? []
}

export async function createTeam(teamName) {
  const response = await api.post('/teams', { teamName })
  return response.data
}

export async function addTeamMember(teamName, memberName, role) {
  const response = await api.post(`/teams/${encodeURIComponent(teamName)}/members`, {
    memberName,
    role,
  })
  return response.data
}

export async function removeTeamMember(teamName, memberName) {
  const response = await api.delete(`/teams/${encodeURIComponent(teamName)}/members/${encodeURIComponent(memberName)}`)
  return response.data
}

export async function deleteTeam(teamName) {
  const response = await api.delete(`/teams/${encodeURIComponent(teamName)}`)
  return response.data
}
