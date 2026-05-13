import { api } from './axios.js'

export const getFolders = () => api.get('/folders').then(r => r.data)
export const createFolder = (name) => api.post('/folders', { name }).then(r => r.data)
export const updateFolder = (id, name) => api.put(`/folders/${id}`, { name }).then(r => r.data)
export const deleteFolder = (id) => api.delete(`/folders/${id}`).then(r => r.data)
export const getFolderNotes = (folderId) => api.get(`/folders/${folderId}/notes`).then(r => r.data)
export const createNote = (folderId, title, content) =>
  api.post(`/folders/${folderId}/notes`, { title, content }).then(r => r.data)
