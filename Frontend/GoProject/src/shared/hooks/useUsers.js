import { useQuery } from '@tanstack/react-query'
import { useAuthStore } from '../../stores/authStore.js'
import { getUsers } from '../services/usersApi.js'

export function useUsers() {
  const user = useAuthStore((state) => state.user)

  return useQuery({
    queryKey: ['users'],
    queryFn: getUsers,
    enabled: Boolean(user),
  })
}
