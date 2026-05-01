import { useMutation } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '../../../app/store/authStore.js'
import { ROUTES } from '../../../shared/constants/routes.js'
import { getApiErrorMessage } from '../../../shared/services/apiError.js'
import { login } from '../api/authApi.js'

export function useLogin() {
  const navigate = useNavigate()
  const setSession = useAuthStore((state) => state.setSession)

  return useMutation({
    mutationFn: login,
    onSuccess(session) {
      setSession(session)
      navigate(ROUTES.dashboard, { replace: true })
    },
    throwOnError: false,
    meta: {
      getErrorMessage: getApiErrorMessage,
    },
  })
}
