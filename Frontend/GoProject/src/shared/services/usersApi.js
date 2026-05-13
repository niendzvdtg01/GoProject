import { api } from './axios.js'
import { normalizeUser } from '../types/user.js'

export async function getUsers() {
  const response = await api.get('/users')
  return (response.data.users ?? []).map(normalizeUser)
}

export const importUsers = (file) => {
  const formData = new FormData()
  formData.append('file', file)
  return api.post('/users/import', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  }).then(r => r.data)
}
