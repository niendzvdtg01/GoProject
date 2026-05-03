import { useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import { useAuthStore } from '../../stores/authStore.js'
import { ROUTES } from '../constants/routes.js'
import { logout } from '../services/authApi.js'

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
