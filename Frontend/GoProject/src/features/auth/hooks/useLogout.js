import { useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import { useAuthStore } from '../../../app/store/authStore.js'
import { ROUTES } from '../../../shared/constants/routes.js'
import { logout } from '../api/authApi.js'

export function useLogout() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const clearSession = useAuthStore((state) => state.clearSession)

  return useCallback(async () => {
    try {
      await logout()
    } finally {
      clearSession()
      queryClient.clear()
      navigate(ROUTES.login, { replace: true })
    }
  }, [clearSession, navigate, queryClient])
}
