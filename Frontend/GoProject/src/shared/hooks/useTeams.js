import { useQuery } from '@tanstack/react-query'
import { getTeams } from '../services/teamsApi.js'
import { useAuthStore } from '../../stores/authStore.js'

export const useTeams = ({ enabled = true } = {}) => {
  const accessToken = useAuthStore(s => s.accessToken)
  return useQuery({
    queryKey: ['teams'],
    queryFn: getTeams,
    enabled: enabled && !!accessToken,
    select: (data) => data.teams ?? data,
  })
}
