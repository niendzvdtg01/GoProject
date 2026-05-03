import { useQuery } from '@tanstack/react-query'
import { getTeams } from '../services/teamsApi.js'

export function useTeams({ enabled = false } = {}) {
  return useQuery({
    queryKey: ['teams'],
    queryFn: getTeams,
    enabled,
  })
}
