export function getApiErrorMessage(error) {
  const payload = error.response?.data

  if (payload?.error) {
    return payload.error
  }

  if (payload && typeof payload === 'object') {
    return Object.values(payload).flat().join(', ')
  }

  return error.message ?? 'Request failed'
}
