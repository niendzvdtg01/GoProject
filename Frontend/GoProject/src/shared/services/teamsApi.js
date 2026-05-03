import { api } from './axios.js'

export async function getTeams() {
  const response = await api.get('/teams')
  return response.data.teams ?? response.data ?? []
}
