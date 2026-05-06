export function normalizeUser(rawUser) {
  if (!rawUser) {
    return null
  }

  return {
    userId: rawUser.userId ?? rawUser.user_id ?? '',
    username: rawUser.username ?? '',
    email: rawUser.email ?? '',
    role: rawUser.role ?? rawUser.user_role ?? null,
    createdAt: rawUser.createdAt ?? rawUser.created_at ?? null,
  }
}
