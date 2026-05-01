import { api } from '../../../shared/services/axios.js'
import { normalizeUser } from '../../../shared/types/user.js'

export async function login(payload) {
  const response = await api.post('/auth/login', payload)
  return toSession(response.data)
}

export async function register(payload) {
  const response = await api.post('/users/register', payload)
  return toSession(response.data)
}

export async function logout() {
  const response = await api.post('/auth/logout')
  return response.data
}

function toSession(payload) {
  return {
    accessToken: payload.token,
    user: normalizeUser(payload.user),
  }
}
