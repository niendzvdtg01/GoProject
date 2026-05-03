import { useAuthStore } from '../../stores/authStore.js'
import { api } from './axios.js'

let isConfigured = false

export function setupApiInterceptors() {
  if (isConfigured) {
    return
  }

  isConfigured = true

  api.interceptors.request.use((config) => {
    const token = useAuthStore.getState().accessToken

    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }

    return config
  })

  api.interceptors.response.use(
    (response) => response,
    (error) => {
      if (error.response?.status === 401) {
        useAuthStore.getState().clearSession()
      }

      return Promise.reject(error)
    },
  )
}
