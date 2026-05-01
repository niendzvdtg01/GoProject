import { api } from '../../../shared/services/axios.js'

export async function getTeams() {
  const response = await api.get('/teams')
  return response.data.teams ?? response.data ?? []
}
