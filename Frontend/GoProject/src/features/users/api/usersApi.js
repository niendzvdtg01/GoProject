import { api } from '../../../shared/services/axios.js'
import { normalizeUser } from '../../../shared/types/user.js'

export async function getUsers() {
  const response = await api.get('/users')
  return (response.data.users ?? []).map(normalizeUser)
}
