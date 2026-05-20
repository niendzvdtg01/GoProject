import { api } from './axios.js'
import { normalizeUser } from '../types/user.js'

export async function getUsers() {
  const response = await api.get('/users')
  return (response.data.users ?? []).map(normalizeUser)
}

export async function startImport(file) {
  const formData = new FormData()
  formData.append('file', file)
  const r = await api.post('/users/import', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
  return r.data
}

export async function getImportTask(taskId) {
  const r = await api.get(`/import-tasks/${taskId}`)
  return r.data
}
