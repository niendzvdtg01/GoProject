import { api } from './axios.js'

export async function getTeams() {
  const response = await api.get('/teams')
  return response.data.teams ?? response.data ?? []
}

export async function createTeam(teamName, members = []) {
  const body = {
    teamName,
    members: members.map((member) => ({ user_id: member.userID || member.user_id, role: member.role })),
  }

  const response = await api.post('/teams', body)
  return response.data
}

export async function addTeamMember(teamName, memberName, role = 'manager') {
  const response = await api.post(`/teams/${encodeURIComponent(teamName)}/members`, {
    user_id: memberName,
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
