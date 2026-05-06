import { api } from './axios.js'
import { normalizeUser } from '../types/user.js'

function parseJwt(token) {
  if (!token) {
    return null
  }

  const parts = token.split('.')
  if (parts.length !== 3) {
    return null
  }

  try {
    const payload = atob(parts[1].replace(/-/g, '+').replace(/_/g, '/'))
    return JSON.parse(decodeURIComponent(
      payload
        .split('')
        .map((c) => `%${(`00${c.charCodeAt(0).toString(16)}`).slice(-2)}`)
        .join(''),
    ))
  } catch {
    return null
  }
}

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
  const token = payload.token
  const jwtClaims = parseJwt(token) ?? {}
  const role = payload.user?.role ?? jwtClaims.role ?? null

  return {
    accessToken: token,
    user: normalizeUser({ ...payload.user, role }),
  }
}
