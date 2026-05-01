import { USER_ROLES } from '../constants/roles.js'

export function normalizeUser(rawUser) {
  if (!rawUser) {
    return null
  }

  return {
    userId: rawUser.userId ?? rawUser.user_id ?? '',
    username: rawUser.username ?? '',
    email: rawUser.email ?? '',
    role: rawUser.role ?? USER_ROLES.member,
    createdAt: rawUser.createdAt ?? rawUser.created_at ?? null,
  }
}

export function isManager(user) {
  return user?.role === USER_ROLES.manager
}
