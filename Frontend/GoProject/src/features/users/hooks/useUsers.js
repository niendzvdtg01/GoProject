import { useQuery } from '@tanstack/react-query'
import { isManager } from '../../../shared/types/user.js'
import { useAuthStore } from '../../../app/store/authStore.js'
import { getUsers } from '../api/usersApi.js'

export function useUsers() {
  const user = useAuthStore((state) => state.user)

  return useQuery({
    queryKey: ['users'],
    queryFn: getUsers,
    enabled: isManager(user),
  })
}
