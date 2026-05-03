import { api } from './axios.js'
import { normalizeUser } from '../types/user.js'

export async function getUsers() {
  const response = await api.get('/users')
  return (response.data.users ?? []).map(normalizeUser)
}
