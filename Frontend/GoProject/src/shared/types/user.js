export function normalizeUser(rawUser) {
  if (!rawUser) {
    return null
  }

  return {
    userId: rawUser.userId ?? rawUser.user_id ?? '',
    username: rawUser.username ?? '',
    email: rawUser.email ?? '',
    createdAt: rawUser.createdAt ?? rawUser.created_at ?? null,
  }
}
