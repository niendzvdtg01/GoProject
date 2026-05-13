import { api } from './axios.js'

export const shareAsset = (data) => api.post('/share', data).then(r => r.data)
// data: { email, note_id?, folder_id?, permission_type: 'read'|'write' }
export const revokeAccess = (data) => api.delete('/share', { data }).then(r => r.data)
// data: { email, note_id?, folder_id? }
