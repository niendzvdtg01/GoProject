import { api } from './axios.js'

export const getNote = (id) => api.get(`/notes/${id}`).then(r => r.data)
export const updateNote = (id, title, content) =>
  api.put(`/notes/${id}`, { title, content }).then(r => r.data)
export const deleteNote = (id) => api.delete(`/notes/${id}`).then(r => r.data)
