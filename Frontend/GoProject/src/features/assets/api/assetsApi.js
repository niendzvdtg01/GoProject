import { api } from '../../../shared/services/axios.js'

export async function getAssets() {
  const response = await api.get('/assets')
  return response.data.assets ?? response.data ?? []
}
