import { useQuery } from '@tanstack/react-query'
import { getTeams } from '../api/teamsApi.js'

export function useTeams({ enabled = false } = {}) {
  return useQuery({
    queryKey: ['teams'],
    queryFn: getTeams,
    enabled,
  })
}
